// Copyright (C) 2025 ScyllaDB

package validation

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/pointer"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestValidateScyllaDBManagerTask(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name                string
		scyllaDBManagerTask *scyllav1alpha1.ScyllaDBManagerTask
		expectedErrorList   field.ErrorList
		expectedErrorString string
	}{
		{
			name: "valid repair",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type:   "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{},
				},
			},
			expectedErrorList:   field.ErrorList{},
			expectedErrorString: ``,
		},
		{
			name: "valid backup",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup",
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Backup",
					Backup: &scyllav1alpha1.ScyllaDBManagerBackupTaskOptions{
						Location: []string{
							"gcs:test",
						},
					},
				},
			},
			expectedErrorList:   field.ErrorList{},
			expectedErrorString: ``,
		},
		{
			name: "unsupported task type",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "random",
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Unsupported",
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeNotSupported,
					Field:    "spec.type",
					BadValue: scyllav1alpha1.ScyllaDBManagerTaskType("Unsupported"),
					Detail:   `supported values: "Backup", "Repair"`,
				},
			},
			expectedErrorString: `spec.type: Unsupported value: "Unsupported": supported values: "Backup", "Repair"`,
		},

		{
			name: "missing required options for repair type",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Repair",
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeRequired,
					Field:    "spec.repair",
					BadValue: "",
					Detail:   `repair options are required when task type is "Repair"`,
				},
			},
			expectedErrorString: `spec.repair: Required value: repair options are required when task type is "Repair"`,
		},
		{
			name: "missing required options for backup type",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup",
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Backup",
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeRequired,
					Field:    "spec.backup",
					BadValue: "",
					Detail:   `backup options are required when task type is "Backup"`,
				},
			},
			expectedErrorString: `spec.backup: Required value: backup options are required when task type is "Backup"`,
		},
		{
			name: "repair options for backup task type",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup",
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Backup",
					Backup: &scyllav1alpha1.ScyllaDBManagerBackupTaskOptions{
						Location: []string{
							"gcs:test",
						},
					},
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeForbidden,
					Field:    "spec.repair",
					BadValue: "",
					Detail:   `repair options are forbidden when task type is not "Repair"`,
				},
			},
			expectedErrorString: `spec.repair: Forbidden: repair options are forbidden when task type is not "Repair"`,
		},
		{
			name: "backup options for repair task type",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type:   "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{},
					Backup: &scyllav1alpha1.ScyllaDBManagerBackupTaskOptions{
						Location: []string{
							"gcs:test",
						},
					},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeForbidden,
					Field:    "spec.backup",
					BadValue: "",
					Detail:   `backup options are forbidden when task type is not "Backup"`,
				},
			},
			expectedErrorString: `spec.backup: Forbidden: backup options are forbidden when task type is not "Backup"`,
		},
		{
			name: "empty backup location",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup",
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Backup",
					Backup: &scyllav1alpha1.ScyllaDBManagerBackupTaskOptions{
						Location: []string{},
					},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeRequired,
					Field:    "spec.backup.location",
					BadValue: "",
					Detail:   "location must not be empty",
				},
			},
			expectedErrorString: `spec.backup.location: Required value: location must not be empty`,
		},
		{
			name: "empty backup location item",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup",
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Backup",
					Backup: &scyllav1alpha1.ScyllaDBManagerBackupTaskOptions{
						Location: []string{
							"",
						},
					},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeRequired,
					Field:    "spec.backup.location[0]",
					BadValue: "",
					Detail:   "location must not be empty",
				},
			},
			expectedErrorString: `spec.backup.location[0]: Required value: location must not be empty`,
		},
		{
			name: "disabled location validation via annotation",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskBackupLocationDisableValidationAnnotation: naming.AnnotationValueTrue,
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type:   "Backup",
					Backup: &scyllav1alpha1.ScyllaDBManagerBackupTaskOptions{},
				},
			},
			expectedErrorList:   field.ErrorList{},
			expectedErrorString: ``,
		},
		{
			name: "invalid repair with negative intensity",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
						Intensity: pointer.Ptr[int64](-1),
					},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.repair.intensity",
					BadValue: int64(-1),
					Detail:   "can't be negative",
				},
			},
			expectedErrorString: `spec.repair.intensity: Invalid value: -1: can't be negative`,
		},
		{
			name: "valid repair intensity override",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskRepairIntensityOverrideAnnotation: "0.5",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type:   "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{},
				},
			},
			expectedErrorList:   field.ErrorList{},
			expectedErrorString: ``,
		},
		{
			name: "invalid repair intensity override",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskRepairIntensityOverrideAnnotation: "invalid",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type:   "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-repair-intensity-override]",
					BadValue: "invalid",
					Detail:   "must be a float",
				},
			},
			expectedErrorString: `metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-repair-intensity-override]: Invalid value: "invalid": must be a float`,
		},
		{
			name: "valid repair with small table threshold",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
						SmallTableThreshold: pointer.Ptr(resource.MustParse("1Gi")),
					},
				},
			},
			expectedErrorList:   field.ErrorList{},
			expectedErrorString: ``,
		},
		{
			name: "invalid repair with malformed small table threshold",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
						// TODO: look into this, why is this working correctly?
						// TODO: try with PVC or some other resource
						SmallTableThreshold: pointer.Ptr(resource.MustParse("100m")),
					},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.repair.smallTableThreshold",
					BadValue: "invalid",
					Detail:   "quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'",
				},
			},
			expectedErrorString: `spec.repair.smallTableThreshold: Invalid value: "invalid": quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'`,
		},
		{
			name: "repair intensity override with options intensity",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskRepairIntensityOverrideAnnotation: "0.5",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
						Intensity: pointer.Ptr[int64](1),
					},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeForbidden,
					Field:    "metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-repair-intensity-override]",
					BadValue: "",
					Detail:   "can't be used together with repair intensity",
				},
			},
			expectedErrorString: `metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-repair-intensity-override]: Forbidden: can't be used together with repair intensity`,
		},
		{
			name: "invalid name override annotation",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskNameOverrideAnnotation: "invalid-with-trailing-dash-",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type:   "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-name-override]",
					BadValue: "invalid-with-trailing-dash-",
					Detail:   `a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`,
				},
			},
			expectedErrorString: `metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-name-override]: Invalid value: "invalid-with-trailing-dash-": a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`,
		},
		{
			name: "invalid interval override without cron",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskScheduleIntervalOverrideAnnotation: "invalid",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type:   "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{},
				},
			},
			expectedErrorList:   field.ErrorList{},
			expectedErrorString: ``,
		},
		{
			name: "invalid interval override with repair cron",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskScheduleIntervalOverrideAnnotation: "invalid",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
						ScyllaDBManagerTaskSchedule: scyllav1alpha1.ScyllaDBManagerTaskSchedule{
							Cron: pointer.Ptr("0 0 * * *"),
						},
					},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-schedule-interval-override]",
					BadValue: "invalid",
					Detail:   "valid units are d, h, m, s",
				},
			},
			expectedErrorString: `metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-schedule-interval-override]: Invalid value: "invalid": valid units are d, h, m, s`,
		},
		{
			name: "non-zero interval override with repair cron",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskScheduleIntervalOverrideAnnotation: "24h",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
						ScyllaDBManagerTaskSchedule: scyllav1alpha1.ScyllaDBManagerTaskSchedule{
							Cron: pointer.Ptr("0 0 * * *"),
						},
					},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeForbidden,
					Field:    "metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-schedule-interval-override]",
					BadValue: "",
					Detail:   "can't be non-zero when cron is specified",
				},
			},
			expectedErrorString: `metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-schedule-interval-override]: Forbidden: can't be non-zero when cron is specified`,
		},
		{
			name: "zero interval override with repair cron",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskScheduleIntervalOverrideAnnotation: "0s",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
						ScyllaDBManagerTaskSchedule: scyllav1alpha1.ScyllaDBManagerTaskSchedule{
							Cron: pointer.Ptr("0 0 * * *"),
						},
					},
				},
			},
			expectedErrorList:   field.ErrorList{},
			expectedErrorString: ``,
		},
		{
			name: "invalid interval override with backup cron",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskScheduleIntervalOverrideAnnotation: "invalid",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Backup",
					Backup: &scyllav1alpha1.ScyllaDBManagerBackupTaskOptions{
						ScyllaDBManagerTaskSchedule: scyllav1alpha1.ScyllaDBManagerTaskSchedule{
							Cron: pointer.Ptr("0 0 * * *"),
						},
						Location: []string{
							"gcs:test",
						},
					},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-schedule-interval-override]",
					BadValue: "invalid",
					Detail:   "valid units are d, h, m, s",
				},
			},
			expectedErrorString: `metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-schedule-interval-override]: Invalid value: "invalid": valid units are d, h, m, s`,
		},
		{
			name: "non-zero interval override with repair cron",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskScheduleIntervalOverrideAnnotation: "24h",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Backup",
					Backup: &scyllav1alpha1.ScyllaDBManagerBackupTaskOptions{
						ScyllaDBManagerTaskSchedule: scyllav1alpha1.ScyllaDBManagerTaskSchedule{
							Cron: pointer.Ptr("0 0 * * *"),
						},
						Location: []string{
							"gcs:test",
						},
					},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeForbidden,
					Field:    "metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-schedule-interval-override]",
					BadValue: "",
					Detail:   "can't be non-zero when cron is specified",
				},
			},
			expectedErrorString: `metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-schedule-interval-override]: Forbidden: can't be non-zero when cron is specified`,
		},
		{
			name: "zero interval override with repair cron",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskScheduleIntervalOverrideAnnotation: "0s",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Backup",
					Backup: &scyllav1alpha1.ScyllaDBManagerBackupTaskOptions{
						ScyllaDBManagerTaskSchedule: scyllav1alpha1.ScyllaDBManagerTaskSchedule{
							Cron: pointer.Ptr("0 0 * * *"),
						},
						Location: []string{
							"gcs:test",
						},
					},
				},
			},
			expectedErrorList:   field.ErrorList{},
			expectedErrorString: ``,
		},
		{
			name: "valid timezone override with repair cron",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskScheduleTimezoneOverrideAnnotation: "Europe/Warsaw",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
						ScyllaDBManagerTaskSchedule: scyllav1alpha1.ScyllaDBManagerTaskSchedule{
							Cron: pointer.Ptr("0 0 * * *"),
						},
					},
				},
			},
			expectedErrorList:   field.ErrorList{},
			expectedErrorString: ``,
		},
		{
			name: "valid timezone override with backup cron",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskScheduleTimezoneOverrideAnnotation: "Europe/Warsaw",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Backup",
					Backup: &scyllav1alpha1.ScyllaDBManagerBackupTaskOptions{
						ScyllaDBManagerTaskSchedule: scyllav1alpha1.ScyllaDBManagerTaskSchedule{
							Cron: pointer.Ptr("0 0 * * *"),
						},
						Location: []string{
							"gcs:test",
						},
					},
				},
			},
			expectedErrorList:   field.ErrorList{},
			expectedErrorString: ``,
		},
		{
			name: "invalid timezone override",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskScheduleTimezoneOverrideAnnotation: "invalid",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type: "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
						ScyllaDBManagerTaskSchedule: scyllav1alpha1.ScyllaDBManagerTaskSchedule{
							Cron: pointer.Ptr("0 0 * * *"),
						},
					},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-schedule-timezone-override]",
					BadValue: "invalid",
					Detail:   "unknown time zone invalid",
				},
			},
			expectedErrorString: `metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-schedule-timezone-override]: Invalid value: "invalid": unknown time zone invalid`,
		},
		{
			name: "timezone override without cron",
			scyllaDBManagerTask: &scyllav1alpha1.ScyllaDBManagerTask{
				ObjectMeta: metav1.ObjectMeta{
					Name: "repair",
					Annotations: map[string]string{
						naming.ScyllaDBManagerTaskScheduleTimezoneOverrideAnnotation: "UTC",
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Name: "basic",
						Kind: "ScyllaDBDatacenter",
					},
					Type:   "Repair",
					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{},
				},
			},
			expectedErrorList: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeForbidden,
					Field:    "metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-schedule-timezone-override]",
					BadValue: "",
					Detail:   "can't be set when cron is not specified",
				},
			},
			expectedErrorString: `metadata.annotations[internal.scylla-operator.scylladb.com/scylladb-manager-task-schedule-timezone-override]: Forbidden: can't be set when cron is not specified`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			errList := ValidateScyllaDBManagerTask(tc.scyllaDBManagerTask)
			if !reflect.DeepEqual(errList, tc.expectedErrorList) {
				t.Errorf("expected and actual error lists differ: %s", cmp.Diff(tc.expectedErrorList, errList))
			}

			var errStr string
			if agg := errList.ToAggregate(); agg != nil {
				errStr = agg.Error()
			}
			if !reflect.DeepEqual(errStr, tc.expectedErrorString) {
				t.Errorf("expected and actual error strings differ: %s", cmp.Diff(tc.expectedErrorString, errStr))
			}
		})
	}
}

