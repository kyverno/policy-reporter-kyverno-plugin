package kyverno

import (
	"sync"
)

type PolicyViolationListener = func(PolicyViolation)

// ViolationPublisher for LifecycleEvents
type ViolationPublisher interface {
	// RegisterListener register Handlers called on each PolicyViolation
	RegisterListener(PolicyViolationListener)
	// GetListener returns a list of all registered Listeners
	GetListener() []PolicyViolationListener
	// Publish events to the registered listeners
	Publish(PolicyViolation)
}

type publisher struct {
	listeners []PolicyViolationListener
}

func (p *publisher) RegisterListener(listener PolicyViolationListener) {
	p.listeners = append(p.listeners, listener)
}

func (p *publisher) GetListener() []PolicyViolationListener {
	return p.listeners
}

func (p *publisher) Publish(event PolicyViolation) {
	wg := sync.WaitGroup{}
	wg.Add(len(p.listeners))

	for _, listener := range p.listeners {
		go func(li PolicyViolationListener, ev PolicyViolation) {
			li(event)
			wg.Done()
		}(listener, event)
	}

	wg.Wait()
}

// NewViolationPublisher constructure for EventPublisher
func NewViolationPublisher() ViolationPublisher {
	return &publisher{}
}
