package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PolicyInterface abstracts the concrete policy type (Policy vs ClusterPolicy)
// +kubebuilder:object:generate=false
type PolicyInterface interface {
	metav1.Object
	HasAutoGenAnnotation() bool
	IsNamespaced() bool
	GetSpec() *Spec
	GetStatus() *PolicyStatus
	GetKind() string
	SetKind(string)
	GetAPIVersion() string
	SetAPIVersion(string)
}
