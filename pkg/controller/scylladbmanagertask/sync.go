// Copyright (C) 2025 ScyllaDB

package scylladbmanagertask

import (
	"context"
	"fmt"
	"time"

	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"github.com/scylladb/scylla-operator/pkg/helpers/slices"
	"github.com/scylladb/scylla-operator/pkg/internalapi"
	"github.com/scylladb/scylla-operator/pkg/naming"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func getLabels(smt *scyllav1alpha1.ScyllaDBManagerTask) labels.Set {
	return labels.Set{
		// TODO: improve
		"scylla-operator.scylladb.com/scylladb-manager-task-name": smt.Name,
	}
}

func getSelector(smt *scyllav1alpha1.ScyllaDBManagerTask) labels.Selector {
	return labels.SelectorFromSet(getLabels(smt))
}

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

		return fmt.Errorf("caniln't get ScyllaDBManagerTask %q: %w", naming.ManualRef(namespace, name), err)
	}

	selector := getSelector(smt)

	type CT = *scyllav1alpha1.ScyllaDBManagerTask
	var objectErrs []error

	jobs, err := controllerhelpers.GetObjects[CT, *batchv1.Job](
		ctx,
		smt,
		scyllaDBManagerTaskControllerGVK,
		selector,
		controllerhelpers.ControlleeManagerGetObjectsFuncs[CT, *batchv1.Job]{
			GetControllerUncachedFunc: smtc.scyllaClient.ScyllaDBManagerTasks(smt.Namespace).Get,
			ListObjectsFunc:           smtc.jobLister.Jobs(smt.Namespace).List,
			PatchObjectFunc:           smtc.kubeClient.BatchV1().Jobs(smt.Namespace).Patch,
		},
	)
	if err != nil {
		objectErrs = append(objectErrs, fmt.Errorf("can't get jobs: %w", err))
	}

	err = utilerrors.NewAggregate(objectErrs)
	if err != nil {
		return err
	}

	status := smtc.calculateStatus(smt)

	if smt.DeletionTimestamp != nil {
		return smtc.updateStatus(ctx, smt, status)
	}

	//if !smtc.hasFinalizer(smt.GetFinalizers()) {
	//err = smtc.addFinalizer(ctx, smt)
	//if err != nil {
	//	return fmt.Errorf("can't add finalizer: %w", err)
	//}
	//return nil
	//}

	var errs []error
	//err = controllerhelpers.RunSync(
	//	&status.Conditions,
	//	managerControllerProgressingCondition,
	//	managerControllerDegradedCondition,
	//	smt.Generation,
	//	func() ([]metav1.Condition, error) {
	//		return smtc.syncManager(ctx, smt, status)
	//	},
	//)
	//if err != nil {
	//	errs = append(errs, fmt.Errorf("can't sync manager: %w", err))
	//}

	err = controllerhelpers.RunSync(
		&status.Conditions,
		jobControllerProgressingCondition,
		jobControllerDegradedCondition,
		smt.Generation,
		func() ([]metav1.Condition, error) {
			return smtc.syncJobs(ctx, smt, jobs, status)
		},
	)
	if err != nil {
		errs = append(errs, fmt.Errorf("can't sync jobs: %w", err))
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

func (smtc *Controller) hasFinalizer(finalizers []string) bool {
	return slices.ContainsItem(finalizers, naming.ScyllaDBManagerTaskFinalizer)
}

func (smtc *Controller) addFinalizer(ctx context.Context, smt *scyllav1alpha1.ScyllaDBManagerTask) error {
	patch, err := controllerhelpers.AddFinalizerPatch(smt, naming.ScyllaDBManagerTaskFinalizer)
	if err != nil {
		return fmt.Errorf("can't create add finalizer patch: %w", err)
	}

	_, err = smtc.scyllaClient.ScyllaDBManagerTasks(smt.Namespace).Patch(ctx, smt.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("can't patch ScyllaDBManagerTask %q: %w", naming.ObjRef(smt), err)
	}

	klog.V(2).InfoS("Added finalizer to ScyllaDBManagerTask", "ScyllaDBManagerTask", klog.KObj(smt))
	return nil
}
