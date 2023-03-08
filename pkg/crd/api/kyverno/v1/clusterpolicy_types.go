package v1

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clusterpolicies,scope="Cluster",shortName=cpol,categories=kyverno
// +kubebuilder:printcolumn:name="Background",type=boolean,JSONPath=".spec.background"
// +kubebuilder:printcolumn:name="Validate Action",type=string,JSONPath=".spec.validationFailureAction"
// +kubebuilder:printcolumn:name="Failure Policy",type=string,JSONPath=".spec.failurePolicy",priority=1
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type == "Ready")].status`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Validate",type=integer,JSONPath=`.status.rulecount.validate`,priority=1
// +kubebuilder:printcolumn:name="Mutate",type=integer,JSONPath=`.status.rulecount.mutate`,priority=1
// +kubebuilder:printcolumn:name="Generate",type=integer,JSONPath=`.status.rulecount.generate`,priority=1
// +kubebuilder:printcolumn:name="Verifyimages",type=integer,JSONPath=`.status.rulecount.verifyimages`,priority=1
// +kubebuilder:storageversion

// ClusterPolicy declares validation, mutation, and generation behaviors for matching resources.
type ClusterPolicy struct {
	metav1.TypeMeta   `json:",inline,omitempty" yaml:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Spec declares policy behaviors.
	Spec Spec `json:"spec" yaml:"spec"`

	// Status contains policy runtime data.
	// +optional
	Status PolicyStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// HasAutoGenAnnotation checks if a policy has auto-gen annotation
func (p *ClusterPolicy) HasAutoGenAnnotation() bool {
	annotations := p.GetAnnotations()
	val, ok := annotations[PodControllersAnnotation]
	if ok && strings.ToLower(val) != "none" {
		return true
	}
	return false
}

// HasMutateOrValidateOrGenerate checks for rule types
func (p *ClusterPolicy) HasMutateOrValidateOrGenerate() bool {
	for _, rule := range p.Spec.Rules {
		if rule.HasMutate() || rule.HasValidate() || rule.HasGenerate() {
			return true
		}
	}
	return false
}

// HasMutate checks for mutate rule types
func (p *ClusterPolicy) HasMutate() bool {
	return p.Spec.HasMutate()
}

// HasValidate checks for validate rule types
func (p *ClusterPolicy) HasValidate() bool {
	return p.Spec.HasValidate()
}

// HasGenerate checks for generate rule types
func (p *ClusterPolicy) HasGenerate() bool {
	return p.Spec.HasGenerate()
}

// HasVerifyImages checks for image verification rule types
func (p *ClusterPolicy) HasVerifyImages() bool {
	return p.Spec.HasVerifyImages()
}

// GetSpec returns the policy spec
func (p *ClusterPolicy) GetSpec() *Spec {
	return &p.Spec
}

// GetStatus returns the policy status
func (p *ClusterPolicy) GetStatus() *PolicyStatus {
	return &p.Status
}

// IsNamespaced indicates if the policy is namespace scoped
func (p *ClusterPolicy) IsNamespaced() bool {
	return p.GetNamespace() != ""
}

func (p *ClusterPolicy) GetKind() string {
	return p.Kind
}

func (p *ClusterPolicy) GetAPIVersion() string {
	return p.APIVersion
}

func (p *ClusterPolicy) SetKind(value string) {
	p.Kind = value
}

func (p *ClusterPolicy) SetAPIVersion(value string) {
	p.APIVersion = value
}

func (p *ClusterPolicy) CreateDeepCopy() PolicyInterface {
	return p.DeepCopy()
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterPolicyList is a list of ClusterPolicy instances.
type ClusterPolicyList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata" yaml:"metadata"`
	Items           []ClusterPolicy `json:"items" yaml:"items"`
}
