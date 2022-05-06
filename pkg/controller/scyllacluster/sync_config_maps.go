// Copyright (c) 2022 ScyllaDB

package scyllacluster

import (
	"context"
	"fmt"

	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/resourceapply"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

func (scc *Controller) makeConfigMaps(sc *scyllav1.ScyllaCluster) ([]*corev1.ConfigMap, error) {
	configMaps := []*corev1.ConfigMap{}

	// TODO merge with user provided config
	// TODO error aggregate
	for _, rack := range sc.Spec.Datacenter.Racks {
		cmName := naming.StatefulSetNameForRack(rack, sc) // FIXME: naming

		data := map[string]string{
			naming.ScyllaConfigName: "",
		}

		configMaps = append(configMaps, MakeConfigMap(sc, cmName, data))
	}

	return configMaps, nil
}

func (scc *Controller) pruneConfigMaps(ctx context.Context, requiredConfigMaps []*corev1.ConfigMap, configMaps map[string]*corev1.ConfigMap) error {
	var errs []error

	for _, cm := range configMaps {
		if cm.DeletionTimestamp != nil {
			continue
		}

		isRequired := false
		for _, r := range requiredConfigMaps {
			if cm.Name == r.Name {
				isRequired = true
			}
		}
		if isRequired {
			continue
		}

		propagationPolicy := metav1.DeletePropagationBackground
		err := scc.kubeClient.CoreV1().ConfigMaps(cm.Namespace).Delete(ctx, cm.Name, metav1.DeleteOptions{
			Preconditions: &metav1.Preconditions{
				UID: &cm.UID,
			},
			PropagationPolicy: &propagationPolicy,
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	return utilerrors.NewAggregate(errs)
}

func (scc *Controller) applyConfigMaps(ctx context.Context, requiredConfigMaps []*corev1.ConfigMap) error {
	// TODO status?
	var errs []error

	for _, r := range requiredConfigMaps {
		_, _, err := resourceapply.ApplyConfigMap(ctx, scc.kubeClient.CoreV1(), scc.configMapLister, scc.eventRecorder, r)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return utilerrors.NewAggregate(errs)
}

func (scc *Controller) syncConfigMaps(
	ctx context.Context,
	sc *scyllav1.ScyllaCluster,
	configMaps map[string]*corev1.ConfigMap,
) error {
	// TODO status??

	requiredConfigMaps, err := scc.makeConfigMaps(sc)
	if err != nil {
		// TODO event recorder?
		return err // TODO fmt
	}

	err := scc.pruneConfigMaps(ctx, requiredConfigMaps, configMaps)
	if err != nil {
		return fmt.Errorf("can't prune ConfigMap(s): %w", err)
	}

	err = scc.applyConfigMaps(ctx, requiredConfigMaps)
	if err != nil {
		// TODO updare/create
		return fmt.Errorf("can't apply ConfigMap(s): %w", err)
	}

	return nil
}
