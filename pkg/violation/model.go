package violation

import "time"

type Resource struct {
	Kind      string
	Name      string
	Namespace string
}

type Event struct {
	Name string
	UID  string
}

type Policy struct {
	Name     string
	Rule     string
	Message  string
	Category string
	Severity string
}

type PolicyViolation struct {
	Resource  Resource
	Policy    Policy
	Event     Event
	Timestamp time.Time
	Updated   bool
}

// EventClient to watch for PolicyViolations in the cluster
type EventClient interface {
	// Run watches for PolicyViolation Events
	Run(stopper chan struct{}) error
}
