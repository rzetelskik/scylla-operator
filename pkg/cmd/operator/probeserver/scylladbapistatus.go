package probeserver

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/scylladb/scylla-operator/pkg/genericclioptions"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/probeserver/scylladbapistatus"
	"github.com/scylladb/scylla-operator/pkg/signals"
	"github.com/scylladb/scylla-operator/pkg/version"
	"github.com/spf13/cobra"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	apierrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

type ScyllaDBAPIStatusOptions struct {
	ServeProbesOptions

	genericclioptions.ClientConfig
	genericclioptions.InClusterReflection
	ServiceName string
	AwaitPaths  []string

	mux        *http.ServeMux
	kubeClient kubernetes.Interface
}

func NewScyllaDBAPIStatusOptions(streams genericclioptions.IOStreams) *ScyllaDBAPIStatusOptions {
	mux := http.NewServeMux()

	return &ScyllaDBAPIStatusOptions{
		ServeProbesOptions: *NewServeProbesOptions(streams, naming.ScyllaDBAPIStatusProbePort, mux),
		ClientConfig:       genericclioptions.NewClientConfig("scylla-operator-scylladb-api-status-probe"),
		mux:                mux,
	}
}

func (o *ScyllaDBAPIStatusOptions) AddFlags(cmd *cobra.Command) {
	o.ServeProbesOptions.AddFlags(cmd)
	o.ClientConfig.AddFlags(cmd)
	o.InClusterReflection.AddFlags(cmd)

	cmd.Flags().StringVarP(&o.ServiceName, "service-name", "", o.ServiceName, "Name of the service corresponding to the managed node.")
	cmd.Flags().StringSliceVarP(&o.AwaitPaths, "await-paths", "", o.AwaitPaths, "Paths to await existence of. Until all exist, service will be considered healthy and unready.")
}

func NewScyllaDBAPIStatusCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewScyllaDBAPIStatusOptions(streams)

	cmd := &cobra.Command{
		Use:   "scylladb-api-status",
		Short: "Serves probes based on ScyllaDB API status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}

			err = o.Complete(args)
			if err != nil {
				return err
			}

			err = o.Run(streams, cmd)
			if err != nil {
				return err
			}

			return nil
		},

		SilenceErrors: true,
		SilenceUsage:  true,
	}

	o.AddFlags(cmd)

	return cmd
}

func (o *ScyllaDBAPIStatusOptions) Validate(args []string) error {
	var err error
	var errs []error

	errs = append(errs, o.ServeProbesOptions.Validate(args))
	errs = append(errs, o.ClientConfig.Validate())
	errs = append(errs, o.InClusterReflection.Validate())

	if len(o.ServiceName) == 0 {
		errs = append(errs, fmt.Errorf("service-name can't be empty"))
	} else {
		serviceNameValidationErrs := apimachineryvalidation.NameIsDNS1035Label(o.ServiceName, false)
		if len(serviceNameValidationErrs) != 0 {
			errs = append(errs, fmt.Errorf("invalid service name %q: %v", o.ServiceName, serviceNameValidationErrs))
		}
	}

	for _, path := range o.AwaitPaths {
		_, err = os.Stat(path)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			errs = append(errs, fmt.Errorf("can't stat %q: %w", path, err))
		}
	}

	return apierrors.NewAggregate(errs)
}

func (o *ScyllaDBAPIStatusOptions) Complete(args []string) error {
	err := o.ServeProbesOptions.Complete(args)
	if err != nil {
		return err
	}

	err = o.ClientConfig.Complete()
	if err != nil {
		return err
	}

	err = o.InClusterReflection.Complete()
	if err != nil {
		return err
	}

	o.kubeClient, err = kubernetes.NewForConfig(o.ProtoConfig)
	if err != nil {
		return fmt.Errorf("can't build kubernetes clientset: %w", err)
	}

	return nil
}

func (o *ScyllaDBAPIStatusOptions) Run(originalStreams genericclioptions.IOStreams, cmd *cobra.Command) (returnErr error) {
	klog.Infof("%s version %s", cmd.Name(), version.Get())
	cliflag.PrintFlags(cmd.Flags())

	stopCh := signals.StopChannel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-stopCh
		cancel()
	}()

	return o.Execute(ctx, originalStreams, cmd)
}

func (o *ScyllaDBAPIStatusOptions) Execute(ctx context.Context, originalStreams genericclioptions.IOStreams, cmd *cobra.Command) (returnErr error) {
	singleServiceKubeInformers := informers.NewSharedInformerFactoryWithOptions(
		o.kubeClient,
		12*time.Hour,
		informers.WithNamespace(o.Namespace),
		informers.WithTweakListOptions(
			func(options *metav1.ListOptions) {
				options.FieldSelector = fields.OneTermEqualSelector("metadata.name", o.ServiceName).String()
			},
		),
	)
	singleServiceInformer := singleServiceKubeInformers.Core().V1().Services()

	prober := scylladbapistatus.NewProber(
		o.Namespace,
		o.ServiceName,
		singleServiceInformer.Lister(),
		o.AwaitPaths,
	)

	o.mux.HandleFunc(naming.LivenessProbePath, prober.Healthz)
	o.mux.HandleFunc(naming.ReadinessProbePath, prober.Readyz)

	// Start informers.
	singleServiceKubeInformers.Start(ctx.Done())
	defer singleServiceKubeInformers.Shutdown()

	ok := cache.WaitForNamedCacheSync("Prober", ctx.Done(), singleServiceInformer.Informer().HasSynced)
	if !ok {
		return fmt.Errorf("error waiting for service informer cache to sync")
	}

	return o.ServeProbesOptions.Execute(ctx, originalStreams, cmd)
}
