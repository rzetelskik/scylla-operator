// Copyright (C) 2025 ScyllaDB

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type LocalScyllaDBReference struct {
	// kind specifies the type of the resource.
	Kind string `json:"kind"`
	// name specifies the name of the resource.
	Name string `json:"name"`
}

type ScyllaDBManagerClusterRegistrationSpec struct {
	// scyllaDBManagerRef specifies the reference to the local ScyllaDBManager that the cluster should be registered with.
	ScyllaDBManagerRef string `json:"scyllaDBManagerRef"`

	// scyllaDBClusterRef specifies the typed reference to the local ScyllaDB cluster.
	// Supported kinds are ScyllaDBCluster (scylla.scylladb.com/v1alpha1) and ScyllaDBDatacenter (scylla.scylladb.com/v1alpha1).
	ScyllaDBClusterRef LocalScyllaDBReference `json:"scyllaDBClusterRef"`
}

type ScyllaDBManagerClusterRegistrationStatus struct {
	// clusterID reflects the internal identification number of the cluster in ScyllaDB Manager state.
	// +optional
	ClusterID *string `json:"clusterID,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="PROGRESSING",type=string,JSONPath=".status.conditions[?(@.type=='Progressing')].status"
// +kubebuilder:printcolumn:name="DEGRADED",type=string,JSONPath=".status.conditions[?(@.type=='Degraded')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

type ScyllaDBManagerClusterRegistration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of ScyllaDBManagerClusterRegistration.
	Spec ScyllaDBManagerClusterRegistrationSpec `json:"spec,omitempty"`

	// status reflects the observed state of ScyllaDBManagerClusterRegistration.
	Status ScyllaDBManagerClusterRegistrationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ScyllaDBManagerClusterRegistrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScyllaDBManagerClusterRegistration `json:"items"`
}
