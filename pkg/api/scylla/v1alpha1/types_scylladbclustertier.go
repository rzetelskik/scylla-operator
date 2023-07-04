/*
Copyright (C) 2023 ScyllaDB

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	v1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ScyllaDBClusterTemplateSpec struct {
	// Standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the specification of the desired ScyllaCluster.
	// TODO: how to deal with volume claims?
	Spec v1.ScyllaClusterSpec `json:"spec"`
}

type ScyllaDBClusterTierSpec struct {
	// template describes the basis for creating ScyllaClusters.
	Template ScyllaDBClusterTemplateSpec `json:"template"`
}

// TODO: namespaced or non-namespaced?
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ScyllaDBClusterTier struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ScyllaDBClusterTierSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ScyllaDBClusterTierList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScyllaDBClusterTier `json:"items"`
}
