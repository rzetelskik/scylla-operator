// Copyright (C) 2025 ScyllaDB

package scylladbmanager

import (
	"context"

	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/pointer"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (smc *Controller) calculateStatus(sm *scyllav1alpha1.ScyllaDBManager) *scyllav1alpha1.ScyllaDBManagerStatus {
	status := sm.Status.DeepCopy()
	status.ObservedGeneration = pointer.Ptr(sm.Generation)

	return status
}

func (smc *Controller) updateStatus(ctx context.Context, currentSM *scyllav1alpha1.ScyllaDBManager, status *scyllav1alpha1.ScyllaDBManagerStatus) error {
	if apiequality.Semantic.DeepEqual(&currentSM.Status, status) {
		return nil
	}

	sm := currentSM.DeepCopy()
	sm.Status = *status

	klog.V(2).InfoS("Updating status", "ScyllaDBManager", klog.KObj(sm))
	_, err := smc.scyllaClient.ScyllaDBManagers(sm.Namespace).UpdateStatus(ctx, sm, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	klog.V(2).InfoS("Status updated", "ScyllaDBManager", klog.KObj(sm))

	return nil
}
