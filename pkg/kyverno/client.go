package kyverno

import "context"

// PolicyListener is called whenver a new Policy comes in
type PolicyListener = func(LifecycleEvent)

// PolicyClient to watch for LifecycleEvents in the cluster
type PolicyClient interface {
	// StartWatching watches for Policy Events and return a channel with incomming results
	StartWatching(ctx context.Context) <-chan LifecycleEvent
	// GetFoundResources as Map of Names
	GetFoundResources() map[string]bool
}
