// Copyright (C) 2024 ScyllaDB

package operator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/scylladb/scylla-operator/pkg/controller/forcevolumesync"
	"github.com/scylladb/scylla-operator/pkg/genericclioptions"
	"github.com/scylladb/scylla-operator/pkg/helpers/slices"
	"github.com/scylladb/scylla-operator/pkg/signals"
	"github.com/scylladb/scylla-operator/pkg/version"
	"github.com/spf13/cobra"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

type ForceVolumeSyncOptions struct {
	genericclioptions.ClientConfig
	genericclioptions.InClusterReflection

	PodName       string
	VolumesToSync []string

	kubeClient kubernetes.Interface
}

func NewForceVolumeSyncOptions(streams genericclioptions.IOStreams) *ForceVolumeSyncOptions {
	return &ForceVolumeSyncOptions{
		ClientConfig: genericclioptions.NewClientConfig("scylla-operator-forcevolumesync"),

		PodName:       "",
		VolumesToSync: []string{},
	}
}

func (o *ForceVolumeSyncOptions) AddFlags(cmd *cobra.Command) {
	o.ClientConfig.AddFlags(cmd)
	o.InClusterReflection.AddFlags(cmd)

	cmd.Flags().StringVarP(&o.PodName, "pod-name", "", o.PodName, "")
	cmd.Flags().StringSliceVarP(&o.VolumesToSync, "volumes-to-sync", "", o.VolumesToSync, "")
}

func NewForceVolumeSyncCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewForceVolumeSyncOptions(streams)

	cmd := &cobra.Command{
		Use:   "run-forcevolumesync",
		Short: "Forces immediate content propagation of ConfigMap and Secret volumes to self when run from within a Pod.",
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
		Hidden:        true,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	o.AddFlags(cmd)

	return cmd
}

func (o *ForceVolumeSyncOptions) Validate(args []string) error {
	var errs []error

	errs = append(errs, o.ClientConfig.Validate())
	errs = append(errs, o.InClusterReflection.Validate())

	if len(o.PodName) == 0 {
		errs = append(errs, fmt.Errorf("pod-name can't be empty"))
	} else {
		podNameValidationErrs := apimachineryvalidation.NameIsDNS1035Label(o.PodName, false)
		if len(podNameValidationErrs) != 0 {
			errs = append(errs, fmt.Errorf("invalid pod name %q: %v", o.PodName, podNameValidationErrs))
		}
	}

	if len(o.VolumesToSync) == 0 {
		errs = append(errs, fmt.Errorf("volumes-to-sync can't be empty"))
	} else {
		volumesToSyncSet := sets.New[string]()
		for _, volumeToSync := range o.VolumesToSync {
			if volumesToSyncSet.Has(volumeToSync) {
				errs = append(errs, fmt.Errorf("volumes-to-sync has duplicate value: %q", volumeToSync))
			}

			volumesToSyncSet.Insert(volumeToSync)
		}

		volumesToSyncNameValidationErrs := slices.ConvertSlice(volumesToSyncSet.UnsortedList(), func(volumeToSync string) error {
			volumeNameValidationErrs := apimachineryvalidation.NameIsDNS1035Label(volumeToSync, false)
			if len(volumeNameValidationErrs) != 0 {
				return fmt.Errorf("invalid volume name %q: %v", volumeToSync, volumeNameValidationErrs)
			}

			return nil
		})

		errs = append(errs, volumesToSyncNameValidationErrs...)
	}

	return utilerrors.NewAggregate(errs)
}

func (o *ForceVolumeSyncOptions) Complete(args []string) error {
	err := o.ClientConfig.Complete()
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

func (o *ForceVolumeSyncOptions) Run(originalStreams genericclioptions.IOStreams, cmd *cobra.Command) (returnErr error) {
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

func (o *ForceVolumeSyncOptions) Execute(cmdCtx context.Context, originalStreams genericclioptions.IOStreams, cmd *cobra.Command) error {
	identityKubeInformers := informers.NewSharedInformerFactoryWithOptions(
		o.kubeClient,
		12*time.Hour,
		informers.WithNamespace(o.Namespace),
		informers.WithTweakListOptions(
			func(options *metav1.ListOptions) {
				options.FieldSelector = fields.OneTermEqualSelector("metadata.name", o.PodName).String()
			},
		),
	)

	namespacedKubeInformers := informers.NewSharedInformerFactoryWithOptions(
		o.kubeClient,
		12*time.Hour,
		informers.WithNamespace(o.Namespace),
	)

	forceVolumeSyncController, err := forcevolumesync.NewController(
		o.Namespace,
		o.PodName,
		o.VolumesToSync,
		o.kubeClient,
		identityKubeInformers.Core().V1().Pods(),
		namespacedKubeInformers.Core().V1().ConfigMaps(),
		namespacedKubeInformers.Core().V1().Secrets(),
	)
	if err != nil {
		return fmt.Errorf("can't create force volume sync controller: %w", err)
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, ctxCancel := context.WithCancel(cmdCtx)
	defer ctxCancel()

	/* Start informers */

	wg.Add(1)
	go func() {
		defer wg.Done()

		identityKubeInformers.Start(ctx.Done())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		namespacedKubeInformers.Start(ctx.Done())
	}()

	/* Start controllers */
	wg.Add(1)
	go func() {
		defer wg.Done()

		forceVolumeSyncController.Run(ctx)
	}()

	<-ctx.Done()

	return nil
}
