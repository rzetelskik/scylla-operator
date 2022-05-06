// Copyright (c) 2022 ScyllaDB

package scyllacluster

import (
	"context"
	"fmt"

	"github.com/magiconair/properties"
	"k8s.io/klog/v2"

	"github.com/pkg/errors"

	"github.com/ghodss/yaml"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	"github.com/scylladb/scylla-operator/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/resourceapply"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

// mergeYAMLs merges two arbitrary YAML structures at the top level.
func mergeYAMLs(initialYAML, overrideYAML []byte) ([]byte, error) {

	var initial, override map[string]interface{}
	if err := yaml.Unmarshal(initialYAML, &initial); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := yaml.Unmarshal(overrideYAML, &override); err != nil {
		return nil, errors.WithStack(err)
	}

	if initial == nil {
		initial = make(map[string]interface{})
	}
	// Overwrite the values onto initial
	for k, v := range override {
		initial[k] = v
	}
	return yaml.Marshal(initial)
}

func makeScyllaConfig(configMapBytes []byte, clusterName string) ([]byte, error) { // FIXME fix merging logic
	configFileBytes := []byte{} // FIXME get actual scylla config

	// Custom options
	var cfg = make(map[string]interface{}) // TODO move outside of func and only do once?
	cfg["cluster_name"] = clusterName
	cfg["rpc_address"] = "0.0.0.0"
	cfg["endpoint_snitch"] = "GossipingPropertyFileSnitch"

	overrideBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("can't parse override options for scylla.yaml: %w", err) // FIXME naming instead of scylla.yaml
	}

	overwrittenBytes, err := mergeYAMLs(configFileBytes, overrideBytes)
	if err != nil {
		return nil, fmt.Errorf("can't merge scylla yaml with operator pre-sets: %w", err)
	}

	customConfigBytesBytes, err := mergeYAMLs(overwrittenBytes, configMapBytes)
	if err != nil {
		return nil, fmt.Errorf("can't merge overwritten scylla yaml with user config map: %w", err)
	}

	return customConfigBytesBytes, nil
}

func loadProperties(configMapBytes []byte) *properties.Properties {
	l := &properties.Loader{Encoding: properties.UTF8}

	p, err := l.LoadBytes(configMapBytes)
	if err != nil {
		klog.InfoS("unable to read properties") // FIXME add configMap name
		return properties.NewProperties()
	}

	return p
}

func makeRackDCProperties(configMapBytes []byte, dc, rack string) *properties.Properties {
	configMapProperties := loadProperties(configMapBytes)

	rackDCProperties := properties.NewProperties()
	rackDCProperties.DisableExpansion = true
	rackDCProperties.Set("dc", dc)
	rackDCProperties.Set("rack", rack)
	rackDCProperties.Set("prefer_local", configMapProperties.GetString("prefer_local", "false"))
	if dcSuffix, ok := configMapProperties.Get("dc_suffix"); ok {
		rackDCProperties.Set("dc_suffix", dcSuffix)
	}

	return rackDCProperties
}

func (scc *Controller) makeConfigMaps(ctx context.Context, sc *scyllav1.ScyllaCluster) ([]*corev1.ConfigMap, error) {
	configMaps := []*corev1.ConfigMap{}

	// TODO merge with user provided config
	// TODO error aggregate
	for _, rack := range sc.Spec.Datacenter.Racks {
		cmName := naming.StatefulSetNameForRack(rack, sc) // FIXME: naming

		suppliedCM, err := scc.kubeClient.CoreV1().ConfigMaps(sc.Namespace).Get(ctx, rack.ScyllaConfig, metav1.GetOptions{}) // FIXME get or default
		if err != nil {
			return nil, err // FIXME fmt, err aggregate
		}

		scyllaConfigBytes, err := makeScyllaConfig(suppliedCM.BinaryData[naming.ScyllaConfigName], sc.Name)
		if err != nil {
			return nil, err // FIXME fmt, err aggregate
		}

		rackDCProperties := makeRackDCProperties(suppliedCM.BinaryData[naming.ScyllaRackDCPropertiesName], sc.Spec.Datacenter.Name, rack.Name)

		data := map[string]string{
			naming.ScyllaConfigName:           string(scyllaConfigBytes),
			naming.ScyllaRackDCPropertiesName: rackDCProperties.String(),
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

	requiredConfigMaps, err := scc.makeConfigMaps(ctx, sc)
	if err != nil {
		// TODO event recorder?
		return err // TODO fmt
	}

	err = scc.pruneConfigMaps(ctx, requiredConfigMaps, configMaps)
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
