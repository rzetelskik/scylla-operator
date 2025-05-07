// Copyright (C) 2025 ScyllaDB

package scylladbmanagertask

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/go-openapi/strfmt"
	"github.com/scylladb/scylla-manager/v3/pkg/managerclient"
	"github.com/scylladb/scylla-manager/v3/pkg/util/uuid"
	"github.com/scylladb/scylla-manager/v3/swagger/gen/scylla-manager/models"
	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/pointer"
	hashutil "github.com/scylladb/scylla-operator/pkg/util/hash"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (smtc *Controller) syncManager(
	ctx context.Context,
	smt *scyllav1alpha1.ScyllaDBManagerTask,
	status *scyllav1alpha1.ScyllaDBManagerTaskStatus,
) ([]metav1.Condition, error) {
	var progressingConditions []metav1.Condition

	smcrName, err := getScyllaDBManagerClusterRegistrationName(smt)
	if err != nil {
		return progressingConditions, fmt.Errorf("can't get ScyllaDBManagerClusterRegistration name: %w", err)
	}

	smcr, err := smtc.scyllaDBManagerClusterRegistrationLister.ScyllaDBManagerClusterRegistrations(smt.Namespace).Get(smcrName)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return progressingConditions, fmt.Errorf("can't get ScyllaDBManagerClusterRegistration: %w", err)
		}

		progressingConditions = append(progressingConditions, metav1.Condition{
			Type:               managerControllerProgressingCondition,
			Status:             metav1.ConditionTrue,
			ObservedGeneration: smt.Generation,
			Reason:             "AwaitingScyllaDBManagerClusterRegistrationCreation",
			Message:            fmt.Sprintf("Awaiting creation of ScyllaDBManagerClusterRegistration: %q.", naming.ManualRef(smt.Namespace, smcrName)),
		})

		return progressingConditions, nil
	}

	if smcr.Status.ClusterID == nil || len(*smcr.Status.ClusterID) == 0 {
		progressingConditions = append(progressingConditions, metav1.Condition{
			Type:               managerControllerProgressingCondition,
			Status:             metav1.ConditionTrue,
			ObservedGeneration: smt.Generation,
			Reason:             "AwaitingScyllaDBManagerClusterRegistrationClusterIDPropagation",
			Message:            fmt.Sprintf("Awaiting the ScyllaDB Manager's cluster ID to be propagated to the status of ScyllaDBManagerClusterRegistration: %q.", naming.ManualRef(smt.Namespace, smcrName)),
		})

		return progressingConditions, nil
	}

	clusterID := *smcr.Status.ClusterID
	requiredManagerTask, err := makeRequiredScyllaDBManagerClientTask(smt, clusterID)
	if err != nil {
		return progressingConditions, fmt.Errorf("can't make required ScyllaDB Manager task: %w", err)
	}

	managerClient, err := controllerhelpers.GetScyllaDBManagerClient(ctx, smcr)
	if err != nil {
		return progressingConditions, fmt.Errorf("can't get manager client: %w", err)
	}

	managerTask, found, err := getScyllaDBManagerTask(ctx, smt, clusterID, managerClient)
	if err != nil {
		return progressingConditions, fmt.Errorf("can't get ScyllaDB Manager task: %w", err)
	}

	if !found {
		// TODO: log name if different from smt.Name
		klog.V(2).InfoS("Creating ScyllaDB Manager task.", "ScyllaDBManagerTask", klog.KObj(smt))

		var managerTaskID uuid.UUID
		managerTaskID, err = managerClient.CreateTask(ctx, clusterID, requiredManagerTask)
		if err != nil {
			return progressingConditions, fmt.Errorf("can't create ScyllaDB Manager task: %w", err)
		}

		status.TaskID = pointer.Ptr(managerTaskID.String())
		return progressingConditions, nil
	}

	var managerTaskID uuid.UUID
	managerTaskID, err = uuid.Parse(managerTask.ID)
	if err != nil {
		return progressingConditions, fmt.Errorf("can't parse ScyllaDB Manager task ID: %w", err)
	}

	ownerUIDLabelValue, hasOwnerUIDLabel := managerTask.Labels[naming.OwnerUIDLabel]
	if !hasOwnerUIDLabel {
		klog.Warningf("ScyllaDB Manager task %q is missing the owner UID label. Deleting it to avoid a name collision.", managerTask.Name)

		err = managerClient.DeleteTask(ctx, clusterID, managerTask.Type, managerTaskID)
		if err != nil {
			// TODO: check if name differs
			return progressingConditions, fmt.Errorf("can't delete ScyllaDB Manager task %q: %w", managerTask.Name, err)
		}

		progressingConditions = append(progressingConditions, metav1.Condition{
			Type:               managerControllerProgressingCondition,
			Status:             metav1.ConditionTrue,
			ObservedGeneration: smcr.Generation,
			Reason:             "DeletedCollidingScyllaDBManagerTask",
			Message:            "Deleted a colliding ScyllaDB Manager task with no OwnerUID label.",
		})
		return progressingConditions, nil
	}

	if ownerUIDLabelValue == string(smt.UID) && requiredManagerTask.Labels[naming.ManagedHash] == managerTask.Labels[naming.ManagedHash] {
		// Cluster matches the desired state, nothing to do.
		return progressingConditions, nil
	}

	if ownerUIDLabelValue != string(smt.UID) {
		// Ideally, we wouldn't do anything here as this is error-prone and might hinder discovering bugs.
		// However, the task could have been created by the legacy component (manager-controller), so we update it to become a new owner without disrupting the state.
		klog.Warningf("Task %q already exists in ScyllaDB Manager state and has an owner UID label (%q), but it has a different owner. ScyllaDBManagerTask %q will adopt it.", managerTask.Name, ownerUIDLabelValue, klog.KObj(smt))
	}

	requiredManagerTask.ID = managerTask.ID

	// TODO: check if name differs
	klog.V(2).InfoS("Updating ScyllaDB Manager task.", "ScyllaDBManagerTask", klog.KObj(smt), "ScyllaDBManagerTaskName", requiredManagerTask.Name, "ScyllaDBManagerTaskID", requiredManagerTask.ID)
	err = managerClient.UpdateTask(ctx, clusterID, requiredManagerTask)
	if err != nil {
		return progressingConditions, fmt.Errorf("can't update ScyllaDB Manager task %q: %w", requiredManagerTask.Name, err)
	}

	return progressingConditions, nil
}

