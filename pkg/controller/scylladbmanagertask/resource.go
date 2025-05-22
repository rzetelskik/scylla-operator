// Copyright (C) 2025 ScyllaDB

package scylladbmanagertask

import (
	"fmt"

	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/pointer"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makeJobForScyllaDBManagerTask(
	smt *scyllav1alpha1.ScyllaDBManagerTask,
	host string,
	authToken string,
) (*batchv1.Job, error) {
	switch smt.Spec.Type {
	case scyllav1alpha1.ScyllaDBManagerTaskTypeRepair:
		backoffLimit := int32(0)
		if smt.Spec.Repair.ScyllaDBManagerTaskSchedule.NumRetries != nil {
			backoffLimit = int32(*smt.Spec.Repair.ScyllaDBManagerTaskSchedule.NumRetries)
		}

		restartPolicy := corev1.RestartPolicyNever
		if backoffLimit > 0 {
			restartPolicy = corev1.RestartPolicyOnFailure
		}

		return &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-job", smt.Name),
				Namespace: smt.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(smt, scyllaDBManagerTaskControllerGVK),
				},
				Labels: getLabels(smt),
				// TODO: annotations
			},
			Spec: batchv1.JobSpec{
				Selector:       nil,
				ManualSelector: pointer.Ptr(false),
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: getLabels(smt),
						// TODO: annotations
					},
					Spec: corev1.PodSpec{
						// TODO: switch to on failure?
						RestartPolicy: restartPolicy,
						Containers: []corev1.Container{
							{
								Name:            "task",
								Image:           "docker.io/scylladb/scylla-manager:3.6.0-dev-0.20250520.1a372cfb1@sha256:04e8a9b290f2caf3b386a479ae543d9bbb73d931d2e87bc254e04a767f729d9c",
								ImagePullPolicy: corev1.PullIfNotPresent,
								Command: []string{
									"sctool",
									"repair",
									"unsupported",
									fmt.Sprintf("--auth-token=%s", authToken),
									"--data-path=TODO",
									"--force-non-ssl-session-port",
									"--force-tls-disabled",
									fmt.Sprintf("--host=%s", host),
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "data",
										MountPath: "/var/lib/scylladb-manager",
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "data",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
					},
				},
				//TTLSecondsAfterFinished: nil,
				//CompletionMode:          nil,
				//Suspend: 				   nil,
				//PodReplacementPolicy:    nil,
				//ManagedBy:               nil,
				//Parallelism:             nil,
				//Completions:             nil,
				//ActiveDeadlineSeconds:   nil,
				//PodFailurePolicy:        nil,
				//SuccessPolicy:           nil,
				BackoffLimit: &backoffLimit,
				//BackoffLimitPerIndex:    nil,
				//MaxFailedIndexes:        nil,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported ScyllaDBManagerTask type: %q", smt.Spec.Type)

	}
}
