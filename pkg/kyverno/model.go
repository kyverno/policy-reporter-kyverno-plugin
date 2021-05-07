package kyverno

import (
	"time"
)

// Rule from the Policy spec clusterpolicies.kyverno.io/v1.Policy
type Rule struct {
	ValidateMessage string `json:"message,omitempty"`
	Name            string `json:"name"`
	Type            string `json:"type"`
}

// Policy spec clusterpolicies.kyverno.io/v1.Policy
type Policy struct {
	Kind                    string    `json:"kind"`
	Name                    string    `json:"name"`
	Namespace               string    `json:"namespace,omitempty"`
	AutogenControllers      []string  `json:"autogenControllers,omitempty"`
	ValidationFailureAction string    `json:"validationFailureAction,omitempty"`
	Background              bool      `json:"background"`
	Rules                   []Rule    `json:"rules"`
	Category                string    `json:"category,omitempty"`
	Description             string    `json:"description,omitempty"`
	Severity                string    `json:"severity,omitempty"`
	CreationTimestamp       time.Time `json:"creationTimestamp,omitempty"`
	UID                     string    `json:"uid,omitempty"`
	Content                 string    `json:"content"`
}

// GetIdentifier returns a global unique Policy identifier
func (pr Policy) GetIdentifier() string {
	return pr.UID
}
