// Copyright (C) 2025 ScyllaDB

package scylladbmanagertask

import (
	"context"
	"fmt"
	"time"

	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"github.com/scylladb/scylla-operator/pkg/internalapi"
	"github.com/scylladb/scylla-operator/pkg/naming"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (smtc *Controller) sync(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.ErrorS(err, "Failed to split meta namespace cache key", "cacheKey", key)
		return err
	}

	startTime := time.Now()
	klog.V(4).InfoS("Started syncing ScyllaDBManagerTask", "ScyllaDBManagerTask", klog.KRef(namespace, name), "startTime", startTime)
	defer func() {
		klog.V(4).InfoS("Finished syncing ScyllaDBManagerTask", "ScyllaDBManagerTask", klog.KRef(namespace, name), "duration", time.Since(startTime))
	}()

	smt, err := smtc.scyllaDBManagerTaskLister.ScyllaDBManagerTasks(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(2).InfoS("ScyllaDBManagerTask has been deleted", "ScyllaDBManagerTask", klog.KRef(namespace, name))
			return nil
		}

		return fmt.Errorf("can't get ScyllaDBManagerTask %q: %w", naming.ManualRef(namespace, name), err)
	}

	status := smtc.calculateStatus(smt)

	if smt.DeletionTimestamp != nil {
		// TODO: sync finalizer
		return smtc.updateStatus(ctx, smt, status)
	}

	// TODO: set finalizer

	var errs []error
	err = controllerhelpers.RunSync(
		&status.Conditions,
		managerControllerProgressingCondition,
		managerControllerDegradedCondition,
		smt.Generation,
		func() ([]metav1.Condition, error) {
			return smtc.syncManager(ctx, smt)
		},
	)
	if err != nil {
		errs = append(errs, fmt.Errorf("can't sync manager: %w", err))
	}

	var aggregationErrs []error
	progressingCondition, err := controllerhelpers.AggregateStatusConditions(
		controllerhelpers.FindStatusConditionsWithSuffix(status.Conditions, scyllav1alpha1.ProgressingCondition),
		metav1.Condition{
			Type:               scyllav1alpha1.ProgressingCondition,
			Status:             metav1.ConditionFalse,
			Reason:             internalapi.AsExpectedReason,
			Message:            "",
			ObservedGeneration: smt.Generation,
		},
	)
	if err != nil {
		aggregationErrs = append(aggregationErrs, fmt.Errorf("can't aggregate progressing conditions: %w", err))
	}

	degradedCondition, err := controllerhelpers.AggregateStatusConditions(
		controllerhelpers.FindStatusConditionsWithSuffix(status.Conditions, scyllav1alpha1.DegradedCondition),
		metav1.Condition{
			Type:               scyllav1alpha1.DegradedCondition,
			Status:             metav1.ConditionFalse,
			Reason:             internalapi.AsExpectedReason,
			Message:            "",
			ObservedGeneration: smt.Generation,
		},
	)
	if err != nil {
		aggregationErrs = append(aggregationErrs, fmt.Errorf("can't aggregate degraded conditions: %w", err))
	}

	if len(aggregationErrs) > 0 {
		errs = append(errs, aggregationErrs...)
		return utilerrors.NewAggregate(errs)
	}

	apimeta.SetStatusCondition(&status.Conditions, progressingCondition)
	apimeta.SetStatusCondition(&status.Conditions, degradedCondition)

	err = smtc.updateStatus(ctx, smt, status)
	if err != nil {
		errs = append(errs, fmt.Errorf("can't update status: %w", err))
	}

	return utilerrors.NewAggregate(errs)
}
