package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Mirror is the Schema for the mirrors API
// +k8s:openapi-gen=true
type Mirror struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MirrorSpec   `json:"spec,omitempty"`
	Status MirrorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MirrorList contains a list of Mirror
type MirrorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mirror `json:"items"`
}

// MirrorSpec defines the desired state of Mirror
type MirrorSpec struct {
	ClusterName string `json:"clusterName"`
}

// MirrorStatus defines the observed state of Mirror
type MirrorStatus struct{}

func init() {
	SchemeBuilder.Register(&Mirror{}, &MirrorList{})
}