func getScyllaDBManagerClusterRegistrationName(smt *scyllav1alpha1.ScyllaDBManagerTask) (string, error) {
	switch smt.Spec.ScyllaDBClusterRef.Kind {
	case naming.ScyllaDBDatacenterKind:
		return naming.ScyllaDBManagerClusterRegistrationNameForScyllaDBDatacenter(&scyllav1alpha1.ScyllaDBDatacenter{
			ObjectMeta: metav1.ObjectMeta{
				Name: smt.Spec.ScyllaDBClusterRef.Name,
			},
		})

	default:
		return "", fmt.Errorf("unsupported scyllaDBClusterRef kind: %q", smt.Spec.ScyllaDBClusterRef.Kind)

	}
}

func getScyllaDBManagerTask(ctx context.Context, smt *scyllav1alpha1.ScyllaDBManagerTask, clusterID string, managerClient *managerclient.Client) (*managerclient.TaskListItem, bool, error) {
	taskType, err := scyllaDBManagerClientTaskType(smt)
	if err != nil {
		return nil, false, fmt.Errorf("can't get ScyllaDB Manager task type: %w", err)
	}

	var taskID string
	if smt.Status.TaskID != nil {
		taskID = *smt.Status.TaskID
	}

	tasks, err := managerClient.ListTasks(ctx, clusterID, taskType, true, "", taskID)
	if err != nil {
		return nil, false, fmt.Errorf("can't list ScyllaDB Manager tasks: %w", err)
	}

	if len(tasks.TaskListItemSlice) == 0 {
		return nil, false, nil
	}

	if len(taskID) > 0 && len(tasks.TaskListItemSlice) > 1 {
		return nil, false, fmt.Errorf("more than one task found in ScyllaDB Manager state with taskID: %s", taskID)
	}

	idx := slices.IndexFunc(tasks.TaskListItemSlice, func(item *models.TaskListItem) bool {
		// TODO: name
		return item.Name == smt.Name
	})

	if idx >= 0 {
		return tasks.TaskListItemSlice[idx], true, nil
	}

	return nil, false, nil
}

