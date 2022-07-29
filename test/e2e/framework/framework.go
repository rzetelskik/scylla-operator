// Copyright (C) 2021 ScyllaDB

package framework

import (
	"archive/tar"
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	scyllaclientset "github.com/scylladb/scylla-operator/pkg/client/scylla/clientset/versioned"
	"github.com/scylladb/scylla-operator/pkg/naming"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/klog/v2"
)

const (
	ServiceAccountName        = "e2e-user"
	serviceAccountWaitTimeout = 1 * time.Minute
)

type Framework struct {
	name                string
	namespace           *corev1.Namespace
	watchLogsCancel     context.CancelFunc
	watchTCPDumpsCancel context.CancelFunc

	adminClientConfig *restclient.Config
	clientConfig      *restclient.Config
	username          string
}

func NewFramework(name string) *Framework {
	uniqueName := names.SimpleNameGenerator.GenerateName(fmt.Sprintf("%s-", name))

	adminClientConfig := restclient.CopyConfig(TestContext.RestConfig)
	adminClientConfig.UserAgent = "scylla-operator-e2e"
	adminClientConfig.QPS = 20
	adminClientConfig.Burst = 50

	f := &Framework{
		name:              uniqueName,
		username:          "admin",
		adminClientConfig: adminClientConfig,
	}

	g.BeforeEach(f.beforeEach)
	g.AfterEach(f.afterEach)

	return f
}

func (f *Framework) Namespace() string {
	return f.namespace.Name
}

func (f *Framework) Username() string {
	return f.username
}

func (f *Framework) ClientConfig() *restclient.Config {
	return f.clientConfig
}

func (f *Framework) AdminClientConfig() *restclient.Config {
	return f.adminClientConfig
}

func (f *Framework) DiscoveryClient() *discovery.DiscoveryClient {
	client, err := discovery.NewDiscoveryClientForConfig(f.ClientConfig())
	o.Expect(err).NotTo(o.HaveOccurred())
	return client
}

func (f *Framework) DynamicClient() dynamic.Interface {
	client, err := dynamic.NewForConfig(f.ClientConfig())
	o.Expect(err).NotTo(o.HaveOccurred())
	return client
}

func (f *Framework) DynamicAdminClient() dynamic.Interface {
	client, err := dynamic.NewForConfig(f.AdminClientConfig())
	o.Expect(err).NotTo(o.HaveOccurred())
	return client
}

func (f *Framework) KubeClient() *kubernetes.Clientset {
	client, err := kubernetes.NewForConfig(f.ClientConfig())
	o.Expect(err).NotTo(o.HaveOccurred())
	return client
}

func (f *Framework) KubeAdminClient() *kubernetes.Clientset {
	client, err := kubernetes.NewForConfig(f.AdminClientConfig())
	o.Expect(err).NotTo(o.HaveOccurred())
	return client
}

func (f *Framework) ScyllaClient() *scyllaclientset.Clientset {
	client, err := scyllaclientset.NewForConfig(f.ClientConfig())
	o.Expect(err).NotTo(o.HaveOccurred())
	return client
}
func (f *Framework) ScyllaAdminClient() *scyllaclientset.Clientset {
	client, err := scyllaclientset.NewForConfig(f.AdminClientConfig())
	o.Expect(err).NotTo(o.HaveOccurred())
	return client
}

