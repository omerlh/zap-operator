package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ZaproxySpec defines the desired state of Zaproxy
// +k8s:openapi-gen=true
type ZaproxySpec struct {
	AttackType string `json:"attackType"`
	TargetNamespace string `json:"tragetNamespace"`
	TargetIngress string `json:"tragetIngress"`
}

// ZaproxyStatus defines the observed state of Zaproxy
// +k8s:openapi-gen=true
type ZaproxyStatus struct {
	AttackStatus string `json:"attackStatus"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Zaproxy is the Schema for the zaproxies API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=zaproxies,scope=Namespaced
type Zaproxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZaproxySpec   `json:"spec,omitempty"`
	Status ZaproxyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ZaproxyList contains a list of Zaproxy
type ZaproxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Zaproxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Zaproxy{}, &ZaproxyList{})
}
