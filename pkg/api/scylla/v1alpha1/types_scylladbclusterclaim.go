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

type ScyllaDBClusterClaimSpec struct {
	// selector is a label query over ScyllaDBClusterPools to consider for binding a ScyllaCluster from.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// scyllaDBClusterTierName is the name of the ScyllaDBClusterTier required by the claim.
	ScyllaDBClusterTierName string `json:"scyllaDBClusterTierName"`

	// scyllaClusterName is the binding reference to the ScyllaCluster backing this claim.
	// +optional
	ScyllaClusterName string `json:"scyllaClusterName,omitempty"`
}

// +enum
type ScyllaDBClusterClaimPhase string

const (
	// used for ScyllaDBClusterClaims that are not yet bound
	ClaimPending ScyllaDBClusterClaimPhase = "Pending"
	// used for PersistentVolumeClaims that are bound
	ClaimBound ScyllaDBClusterClaimPhase = "Bound"

	// TODO: lost?
)

type ScyllaDBClusterClaimStatus struct {
	// phase represents the current phase of ScyllaDBClusterClaim.
	// +optional
	Phase ScyllaDBClusterClaimPhase `json:"phase,omitempty"`

	// TODO: propagate conditions?
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ScyllaDBClusterClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScyllaDBClusterClaimSpec   `json:"spec,omitempty"`
	Status ScyllaDBClusterClaimStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ScyllaDBClusterClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScyllaDBClusterClaim `json:"items"`
}