func (f *Framework) setupNamespace(ctx context.Context) {
	By("Creating a new namespace")
	var ns *corev1.Namespace
	generateName := func() string {
		return names.SimpleNameGenerator.GenerateName(fmt.Sprintf("e2e-test-%s-", f.name))
	}
	name := generateName()
	err := wait.PollImmediate(2*time.Second, 30*time.Second, func() (bool, error) {
		var err error
		// We want to know the name ahead, even if the api call fails.
		ns, err = f.KubeAdminClient().CoreV1().Namespaces().Create(
			ctx,
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						"e2e":       "scylla-operator",
						"framework": f.name,
					},
				},
			},
			metav1.CreateOptions{},
		)
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				// regenerate on conflict
				Infof("Namespace name %q was already taken, generating a new name and retrying", name)
				name = generateName()
				return false, nil
			}
			return true, err
		}
		return true, nil
	})
	o.Expect(err).NotTo(o.HaveOccurred())

	Infof("Created namespace %q.", ns.Name)

	f.namespace = ns

	// Create user service account.
	userSA, err := f.KubeAdminClient().CoreV1().ServiceAccounts(ns.Name).Create(ctx, &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: ServiceAccountName,
		},
	}, metav1.CreateOptions{})
	o.Expect(err).NotTo(o.HaveOccurred())

	// Grant it edit permission in this namespace.
	_, err = f.KubeAdminClient().RbacV1().RoleBindings(ns.Name).Create(ctx, &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: userSA.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup:  corev1.GroupName,
				Kind:      rbacv1.ServiceAccountKind,
				Namespace: userSA.Namespace,
				Name:      userSA.Name,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "admin",
		},
	}, metav1.CreateOptions{})
	o.Expect(err).NotTo(o.HaveOccurred())

	// Wait for user ServiceAccount.
	By(fmt.Sprintf("Waiting for ServiceAccount %q in namespace %q.", userSA.Name, userSA.Namespace))
	ctxUserSa, ctxUserSaCancel := watchtools.ContextWithOptionalTimeout(ctx, serviceAccountWaitTimeout)
	defer ctxUserSaCancel()
	userSA, err = WaitForServiceAccount(ctxUserSa, f.KubeAdminClient().CoreV1(), userSA.Namespace, userSA.Name)
	o.Expect(err).NotTo(o.HaveOccurred())

	// Create a restricted client using the default SA.
	var token []byte
	for _, secretName := range userSA.Secrets {
		secret, err := f.KubeAdminClient().CoreV1().Secrets(userSA.Namespace).Get(ctx, secretName.Name, metav1.GetOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())

		if secret.Type == corev1.SecretTypeServiceAccountToken {
			name := secret.Annotations[corev1.ServiceAccountNameKey]
			uid := secret.Annotations[corev1.ServiceAccountUIDKey]
			if name == userSA.Name && uid == string(userSA.UID) {
				t, found := secret.Data[corev1.ServiceAccountTokenKey]
				if found {
					token = t
					break
				}
			}
		}
	}
	o.Expect(token).NotTo(o.HaveLen(0))

	f.clientConfig = restclient.AnonymousClientConfig(f.AdminClientConfig())
	f.clientConfig.BearerToken = string(token)

	// Wait for default ServiceAccount.
	By(fmt.Sprintf("Waiting for default ServiceAccount in namespace %q.", ns.Name))
	ctxSa, ctxSaCancel := watchtools.ContextWithOptionalTimeout(ctx, serviceAccountWaitTimeout)
	defer ctxSaCancel()
	_, err = WaitForServiceAccount(ctxSa, f.KubeAdminClient().CoreV1(), ns.Namespace, "default")
	o.Expect(err).NotTo(o.HaveOccurred())

	/* logs */
	d := path.Join(TestContext.ArtifactsDir, "e2e-test-logs")
	namespaceDir := path.Join(d, name)
	err = os.MkdirAll(namespaceDir, 0777)
	if err != nil && !os.IsExist(err) {
		o.Expect(err).NotTo(o.HaveOccurred())
	}

	f.watchLogsCancel, err = gatherLogs(ctx, f.KubeAdminClient(), name, namespaceDir)
	o.Expect(err).NotTo(o.HaveOccurred())

	/* tcpdump */
	d = path.Join(TestContext.ArtifactsDir, "e2e-test-tcpdumps")
	namespaceDir = path.Join(d, name)
	err = os.MkdirAll(namespaceDir, 0777)
	if err != nil && !os.IsExist(err) {
		o.Expect(err).NotTo(o.HaveOccurred())
	}

	f.watchTCPDumpsCancel, err = gatherTCPDumps(ctx, f.KubeAdminClient(), TestContext.RestConfig, name, namespaceDir)
	o.Expect(err).NotTo(o.HaveOccurred())
}

func gatherTCPDumps(ctx context.Context, cs kubernetes.Interface, csConfig *restclient.Config, ns string, dir string) (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(ctx)

	err := watchTCPDumps(ctx, cs, csConfig, ns, dir)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("can't watch for logs: %w", err)
	}
	return cancel, nil
}

