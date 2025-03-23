// Copyright (C) 2025 ScyllaDB

package validation

import (
	"fmt"
	"strings"

	"github.com/robfig/cron/v3"
	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/helpers/slices"
	corevalidation "github.com/scylladb/scylla-operator/pkg/thirdparty/k8s.io/kubernetes/pkg/apis/core/validation"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var (
	supportedLocalScyllaDBReferenceKinds = []string{
		"ScyllaDBCluster",
		"ScyllaDBDatacenter",
	}

	supportedScyllaDBManagerTaskTypes = []scyllav1alpha1.ScyllaDBManagerTaskType{
		scyllav1alpha1.ScyllaDBManagerTaskTypeBackup,
		scyllav1alpha1.ScyllaDBManagerTaskTypeRepair,
	}
)

func ValidateScyllaDBManagerTask(smt *scyllav1alpha1.ScyllaDBManagerTask) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, ValidateScyllaDBManagerTaskSpec(&smt.Spec, field.NewPath("spec"))...)

	return allErrs
}

func ValidateScyllaDBManagerTaskSpec(spec *scyllav1alpha1.ScyllaDBManagerTaskSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, ValidateLocalScyllaDBReference(&spec.ScyllaDBCluster, fldPath.Child("scyllaDBCluster"))...)

	switch spec.Type {
	case scyllav1alpha1.ScyllaDBManagerTaskTypeBackup:
		if spec.Backup == nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("backup"), fmt.Sprintf("backup options are required when task type is %q", scyllav1alpha1.ScyllaDBManagerTaskTypeBackup)))
			break
		}

		allErrs = append(allErrs, ValidateScyllaDBManagerBackupTaskOptions(spec.Backup, fldPath.Child("backup"))...)

	case scyllav1alpha1.ScyllaDBManagerTaskTypeRepair:
		if spec.Repair == nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("repair"), fmt.Sprintf("repair options are required when task type is %q", scyllav1alpha1.ScyllaDBManagerTaskTypeRepair)))
			break
		}

		allErrs = append(allErrs, ValidateScyllaDBManagerRepairTaskOptions(spec.Repair, fldPath.Child("repair"))...)

	case "":
		allErrs = append(allErrs, field.Required(fldPath.Child("type"), ""))

	default:
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("type"), spec.Type, slices.ConvertSlice(supportedScyllaDBManagerTaskTypes, slices.ToString)))

	}

	if spec.Type != scyllav1alpha1.ScyllaDBManagerTaskTypeBackup && spec.Backup != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("backup"), fmt.Sprintf("backup options are forbidden when task type is not %q", scyllav1alpha1.ScyllaDBManagerTaskTypeBackup)))
	}

	if spec.Type != scyllav1alpha1.ScyllaDBManagerTaskTypeRepair && spec.Repair != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("repair"), fmt.Sprintf("repair options are forbidden when task type is not %q", scyllav1alpha1.ScyllaDBManagerTaskTypeRepair)))
	}

	return allErrs
}

func ValidateScyllaDBManagerBackupTaskOptions(backupOptions *scyllav1alpha1.ScyllaDBManagerBackupTaskOptions, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, ValidateScyllaDBManagerTaskSchedule(&backupOptions.ScyllaDBManagerTaskSchedule, fldPath)...)

	return allErrs
}

func ValidateScyllaDBManagerRepairTaskOptions(repairOptions *scyllav1alpha1.ScyllaDBManagerRepairTaskOptions, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, ValidateScyllaDBManagerTaskSchedule(&repairOptions.ScyllaDBManagerTaskSchedule, fldPath)...)

	if repairOptions.Intensity != nil && *repairOptions.Intensity < 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("intensity"), *repairOptions.Intensity, "can't be negative"))
	}

	if repairOptions.SmallTableThreshold != nil {
		allErrs = append(allErrs, corevalidation.ValidateResourceQuantityValue(corev1.ResourceStorage, *repairOptions.SmallTableThreshold, fldPath.Child("smallTableThreshold"))...)
	}

	return allErrs
}

func ValidateScyllaDBManagerTaskSchedule(schedule *scyllav1alpha1.ScyllaDBManagerTaskSchedule, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if schedule.Cron != nil {
		_, err := cron.NewParser(schedulerTaskSpecCronParseOptions).Parse(*schedule.Cron)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("cron"), schedule.Cron, err.Error()))
		}

		if strings.Contains(*schedule.Cron, "TZ") {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("cron"), schedule.Cron, "TZ and CRON_TZ prefixes are forbidden"))
		}
	}

	return allErrs
}

func ValidateLocalScyllaDBReference(localScyllaDBReference *scyllav1alpha1.LocalScyllaDBReference, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(localScyllaDBReference.Name) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), ""))
	} else {
		for _, msg := range validation.IsDNS1123Subdomain(localScyllaDBReference.Name) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), localScyllaDBReference.Name, msg))
		}
	}

	if len(localScyllaDBReference.Kind) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("kind"), ""))
	} else {
		allErrs = append(allErrs, validateEnum(localScyllaDBReference.Kind, supportedLocalScyllaDBReferenceKinds, fldPath.Child("kind"))...)
	}

	return allErrs
}

func ValidateScyllaDBManagerTaskUpdate(new, old *scyllav1alpha1.ScyllaDBManagerTask) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}
