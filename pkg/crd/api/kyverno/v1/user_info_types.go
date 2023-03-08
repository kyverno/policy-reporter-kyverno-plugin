package v1

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

// UserInfo contains information about the user performing the operation.
type UserInfo struct {
	// Roles is the list of namespaced role names for the user.
	// +optional
	Roles []string `json:"roles,omitempty" yaml:"roles,omitempty"`

	// ClusterRoles is the list of cluster-wide role names for the user.
	// +optional
	ClusterRoles []string `json:"clusterRoles,omitempty" yaml:"clusterRoles,omitempty"`

	// Subjects is the list of subject names like users, user groups, and service accounts.
	// +optional
	Subjects []rbacv1.Subject `json:"subjects,omitempty" yaml:"subjects,omitempty"`
}
