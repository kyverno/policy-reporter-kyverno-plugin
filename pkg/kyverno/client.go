package kyverno

import (
	"k8s.io/apimachinery/pkg/watch"
)

// PolicyCallback is called whenver a new PolicyReport comes in
type PolicyCallback = func(watch.EventType, Policy, Policy)

type PolicyClient interface {
	// RegisterPolicyReportCallback register Handlers called on each PolicyReport watch.Event
	RegisterCallback(PolicyCallback)
	// FetchPolicies from the unterlying API
	FetchPolicies() ([]Policy, error)
	// StartWatching calls the WatchAPI, waiting for incoming PolicyReport watch.Events and call the registered Handlers
	StartWatching() error
}