//func TestValidateScyllaDBManagerTaskUpdate(t *testing.T) {
//	t.Parallel()
//
//	tests := []struct {
//		name                string
//		oldTask             *scyllav1alpha1.ScyllaDBManagerTask
//		newTask             *scyllav1alpha1.ScyllaDBManagerTask
//		expectedErrorList   field.ErrorList
//		expectedErrorString string
//	}{
//		{
//			name: "valid update - identical tasks",
//			oldTask: &scyllav1alpha1.ScyllaDBManagerTask{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "repair",
//				},
//				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
//					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
//						Name: "basic",
//						Kind: "ScyllaDBDatacenter",
//					},
//					Type:   "Repair",
//					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{},
//				},
//			},
//			newTask: &scyllav1alpha1.ScyllaDBManagerTask{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "repair",
//				},
//				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
//					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
//						Name: "basic",
//						Kind: "ScyllaDBDatacenter",
//					},
//					Type:   "Repair",
//					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{},
//				},
//			},
//			expectedErrorList:   field.ErrorList{},
//			expectedErrorString: "",
//		},
//		{
//			name: "valid update - changing task intensity",
//			oldTask: &scyllav1alpha1.ScyllaDBManagerTask{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "repair",
//				},
//				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
//					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
//						Name: "basic",
//						Kind: "ScyllaDBDatacenter",
//					},
//					Type: "Repair",
//					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
//						Intensity: "0.5",
//					},
//				},
//			},
//			newTask: &scyllav1alpha1.ScyllaDBManagerTask{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "repair",
//				},
//				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
//					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
//						Name: "basic",
//						Kind: "ScyllaDBDatacenter",
//					},
//					Type: "Repair",
//					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{
//						Intensity: "0.75",
//					},
//				},
//			},
//			expectedErrorList:   field.ErrorList{},
//			expectedErrorString: "",
//		},
//		{
//			name: "invalid update - changing task type",
//			oldTask: &scyllav1alpha1.ScyllaDBManagerTask{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "task",
//				},
//				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
//					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
//						Name: "basic",
//						Kind: "ScyllaDBDatacenter",
//					},
//					Type:   "Repair",
//					Repair: &scyllav1alpha1.ScyllaDBManagerRepairTaskOptions{},
//				},
//			},
//			newTask: &scyllav1alpha1.ScyllaDBManagerTask{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "task",
//				},
//				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
//					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
//						Name: "basic",
//						Kind: "ScyllaDBDatacenter",
//					},
//					Type: "Backup",
//					Backup: &scyllav1alpha1.ScyllaDBManagerBackupTaskOptions{
//						Location: []string{"gcs:test"},
//					},
//				},
//			},
//			expectedErrorList: field.ErrorList{
//				&field.Error{
//					Type:     field.ErrorTypeInvalid,
//					Field:    "spec.type",
//					BadValue: "Backup",
//					Detail:   "field is immutable",
//				},
//			},
//			expectedErrorString: "spec.type: Invalid value: \"Backup\": field is immutable",
//		},
//		{
//			name: "invalid update - invalid new task",
//			oldTask: &scyllav1alpha1.ScyllaDBManagerTask{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "backup",
//				},
//				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
//					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
//						Name: "basic",
//						Kind: "ScyllaDBDatacenter",
//					},
//					Type: "Backup",
//					Backup: &scyllav1alpha1.ScyllaDBManagerBackupTaskOptions{
//						Location: []string{"gcs:test"},
//					},
//				},
//			},
//			newTask: &scyllav1alpha1.ScyllaDBManagerTask{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "backup",
//				},
//				Spec: scyllav1alpha1.ScyllaDBManagerTaskSpec{
//					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
//						Name: "basic",
//						Kind: "ScyllaDBDatacenter",
//					},
//					Type:   "Backup",
//					Backup: &scyllav1alpha1.ScyllaDBManagerBackupTaskOptions{},
//				},
//			},
//			expectedErrorList: field.ErrorList{
//				&field.Error{
//					Type:     field.ErrorTypeRequired,
//					Field:    "spec.backup.location",
//					BadValue: "",
//					Detail:   "location must not be empty",
//				},
//			},
//			expectedErrorString: "spec.backup.location: Required value: location must not be empty",
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			t.Parallel()
//
//			errList := ValidateScyllaDBManagerTaskUpdate(tc.newTask, tc.oldTask)
//			if !reflect.DeepEqual(errList, tc.expectedErrorList) {
//				t.Errorf("expected and actual error lists differ: %s", cmp.Diff(tc.expectedErrorList, errList))
//			}
//
//			var errStr string
//			if agg := errList.ToAggregate(); agg != nil {
//				errStr = agg.Error()
//			}
//			if !reflect.DeepEqual(errStr, tc.expectedErrorString) {
//				t.Errorf("expected and actual error strings differ: %s", cmp.Diff(tc.expectedErrorString, errStr))
//			}
//		})
//	}
//}
