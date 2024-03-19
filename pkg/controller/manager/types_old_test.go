// Copyright (C) 2024 ScyllaDB

package manager

import (
	"reflect"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/scylladb/scylla-manager/v3/pkg/managerclient"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	"github.com/scylladb/scylla-operator/pkg/helpers"
	"github.com/scylladb/scylla-operator/pkg/pointer"
	"k8s.io/apimachinery/pkg/api/equality"
)

func TestRepairTask_ToManager(t *testing.T) {
	t.Parallel()

	dateTimeFunc := func(t string) *strfmt.DateTime {
		return pointer.Ptr(helpers.Must(strfmt.ParseDateTime(t)))
	}

	tt := []struct {
		name          string
		repairTask    RepairTask
		expected      *managerclient.Task
		expectedError error
	}{
		{
			name: "fields and properties are propagated with NumRetries",
			repairTask: RepairTask{
				RepairTaskSpec: scyllav1.RepairTaskSpec{
					SchedulerTaskSpec: scyllav1.SchedulerTaskSpec{
						Name:       "repair_task_name",
						StartDate:  "1970-01-01T00:00:00.000Z",
						NumRetries: pointer.Ptr[int64](3),
						Cron:       "0 23 * * SAT",
						RetryWait:  "10m",
						Timezone:   "CET",
						Window:     []string{"22:00,00:00"},
					},
					DC:                  []string{"us-east1"},
					FailFast:            false,
					Intensity:           "1",
					Parallel:            1,
					Keyspace:            []string{"test"},
					SmallTableThreshold: "1GiB",
					Host:                pointer.Ptr("10.0.0.1"),
					Tags:                []string{"tag1", "tag2"},
				},
				ID: "repair_task_id",
			},
			expected: &managerclient.Task{
				ClusterID: "",
				Enabled:   true,
				ID:        "repair_task_id",
				Name:      "repair_task_name",
				Properties: map[string]interface{}{
					"dc": []string{
						"us-east1",
					},
					"intensity": float64(1),
					"parallel":  int64(1),
					"keyspace": []string{
						"test",
					},
					"small_table_threshold": int64(1073741824),
					"host":                  "10.0.0.1",
				},
				Schedule: &managerclient.Schedule{
					Cron:       "0 23 * * SAT",
					Interval:   "",
					NumRetries: 3,
					RetryWait:  "10m",
					StartDate:  dateTimeFunc("1970-01-01T00:00:00.000Z"),
					Timezone:   "CET",
					Window:     []string{"22:00,00:00"},
				},
				Tags: []string{"tag1", "tag2"},
				Type: "repair",
			},
			expectedError: nil,
		},
		{
			name: "fields and properties are propagated with FailFast",
			repairTask: RepairTask{
				RepairTaskSpec: scyllav1.RepairTaskSpec{
					SchedulerTaskSpec: scyllav1.SchedulerTaskSpec{
						Name:       "repair_task_name",
						StartDate:  "1970-01-01T00:00:00.000Z",
						NumRetries: pointer.Ptr[int64](3),
						Cron:       "0 23 * * SAT",
						RetryWait:  "10m",
						Timezone:   "CET",
						Window:     []string{"22:00,00:00"},
					},
					DC:                  []string{"us-east1"},
					FailFast:            true,
					Intensity:           "1",
					Parallel:            1,
					Keyspace:            []string{"test"},
					SmallTableThreshold: "1GiB",
					Host:                pointer.Ptr("10.0.0.1"),
					Tags:                []string{"tag1", "tag2"},
				},
				ID: "repair_task_id",
			},
			expected: &managerclient.Task{
				ClusterID: "",
				Enabled:   true,
				ID:        "repair_task_id",
				Name:      "repair_task_name",
				Properties: map[string]interface{}{
					"dc": []string{
						"us-east1",
					},
					"fail_fast": true,
					"intensity": float64(1),
					"parallel":  int64(1),
					"keyspace": []string{
						"test",
					},
					"small_table_threshold": int64(1073741824),
					"host":                  "10.0.0.1",
				},
				Schedule: &managerclient.Schedule{
					Cron:       "0 23 * * SAT",
					Interval:   "",
					NumRetries: 0,
					RetryWait:  "10m",
					StartDate:  dateTimeFunc("1970-01-01T00:00:00.000Z"),
					Timezone:   "CET",
					Window:     []string{"22:00,00:00"},
				},
				Tags: []string{"tag1", "tag2"},
				Type: "repair",
			},
			expectedError: nil,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			managerClientTask, err := tc.repairTask.ToManager()
			if !equality.Semantic.DeepEqual(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(managerClientTask, tc.expected) {
				t.Errorf("expected and got manager client tasks differ: %s", cmp.Diff(tc.expected, managerClientTask))
			}
		})
	}
}

