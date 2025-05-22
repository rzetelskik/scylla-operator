// Copyright (C) 2025 ScyllaDB

package scylladbmanagertask

import (
	"context"
	"fmt"

	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"github.com/scylladb/scylla-operator/pkg/controllertools"
	"github.com/scylladb/scylla-operator/pkg/helpers"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/resourceapply"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (smtc *Controller) syncJobs(
	ctx context.Context,
	smt *scyllav1alpha1.ScyllaDBManagerTask,
	jobs map[string]*batchv1.Job,
	status *scyllav1alpha1.ScyllaDBManagerTaskStatus,
) ([]metav1.Condition, error) {
	var progressingConditions []metav1.Condition

	var host, authTokenSecretName string
	switch smt.Spec.ScyllaDBClusterRef.Kind {
	case scyllav1alpha1.ScyllaDBDatacenterGVK.Kind:
		sdc, err := smtc.scyllaDBDatacenterLister.ScyllaDBDatacenters(smt.Namespace).Get(smt.Spec.ScyllaDBClusterRef.Name)
		if err != nil {
			return progressingConditions, fmt.Errorf("can't get ScyllaDBDatacenter %q: %w", naming.ManualRef(smt.Namespace, smt.Spec.ScyllaDBClusterRef.Name), err)
		}

		isScyllaDBDatacenterAvailable := sdc.Status.AvailableNodes != nil && *sdc.Status.AvailableNodes > 0
		if !isScyllaDBDatacenterAvailable {
			progressingConditions = append(progressingConditions, metav1.Condition{
				Type:               jobControllerProgressingCondition,
				Status:             metav1.ConditionTrue,
				ObservedGeneration: smt.Generation,
				Reason:             "AwaitingScyllaDBDatacenterAvailability",
				Message:            fmt.Sprintf("Awaiting ScyllaDBDatacenter %q availability.", naming.ObjRef(sdc)),
			})

			return progressingConditions, nil
		}

		host = naming.CrossNamespaceServiceName(sdc)
		authTokenSecretName = naming.AgentAuthTokenSecretName(sdc)

	default:
		return progressingConditions, fmt.Errorf("unsupported scyllaDBClusterRef Kind: %q", smt.Spec.ScyllaDBClusterRef.Kind)

	}

	authToken, err := smtc.getAuthToken(ctx, smt.Namespace, authTokenSecretName)
	if err != nil {
		return progressingConditions, fmt.Errorf("can't get auth token: %w", err)
	}

	required, err := makeJobForScyllaDBManagerTask(
		smt,
		host,
		authToken,
	)
	if err != nil {
		return progressingConditions, fmt.Errorf("can't make Job(s) for ScyllaDBManagerTask %q: %w", naming.ObjRef(smt), err)
	}

	err = controllerhelpers.Prune(
		ctx,
		[]*batchv1.Job{required},
		jobs,
		&controllerhelpers.PruneControlFuncs{
			DeleteFunc: smtc.kubeClient.BatchV1().Jobs(smt.Namespace).Delete,
		},
		smtc.eventRecorder,
	)
	if err != nil {
		return progressingConditions, fmt.Errorf("can't prune Job(s) for ScyllaDBManagerTask %q: %w", err)
	}

	job, changed, err := resourceapply.ApplyJob(ctx, smtc.kubeClient.BatchV1(), smtc.jobLister, smtc.eventRecorder, required, resourceapply.ApplyOptions{
		AllowMissingControllerRef: true,
	})
	if changed {
		controllerhelpers.AddGenericProgressingStatusCondition(&progressingConditions, jobControllerProgressingCondition, required, "apply", smt.Generation)
	}
	if err != nil {
		return progressingConditions, fmt.Errorf("can't apply Job(s) for ScyllaDBManagerTask %q: %w", naming.ObjRef(smt), err)
	}

	if job.Status.Failed > 0 {
		return progressingConditions, controllertools.NewNonRetriable(fmt.Sprintf("task Job %q failed", naming.ObjRef(job)))
	}

	if job.Status.CompletionTime == nil {
		progressingConditions = append(progressingConditions, metav1.Condition{
			Type:               jobControllerProgressingCondition,
			Status:             metav1.ConditionTrue,
			Reason:             "AwaitingJobCompletion",
			Message:            fmt.Sprintf("Waiting for Job %q to complete.", naming.ObjRef(job)),
			ObservedGeneration: smt.Generation,
		})
	}

	return progressingConditions, nil
}

func (smtc *Controller) getAuthToken(ctx context.Context, authTokenSecretNamespace, authTokenSecretName string) (string, error) {
	authTokenSecret, err := smtc.secretLister.Secrets(authTokenSecretNamespace).Get(authTokenSecretName)
	if err != nil {
		return "", fmt.Errorf("can't get secret %q: %w", naming.ManualRef(authTokenSecretNamespace, authTokenSecretName), err)
	}

	authToken, err := helpers.GetAgentAuthTokenFromSecret(authTokenSecret)
	if err != nil {
		return "", fmt.Errorf("can't get agent auth token from secret %q: %w", naming.ObjRef(authTokenSecret), err)
	}

	return authToken, nil
}
