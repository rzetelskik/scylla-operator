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

type PausableScyllaDBClusterSpec struct {
	// paused indicates that the PausableScyllaDBCluster is paused.
	// +optional
	Paused bool `json:"paused,omitempty"`

	// scyllaDBClusterTierName is the name of the ScyllaDBClusterTier in which the required ScyllaCluster is.
	ScyllaDBClusterTierName string `json:"scyllaDBClusterTierName"`

	// TODO: dnsDomains and exposeOptions need to be verified

	// dnsDomains is a list of DNS domains the bound ScyllaCluster is reachable by when unpaused.
	// These domains are used when setting up the infrastructure of the bound ScyllaCluster, like certificates.
	// EXPERIMENTAL. Do not rely on any particular behaviour controlled by this field.
	// +optional
	DNSDomains []string `json:"dnsDomains,omitempty"`

	// exposeOptions specifies options for exposing services of a bound ScyllaCluster.
	// EXPERIMENTAL. Do not rely on any particular behaviour controlled by this field.
	// +optional
	ExposeOptions *v1.ExposeOptions `json:"exposeOptions,omitempty"`
}

type PausableScyllaDBClusterStatus struct {
	// TODO: propagate conditions?
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PausableScyllaDBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PausableScyllaDBClusterSpec   `json:"spec,omitempty"`
	Status PausableScyllaDBClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PausableScyllaDBClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PausableScyllaDBCluster `json:"items"`
}
