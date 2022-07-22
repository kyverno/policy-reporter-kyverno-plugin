package kyverno

import "context"

// PolicyListener is called whenver a new Policy comes in
type PolicyListener = func(LifecycleEvent)

// PolicyClient to watch for LifecycleEvents in the cluster
type PolicyClient interface {
	// Run watches for Policy Events
	Run(chan struct{}) error
	// HasSynced all CRDs
	HasSynced() bool
}

// EventClient to watch for PolicyViolations in the cluster
type EventClient interface {
	// StartWatching watches for PolicyViolation Events and return a channel with incomming events
	StartWatching(ctx context.Context) <-chan PolicyViolation
}
