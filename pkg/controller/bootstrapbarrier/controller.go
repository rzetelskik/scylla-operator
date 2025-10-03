// Copyright (C) 2025 ScyllaDB

package bootstrapbarrier

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	scyllav1alpha1informers "github.com/scylladb/scylla-operator/pkg/client/scylla/informers/externalversions/scylla/v1alpha1"
	scyllav1alpha1listers "github.com/scylladb/scylla-operator/pkg/client/scylla/listers/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/controllertools"
	"github.com/scylladb/scylla-operator/pkg/naming"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
)

// TODO: deduplicate it
var serviceOrdinalRegex = regexp.MustCompile("^.*-([0-9]+)$")

type Controller struct {
	*controllertools.Observer

	namespace                              string
	serviceName                            string
	scyllaDBStatusReportSelectorLabelValue string
	bootstrapPreconditionCh                chan struct{}

	serviceLister              corev1listers.ServiceLister
	scyllaDBStatusReportLister scyllav1alpha1listers.ScyllaDBStatusReportLister
}

func NewController(
	namespace string,
	serviceName string,
	scyllaDBStatusReportSelectorLabelValue string,
	bootstrapPreconditionCh chan struct{},
	kubeClient kubernetes.Interface,
	serviceInformer corev1informers.ServiceInformer,
	scyllaDBStatusReportInformer scyllav1alpha1informers.ScyllaDBStatusReportInformer,
) (*Controller, error) {
	c := &Controller{
		namespace:                              namespace,
		serviceName:                            serviceName,
		scyllaDBStatusReportSelectorLabelValue: scyllaDBStatusReportSelectorLabelValue,
		bootstrapPreconditionCh:                bootstrapPreconditionCh,
		serviceLister:                          serviceInformer.Lister(),
		scyllaDBStatusReportLister:             scyllaDBStatusReportInformer.Lister(),
	}

	observer := controllertools.NewObserver(
		"scylladb-bootstrap-barrier",
		kubeClient.CoreV1().Events(corev1.NamespaceAll),
		c.Sync,
	)

	serviceHandler, err := serviceInformer.Informer().AddEventHandler(observer.GetGenericHandlers())
	if err != nil {
		return nil, fmt.Errorf("can't add Service event handler: %w", err)
	}
	observer.AddCachesToSync(serviceHandler.HasSynced)

	scyllaDBStatusReportHandler, err := scyllaDBStatusReportInformer.Informer().AddEventHandler(observer.GetGenericHandlers())
	if err != nil {
		return nil, fmt.Errorf("can't add ScyllaDBStatusReport event handler: %w", err)
	}
	observer.AddCachesToSync(scyllaDBStatusReportHandler.HasSynced)

	c.Observer = observer

	return c, nil
}

func (c *Controller) Sync(ctx context.Context) error {
	startTime := time.Now()
	klog.V(4).InfoS("Started syncing observer", "Name", c.Observer.Name(), "startTime", startTime)
	defer func() {
		klog.V(4).InfoS("Finished syncing observer", "Name", c.Observer.Name(), "duration", time.Since(startTime))
	}()

	svc, err := c.serviceLister.Services(c.namespace).Get(c.serviceName)
	if err != nil {
		return fmt.Errorf("can't get service %q: %w", c.serviceName, err)
	}

	// TODO: if override bootstrap annotation is true, write to channel and return

	svcDC, ok := svc.Labels[naming.DatacenterNameLabel]
	if !ok {
		return fmt.Errorf("service %q is missing label %q", naming.ObjRef(svc), naming.DatacenterNameLabel)
	}

	svcRack, ok := svc.Labels[naming.RackNameLabel]
	if !ok {
		return fmt.Errorf("service %q is missing label %q", naming.ObjRef(svc), naming.RackNameLabel)
	}

	svcOrdinalStrings := serviceOrdinalRegex.FindStringSubmatch(svc.Name)
	if len(svcOrdinalStrings) != 2 {
		return fmt.Errorf("can't parse ordinal from service %q", naming.ObjRef(svc))
	}

	svcOrdinal, err := strconv.Atoi(svcOrdinalStrings[1])
	if err != nil {
		return fmt.Errorf("can't parse ordinal from service %q: %w", naming.ObjRef(svc), err)
	}

	if _, ok := svc.Labels[naming.ReplacingNodeHostIDLabel]; ok {
		klog.V(2).InfoS("Node is replacing another node, proceeding without verifying the precondition.", "Service", c.serviceName)
		close(c.bootstrapPreconditionCh)
		return nil
	}

	scyllaDBStatusReports, err := c.scyllaDBStatusReportLister.ScyllaDBStatusReports(c.namespace).List(labels.SelectorFromSet(labels.Set{
		naming.ScyllaDBStatusReportSelectorLabel: c.scyllaDBStatusReportSelectorLabelValue,
	}))
	if err != nil {
		return fmt.Errorf("can't list ScyllaDBStatusReports: %w", err)
	}

	klog.V(4).InfoS("Verifying if bootstrap precondition is satisfied.", "Service", c.serviceName, "Datacenter", svcDC, "Rack", svcRack, "Ordinal", svcOrdinal)
	bootstrapPreconditionSatisfied := isBootstrapPreconditionSatisfied(scyllaDBStatusReports, svcDC, svcRack, svcOrdinal)
	if !bootstrapPreconditionSatisfied {
		klog.V(4).InfoS("Bootstrap precondition is not yet satisfied.", "Service", c.serviceName)
		return nil
	}

	klog.V(2).InfoS("Bootstrap precondition is satisfied, proceeding.", "Service", c.serviceName)
	close(c.bootstrapPreconditionCh)
	return nil
}

