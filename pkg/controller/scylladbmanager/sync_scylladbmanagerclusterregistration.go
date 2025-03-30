// Copyright (C) 2025 ScyllaDB

package scylladbmanager

import (
	"context"
	"fmt"

	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/resourceapply"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryvalidationutils "k8s.io/apimachinery/pkg/util/validation"
)

func (smc *Controller) syncScyllaDBManagerClusterRegistrations(ctx context.Context, sm *scyllav1alpha1.ScyllaDBManager, scyllaDBManagerClusterRegistrations map[string]*scyllav1alpha1.ScyllaDBManagerClusterRegistration) ([]metav1.Condition, error) {
	var progressingConditions []metav1.Condition

	// TODO: prune remaining

	//var scyllaDBDatacenters []*scyllav1alpha1.ScyllaDBDatacenter
	for _, kindSelector := range sm.Spec.Selectors {
		if kindSelector.Kind != "ScyllaDBDatacenter" {
			continue
		}

		selector, err := metav1.LabelSelectorAsSelector(&kindSelector.LabelSelector)
		if err != nil {
			return progressingConditions, fmt.Errorf("can't convert label selector to selector: %w", err)
		}

		sdcs, err := smc.scyllaDBDatacenterLister.List(selector)
		if err != nil {
			return progressingConditions, fmt.Errorf("can't list ScyllaDBDatacenters: %w", err)
		}

		for _, sdc := range sdcs {
			smcrName, err := generateScyllaDBManagerClusterRegistrationNameForGlobalScyllaDBManager(sdc.GetNamespace(), "ScyllaDBDatacenter", sdc.GetName())
			if err != nil {
				// TODO: append
				return progressingConditions, fmt.Errorf("can't generate ScyllaDBManagerClusterRegistration name for ScyllaDBDatacenter %q: %w", naming.ObjRef(sdc), err)
			}

			// TODO: labels, annotations

			smcr := &scyllav1alpha1.ScyllaDBManagerClusterRegistration{
				ObjectMeta: metav1.ObjectMeta{
					Name:      smcrName,
					Namespace: sm.Namespace,
					Labels: map[string]string{
						naming.ScyllaDBManagerNameLabel: sm.Name,
					},
					Annotations: map[string]string{
						naming.ScyllaDBManagerClusterRegistrationOverrideScyllaDBClusterNamespaceAnnotation: sdc.Namespace,
					},
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(sm, scyllaDBManagerControllerGVK),
					},
				},
				Spec: scyllav1alpha1.ScyllaDBManagerClusterRegistrationSpec{
					ScyllaDBManagerRef: sm.Name,
					ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
						Kind: "ScyllaDBDatacenter",
						Name: sdc.Name,
					},
				},
			}

			_, changed, err := resourceapply.ApplyScyllaDBManagerClusterRegistration(ctx, smc.scyllaClient, smc.scyllaDBManagerClusterRegistrationLister, smc.eventRecorder, smcr, resourceapply.ApplyOptions{})
			if changed {
				controllerhelpers.AddGenericProgressingStatusCondition(&progressingConditions, scyllaDBManagerClusterRegistrationProgressingCondition, smcr, "apply", sm.Generation)
			}
			if err != nil {
				return progressingConditions, err
			}
		}

	}

	return progressingConditions, nil
}

func generateScyllaDBManagerClusterRegistrationNameForGlobalScyllaDBManager(namespace, kind, name string) (string, error) {
	nameSuffix, err := naming.GenerateNameHash(namespace, kind, name)
	if err != nil {
		return "", fmt.Errorf("can't generate name hash: %w", err)
	}

	fullName := fmt.Sprintf("%s-%s-%s", namespace, kind, name)
	fullNameWithSuffix := fmt.Sprintf("%s-%s", fullName[:min(len(fullName), apimachineryvalidationutils.DNS1123SubdomainMaxLength-len(nameSuffix)-1)], nameSuffix)
	return fullNameWithSuffix, nil
}
