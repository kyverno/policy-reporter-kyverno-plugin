package kyverno

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
	// Run watches for PolicyViolation Events
	Run(stopper chan struct{}) error
}
