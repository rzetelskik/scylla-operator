// Copyright (C) 2025 ScyllaDB

package scylladbmanagertask

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/scylladb/scylla-manager/v3/pkg/managerclient"
	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/helpers"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/pointer"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	validTimeValue = "2024-03-20T14:49:33.590Z"
	validTime      = helpers.Must(time.Parse(time.RFC3339, validTimeValue))
)

func Test_makeRequiredScyllaDBManagerClientTask(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name        string
		smt         *scyllav1alpha1.ScyllaDBManagerTask
		clusterID   string
		expected    *managerclient.Task
		expectedErr error
	}{
		{
			name: "repair, without properties",
			smt: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "repair",
					Namespace: "default",
					UID:       "uid",
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Kind: naming.ScyllaDBDatacenterKind,
						Name: "basic",
					},
					Type:   scyllav1alpha1.ScyllaDBManagerTaskTypeRepair,
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{},
				},
			},
			clusterID: "cluster-id",
			expected: &managerclient.Task{
				ClusterID: "cluster-id",
				Enabled:   true,
				ID:        "",
				Labels: map[string]string{
					"scylla-operator.scylladb.com/managed-hash": "+pVEWoxOjM5yK3A5D8GMUrmz6Gcgq3eDR2tEL6VBQboBk5/1jK554gpYYp90ukw1Z+DV3N7FHDgGFweULIeZsg==",
					"scylla-operator.scylladb.com/owner-uid":    "uid",
				},
				Name:       "repair",
				Properties: map[string]any{},
				Schedule:   &managerclient.Schedule{},
				Tags:       nil,
				Type:       "repair",
			},
			expectedErr: nil,
		},
		{
			name: "repair, with name override annotation",
			smt: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "repair",
					Namespace: "default",
					UID:       "uid",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskNameOverrideAnnotation: "override",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Kind: naming.ScyllaDBDatacenterKind,
						Name: "basic",
					},
					Type:   scyllav1alpha1.ScyllaDBManagerTaskTypeRepair,
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{},
				},
			},
			clusterID: "cluster-id",
			expected: &managerclient.Task{
				ClusterID: "cluster-id",
				Enabled:   true,
				ID:        "",
				Labels: map[string]string{
					"scylla-operator.scylladb.com/managed-hash": "+pVEWoxOjM5yK3A5D8GMUrmz6Gcgq3eDR2tEL6VBQboBk5/1jK554gpYYp90ukw1Z+DV3N7FHDgGFweULIeZsg==",
					"scylla-operator.scylladb.com/owner-uid":    "uid",
				},
				Name:       "override",
				Properties: map[string]any{},
				Schedule:   &managerclient.Schedule{},
				Tags:       nil,
				Type:       "repair",
			},
			expectedErr: nil,
		},
		{
			name: "repair, with properties",
			smt: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "repair",
					Namespace: "default",
					UID:       "uid",
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Kind: naming.ScyllaDBDatacenterKind,
						Name: "basic",
					},
					Type: scyllav1alpha1.ScyllaDBManagerTaskTypeRepair,
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
						ScyllaDBManagerTaskSchedule: scyllav1alpha1.ScyllaDBManagerTaskSchedule{
							Cron:       pointer.Ptr("0 23 * * SAT"),
							NumRetries: pointer.Ptr[int64](3),
							StartDate:  pointer.Ptr(metav1.NewTime(validTime)),
						},
						DC:                  []string{"dc1", "!otherdc*"},
						Keyspace:            []string{"keyspace", "!keyspace.table_prefix_*"},
						FailFast:            pointer.Ptr(true),
						Host:                pointer.Ptr("10.0.0.1"),
						Intensity:           pointer.Ptr[int64](1),
						Parallel:            pointer.Ptr[int64](1),
						SmallTableThreshold: pointer.Ptr(resource.MustParse("1Gi")),
					},
				},
			},
			clusterID: "cluster-id",
			expected: &managerclient.Task{
				ClusterID: "cluster-id",
				Enabled:   true,
				ID:        "",
				Labels: map[string]string{
					"scylla-operator.scylladb.com/managed-hash": "+pVEWoxOjM5yK3A5D8GMUrmz6Gcgq3eDR2tEL6VBQboBk5/1jK554gpYYp90ukw1Z+DV3N7FHDgGFweULIeZsg==",
					"scylla-operator.scylladb.com/owner-uid":    "uid",
				},
				Name: "repair",
				Properties: map[string]any{
					"dc":                    []string{"dc1", "!otherdc*"},
					"keyspace":              []string{"keyspace", "!keyspace.table_prefix_*"},
					"fail_fast":             true,
					"host":                  "10.0.0.1",
					"intensity":             int64(1),
					"parallel":              int64(1),
					"small_table_threshold": int64(1073741824),
				},
				Schedule: &managerclient.Schedule{
					Cron:       "0 23 * * SAT",
					Interval:   "",
					NumRetries: 3,
					RetryWait:  "",
					StartDate:  pointer.Ptr(strfmt.DateTime(validTime)),
					Timezone:   "",
					Window:     nil,
				},
				Tags: nil,
				Type: "repair",
			},
			expectedErr: nil,
		},
		{
			name: "repair, with valid intensity override annotation",
			smt: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "repair",
					Namespace: "default",
					UID:       "uid",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskRepairIntensityOverrideAnnotation: "0.5",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Kind: naming.ScyllaDBDatacenterKind,
						Name: "basic",
					},
					Type: scyllav1alpha1.ScyllaDBManagerTaskTypeRepair,
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
						ScyllaDBManagerTaskSchedule: scyllav1alpha1.ScyllaDBManagerTaskSchedule{},
						DC:                          []string{"dc1", "!otherdc*"},
						Keyspace:                    []string{"keyspace", "!keyspace.table_prefix_*"},
						FailFast:                    pointer.Ptr(true),
						Host:                        pointer.Ptr("10.0.0.1"),
						Intensity:                   pointer.Ptr[int64](1),
						Parallel:                    pointer.Ptr[int64](1),
						SmallTableThreshold:         pointer.Ptr(resource.MustParse("1Gi")),
					},
				},
			},
			clusterID: "cluster-id",
			expected: &managerclient.Task{
				ClusterID: "cluster-id",
				Enabled:   true,
				ID:        "",
				Labels: map[string]string{
					"scylla-operator.scylladb.com/managed-hash": "+pVEWoxOjM5yK3A5D8GMUrmz6Gcgq3eDR2tEL6VBQboBk5/1jK554gpYYp90ukw1Z+DV3N7FHDgGFweULIeZsg==",
					"scylla-operator.scylladb.com/owner-uid":    "uid",
				},
				Name: "repair",
				Properties: map[string]any{
					"dc":                    []string{"dc1", "!otherdc*"},
					"keyspace":              []string{"keyspace", "!keyspace.table_prefix_*"},
					"fail_fast":             true,
					"host":                  "10.0.0.1",
					"intensity":             float64(0.5),
					"parallel":              int64(1),
					"small_table_threshold": int64(1073741824),
				},
				Schedule: &managerclient.Schedule{},
				Tags:     nil,
				Type:     "repair",
			},
			expectedErr: nil,
		},
		{
			name: "repair, with invalid intensity override annotation",
			smt: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "repair",
					Namespace: "default",
					UID:       "uid",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskRepairIntensityOverrideAnnotation: "invalid",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Kind: naming.ScyllaDBDatacenterKind,
						Name: "basic",
					},
					Type: scyllav1alpha1.ScyllaDBManagerTaskTypeRepair,
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
						ScyllaDBManagerTaskSchedule: scyllav1alpha1.ScyllaDBManagerTaskSchedule{},
						DC:                          []string{"dc1", "!otherdc*"},
						Keyspace:                    []string{"keyspace", "!keyspace.table_prefix_*"},
						FailFast:                    pointer.Ptr(true),
						Host:                        pointer.Ptr("10.0.0.1"),
						Intensity:                   pointer.Ptr[int64](1),
						Parallel:                    pointer.Ptr[int64](1),
						SmallTableThreshold:         pointer.Ptr(resource.MustParse("1Gi")),
					},
				},
			},
			clusterID:   "cluster-id",
			expected:    nil,
			expectedErr: fmt.Errorf("can't parse intensity: %w", &strconv.NumError{Func: "ParseFloat", Num: "invalid", Err: strconv.ErrSyntax}),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := makeRequiredScyllaDBManagerClientTask(tc.smt, tc.clusterID)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s\n", cmp.Diff(tc.expectedErr, err))
			}

			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got ScyllaDB Manager client tasks differ:\n%s\n", cmp.Diff(tc.expected, got))
			}
		})
	}
}
