// Copyright (C) 2017 ScyllaDB

package manager

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/scylladb/scylla-manager/v3/pkg/managerclient"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	"github.com/scylladb/scylla-operator/pkg/pointer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestManagerSynchronization(t *testing.T) {
	const (
		clusterAuthToken = "token"
		namespace        = "test"
		name             = "cluster"
		clusterName      = "test/cluster"
		clusterID        = "cluster-id"
	)

	tcs := []struct {
		Name   string
		Spec   scyllav1.ScyllaClusterSpec
		Status scyllav1.ScyllaClusterStatus
		State  state

		Actions []action
		Requeue bool
	}{
		{
			Name:   "Empty manager, empty spec, add cluster and requeue",
			Spec:   scyllav1.ScyllaClusterSpec{},
			Status: scyllav1.ScyllaClusterStatus{},
			State:  state{},

			Requeue: true,
			Actions: []action{&addClusterAction{cluster: &managerclient.Cluster{Name: clusterName}}},
		},
		{
			Name: "Empty manager, task in spec, add cluster and requeue",
			Spec: scyllav1.ScyllaClusterSpec{
				Repairs: []scyllav1.RepairTaskSpec{},
			},
			Status: scyllav1.ScyllaClusterStatus{},
			State:  state{},

			Requeue: true,
			Actions: []action{&addClusterAction{cluster: &managerclient.Cluster{Name: clusterName}}},
		},
		{
			Name:   "Cluster registered in manager do nothing",
			Spec:   scyllav1.ScyllaClusterSpec{},
			Status: scyllav1.ScyllaClusterStatus{},
			State: state{
				Clusters: []*managerclient.Cluster{{
					Name:      clusterName,
					AuthToken: clusterAuthToken,
				}},
			},

			Requeue: false,
			Actions: nil,
		},
		{
			Name: "Cluster registered in manager but auth token is different, update and requeue",
			Spec: scyllav1.ScyllaClusterSpec{},
			Status: scyllav1.ScyllaClusterStatus{
				ManagerID: pointer.Ptr(clusterID),
			},
			State: state{
				Clusters: []*managerclient.Cluster{{
					ID:        clusterID,
					Name:      clusterName,
					AuthToken: "different-auth-token",
				}},
			},

			Requeue: true,
			Actions: []action{&updateClusterAction{cluster: &managerclient.Cluster{ID: clusterID}}},
		},
		{
			Name: "Name collision, delete old one, add new and requeue",
			Spec: scyllav1.ScyllaClusterSpec{},
			Status: scyllav1.ScyllaClusterStatus{
				ManagerID: pointer.Ptr(clusterID),
			},
			State: state{
				Clusters: []*managerclient.Cluster{{
					ID:   "different-id",
					Name: clusterName,
				}},
			},

			Requeue: true,
			Actions: []action{
				&deleteClusterAction{clusterID: "different-id"},
				&addClusterAction{cluster: &managerclient.Cluster{Name: clusterName}},
			},
		},
		{
			Name: "Schedule repair task",
			Spec: scyllav1.ScyllaClusterSpec{
				Repairs: []scyllav1.RepairTaskSpec{
					{
						TaskSpec: scyllav1.TaskSpec{
							Name: "my-repair",
							SchedulerTaskSpec: scyllav1.SchedulerTaskSpec{
								StartDate: "2006-01-02T15:04:05Z",
							},
						},
						SmallTableThreshold: "1GiB",
						DC:                  []string{"dc1"},
						FailFast:            false,
						Intensity:           "0.5",
						Keyspace:            []string{"keyspace1"},
					},
				},
			},
			Status: scyllav1.ScyllaClusterStatus{
				ManagerID: pointer.Ptr(clusterID),
			},
			State: state{
				Clusters: []*managerclient.Cluster{{
					ID:        clusterID,
					Name:      clusterName,
					AuthToken: clusterAuthToken,
				}},
			},

			Actions: []action{&addTaskAction{clusterID: clusterID, task: &managerclient.Task{Name: "my-repair"}}},
		},
		{
			Name: "Schedule backup task",
			Spec: scyllav1.ScyllaClusterSpec{
				Backups: []scyllav1.BackupTaskSpec{
					{
						TaskSpec: scyllav1.TaskSpec{
							Name: "my-backup",
							SchedulerTaskSpec: scyllav1.SchedulerTaskSpec{
								StartDate: "2006-01-02T15:04:05Z",
							},
						},
						DC:               []string{"dc1"},
						Keyspace:         []string{"keyspace1"},
						Location:         []string{"s3:abc"},
						RateLimit:        []string{"dc1:1"},
						Retention:        3,
						SnapshotParallel: []string{"dc1:1"},
						UploadParallel:   []string{"dc1:1"},
					},
				},
			},
			Status: scyllav1.ScyllaClusterStatus{
				ManagerID: pointer.Ptr(clusterID),
			},
			State: state{
				Clusters: []*managerclient.Cluster{{
					ID:        clusterID,
					Name:      clusterName,
					AuthToken: clusterAuthToken,
				}},
			},

			Actions: []action{&addTaskAction{clusterID: clusterID, task: &managerclient.Task{Name: "my-backup"}}},
		},
		{
			Name: "Update repair if it's already registered in Manager",
			Spec: scyllav1.ScyllaClusterSpec{
				Repairs: []scyllav1.RepairTaskSpec{
					{
						TaskSpec: scyllav1.TaskSpec{
							Name: "repair",
							SchedulerTaskSpec: scyllav1.SchedulerTaskSpec{
								StartDate: "2006-01-02T15:04:05Z",
							},
						},
						SmallTableThreshold: "1GiB",
						Intensity:           "0",
					},
				},
			},
			Status: scyllav1.ScyllaClusterStatus{
				ManagerID: pointer.Ptr(clusterID),
			},
			State: state{
				Clusters: []*managerclient.Cluster{{
					ID:        clusterID,
					Name:      clusterName,
					AuthToken: clusterAuthToken,
				}},
				RepairTasks: map[string]RepairTaskStatus{
					"repair": {
						TaskStatus: scyllav1.TaskStatus{
							SchedulerTaskStatus: scyllav1.SchedulerTaskStatus{
								StartDate: pointer.Ptr("2006-01-02T15:04:05Z"),
							},
							Name: "repair",
							ID:   pointer.Ptr("repair-id"),
						},
						Intensity:           pointer.Ptr("123"),
						SmallTableThreshold: pointer.Ptr("1GiB"),
					},
				},
			},

			Actions: []action{&updateTaskAction{clusterID: clusterID, task: &managerclient.Task{ID: "repair-id"}}},
		},
		{
			Name: "Do not update task when it didn't change",
			Spec: scyllav1.ScyllaClusterSpec{
				Repairs: []scyllav1.RepairTaskSpec{
					{
						TaskSpec: scyllav1.TaskSpec{
							Name: "repair",
							SchedulerTaskSpec: scyllav1.SchedulerTaskSpec{
								StartDate: "2021-01-01T11:11:11Z",
								Interval:  "0",
							},
						},
						SmallTableThreshold: "1GiB",
						Intensity:           "666",
						Parallel:            0,
					},
				},
			},
			Status: scyllav1.ScyllaClusterStatus{
				ManagerID: pointer.Ptr(clusterID),
			},
			State: state{
				Clusters: []*managerclient.Cluster{{
					ID:        clusterID,
					Name:      clusterName,
					AuthToken: clusterAuthToken,
				}},
				RepairTasks: map[string]RepairTaskStatus{
					"repair": {
						TaskStatus: scyllav1.TaskStatus{
							SchedulerTaskStatus: scyllav1.SchedulerTaskStatus{
								StartDate: pointer.Ptr("2021-01-01T11:11:11Z"),
								Interval:  pointer.Ptr("0"),
							},
							Name: "repair",
							ID:   pointer.Ptr("repair-id"),
						},
						FailFast:            pointer.Ptr(false),
						Intensity:           pointer.Ptr("666"),
						Parallel:            pointer.Ptr[int64](0),
						SmallTableThreshold: pointer.Ptr("1GiB"),
					},
				},
			},
			Actions: nil,
		},
		{
			Name: "Delete tasks from Manager unknown to spec",
			Spec: scyllav1.ScyllaClusterSpec{},
			Status: scyllav1.ScyllaClusterStatus{
				ManagerID: pointer.Ptr(clusterID),
			},
			State: state{
				Clusters: []*managerclient.Cluster{{
					ID:        clusterID,
					Name:      clusterName,
					AuthToken: clusterAuthToken,
				}},
				RepairTasks: map[string]RepairTaskStatus{
					"other-repair": {
						TaskStatus: scyllav1.TaskStatus{
							Name: "other-repair",
							SchedulerTaskStatus: scyllav1.SchedulerTaskStatus{
								StartDate: pointer.Ptr("2006-01-02T15:04:05Z"),
							},
							ID: pointer.Ptr("other-repair-id"),
						},
					},
				},
			},

			Actions: []action{&deleteTaskAction{clusterID: clusterID, taskID: "other-repair-id"}},
		},
		{
			Name: "Special 'now' startDate is not compared during update decision",
			Spec: scyllav1.ScyllaClusterSpec{
				Repairs: []scyllav1.RepairTaskSpec{
					{
						TaskSpec: scyllav1.TaskSpec{
							Name: "repair",
							SchedulerTaskSpec: scyllav1.SchedulerTaskSpec{
								StartDate: "now",
								Interval:  "0",
							},
						},
						Intensity:           "666",
						SmallTableThreshold: "1GiB",
						Parallel:            0,
					},
				},
			},
			Status: scyllav1.ScyllaClusterStatus{
				ManagerID: pointer.Ptr(clusterID),
			},
			State: state{
				Clusters: []*managerclient.Cluster{{
					ID:        clusterID,
					Name:      clusterName,
					AuthToken: clusterAuthToken,
				}},
				RepairTasks: map[string]RepairTaskStatus{
					"repair": {
						TaskStatus: scyllav1.TaskStatus{
							SchedulerTaskStatus: scyllav1.SchedulerTaskStatus{
								StartDate: pointer.Ptr("2021-01-01T11:11:11Z"),
								Interval:  pointer.Ptr("0"),
							},
							Name: "repair",
							ID:   pointer.Ptr("repair-id"),
						},
						FailFast:            pointer.Ptr(false),
						Intensity:           pointer.Ptr("666"),
						Parallel:            pointer.Ptr[int64](0),
						SmallTableThreshold: pointer.Ptr("1GiB"),
					},
				},
			},

			Actions: nil,
		},
		{
			Name: "Task gets updated when startDate is changed",
			Spec: scyllav1.ScyllaClusterSpec{
				Repairs: []scyllav1.RepairTaskSpec{
					{
						TaskSpec: scyllav1.TaskSpec{
							Name: "repair",
							SchedulerTaskSpec: scyllav1.SchedulerTaskSpec{
								StartDate: "2006-01-02T15:04:05Z",
								Interval:  "0",
							},
						},
						Intensity:           "666",
						SmallTableThreshold: "1GiB",
						Parallel:            0,
					},
				},
			},
			Status: scyllav1.ScyllaClusterStatus{
				ManagerID: pointer.Ptr(clusterID),
			},
			State: state{
				Clusters: []*managerclient.Cluster{{
					ID:        clusterID,
					Name:      clusterName,
					AuthToken: clusterAuthToken,
				}},
				RepairTasks: map[string]RepairTaskStatus{
					"repair": {
						TaskStatus: scyllav1.TaskStatus{
							SchedulerTaskStatus: scyllav1.SchedulerTaskStatus{
								StartDate: pointer.Ptr("2021-01-01T11:11:11Z"),
								Interval:  pointer.Ptr("0"),
							},
							Name: "repair",
							ID:   pointer.Ptr("repair-id"),
						},
						FailFast:            pointer.Ptr(false),
						Intensity:           pointer.Ptr("666"),
						Parallel:            pointer.Ptr[int64](0),
						SmallTableThreshold: pointer.Ptr("1GiB"),
					},
				},
			},

			Actions: []action{&updateTaskAction{clusterID: clusterID, task: &managerclient.Task{ID: "repair-id"}}},
		},
	}

	for i := range tcs {
		test := tcs[i]
		t.Run(test.Name, func(t *testing.T) {
			ctx := context.Background()
			cluster := &scyllav1.ScyllaCluster{
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
				Spec:       test.Spec,
				Status:     test.Status,
			}
			actions, requeue, err := runSync(ctx, cluster, clusterAuthToken, &test.State)
			if err != nil {
				t.Error(err)
			}
			if requeue != test.Requeue {
				t.Error(err, "Unexpected requeue")
			}
			if !cmp.Equal(actions, test.Actions, cmp.Comparer(actionComparer)) {
				t.Error(err, "Unexpected actions", cmp.Diff(actions, test.Actions, cmp.Comparer(actionComparer)))
			}
		})
	}
}

