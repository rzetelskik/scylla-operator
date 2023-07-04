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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ScyllaDBClusterPoolSpec struct {
	// replicas is the desired number of ready, prewarmed ScyllaCluster instances in the pool.
	// The replicas instantiate the same template, but they are all logically separate ScyllaClusters.
	// +kubebuilder:validation:Minimum=0
	Replicas int32 `json:"replicas"`

	// scyllaDBClusterTierName is the name of the ScyllaDBClusterTier in which the ScyllaClusters provided by the pool are.
	ScyllaDBClusterTierName string `json:"scyllaDBClusterTierName"`

	// parentStorageClassName specifies a name of the StorageClass used to replace the underlying storage of each ScyllaCluster in the pool.
	// TODO: further comment
	ParentStorageClassName string `json:"parentStorageClassName"`
}

type ScyllaDBClusterPoolStatus struct {
	// observedGeneration is the most recent generation observed for this ScyllaDBClusterPool. It corresponds to the
	// ScyllaDBClusterPool's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ScyllaDBClusterPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScyllaDBClusterPoolSpec   `json:"spec,omitempty"`
	Status ScyllaDBClusterPoolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ScyllaDBClusterPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScyllaDBClusterPool `json:"items"`
}
