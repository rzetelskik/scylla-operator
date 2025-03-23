// Copyright (C) 2025 ScyllaDB

package scyllacluster

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/pointer"
	"github.com/scylladb/scylla-operator/pkg/resourceapply"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const taskScheduleStartDateNowPrefix = "now"

func (scmc *Controller) syncScyllaDBManagerTasks(ctx context.Context, sc *scyllav1.ScyllaCluster, smts map[string]*scyllav1alpha1.ScyllaDBManagerTask) ([]metav1.Condition, error) {
	var progressingConditions []metav1.Condition

	for _, backupSpec := range sc.Spec.Backups {
		nameSuffix, err := naming.GenerateNameHash(string(scyllav1alpha1.ScyllaDBManagerTaskTypeBackup), backupSpec.Name)
		if err != nil {
			return progressingConditions, fmt.Errorf("can't generate ScyllaDBManagerTask name suffix: %w", err)
		}

		smt := &scyllav1alpha1.ScyllaDBManagerTask{
			ObjectMeta: metav1.ObjectMeta{
				Name:        fmt.Sprintf("%s-%s", sc.Name, nameSuffix),
				Namespace:   sc.Namespace,
				Annotations: map[string]string{},
				Labels:      map[string]string{},
			},
			Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
				ScyllaDBCluster: scyllav1alpha1.LocalScyllaDBReference{
					Name: sc.Name,
					Kind: "ScyllaDBDatacenter",
				},
				Type: scyllav1alpha1.ScyllaDBManagerTaskTypeBackup,
				Backup: &scyllav1alpha1.ScyllaDBManagerBackupTaskOptions{
					ScyllaDBManagerTaskSchedule: scyllav1alpha1.ScyllaDBManagerTaskSchedule{},
					DC:                          slices.Clone(backupSpec.DC),
					Keyspace:                    slices.Clone(backupSpec.Keyspace),
					Location:                    slices.Clone(backupSpec.Location),
					RateLimit:                   slices.Clone(backupSpec.RateLimit),
					Retention:                   pointer.Ptr(backupSpec.Retention),
					SnapshotParallel:            slices.Clone(backupSpec.SnapshotParallel),
					UploadParallel:              slices.Clone(backupSpec.UploadParallel),
				},
			},
			// Status is reconciled by the controllers.
			Status: scyllav1alpha1.ScyllaDBManagerTaskStatus{},
		}

		if backupSpec.Cron != nil {
			smt.Spec.Backup.Cron = pointer.Ptr(*backupSpec.Cron)
		}

		if backupSpec.NumRetries != nil {
			smt.Spec.Backup.NumRetries = pointer.Ptr(*backupSpec.NumRetries)
		}

		// TODO: make common for schedule
		if backupSpec.StartDate != nil {
			smt.Annotations[naming.TransformScyllaClusterToScyllaDBManagerTaskScheduleStartDateAnnotation] = *backupSpec.StartDate
		}

		if backupSpec.Interval != nil {
			smt.Annotations[naming.TransformScyllaClusterToScyllaDBManagerTaskScheduleIntervalAnnotation] = *backupSpec.Interval
		}

		if backupSpec.Timezone != nil {
			smt.Annotations[naming.TransformScyllaClusterToScyllaDBManagerTaskScheduleTimezoneAnnotation] = *backupSpec.Timezone
		}

		maps.Copy(smt.Labels, naming.ClusterLabelsForScyllaCluster(sc))
		smt.SetOwnerReferences([]metav1.OwnerReference{
			*metav1.NewControllerRef(sc, scyllaClusterControllerGVK),
		})

		_, changed, err := resourceapply.ApplyScyllaDBManagerTask(ctx, scmc.scyllaClient.ScyllaV1alpha1(), scmc.scyllaDBManagerTaskLister, scmc.eventRecorder, smt, resourceapply.ApplyOptions{})
		if changed {
			controllerhelpers.AddGenericProgressingStatusCondition(&progressingConditions, scyllaDBManagerTaskControllerProgressingCondition, smt, "apply", sc.Generation)
		}
		if err != nil {
			return progressingConditions, fmt.Errorf("can't apply ScyllaDBManagerTask: %w", err)
		}
	}

	for _, repairSpec := range sc.Spec.Repairs {
		nameSuffix, err := naming.GenerateNameHash(string(scyllav1alpha1.ScyllaDBManagerTaskTypeRepair), backupSpec.Name)
		if err != nil {
			return progressingConditions, fmt.Errorf("can't generate ScyllaDBManagerTask name suffix: %w", err)
		}

		smt := &scyllav1alpha1.ScyllaDBManagerTask{
			ObjectMeta: metav1.ObjectMeta{
				Name:        fmt.Sprintf("%s-%s", sc.Name, nameSuffix),
				Namespace:   sc.Namespace,
				Annotations: map[string]string{},
				Labels:      map[string]string{},
			},
			Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
				ScyllaDBCluster: scyllav1alpha1.LocalScyllaDBReference{
					Name: sc.Name,
					Kind: "ScyllaDBDatacenter",
				},
				Type: scyllav1alpha1.ScyllaDBManagerTaskTypeRepair,
				Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
					ScyllaDBManagerTaskSchedule: scyllav1alpha1.ScyllaDBManagerTaskSchedule{},
					DC:                          slices.Clone(repairSpec.DC),
					Keyspace:                    slices.Clone(repairSpec.Keyspace),
					FailFast:                    pointer.Ptr(repairSpec.FailFast),
					Parallel:                    pointer.Ptr(repairSpec.Parallel),
				},
			},
			// Status is reconciled by the controllers.
			Status: scyllav1alpha1.ScyllaDBManagerTaskStatus{},
		}

		if repairSpec.Host != nil {
			smt.Spec.Repair.Host = pointer.Ptr(*repairSpec.Host)
		}

		if len(repairSpec.Intensity) != 0 {
			smt.Annotations[naming.TransformScyllaClusterToScyllaDBManagerTaskRepairIntensityAnnotation] = repairSpec.Intensity
		}

		if len(repairSpec.SmallTableThreshold) != 0 {
			smt.Annotations[naming.TransformScyllaClusterToScyllaDBManagerTaskRepairSmallTableThresholdAnnotation] = repairSpec.SmallTableThreshold
		}

		if repairSpec.StartDate != nil {
			smt.Annotations[naming.TransformScyllaClusterToScyllaDBManagerTaskScheduleStartDateAnnotation] = *repairSpec.StartDate
		}

		if repairSpec.Interval != nil {
			smt.Annotations[naming.TransformScyllaClusterToScyllaDBManagerTaskScheduleIntervalAnnotation] = *repairSpec.Interval
		}

		if repairSpec.Timezone != nil {
			smt.Annotations[naming.TransformScyllaClusterToScyllaDBManagerTaskScheduleTimezoneAnnotation] = *repairSpec.Timezone
		}
	}

	return progressingConditions, nil
}