func actionComparer(a action, b action) bool {
	switch va := a.(type) {
	case *addClusterAction:
		vb := b.(*addClusterAction)
		return va.cluster.Name == vb.cluster.Name
	case *updateClusterAction:
		vb := b.(*updateClusterAction)
		return va.cluster.ID == vb.cluster.ID
	case *deleteClusterAction:
		vb := b.(*deleteClusterAction)
		return va.clusterID == vb.clusterID
	case *updateTaskAction:
		vb := b.(*updateTaskAction)
		return va.clusterID == vb.clusterID && va.task.ID == vb.task.ID
	case *addTaskAction:
		vb := b.(*addTaskAction)
		return va.clusterID == vb.clusterID && va.task.Name == vb.task.Name
	case *deleteTaskAction:
		vb := b.(*deleteTaskAction)
		return va.clusterID == vb.clusterID && va.taskID == vb.taskID
	default:
	}
	return false
}

func TestBackupTaskChanged(t *testing.T) {
	ts := []struct {
		name        string
		spec        *BackupTaskSpec
		managerTask *BackupTaskStatus
		expected    *BackupTaskStatus
	}{
		{
			name: "Task startDate is changed to one from manager state when prefix is 'now'",
			spec: &BackupTaskSpec{
				TaskSpec: scyllav1.TaskSpec{
					SchedulerTaskSpec: scyllav1.SchedulerTaskSpec{
						StartDate: "now",
						Interval:  "0",
					},
				},
				Retention: 3,
			},
			managerTask: &BackupTaskStatus{
				TaskStatus: scyllav1.TaskStatus{
					SchedulerTaskStatus: scyllav1.SchedulerTaskStatus{
						StartDate: pointer.Ptr("2021-01-01T11:11:11Z"),
						Interval:  pointer.Ptr("0"),
					},
				},
				Retention: pointer.Ptr[int64](3),
			},
			expected: &BackupTaskStatus{
				TaskStatus: scyllav1.TaskStatus{
					SchedulerTaskStatus: scyllav1.SchedulerTaskStatus{
						StartDate: pointer.Ptr("2021-01-01T11:11:11Z"),
						Interval:  pointer.Ptr("0"),
					},
				},
				Retention: pointer.Ptr[int64](3),
			},
		},
	}

	for i := range ts {
		test := ts[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			evaluateDates(test.spec, test.managerTask)
			got := test.spec.ToStatus()
			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("expected and got repair task statuses differ: %s", cmp.Diff(test.expected, got))
			}
		})
	}
}

