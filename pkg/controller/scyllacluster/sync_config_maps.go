// Copyright (c) 2022 ScyllaDB

package scyllacluster

import (
	"context"
	"fmt"

	"github.com/scylladb/scylla-operator/pkg/resourceapply"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	corev1 "k8s.io/api/core/v1"
)

func (scc *Controller) makeConfigMap(sc *scyllav1.ScyllaCluster) *corev1.ConfigMap {
	return makeConfigMap(sc, map[string]string{})
}

func (scc *Controller) pruneConfigMaps(ctx context.Context, required *corev1.ConfigMap, configMaps map[string]*corev1.ConfigMap) error {
	var errs []error

	for _, cm := range configMaps {
		if cm.DeletionTimestamp != nil {
			continue
		}

		if cm.Name == required.Name {
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

func (scc *Controller) syncConfigMaps(
	ctx context.Context,
	sc *scyllav1.ScyllaCluster,
	configMaps map[string]*corev1.ConfigMap,
) error {
	required := scc.makeConfigMap(sc)

	err := scc.pruneConfigMaps(ctx, required, configMaps)
	if err != nil {
		return fmt.Errorf("can't prune ConfigMap(s): %w", err)
	}

	if required != nil {
		_, _, err := resourceapply.ApplyConfigMap(ctx, scc.kubeClient.CoreV1(), scc.configMapLister, scc.eventRecorder, required)
		if err != nil {
			return fmt.Errorf("can't apply ConfigMap: %w", err) // TODO plural
		}
	}

	return nil
}
