// Copyright (C) 2025 ScyllaDB

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type LocalScyllaDBSelector struct {
	// kind specifies the type of the resource.
	Kind string `json:"kind"`

	// labelSelector specifies the label selector for resources of the specified kind.
	LabelSelector metav1.LabelSelector `json:"labelSelector"`
}

type ScyllaDBManagerSpec struct {
	// selectors specify which ScyllaDB clusters should be registered with ScyllaDBManager.
	// Supported kinds are ScyllaDBCluster (scylla.scylladb.com/v1alpha1) and ScyllaDBDatacenter (scylla.scylladb.com/v1alpha1).
	// A disjunction is used to combine all selectors.
	// +optional
	Selectors []LocalScyllaDBSelector `json:"selectors,omitempty"`
}

type ScyllaDBManagerStatus struct {
	// observedGeneration is the most recent generation observed for this ScyllaDBManager. It corresponds to the
	// ScyllaDBManager's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// conditions hold conditions describing ScyllaDBManager state.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="PROGRESSING",type=string,JSONPath=".status.conditions[?(@.type=='Progressing')].status"
// +kubebuilder:printcolumn:name="DEGRADED",type=string,JSONPath=".status.conditions[?(@.type=='Degraded')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

type ScyllaDBManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of ScyllaDBManager.
	Spec ScyllaDBManagerSpec `json:"spec,omitempty"`

	// status reflects the observed state of ScyllaDBManager.
	Status ScyllaDBManagerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ScyllaDBManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScyllaDBManager `json:"items"`
}
