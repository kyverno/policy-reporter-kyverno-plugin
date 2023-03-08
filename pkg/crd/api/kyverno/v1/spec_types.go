package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidationFailureAction defines the policy validation failure action
type ValidationFailureAction string

// Policy Reporting Modes
const (
	// auditOld doesn't block the request on failure
	// DEPRECATED: use Audit instead
	auditOld ValidationFailureAction = "audit"
	// enforceOld blocks the request on failure
	// DEPRECATED: use Enforce instead
	enforceOld ValidationFailureAction = "enforce"
	// Enforce blocks the request on failure
	Enforce ValidationFailureAction = "Enforce"
	// Audit doesn't block the request on failure
	Audit ValidationFailureAction = "Audit"
)

func (a ValidationFailureAction) Enforce() bool {
	return a == Enforce || a == enforceOld
}

func (a ValidationFailureAction) Audit() bool {
	return !a.Enforce()
}

func (a ValidationFailureAction) IsValid() bool {
	return a == enforceOld || a == auditOld || a == Enforce || a == Audit
}

type ValidationFailureActionOverride struct {
	// +kubebuilder:validation:Enum=audit;enforce;Audit;Enforce
	Action            ValidationFailureAction `json:"action,omitempty" yaml:"action,omitempty"`
	Namespaces        []string                `json:"namespaces,omitempty" yaml:"namespaces,omitempty"`
	NamespaceSelector *metav1.LabelSelector   `json:"namespaceSelector,omitempty" yaml:"namespaceSelector,omitempty"`
}

// Spec contains a list of Rule instances and other policy controls.
type Spec struct {
	// Rules is a list of Rule instances. A Policy contains multiple rules and
	// each rule can validate, mutate, or generate resources.
	Rules []Rule `json:"rules,omitempty" yaml:"rules,omitempty"`

	// ApplyRules controls how rules in a policy are applied. Rule are processed in
	// the order of declaration. When set to `One` processing stops after a rule has
	// been applied i.e. the rule matches and results in a pass, fail, or error. When
	// set to `All` all rules in the policy are processed. The default is `All`.
	// +optional
	ApplyRules *ApplyRulesType `json:"applyRules,omitempty" yaml:"applyRules,omitempty"`

	// FailurePolicy defines how unexpected policy errors and webhook response timeout errors are handled.
	// Rules within the same policy share the same failure behavior.
	// This field should not be accessed directly, instead `GetFailurePolicy()` should be used.
	// Allowed values are Ignore or Fail. Defaults to Fail.
	// +optional
	FailurePolicy *FailurePolicyType `json:"failurePolicy,omitempty" yaml:"failurePolicy,omitempty"`

	// ValidationFailureAction defines if a validation policy rule violation should block
	// the admission review request (enforce), or allow (audit) the admission review request
	// and report an error in a policy report. Optional.
	// Allowed values are audit or enforce. The default value is "Audit".
	// +optional
	// +kubebuilder:validation:Enum=audit;enforce;Audit;Enforce
	// +kubebuilder:default=Audit
	ValidationFailureAction ValidationFailureAction `json:"validationFailureAction,omitempty" yaml:"validationFailureAction,omitempty"`

	// ValidationFailureActionOverrides is a Cluster Policy attribute that specifies ValidationFailureAction
	// namespace-wise. It overrides ValidationFailureAction for the specified namespaces.
	// +optional
	ValidationFailureActionOverrides []ValidationFailureActionOverride `json:"validationFailureActionOverrides,omitempty" yaml:"validationFailureActionOverrides,omitempty"`

	// Background controls if rules are applied to existing resources during a background scan.
	// Optional. Default value is "true". The value must be set to "false" if the policy rule
	// uses variables that are only available in the admission review request (e.g. user name).
	// +optional
	// +kubebuilder:default=true
	Background *bool `json:"background,omitempty" yaml:"background,omitempty"`

	// SchemaValidation skips validation checks for policies as well as patched resources.
	// Optional. The default value is set to "true", it must be set to "false" to disable the validation checks.
	// +optional
	SchemaValidation *bool `json:"schemaValidation,omitempty" yaml:"schemaValidation,omitempty"`

	// WebhookTimeoutSeconds specifies the maximum time in seconds allowed to apply this policy.
	// After the configured time expires, the admission request may fail, or may simply ignore the policy results,
	// based on the failure policy. The default timeout is 10s, the value must be between 1 and 30 seconds.
	WebhookTimeoutSeconds *int32 `json:"webhookTimeoutSeconds,omitempty" yaml:"webhookTimeoutSeconds,omitempty"`

	// MutateExistingOnPolicyUpdate controls if a mutateExisting policy is applied on policy events.
	// Default value is "false".
	// +optional
	MutateExistingOnPolicyUpdate bool `json:"mutateExistingOnPolicyUpdate,omitempty" yaml:"mutateExistingOnPolicyUpdate,omitempty"`

	// GenerateExistingOnPolicyUpdate controls whether to trigger generate rule in existing resources
	// If is set to "true" generate rule will be triggered and applied to existing matched resources.
	// Defaults to "false" if not specified.
	// +optional
	GenerateExistingOnPolicyUpdate bool `json:"generateExistingOnPolicyUpdate,omitempty" yaml:"generateExistingOnPolicyUpdate,omitempty"`
}

// HasMutate checks for mutate rule types
func (s *Spec) HasMutate() bool {
	for _, rule := range s.Rules {
		if rule.HasMutate() {
			return true
		}
	}

	return false
}

// HasValidate checks for validate rule types
func (s *Spec) HasValidate() bool {
	for _, rule := range s.Rules {
		if rule.HasValidate() {
			return true
		}
	}

	return false
}

// HasGenerate checks for generate rule types
func (s *Spec) HasGenerate() bool {
	for _, rule := range s.Rules {
		if rule.HasGenerate() {
			return true
		}
	}

	return false
}

// HasImagesValidationChecks checks for image verification rules invoked during resource validation
func (s *Spec) HasImagesValidationChecks() bool {
	for _, rule := range s.Rules {
		if rule.HasImagesValidationChecks() {
			return true
		}
	}

	return false
}

// HasVerifyImages checks for image verification rules invoked during resource mutation
func (s *Spec) HasVerifyImages() bool {
	for _, rule := range s.Rules {
		if rule.HasVerifyImages() {
			return true
		}
	}

	return false
}

// HasYAMLSignatureVerify checks for image verification rules invoked during resource mutation
func (s *Spec) HasYAMLSignatureVerify() bool {
	for _, rule := range s.Rules {
		if rule.HasYAMLSignatureVerify() {
			return true
		}
	}

	return false
}
