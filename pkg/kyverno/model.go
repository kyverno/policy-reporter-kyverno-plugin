package kyverno

import (
	"time"
)

// Event Enum
type Event = int

// Possible Policy Event Enums
const (
	Added Event = iota
	Updated
	Deleted
)

const (
	PolicyKind        = "Policy"
	ClusterPolicyKind = "ClusterPolicy"
)

// LifecycleEvent of Policys
type LifecycleEvent struct {
	Type      Event
	NewPolicy Policy
	OldPolicy Policy
}

// VerifyImage from the Policy spec clusterpolicies.kyverno.io/v1.Policy
type VerifyImage struct {
	Attestations string `json:"attestations,omitempty"`
	Repository   string `json:"repository"`
	Image        string `json:"image"`
	Key          string `json:"key"`
}

// Rule from the Policy spec clusterpolicies.kyverno.io/v1.Policy
type Rule struct {
	ValidateMessage string         `json:"message,omitempty"`
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	VerifyImages    []*VerifyImage `json:"verifyImages,omitempty"`
}

// Policy spec clusterpolicies.kyverno.io/v1.Policy
type Policy struct {
	Kind                    string    `json:"kind"`
	Name                    string    `json:"name"`
	Namespace               string    `json:"namespace,omitempty"`
	AutogenControllers      []string  `json:"autogenControllers,omitempty"`
	ValidationFailureAction string    `json:"validationFailureAction,omitempty"`
	Background              bool      `json:"background"`
	Rules                   []*Rule   `json:"rules"`
	Category                string    `json:"category,omitempty"`
	Description             string    `json:"description,omitempty"`
	Severity                string    `json:"severity,omitempty"`
	CreationTimestamp       time.Time `json:"creationTimestamp,omitempty"`
	UID                     string    `json:"uid,omitempty"`
	Content                 string    `json:"content"`
}