func copyFromPod(coreClient corev1client.CoreV1Interface, ns string, podName string, containerName, srcPath string, destPath string) error {
	reader, outStream := io.Pipe()
	cmdArr := []string{"tar", "cf", "-", srcPath}
	req := coreClient.RESTClient().
		Get().
		Namespace(ns).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   cmdArr,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(TestContext.RestConfig, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("copy from pod: can't create executor: %w", err)
	}

	go func() {
		defer outStream.Close()
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  os.Stdin,
			Stdout: outStream,
			Stderr: os.Stderr,
			Tty:    false,
		})
		if err != nil {
			Warnf("copy from pod: %s: can't stream: %v", podName, err)
		}
	}()
	prefix := getPrefix(srcPath)
	prefix = path.Clean(prefix)
	prefix = stripPathShortcuts(prefix)
	destPath = path.Join(destPath, path.Base(prefix))
	err = untarAll(reader, destPath, prefix)
	if err != nil {
		return fmt.Errorf("copy from pod: can't untar: %w", err)
	}
	return nil
}

func untarAll(reader io.Reader, destDir, prefix string) error {
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		if !strings.HasPrefix(header.Name, prefix) {
			return fmt.Errorf("tar contents corrupted")
		}

		mode := header.FileInfo().Mode()
		destFileName := filepath.Join(destDir, header.Name[len(prefix):])

		baseName := filepath.Dir(destFileName)
		if err := os.MkdirAll(baseName, 0755); err != nil {
			return err
		}
		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(destFileName, 0755); err != nil {
				return err
			}
			continue
		}

		evaledPath, err := filepath.EvalSymlinks(baseName)
		if err != nil {
			return err
		}

		if mode&os.ModeSymlink != 0 {
			linkname := header.Linkname

			if !filepath.IsAbs(linkname) {
				_ = filepath.Join(evaledPath, linkname)
			}

			if err := os.Symlink(linkname, destFileName); err != nil {
				return err
			}
		} else {
			outFile, err := os.Create(destFileName)
			if err != nil {
				return err
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			if err := outFile.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func getPrefix(file string) string {
	return strings.TrimLeft(file, "/")
}

// stripPathShortcuts removes any leading or trailing "../" from a given path
func stripPathShortcuts(p string) string {
	newPath := path.Clean(p)
	trimmed := strings.TrimPrefix(newPath, "../")

	for trimmed != newPath {
		newPath = trimmed
		trimmed = strings.TrimPrefix(newPath, "../")
	}

	// trim leftover {".", ".."}
	if newPath == "." || newPath == ".." {
		newPath = ""
	}

	if len(newPath) > 0 && string(newPath[0]) == "/" {
		return newPath[1:]
	}

	return newPath
}

func watchTCPDumps(ctx context.Context, cs kubernetes.Interface, csConfig *restclient.Config, ns string, dir string) error {
	options := metav1.ListOptions{
		LabelSelector: naming.ScyllaSelector().String(),
	}

	watcher, err := cs.CoreV1().Pods(ns).Watch(ctx, options)
	if err != nil {
		return fmt.Errorf("cannot create Pod event watcher: %w", err)
	}

	go func() {
		var m sync.Mutex
		// Key is pod/container name, true if currently logging it.
		//active := map[string]bool{}
		// Key is pod/container/container-id, true if we have ever started to capture its output.
		//started := map[string]bool{}

		check := func() {
			m.Lock()
			defer m.Unlock()

			pods, err := cs.CoreV1().Pods(ns).List(ctx, options)
			if err != nil {
				Errorf("can't get pod list in %s: %v", ns, err)
				return
			}
			for _, pod := range pods.Items {
				podPath := path.Join(dir, pod.ObjectMeta.Name)

				for i, c := range pod.Spec.Containers {
					if c.Name != naming.ScyllaContainerName {
						continue
					}

					if len(pod.Status.ContainerStatuses) <= i {
						continue
					}

					//name := pod.ObjectMeta.Name + "/" + c.Name
					containerID := pod.Status.ContainerStatuses[i].ContainerID
					_, after, ok := strings.Cut(containerID, "://")
					if ok {
						containerID = after
					}
					name := c.Name + "-" + containerID
					//id := name + "/" + pod.Status.ContainerStatuses[i].ContainerID
					//if active[name] ||
					//	// If we have worked on a container before and it has now terminated, then
					//	// there cannot be any new output and we can ignore it.
					//	(pod.Status.ContainerStatuses[i].State.Terminated != nil &&
					//		started[id]) ||
					//	// State.Terminated might not have been updated although the container already
					//	// stopped running. Also check whether the pod is deleted.
					//	(pod.DeletionTimestamp != nil && started[id]) ||
					//	// Don't attempt to get logs for a container unless it is running or has terminated.
					//	// Trying to get a log would just end up with an error that we would have to suppress.
					//	(pod.Status.ContainerStatuses[i].State.Running == nil &&
					//		pod.Status.ContainerStatuses[i].State.Terminated == nil) {
					//	Infof("pod tcpdump: skipping %s", name)
					//	continue
					//}

					src := "/mnt/tcpdump/log.out"
					dst := filepath.Join(podPath, name+".pcap")

					err = copyFromPod(cs.CoreV1(), ns, pod.ObjectMeta.Name, c.Name, src, dst)
					if err != nil {
						Warnf("pod tcpdump: %s: %v", name, err)
						continue
					}
					//active[name] = true
					//started[id] = true
				}
			}
		}
		check()
		for {
			select {
			case <-watcher.ResultChan():
				check()
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func gatherLogs(ctx context.Context, cs kubernetes.Interface, ns string, dir string) (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(ctx)

	err := copyPodLogs(ctx, cs, ns, dir)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("can't watch for logs: %w", err)
	}
	return cancel, nil
}

func copyPodLogs(ctx context.Context, cs kubernetes.Interface, ns string, dir string) error {
	options := metav1.ListOptions{}

	watcher, err := cs.CoreV1().Pods(ns).Watch(ctx, options)
	if err != nil {
		return fmt.Errorf("cannot create Pod event watcher: %w", err)
	}

	go func() {
		var m sync.Mutex
		// Key is pod/container name, true if currently logging it.
		active := map[string]bool{}
		// Key is pod/container/container-id, true if we have ever started to capture its output.
		started := map[string]bool{}

		check := func() {
			m.Lock()
			defer m.Unlock()

			pods, err := cs.CoreV1().Pods(ns).List(ctx, options)
			if err != nil {
				Errorf("can't get pod list in %s: %v", ns, err)
				return
			}

			for _, pod := range pods.Items {
				podPath := path.Join(dir, pod.ObjectMeta.Name)
				for i, c := range pod.Spec.Containers {
					// sanity check, array should have entry for each container
					if len(pod.Status.ContainerStatuses) <= i {
						continue
					}
					name := pod.ObjectMeta.Name + "/" + c.Name
					id := name + "/" + pod.Status.ContainerStatuses[i].ContainerID
					if active[name] ||
						// If we have worked on a container before and it has now terminated, then
						// there cannot be any new output and we can ignore it.
						(pod.Status.ContainerStatuses[i].State.Terminated != nil &&
							started[id]) ||
						// State.Terminated might not have been updated although the container already
						// stopped running. Also check whether the pod is deleted.
						(pod.DeletionTimestamp != nil && started[id]) ||
						// Don't attempt to get logs for a container unless it is running or has terminated.
						// Trying to get a log would just end up with an error that we would have to suppress.
						(pod.Status.ContainerStatuses[i].State.Running == nil &&
							pod.Status.ContainerStatuses[i].State.Terminated == nil) {
						continue
					}
					readCloser, err := logsForPod(ctx, cs, ns, pod.ObjectMeta.Name,
						&v1.PodLogOptions{
							Container: c.Name,
							Follow:    true,
						})
					if err != nil {
						// We do get "normal" errors here, like trying to read too early.
						// We can ignore those.
						Warnf("pod log: %s: %v", name, err)
						continue
					}

					// Determine where we write. If this fails, we intentionally return without clearing
					// the active[name] flag, which prevents trying over and over again to
					// create the output file.
					var out io.Writer
					var closer io.Closer

					filename := path.Join(podPath, c.Name+".log")
					err = os.MkdirAll(path.Dir(filename), 0755)
					if err != nil && !os.IsExist(err) {
						Errorf("pod log: create directory for %s: %v", filename, err)
						return
					}
					// The test suite might run the same test multiple times,
					// so we have to append here.
					file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						Errorf("pod log: create file %s: %v", filename, err)
						return
					}
					closer = file
					out = file
					go func() {
						if closer != nil {
							defer closer.Close()
						}
						first := true
						defer func() {
							m.Lock()
							// If we never printed anything, then also skip the final message.
							if !first {
								fmt.Fprintf(out, "==== end of pod log for container %s ====\n", id)
							}
							active[name] = false
							m.Unlock()
							readCloser.Close()
						}()
						scanner := bufio.NewScanner(readCloser)
						for scanner.Scan() {
							line := scanner.Text()
							// Filter out the expected "end of stream" error message,
							// it would just confuse developers who don't know about it.
							// Same for attempts to read logs from a container that
							// isn't ready (yet?!).
							if !strings.HasPrefix(line, "rpc error: code = Unknown desc = Error: No such container:") &&
								!strings.HasPrefix(line, "rpc error: code = NotFound desc = an error occurred when try to find container") &&
								!strings.HasPrefix(line, "unable to retrieve container logs for ") &&
								!strings.HasPrefix(line, "Unable to retrieve container logs for ") {
								if first {
									// Because the same log might be written to multiple times
									// in different test instances, log an extra line to separate them.
									// Also provides some useful extra information.
									fmt.Fprintf(out, "==== start of pod log for container %s ====\n", id)
									first = false
								}
								fmt.Fprintf(out, "%s\n", line)
							}
						}
					}()
					active[name] = true
					started[id] = true
				}
			}
		}

		// Watch events to see whether we can start logging
		// and log interesting ones.
		check()
		for {
			select {
			case <-watcher.ResultChan():
				check()
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func logsForPod(ctx context.Context, cs kubernetes.Interface, ns, pod string, opts *v1.PodLogOptions) (io.ReadCloser, error) {
	return cs.CoreV1().Pods(ns).GetLogs(pod, opts).Stream(ctx)
}

func (f *Framework) deleteNamespace(ctx context.Context, ns *corev1.Namespace) {
	By("Destroying namespace %q.", ns.Name)
	var gracePeriod int64 = 0
	var propagation = metav1.DeletePropagationForeground
	err := f.KubeAdminClient().CoreV1().Namespaces().Delete(
		ctx,
		ns.Name,
		metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriod,
			PropagationPolicy:  &propagation,
			Preconditions: &metav1.Preconditions{
				UID: &ns.UID,
			},
		},
	)
	o.Expect(err).NotTo(o.HaveOccurred())

	// We have deleted only the namespace object but it is still there with deletionTimestamp set.

	By("Waiting for namespace %q to be removed.", ns.Name)
	err = WaitForObjectDeletion(ctx, f.DynamicAdminClient(), corev1.SchemeGroupVersion.WithResource("namespaces"), "", ns.Name, &ns.UID)
	o.Expect(err).NotTo(o.HaveOccurred())
	klog.InfoS("Namespace removed.", "Namespace", ns.Name)
}

func (f *Framework) beforeEach() {
	f.setupNamespace(context.Background())
}

func (f *Framework) afterEach() {
	if f.namespace == nil {
		return
	}

	f.watchLogsCancel()

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	defer func() {
		keepNamespace := false
		switch TestContext.DeleteTestingNSPolicy {
		case DeleteTestingNSPolicyNever:
			keepNamespace = true
		case DeleteTestingNSPolicyOnSuccess:
			if g.CurrentSpecReport().Failed() {
				keepNamespace = true
			}
		case DeleteTestingNSPolicyAlways:
		default:
		}

		if keepNamespace {
			By("Keeping namespace %q for debugging", f.Namespace())
			return
		}

		f.deleteNamespace(ctx, f.namespace)
		f.namespace = nil
		f.clientConfig = nil
	}()

	// Print events if the test failed.
	if g.CurrentSpecReport().Failed() {
		By(fmt.Sprintf("Collecting events from namespace %q.", f.namespace.Name))
		DumpEventsInNamespace(ctx, f.KubeAdminClient(), f.namespace.Name)
	}

	// CI can't keep namespaces alive because it could get out of resources for the other tests
	// so we need to collect the namespaced dump before destroying the namespace.
	// Collecting artifacts even for successful runs helps to verify if it went
	// as expected and the amount of data is bearable.
	if len(TestContext.ArtifactsDir) != 0 {
		By(fmt.Sprintf("Collecting dumps from namespace %q.", f.namespace.Name))

		d := path.Join(TestContext.ArtifactsDir, "e2e-namespaces")
		err := os.Mkdir(d, 0777)
		if err != nil && !os.IsExist(err) {
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		err = DumpNamespace(ctx, f.KubeAdminClient().Discovery(), f.DynamicAdminClient(), f.KubeAdminClient().CoreV1(), d, f.Namespace())
		o.Expect(err).NotTo(o.HaveOccurred())
	}
}
