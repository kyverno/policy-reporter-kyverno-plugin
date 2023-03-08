package v1

import (
	"reflect"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

type ImageExtractorConfigs map[string][]ImageExtractorConfig

type ImageExtractorConfig struct {
	// Path is the path to the object containing the image field in a custom resource.
	// It should be slash-separated. Each slash-separated key must be a valid YAML key or a wildcard '*'.
	// Wildcard keys are expanded in case of arrays or objects.
	Path string `json:"path" yaml:"path"`
	// Value is an optional name of the field within 'path' that points to the image URI.
	// This is useful when a custom 'key' is also defined.
	// +optional
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
	// Name is the entry the image will be available under 'images.<name>' in the context.
	// If this field is not defined, image entries will appear under 'images.custom'.
	// +optional
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Key is an optional name of the field within 'path' that will be used to uniquely identify an image.
	// Note - this field MUST be unique.
	// +optional
	Key string `json:"key,omitempty" yaml:"key,omitempty"`
}

// Rule defines a validation, mutation, or generation control for matching resources.
// Each rules contains a match declaration to select resources, and an optional exclude
// declaration to specify which resources to exclude.
type Rule struct {
	// Name is a label to identify the rule, It must be unique within the policy.
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Context defines variables and data sources that can be used during rule execution.
	// +optional
	Context []ContextEntry `json:"context,omitempty" yaml:"context,omitempty"`

	// MatchResources defines when this policy rule should be applied. The match
	// criteria can include resource information (e.g. kind, name, namespace, labels)
	// and admission review request information like the user name or role.
	// At least one kind is required.
	MatchResources MatchResources `json:"match,omitempty" yaml:"match,omitempty"`

	// ExcludeResources defines when this policy rule should not be applied. The exclude
	// criteria can include resource information (e.g. kind, name, namespace, labels)
	// and admission review request information like the name or role.
	// +optional
	ExcludeResources MatchResources `json:"exclude,omitempty" yaml:"exclude,omitempty"`

	// ImageExtractors defines a mapping from kinds to ImageExtractorConfigs.
	// This config is only valid for verifyImages rules.
	// +optional
	ImageExtractors ImageExtractorConfigs `json:"imageExtractors,omitempty" yaml:"imageExtractors,omitempty"`

	// Preconditions are used to determine if a policy rule should be applied by evaluating a
	// set of conditions. The declaration can contain nested `any` or `all` statements. A direct list
	// of conditions (without `any` or `all` statements is supported for backwards compatibility but
	// will be deprecated in the next major release.
	// See: https://kyverno.io/docs/writing-policies/preconditions/
	// +optional
	RawAnyAllConditions *apiextv1.JSON `json:"preconditions,omitempty" yaml:"preconditions,omitempty"`

	// Mutation is used to modify matching resources.
	// +optional
	Mutation Mutation `json:"mutate,omitempty" yaml:"mutate,omitempty"`

	// Validation is used to validate matching resources.
	// +optional
	Validation Validation `json:"validate,omitempty" yaml:"validate,omitempty"`

	// Generation is used to create new resources.
	// +optional
	Generation Generation `json:"generate,omitempty" yaml:"generate,omitempty"`

	// VerifyImages is used to verify image signatures and mutate them to add a digest
	// +optional
	VerifyImages []ImageVerification `json:"verifyImages,omitempty" yaml:"verifyImages,omitempty"`
}

// HasMutate checks for mutate rule
func (r *Rule) HasMutate() bool {
	return !reflect.DeepEqual(r.Mutation, Mutation{})
}

// HasVerifyImages checks for verifyImages rule
func (r *Rule) HasVerifyImages() bool {
	return r.VerifyImages != nil && !reflect.DeepEqual(r.VerifyImages, ImageVerification{})
}

// HasYAMLSignatureVerify checks for validate.manifests rule
func (r Rule) HasYAMLSignatureVerify() bool {
	return r.Validation.Manifests != nil && len(r.Validation.Manifests.Attestors) != 0
}

// HasImagesValidationChecks checks whether the verifyImages rule has validation checks
func (r *Rule) HasImagesValidationChecks() bool {
	for _, v := range r.VerifyImages {
		if v.VerifyDigest || v.Required {
			return true
		}
	}

	return false
}

// HasYAMLSignatureVerify checks for validate rule
func (p *ClusterPolicy) HasYAMLSignatureVerify() bool {
	for _, rule := range p.Spec.Rules {
		if rule.HasYAMLSignatureVerify() {
			return true
		}
	}

	return false
}

// HasValidate checks for validate rule
func (r *Rule) HasValidate() bool {
	return !reflect.DeepEqual(r.Validation, Validation{})
}

// HasGenerate checks for generate rule
func (r *Rule) HasGenerate() bool {
	return !reflect.DeepEqual(r.Generation, Generation{})
}

// IsMutateExisting checks if the mutate rule applies to existing resources
func (r *Rule) IsMutateExisting() bool {
	return r.Mutation.Targets != nil
}
