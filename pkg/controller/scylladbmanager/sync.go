// Copyright (C) 2025 ScyllaDB

package scylladbmanager

import (
	"context"
	"fmt"
	"time"

	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"github.com/scylladb/scylla-operator/pkg/internalapi"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (smc *Controller) sync(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.ErrorS(err, "Failed to split meta namespace cache key", "cacheKey", key)
		return err
	}

	startTime := time.Now()
	klog.V(4).InfoS("Started syncing ScyllaDBManager", "ScyllaDBManager", klog.KRef(namespace, name), "startTime", startTime)
	defer func() {
		klog.V(4).InfoS("Finished syncing ScyllaDBManager", "ScyllaDBManager", klog.KRef(namespace, name), "duration", time.Since(startTime))
	}()

	sm, err := smc.scyllaDBManagerLister.ScyllaDBManagers(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(2).InfoS("ScyllaDBManager has been deleted", "ScyllaDBManager", klog.KRef(namespace, name))
			return nil
		}

		return fmt.Errorf("can't get ScyllaDBManager %q: %w", naming.ManualRef(namespace, name), err)
	}

	// TODO: check if global annotation

	smSelector := labels.SelectorFromSet(labels.Set{
		naming.ScyllaDBManagerNameLabel: sm.Name,
	})

	type CT = *scyllav1alpha1.ScyllaDBManager

	scyllaDBManagerClusterRegistrationMap, err := controllerhelpers.GetObjects[CT, *scyllav1alpha1.ScyllaDBManagerClusterRegistration](
		ctx,
		sm,
		scyllaDBManagerControllerGVK,
		smSelector,
		controllerhelpers.ControlleeManagerGetObjectsFuncs[CT, *scyllav1alpha1.ScyllaDBManagerClusterRegistration]{
			GetControllerUncachedFunc: smc.scyllaClient.ScyllaDBManagers(sm.Namespace).Get,
			ListObjectsFunc:           smc.scyllaDBManagerClusterRegistrationLister.ScyllaDBManagerClusterRegistrations(sm.Namespace).List,
			PatchObjectFunc:           smc.scyllaClient.ScyllaDBManagerClusterRegistrations(sm.Namespace).Patch,
		},
	)
	if err != nil {
		return fmt.Errorf("can't get objects: %w", err)
	}

	status := smc.calculateStatus(sm)

	if sm.DeletionTimestamp != nil {
		return smc.updateStatus(ctx, sm, status)
	}

	var errs []error

	err = controllerhelpers.RunSync(
		&status.Conditions,
		scyllaDBManagerClusterRegistrationProgressingCondition,
		scyllaDBManagerClusterRegistrationDegradedCondition,
		sm.Generation,
		func() ([]metav1.Condition, error) {
			return smc.syncScyllaDBManagerClusterRegistrations(ctx, sm, scyllaDBManagerClusterRegistrationMap)
		},
	)
	if err != nil {
		errs = append(errs, fmt.Errorf("can't sync ScyllaDBManagerClusterRegistrations: %w", err))
	}

	var aggregationErrs []error

	progressingCondition, err := controllerhelpers.AggregateStatusConditions(
		controllerhelpers.FindStatusConditionsWithSuffix(status.Conditions, scyllav1alpha1.ProgressingCondition),
		metav1.Condition{
			Type:               scyllav1.ProgressingCondition,
			Status:             metav1.ConditionFalse,
			Reason:             internalapi.AsExpectedReason,
			Message:            "",
			ObservedGeneration: sm.Generation,
		},
	)
	if err != nil {
		return fmt.Errorf("can't aggregate progressing status conditions: %w", err)
	}

	degradedCondition, err := controllerhelpers.AggregateStatusConditions(
		controllerhelpers.FindStatusConditionsWithSuffix(status.Conditions, scyllav1alpha1.DegradedCondition),
		metav1.Condition{
			Type:               scyllav1.DegradedCondition,
			Status:             metav1.ConditionFalse,
			Reason:             internalapi.AsExpectedReason,
			Message:            "",
			ObservedGeneration: sm.Generation,
		},
	)
	if err != nil {
		return fmt.Errorf("can't aggregate degraded status conditions: %w", err)
	}

	if len(aggregationErrs) > 0 {
		errs = append(errs, aggregationErrs...)
		return utilerrors.NewAggregate(errs)
	}

	apimeta.SetStatusCondition(&status.Conditions, progressingCondition)
	apimeta.SetStatusCondition(&status.Conditions, degradedCondition)

	err = smc.updateStatus(ctx, sm, status)
	if err != nil {
		errs = append(errs, fmt.Errorf("can't update status: %w", err))
	}

	return utilerrors.NewAggregate(errs)
}