func isBootstrapPreconditionSatisfied(scyllaDBStatusReports []*scyllav1alpha1.ScyllaDBStatusReport, selfDC string, selfRack string, selfOrdinal int) bool {
	// allHostIDs is a set of host IDs of all nodes which appeared in the status report, including the reportees.
	allHostIDs := map[string]bool{}
	// reportingHostIDToObservedNodeStatusesMap maps a reporting node's host ID to a map of observed nodes' host IDs to their statuses as observed by the reporting node.
	reportingHostIDToObservedNodeStatusesMap := map[string]map[string]bool{}

	for _, ssr := range scyllaDBStatusReports {
		for _, rack := range ssr.Datacenter.Racks {
			for _, node := range rack.Nodes {
				if ssr.Datacenter.Name == selfDC && rack.Name == selfRack && node.Ordinal == selfOrdinal {
					// Skip self.
					// The node is bootstrapping, so it won't have a report nor a host ID propagated.
					continue
				}

				if node.HostID == nil {
					klog.V(4).InfoS("A required node is missing a host ID, can't proceed with bootstrap", "RequiredNodeDatacenter", ssr.Datacenter.Name, "RequiredNodeRack", rack.Name, "RequiredNodeOrdinal", node.Ordinal)
					return false
				}

				allHostIDs[*node.HostID] = true

				observedNodeHostIDToNodeStatusesMap := map[string]bool{}
				for _, observedNode := range node.ObservedNodes {
					allHostIDs[observedNode.HostID] = true
					observedNodeHostIDToNodeStatusesMap[observedNode.HostID] = observedNode.Status == scyllav1alpha1.NodeStatusUp
				}

				reportingHostIDToObservedNodeStatusesMap[*node.HostID] = observedNodeHostIDToNodeStatusesMap
			}
		}
	}

	allowNonReportingHostIDs := false
	if len(scyllaDBStatusReports) == 1 && scyllaDBStatusReports[0].Name == selfDC {
		allowNonReportingHostIDs = true
	}

	for hostID := range allHostIDs {
		nodeStatuses, ok := reportingHostIDToObservedNodeStatusesMap[hostID]
		if !ok {
			if allowNonReportingHostIDs {
				// In non-automated multi-datacenter deployments, we expect nodes from external DCs to appear in the status report as reportees only.
				// Users are expected to manually ensure the cross-DC precondition is satisfied.
				klog.V(4).InfoS("Non-required node's status report is missing. Skipping.", "HostID", hostID)
				continue
			}

			// The node's status report is missing.
			// We don't know what it thinks about other nodes, so we must assume the worst.
			klog.V(4).InfoS("Required node's status report is missing. Bootstrap precondition is not satisfied.", "HostID", hostID)
			return false
		}

		for otherHostID := range allHostIDs {
			if !nodeStatuses[otherHostID] {
				// The other node is either missing from this node's report or is considered DOWN.
				klog.V(4).InfoS("Node's status report is missing another node or considers it DOWN. Bootstrap precondition is not satisfied.", "HostID", hostID, "MissingOrDownHostID", otherHostID)
				return false
			}
		}
	}

	return true
}