func TestRepairTaskChanged(t *testing.T) {
	ts := []struct {
		name        string
		spec        *RepairTaskSpec
		managerTask *RepairTaskStatus
		expected    *RepairTaskStatus
	}{
		{
			name: "Task startDate is changed to one from manager state when prefix is 'now'",
			spec: &RepairTaskSpec{
				TaskSpec: scyllav1.TaskSpec{
					SchedulerTaskSpec: scyllav1.SchedulerTaskSpec{
						StartDate: "now",
						Interval:  "0",
					},
				},
				Intensity:           "1",
				SmallTableThreshold: "1GiB",
				Parallel:            0,
			},
			managerTask: &RepairTaskStatus{
				TaskStatus: scyllav1.TaskStatus{
					SchedulerTaskStatus: scyllav1.SchedulerTaskStatus{
						StartDate: pointer.Ptr("2021-01-01T11:11:11Z"),
					},
				},
			},
			expected: &RepairTaskStatus{
				TaskStatus: scyllav1.TaskStatus{
					SchedulerTaskStatus: scyllav1.SchedulerTaskStatus{
						StartDate: pointer.Ptr("2021-01-01T11:11:11Z"),
						Interval:  pointer.Ptr("0"),
					},
				},
				FailFast:            pointer.Ptr(false),
				Intensity:           pointer.Ptr("1"),
				Parallel:            pointer.Ptr[int64](0),
				SmallTableThreshold: pointer.Ptr("1GiB"),
			},
		},
	}

	for i := range ts {
		test := ts[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			evaluateDates(test.spec, test.managerTask)
			got := test.spec.ToStatus()
			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("expected and got backup task statuses differ: %s", cmp.Diff(test.expected, got))
			}
		})
	}
}