func TestRepairTask_FromManager(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name          string
		managerTask   *managerclient.TaskListItem
		expected      *RepairTask
		expectedError error
	}{
		{
			name: "fields and properties are propagated",
			managerTask: &managerclient.TaskListItem{
				ClusterID:      "cluster_id",
				Enabled:        true,
				ErrorCount:     1,
				ID:             "repair_task_id",
				LastError:      pointer.Ptr(strfmt.NewDateTime()),
				LastSuccess:    pointer.Ptr(strfmt.NewDateTime()),
				Name:           "repair_task_name",
				NextActivation: pointer.Ptr(strfmt.NewDateTime()),
				Properties: map[string]interface{}{
					"dc": []string{
						"us-east1",
					},
					"fail_fast": true,
					"intensity": "1",
					"parallel":  1,
					"keyspace": []string{
						"test",
					},
					"small_table_threshold": "1073741824",
					"host":                  "10.0.0.1",
				},
				Retry: 1,
				Schedule: &managerclient.Schedule{
					Cron:       "0 23 * * SAT",
					NumRetries: 3,
					RetryWait:  "10m",
					StartDate:  pointer.Ptr(strfmt.NewDateTime()),
					Timezone:   "CET",
					Window:     []string{"22:00,00:00"},
				},
				Status:       managerclient.TaskStatusRunning,
				SuccessCount: 0,
				Suspended:    false,
				Tags:         []string{"tag1", "tag2"},
				Type:         "repair",
			},
			expected: &RepairTask{
				RepairTaskSpec: scyllav1.RepairTaskSpec{
					SchedulerTaskSpec: scyllav1.SchedulerTaskSpec{
						Name:       "repair_task_name",
						StartDate:  "1970-01-01T00:00:00.000Z",
						NumRetries: pointer.Ptr[int64](3),
						Cron:       "0 23 * * SAT",
						RetryWait:  "10m",
						Timezone:   "CET",
						Window:     []string{"22:00,00:00"},
					},
					DC:                  []string{"us-east1"},
					FailFast:            true,
					Intensity:           "1",
					Parallel:            1,
					Keyspace:            []string{"test"},
					SmallTableThreshold: "1073741824",
					Host:                pointer.Ptr("10.0.0.1"),
					Tags:                []string{"tag1", "tag2"},
				},
				ID: "repair_task_id",
			},
			expectedError: nil,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rt := &RepairTask{}
			err := rt.FromManager(tc.managerTask)
			if !equality.Semantic.DeepEqual(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(rt, tc.expected) {
				t.Errorf("expected and got repair task differ: %s", cmp.Diff(tc.expected, rt))
			}
		})
	}
}

