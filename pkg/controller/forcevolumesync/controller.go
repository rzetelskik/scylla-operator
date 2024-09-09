// Copyright (C) 2024 ScyllaDB

package forcevolumesync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"github.com/scylladb/scylla-operator/pkg/controllertools"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/pointer"
	hashutil "github.com/scylladb/scylla-operator/pkg/util/hash"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
)

const forceVolumeSyncManagedHashAnnotation = "internal.scylla-operator.scylladb.com/force-volume-sync-managed-hash"

type Controller struct {
	*controllertools.Observer

	namespace     string
	podName       string
	volumesToSync []string

	kubeClient kubernetes.Interface

	podLister       corev1listers.PodLister
	configMapLister corev1listers.ConfigMapLister
	secretLister    corev1listers.SecretLister
}

func NewController(
	namespace string,
	podName string,
	volumesToSync []string,
	kubeClient kubernetes.Interface,
	podInformer corev1informers.PodInformer,
	configMapInformer corev1informers.ConfigMapInformer,
	secretInformer corev1informers.SecretInformer,
) (*Controller, error) {
	controller := &Controller{
		namespace:       namespace,
		podName:         podName,
		volumesToSync:   volumesToSync,
		kubeClient:      kubeClient,
		podLister:       podInformer.Lister(),
		configMapLister: configMapInformer.Lister(),
		secretLister:    secretInformer.Lister(),
	}

	observer := controllertools.NewObserver(
		"forcevolumesync",
		kubeClient.CoreV1().Events(corev1.NamespaceAll),
		controller.sync,
	)

	podHandler, err := podInformer.Informer().AddEventHandler(observer.GetGenericHandlers())
	if err != nil {
		return nil, fmt.Errorf("can't add pod event handler: %w", err)
	}
	observer.AddCachesToSync(podHandler.HasSynced)

	// TODO: write proper handlers
	configMapHandler, err := configMapInformer.Informer().AddEventHandler(observer.GetGenericHandlers())
	if err != nil {
		return nil, fmt.Errorf("can't add config map event handler: %w", err)
	}
	observer.AddCachesToSync(configMapHandler.HasSynced)

	secretHandler, err := secretInformer.Informer().AddEventHandler(observer.GetGenericHandlers())
	if err != nil {
		return nil, fmt.Errorf("can't add secret event handler: %w", err)
	}
	observer.AddCachesToSync(secretHandler.HasSynced)

	controller.Observer = observer

	return controller, nil
}

func (c *Controller) sync(ctx context.Context) error {
	startTime := time.Now()
	klog.V(4).InfoS("Started syncing", "Name", c.Observer.Name(), "startTime", startTime)
	defer func() {
		klog.V(4).InfoS("Finished syncing", "Name", c.Observer.Name(), "duration", time.Since(startTime))
	}()

	pod, err := c.podLister.Pods(c.namespace).Get(c.podName)
	if err != nil {
		return fmt.Errorf("can't get pod %q: %w", naming.ManualRef(c.namespace, c.podName), err)
	}

	volumesToSyncSet := sets.New(c.volumesToSync...)

	var objectsToHash []interface{}
	for _, volume := range pod.Spec.Volumes {
		if !volumesToSyncSet.Has(volume.Name) {
			continue
		}

		if volume.ConfigMap != nil {
			cm, err := c.configMapLister.ConfigMaps(c.namespace).Get(volume.ConfigMap.Name)
			if err != nil {
				if !apierrors.IsNotFound(err) {
					return fmt.Errorf("can't get configmap %q: %w", naming.ManualRef(c.namespace, volume.ConfigMap.Name), err)
				}

				klog.ErrorS(err, "configMap has not been created yet", "Pod", klog.KObj(pod), "Volume", volume.Name, "ConfigMap", klog.KRef(c.namespace, volume.ConfigMap.Name))
				continue
			}

			objectsToHash = append(objectsToHash, cm.Data)
		} else if volume.Secret != nil {
			secret, err := c.secretLister.Secrets(c.namespace).Get(volume.Secret.SecretName)
			if err != nil {
				if !apierrors.IsNotFound(err) {
					return fmt.Errorf("can't get secret %q: %w", naming.ManualRef(c.namespace, volume.Secret.SecretName), err)
				}

				klog.ErrorS(err, "secret has not been created yet", "Pod", klog.KObj(pod), "Volume", volume.Name, "Secret", klog.KRef(c.namespace, volume.Secret.SecretName))
				continue
			}

			objectsToHash = append(objectsToHash, secret.Data)
		} else {
			klog.ErrorS(nil, "volume source is not supported", "Pod", klog.KObj(pod), "Volume", volume.Name, "SupportedVolumeSources", strings.Join([]string{"secret", "configMap"}, ","))
		}
	}

	hash, err := hashutil.HashObjects(objectsToHash)
	if err != nil {
		return fmt.Errorf("can't hash objects: %w", err)
	}

	if controllerhelpers.HasMatchingAnnotation(pod, forceVolumeSyncManagedHashAnnotation, hash) {
		return nil
	}

	patch, err := controllerhelpers.PrepareSetAnnotationPatch(pod, forceVolumeSyncManagedHashAnnotation, pointer.Ptr(hash))
	if err != nil {
		return fmt.Errorf("can't prepate set annotation patch: %w", err)
	}

	_, err = c.kubeClient.CoreV1().Pods(pod.Namespace).Patch(ctx, pod.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("can't patch pod %q: %w", naming.ObjRef(pod), err)
	}

	return nil
}