func makeRequiredScyllaDBManagerClientTask(smt *scyllav1alpha1.ScyllaDBManagerTask, clusterID string) (*managerclient.Task, error) {
	taskType, err := scyllaDBManagerClientTaskType(smt)
	if err != nil {
		return nil, fmt.Errorf("can't get ScyllaDB Manager task type: %w", err)
	}

	schedule := &managerclient.Schedule{}
	properties := map[string]any{}

	switch smt.Spec.Type {
	case scyllav1alpha1.ScyllaDBManagerTaskTypeBackup:

	case scyllav1alpha1.ScyllaDBManagerTaskTypeRepair:
		if smt.Spec.Repair.ScyllaDBManagerTaskSchedule.Cron != nil {
			schedule.Cron = *smt.Spec.Repair.ScyllaDBManagerTaskSchedule.Cron
		}

		if smt.Spec.Repair.ScyllaDBManagerTaskSchedule.StartDate != nil {
			schedule.StartDate = pointer.Ptr(strfmt.DateTime(smt.Spec.Repair.ScyllaDBManagerTaskSchedule.StartDate.Time))
		}

		if smt.Spec.Repair.ScyllaDBManagerTaskSchedule.NumRetries != nil {
			schedule.NumRetries = *smt.Spec.Repair.ScyllaDBManagerTaskSchedule.NumRetries
		}

		// TODO: override interval
		// TODO: override start date
		// TODO: override timezone

		if smt.Spec.Repair.DC != nil {
			properties["dc"] = smt.Spec.Repair.DC
		}

		if smt.Spec.Repair.Keyspace != nil {
			properties["keyspace"] = smt.Spec.Repair.Keyspace
		}

		if smt.Spec.Repair.FailFast != nil {
			properties["fail_fast"] = *smt.Spec.Repair.FailFast
		}

		if smt.Spec.Repair.Host != nil {
			properties["host"] = *smt.Spec.Repair.Host
		}

		if smt.Spec.Repair.Intensity != nil {
			properties["intensity"] = *smt.Spec.Repair.Intensity
		}

		intensityOverrideAnnotation, hasIntensityOverrideAnnotation := smt.Annotations[naming.ScyllaDBManagerTaskRepairIntensityOverrideAnnotation]
		if hasIntensityOverrideAnnotation {
			intensity, err := strconv.ParseFloat(intensityOverrideAnnotation, 64)
			if err != nil {
				return nil, fmt.Errorf("can't parse intensity: %w", err)
			}

			properties["intensity"] = intensity
		}

		if smt.Spec.Repair.Parallel != nil {
			properties["parallel"] = *smt.Spec.Repair.Parallel
		}

		if smt.Spec.Repair.SmallTableThreshold != nil {
			// TODO: make sure this is correct
			properties["small_table_threshold"] = smt.Spec.Repair.SmallTableThreshold.Value()
		}

	default:
		return nil, fmt.Errorf("unsupported scyllaDBManagerTask type: %q", smt.Spec.Type)

	}

	requiredManagerTask := &managerclient.Task{
		ClusterID: clusterID,
		Enabled:   true,
		Labels: map[string]string{
			naming.OwnerUIDLabel: string(smt.UID),
		},
		// TODO: test task name override
		Name: scyllaDBManagerClientTaskName(smt),
		// TODO: properties
		Properties: properties,
		// TODO: schedule
		Schedule: schedule,
		Type:     taskType,
	}

	managedHash, err := hashutil.HashObjects(requiredManagerTask)
	if err != nil {
		return nil, fmt.Errorf("can't calculate managed hash: %w", err)
	}
	requiredManagerTask.Labels[naming.ManagedHash] = managedHash

	return requiredManagerTask, nil
}

func scyllaDBManagerClientTaskName(smt *scyllav1alpha1.ScyllaDBManagerTask) string {
	nameOverrideAnnotationValue, hasNameOverrideAnnotation := smt.Annotations[naming.ScyllaDBManagerTaskNameOverrideAnnotation]
	if hasNameOverrideAnnotation {
		return nameOverrideAnnotationValue
	}

	return smt.Name
}

func scyllaDBManagerClientTaskType(smt *scyllav1alpha1.ScyllaDBManagerTask) (string, error) {
	switch smt.Spec.Type {
	case scyllav1alpha1.ScyllaDBManagerTaskTypeBackup:
		return managerclient.BackupTask, nil

	case scyllav1alpha1.ScyllaDBManagerTaskTypeRepair:
		return managerclient.RepairTask, nil

	default:
		return "", fmt.Errorf("unsupported ScyllaDBManagerTask type: %q", smt.Spec.Type)

	}
}
