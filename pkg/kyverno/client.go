package kyverno

// PolicyListener is called whenver a new Policy comes in
type PolicyListener = func(LifecycleEvent)

// PolicyClient to watch for LifecycleEvents in the cluster
type PolicyClient interface {
	// Run watches for Policy Events
	Run(int, chan struct{}) error
	// HasSynced all CRDs
	HasSynced() bool
}
