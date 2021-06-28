package orphanedpv

import (
	"context"
	"fmt"
	"time"

	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	"github.com/scylladb/scylla-operator/pkg/controller/helpers"
	"github.com/scylladb/scylla-operator/pkg/naming"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type PVItem struct {
	PV          *corev1.PersistentVolume
	ServiceName string
}

func (opc *Controller) getPVsForScyllaCluster(ctx context.Context, sc *scyllav1.ScyllaCluster) ([]*PVItem, bool, error) {
	var errs []error
	requeue := false
	var pis []*PVItem
	for _, rack := range sc.Spec.Datacenter.Racks {
		stsName := naming.StatefulSetNameForRack(rack, sc)
		for i := int32(0); i < rack.Members; i++ {
			svcName := fmt.Sprintf("%s-%d", stsName, i)
			pvcName := fmt.Sprintf("%s-%s", naming.PVCTemplateName, svcName)
			pvc, err := opc.pvcLister.PersistentVolumeClaims(sc.Namespace).Get(pvcName)
			if err != nil {
				if apierrors.IsNotFound(err) {
					klog.V(2).InfoS("PVC not found", "PVC", fmt.Sprintf("%s/%s", sc.Namespace, pvcName))
					// We aren't watching PVCs so we need to requeue manually
					requeue = true
					continue
				}
				errs = append(errs, err)
				continue
			}

			if len(pvc.Spec.VolumeName) == 0 {
				klog.V(2).InfoS("PVC not bound yet", "PVC", klog.KObj(pvc))
				requeue = true
				continue
			}

			pv, err := opc.pvLister.Get(pvc.Spec.VolumeName)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			pis = append(pis, &PVItem{
				PV:          pv,
				ServiceName: svcName,
			})
		}
	}

	return pis, requeue, utilerrors.NewAggregate(errs)
}

func (opc *Controller) sync(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.ErrorS(err, "Failed to split meta namespace cache key", "cacheKey", key)
		return err
	}

	startTime := time.Now()
	klog.V(4).InfoS("Started syncing ScyllaCluster", "ScyllaCluster", klog.KRef(namespace, name), "startTime", startTime)
	defer func() {
		klog.V(4).InfoS("Finished syncing ScyllaCluster", "ScyllaCluster", klog.KRef(namespace, name), "duration", time.Since(startTime))
	}()

	sc, err := opc.scyllaLister.ScyllaClusters(namespace).Get(name)
	if apierrors.IsNotFound(err) {
		klog.V(2).InfoS("ScyllaCluster has been deleted", "ScyllaCluster", klog.KObj(sc))
		return nil
	}
	if err != nil {
		return err
	}

	if sc.DeletionTimestamp != nil {
		return nil
	}

	if !sc.Spec.AutomaticOrphanedNodeCleanup {
		klog.V(4).InfoS("ScyllaCluster doesn't have AutomaticOrphanedNodeCleanup enabled", "ScyllaCluster", klog.KRef(namespace, name))
		return nil
	}

	nodes, err := opc.nodeLister.List(labels.Everything())
	if err != nil {
		return err
	}

	var errs []error

	pis, requeue, err := opc.getPVsForScyllaCluster(ctx, sc)
	// Process at least some PVs even if there were errors retrieving the rest
	if err != nil {
		errs = append(errs, err)
	}

	for _, pi := range pis {
		orphaned, err := helpers.IsOrphanedPV(pi.PV, nodes)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if !orphaned {
			continue
		}

		// Verify that the node doesn't exist with a live call.
		freshNodes, err := opc.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{
			LabelSelector: labels.Everything().String(),
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}

		freshOrphaned, err := helpers.IsOrphanedPV(pi.PV, helpers.GetNodePointerArrayFromArray(freshNodes.Items))
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if !freshOrphaned {
			continue
		}

		_, err = opc.kubeClient.CoreV1().Services(sc.Namespace).Patch(
			ctx,
			pi.ServiceName,
			types.MergePatchType,
			[]byte(fmt.Sprintf(`{"metadata": {"labels": {%q: ""} } }`, naming.ReplaceLabel)),
			metav1.PatchOptions{},
		)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	err = utilerrors.NewAggregate(errs)
	if err != nil {
		return err
	}

	if requeue {
		// TODO: should be exponential backoff (synthetic error)
		opc.queue.AddAfter(key, 30*time.Second)
	}

	return nil
}
