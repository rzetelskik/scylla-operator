// Copyright (C) 2025 ScyllaDB

package scylladbmanagertask

import (
	"context"
	"fmt"

	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/naming"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (smtc *Controller) syncManager(
	ctx context.Context,
	smt *scyllav1alpha1.ScyllaDBManagerTask,
) ([]metav1.Condition, error) {
	var progressingConditions []metav1.Condition

	smcrName, err := scyllaDBManagerClusterRegistrationName(smt)
	if err != nil {
		return progressingConditions, fmt.Errorf("can't get ScyllaDBManagerClusterRegistration name: %w", err)
	}

	smcr, err := smtc.scyllaDBManagerClusterRegistrationLister.ScyllaDBManagerClusterRegistrations(smt.Namespace).Get(smcrName)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return progressingConditions, fmt.Errorf("can't get ScyllaDBManagerClusterRegistration: %w", err)
		}

		//progressingConditions = append(progressingConditions, metav1.Condition{
		//	Type:               managerControllerProgressingCondition,
		//	Status:             metav1.ConditionTrue,
		//	ObservedGeneration: smt.Generation,
		//	Reason:             "AwaitingClusterRegistration",
		//	Message:            fmt.Sprintf("waiting for the ScyllaDBManagerClusterRegistration to become available"),
		//})

		// TODO: progressing condition
	}

	return progressingConditions, nil
}

func scyllaDBManagerClusterRegistrationName(smt *scyllav1alpha1.ScyllaDBManagerTask) (string, error) {
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