func TestBackupTask_ToManager(t *testing.T) {
	t.Parallel()

	dateTimeFunc := func(t string) *strfmt.DateTime {
		return pointer.Ptr(helpers.Must(strfmt.ParseDateTime(t)))
	}

	tt := []struct {
		name          string
		backupTask    BackupTask
		expected      *managerclient.Task
		expectedError error
	}{
		{
			name: "fields and properties are propagated",
			backupTask: BackupTask{
				BackupTaskSpec: scyllav1.BackupTaskSpec{
					SchedulerTaskSpec: scyllav1.SchedulerTaskSpec{
						Name:       "backup_task_name",
						StartDate:  "1970-01-01T00:00:00.000Z",
						NumRetries: pointer.Ptr[int64](3),
						Cron:       "0 23 * * SAT",
						RetryWait:  "10m",
						Timezone:   "CET",
						Window:     []string{"22:00,00:00"},
					},
					DC:       []string{"us-east1"},
					Keyspace: []string{"test"},
					Location: []string{
						"gcs:test",
					},
					RateLimit: []string{
						"10",
						"us-east1:100",
					},
					Retention: 1,
					SnapshotParallel: []string{
						"10",
						"us-east1:100",
					},
					UploadParallel: []string{
						"10",
						"us-east1:100",
					},
					Tags: []string{"tag1", "tag2"},
				},
				ID: "backup_task_id",
			},
			expected: &managerclient.Task{
				ClusterID: "",
				Enabled:   true,
				ID:        "backup_task_id",
				Name:      "backup_task_name",
				Properties: map[string]interface{}{
					"location": []string{
						"gcs:test",
					},
					"keyspace":  []string{"test"},
					"retention": int64(1),
					"dc":        []string{"us-east1"},
					"rate_limit": []string{
						"10",
						"us-east1:100",
					},
					"snapshot_parallel": []string{
						"10",
						"us-east1:100",
					},
					"upload_parallel": []string{
						"10",
						"us-east1:100",
					},
				},
				Schedule: &managerclient.Schedule{
					Cron:       "0 23 * * SAT",
					Interval:   "",
					NumRetries: 3,
					RetryWait:  "10m",
					StartDate:  dateTimeFunc("1970-01-01T00:00:00.000Z"),
					Timezone:   "CET",
					Window:     []string{"22:00,00:00"},
				},
				Tags: []string{"tag1", "tag2"},
				Type: "backup",
			},
			expectedError: nil,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			managerClientTask, err := tc.backupTask.ToManager()
			if !equality.Semantic.DeepEqual(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(managerClientTask, tc.expected) {
				t.Errorf("expected and got manager client tasks differ: %s", cmp.Diff(tc.expected, managerClientTask))
			}
		})
	}
}

func TestBackupTask_FromManager(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name          string
		managerTask   *managerclient.TaskListItem
		expected      *BackupTask
		expectedError error
	}{
		{
			name: "fields and properties are propagated",
			managerTask: &managerclient.TaskListItem{
				ClusterID:      "cluster_id",
				Enabled:        true,
				ErrorCount:     1,
				ID:             "backup_task_id",
				LastError:      pointer.Ptr(strfmt.NewDateTime()),
				LastSuccess:    pointer.Ptr(strfmt.NewDateTime()),
				Name:           "backup_task_name",
				NextActivation: pointer.Ptr(strfmt.NewDateTime()),
				Properties: map[string]interface{}{
					"location": []string{
						"gcs:test",
					},
					"keyspace":  []string{"test"},
					"retention": 1,
					"dc":        []string{"us-east1"},
					"rate_limit": []string{
						"10",
						"us-east1:100",
					},
					"snapshot_parallel": []string{
						"10",
						"us-east1:100",
					},
					"upload_parallel": []string{
						"10",
						"us-east1:100",
					},
				},
				Retry: 1,
				Schedule: &managerclient.Schedule{
					Cron:       "0 23 * * SAT",
					NumRetries: 3,
					RetryWait:  "10m",
					StartDate:  pointer.Ptr(strfmt.NewDateTime()),
					Timezone:   "CET",
					Window:     []string{"22:00,00:00"},
				},
				Status:       managerclient.TaskStatusRunning,
				SuccessCount: 0,
				Suspended:    false,
				Tags:         []string{"tag1", "tag2"},
				Type:         "backup",
			},
			expected: &BackupTask{
				BackupTaskSpec: scyllav1.BackupTaskSpec{
					SchedulerTaskSpec: scyllav1.SchedulerTaskSpec{
						Name:       "backup_task_name",
						StartDate:  "1970-01-01T00:00:00.000Z",
						NumRetries: pointer.Ptr[int64](3),
						Cron:       "0 23 * * SAT",
						RetryWait:  "10m",
						Timezone:   "CET",
						Window:     []string{"22:00,00:00"},
					},
					DC:               []string{"us-east1"},
					Keyspace:         []string{"test"},
					Location:         []string{"gcs:test"},
					RateLimit:        []string{"10", "us-east1:100"},
					Retention:        1,
					SnapshotParallel: []string{"10", "us-east1:100"},
					UploadParallel:   []string{"10", "us-east1:100"},
					Tags:             []string{"tag1", "tag2"},
				},
				ID: "backup_task_id",
			},
			expectedError: nil,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			bt := &BackupTask{}
			err := bt.FromManager(tc.managerTask)
			if !equality.Semantic.DeepEqual(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(bt, tc.expected) {
				t.Errorf("expected and got backup task differ: %s", cmp.Diff(tc.expected, bt))
			}
		})
	}
}
