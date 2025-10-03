// Copyright (C) 2025 ScyllaDB

package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	scyllaversionedclient "github.com/scylladb/scylla-operator/pkg/client/scylla/clientset/versioned"
	"github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"github.com/scylladb/scylla-operator/pkg/genericclioptions"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/signals"
	"github.com/scylladb/scylla-operator/pkg/version"
	"github.com/spf13/cobra"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryutilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

type BootstrappedQueryResult []struct {
	Bootstrapped string `json:"bootstrapped"`
}

type BootstrapBarrierOptions struct {
	genericclioptions.ClientConfig
	genericclioptions.InClusterReflection

	ServiceName              string
	ScyllaDBStatusReportName string
	BootstrappedPath         string

	bootstrapped bool

	kubeClient   kubernetes.Interface
	scyllaClient scyllaversionedclient.Interface
}

func NewBootstrapBarrierOptions(streams genericclioptions.IOStreams) *BootstrapBarrierOptions {
	return &BootstrapBarrierOptions{
		ClientConfig: genericclioptions.NewClientConfig("bootstrap-barrier"),
	}
}

func (o *BootstrapBarrierOptions) AddFlags(cmd *cobra.Command) {
	o.ClientConfig.AddFlags(cmd)
	o.InClusterReflection.AddFlags(cmd)

	cmd.Flags().StringVarP(&o.ServiceName, "service-name", "", o.ServiceName, "Name of the service corresponding to the managed node.")
	cmd.Flags().StringVarP(&o.ScyllaDBStatusReportName, "scylladb-status-report-name", "", o.ScyllaDBStatusReportName, "Name of the ScyllaDBStatusReport corresponding to the cluster.")
	cmd.Flags().StringVarP(&o.BootstrappedPath, "bootstrapped-path", "", o.BootstrappedPath, "Path to containing the JSON result of scylla-sstable query.")
}

func NewBootstrapBarrierCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewBootstrapBarrierOptions(streams)

	cmd := &cobra.Command{
		Use:   "bootstrap-barrier",
		Short: "",
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

func (o *BootstrapBarrierOptions) Validate(args []string) error {
	var errs []error

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

	if len(o.ScyllaDBStatusReportName) == 0 {
		errs = append(errs, fmt.Errorf("cluster-status-configmap-name can't be empty"))
	} else {
		configMapNameValidationErrs := apimachineryvalidation.NameIsDNSSubdomain(o.ScyllaDBStatusReportName, false)
		if len(configMapNameValidationErrs) != 0 {
			errs = append(errs, fmt.Errorf("invalid configmap name %q: %v", o.ScyllaDBStatusReportName, configMapNameValidationErrs))
		}
	}

	if len(o.BootstrappedPath) == 0 {
		errs = append(errs, fmt.Errorf("sstable-system-local-dump-path can't be empty"))
	}

	return apimachineryutilerrors.NewAggregate(errs)
}

func (o *BootstrapBarrierOptions) Complete(args []string) error {
	var err error

	var rawData []byte
	rawData, err = os.ReadFile(o.BootstrappedPath)
	if err != nil {
		return fmt.Errorf("can't read boostrapped query result file %q: %w", o.BootstrappedPath, err)
	}

	var data BootstrappedQueryResult
	if len(rawData) != 0 {
		err = json.Unmarshal(rawData, &data)
		if err != nil {
			return fmt.Errorf("can't unmarshall bootstrapped query result file %q: %w", o.BootstrappedPath, err)
		}
	}

	// TODO: convert OR to AND?
	for _, res := range data {
		if res.Bootstrapped == "COMPLETED" {
			o.bootstrapped = true
		}
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

	o.scyllaClient, err = scyllaversionedclient.NewForConfig(o.RestConfig)
	if err != nil {
		return fmt.Errorf("can't build scylla clientset: %w", err)
	}

	return nil
}

func (o *BootstrapBarrierOptions) Run(originalStreams genericclioptions.IOStreams, cmd *cobra.Command) error {
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

func (o *BootstrapBarrierOptions) Execute(ctx context.Context, originalStreams genericclioptions.IOStreams, cmd *cobra.Command) error {
	startTime := time.Now()
	klog.V(4).InfoS("Started boostrap barrier", "Service", naming.ManualRef(o.Namespace, o.ServiceName), "startTime", startTime)
	defer func() {
		klog.V(4).InfoS("Finished bootstrap barrier", "Service", naming.ManualRef(o.Namespace, o.ServiceName), "duration", time.Since(startTime))
	}()

	if o.bootstrapped {
		klog.V(2).Info("Node is bootstrapped, skipping bootstrap barrier")
		return nil
	}

	svc, err := o.kubeClient.CoreV1().Services(o.Namespace).Get(ctx, o.ServiceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("can't get service %q: %w", naming.ManualRef(o.Namespace, o.ServiceName), err)
	}

	if _, ok := svc.Labels[naming.ReplacingNodeHostIDLabel]; ok {
		klog.V(2).InfoS("Node is replacing another node, skipping bootstrap barrier", "Service", naming.ManualRef(o.Namespace, o.ServiceName))
		return nil
	}

	klog.V(2).InfoS("Waiting for required bootstrap precondition to be satisfied", "Service", naming.ManualRef(o.Namespace, o.ServiceName))
	_, err = controllerhelpers.WaitForScyllaDBStatusReportState(
		ctx,
		o.scyllaClient.ScyllaV1alpha1().ScyllaDBStatusReports(o.Namespace),
		o.ScyllaDBStatusReportName,
		controllerhelpers.WaitForStateOptions{},
		func(ssr *scyllav1alpha1.ScyllaDBStatusReport) (bool, error) {
			return isBootstrapPreconditionMet(ssr), nil
		},
	)

	return nil
}

func isBootstrapPreconditionMet(statusReport *scyllav1alpha1.ScyllaDBStatusReport) bool {
	// hostIDs is a set of all IDs that appeared in the status report.
	hostIDs := map[string]bool{}
	// hostIDToNodeStatusesMap maps a node's host ID to a map of observed nodes' host IDs to their statuses as observed by the node.
	hostIDToNodeStatusesMap := map[string]map[string]bool{}

	for _, dc := range statusReport.Datacenters {
		for _, node := range dc.Nodes {
			hostIDs[node.HostID] = true

			observedNodeHostIDToNodeStatusesMap := map[string]bool{}
			for _, observedNode := range node.ObservedNodes {
				hostIDs[observedNode.HostID] = true
				observedNodeHostIDToNodeStatusesMap[observedNode.HostID] = observedNode.Status == scyllav1alpha1.NodeStatusUp
			}

			hostIDToNodeStatusesMap[node.HostID] = observedNodeHostIDToNodeStatusesMap
		}
	}

	for hostID := range hostIDs {
		nodeStatuses, ok := hostIDToNodeStatusesMap[hostID]
		if !ok {
			// This node did not report a status, but it must be a part of a cluster as it appeared in the status report.
			// We don't know what it thinks about other nodes, so we must assume the worst.
			return false
		}

		for otherHostID := range hostIDs {
			if !nodeStatuses[otherHostID] {
				// The other node is either missing from this node's report or is considered DOWN.
				return false
			}
		}
	}

	return true
}
