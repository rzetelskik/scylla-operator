// Copyright (C) 2025 ScyllaDB

package globalscylladbmanager

import (
	"context"
	"fmt"
	"strings"

	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/resource"
	"github.com/scylladb/scylla-operator/pkg/resourceapply"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	apimachineryvalidationutils "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/klog/v2"
)

func (gsmc *Controller) pruneScyllaDBManagerClusterRegistrations(ctx context.Context, requiredScyllaDBManagerClusterRegistrations []*scyllav1alpha1.ScyllaDBManagerClusterRegistration, scyllaDBManagerClusterRegistrations map[string]*scyllav1alpha1.ScyllaDBManagerClusterRegistration) error {
	var errs []error

	for _, smcr := range scyllaDBManagerClusterRegistrations {
		if smcr.GetDeletionTimestamp() != nil {
			continue
		}

		isRequired := false
		for _, required := range requiredScyllaDBManagerClusterRegistrations {
			if smcr.GetName() == required.GetName() {
				isRequired = true
				break
			}
		}
		if isRequired {
			continue
		}

		uid := smcr.GetUID()
		propagationPolicy := metav1.DeletePropagationBackground
		klog.V(2).InfoS("Pruning resource", "GVK", resource.GetObjectGVKOrUnknown(smcr), "Ref", klog.KObj(smcr))
		err := gsmc.scyllaClient.ScyllaDBManagerClusterRegistrations(smcr.Namespace).Delete(ctx, smcr.GetName(), metav1.DeleteOptions{
			Preconditions: &metav1.Preconditions{
				UID: &uid,
			},
			PropagationPolicy: &propagationPolicy,
		})
		resourceapply.ReportDeleteEvent(gsmc.EventRecorder(), smcr, err)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	return utilerrors.NewAggregate(errs)
}

func (gsmc *Controller) syncScyllaDBManagerClusterRegistrations(ctx context.Context, scyllaDBDatacenters []*scyllav1alpha1.ScyllaDBDatacenter, scyllaDBManagerClusterRegistrations map[string]*scyllav1alpha1.ScyllaDBManagerClusterRegistration) error {
	var requiredScyllaDBManagerClusterRegistrations []*scyllav1alpha1.ScyllaDBManagerClusterRegistration

	requiredScyllaDBManagerClusterRegistrationsForScyllaDBDatacenters, err := makeScyllaDBManagerClusterRegistrationsForScyllaDBDatacenters(scyllaDBDatacenters)
	if err != nil {
		return fmt.Errorf("can't make ScyllaDBManagerClusterRegistrations for ScyllaDBDatacenters: %w", err)
	}

	requiredScyllaDBManagerClusterRegistrations = append(requiredScyllaDBManagerClusterRegistrations, requiredScyllaDBManagerClusterRegistrationsForScyllaDBDatacenters...)

	err = gsmc.pruneScyllaDBManagerClusterRegistrations(
		ctx,
		requiredScyllaDBManagerClusterRegistrations,
		scyllaDBManagerClusterRegistrations,
	)
	if err != nil {
		return fmt.Errorf("can't prune ScyllaDBManagerClusterRegistration(s): %w", err)
	}

	var errs []error
	for _, smcr := range requiredScyllaDBManagerClusterRegistrations {
		_, _, err = resourceapply.ApplyScyllaDBManagerClusterRegistration(ctx, gsmc.scyllaClient, gsmc.scyllaDBManagerClusterRegistrationLister, gsmc.EventRecorder(), smcr, resourceapply.ApplyOptions{
			AllowMissingControllerRef: true,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("can't create ScyllaDBManagerClusterRegistration: %w", err))
		}
	}

	return utilerrors.NewAggregate(errs)
}

func makeScyllaDBManagerClusterRegistrationsForScyllaDBDatacenters(sdcs []*scyllav1alpha1.ScyllaDBDatacenter) ([]*scyllav1alpha1.ScyllaDBManagerClusterRegistration, error) {
	var scyllaDBManagerClusterRegistrations []*scyllav1alpha1.ScyllaDBManagerClusterRegistration
	var errs []error

	for _, sdc := range sdcs {
		name, err := getScyllaDBManagerClusterRegistrationName(naming.ScyllaDBDatacenterKind, sdc.Name)
		if err != nil {
			errs = append(errs, fmt.Errorf("can't get ScyllaDBManagerClusterRegistration name for ScyllaDBDatacenter %q: %w", naming.ObjRef(sdc), err))
			continue
		}

		smcr := &scyllav1alpha1.ScyllaDBManagerClusterRegistration{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: sdc.Namespace,
				Labels: map[string]string{
					naming.GlobalScyllaDBManagerLabel: naming.LabelValueTrue,
				},
				Annotations: map[string]string{},
			},
			Spec: scyllav1alpha1.ScyllaDBManagerClusterRegistrationSpec{
				ScyllaDBClusterRef: scyllav1alpha1.LocalScyllaDBReference{
					Kind: naming.ScyllaDBDatacenterKind,
					Name: sdc.Name,
				},
			},
		}

		nameOverrideAnnotationValue, hasNameOverrideAnnotation := sdc.Annotations[naming.ScyllaDBManagerClusterRegistrationNameOverrideAnnotation]
		if hasNameOverrideAnnotation {
			smcr.Annotations[naming.ScyllaDBManagerClusterRegistrationNameOverrideAnnotation] = nameOverrideAnnotationValue
		}

		scyllaDBManagerClusterRegistrations = append(scyllaDBManagerClusterRegistrations, smcr)
	}

	err := utilerrors.NewAggregate(errs)
	if err != nil {
		return nil, err
	}

	return scyllaDBManagerClusterRegistrations, nil
}

func getScyllaDBManagerClusterRegistrationName(kind, name string) (string, error) {
	nameSuffix, err := naming.GenerateNameHash(kind, name)
	if err != nil {
		return "", fmt.Errorf("can't generate name hash: %w", err)
	}

	fullName := strings.ToLower(fmt.Sprintf("%s-%s", kind, name))
	fullNameWithSuffix := fmt.Sprintf("%s-%s", fullName[:min(len(fullName), apimachineryvalidationutils.DNS1123SubdomainMaxLength-len(nameSuffix)-1)], nameSuffix)
	return fullNameWithSuffix, nil
}
