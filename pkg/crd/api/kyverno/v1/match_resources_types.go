package v1

// MatchResources is used to specify resource and admission review request data for
// which a policy rule is applicable.
type MatchResources struct {
	// Any allows specifying resources which will be ORed
	// +optional
	Any ResourceFilters `json:"any,omitempty" yaml:"any,omitempty"`

	// All allows specifying resources which will be ANDed
	// +optional
	All ResourceFilters `json:"all,omitempty" yaml:"all,omitempty"`

	// UserInfo contains information about the user performing the operation.
	// Specifying UserInfo directly under match is being deprecated.
	// Please specify under "any" or "all" instead.
	// +optional
	UserInfo `json:",omitempty" yaml:",omitempty"`

	// ResourceDescription contains information about the resource being created or modified.
	// Requires at least one tag to be specified when under MatchResources.
	// Specifying ResourceDescription directly under match is being deprecated.
	// Please specify under "any" or "all" instead.
	// +optional
	ResourceDescription `json:"resources,omitempty" yaml:"resources,omitempty"`
}
