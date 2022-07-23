package kyverno

import (
	"sync"
)

type EventPublisher struct {
	listeners []PolicyListener
}

// RegisterListener register Handlers called on each PolicyReport watch.Event
func (p *EventPublisher) RegisterListener(listener PolicyListener) {
	p.listeners = append(p.listeners, listener)
}

// GetListener returns a list of all registered Listeners
func (p *EventPublisher) GetListener() []PolicyListener {
	return p.listeners
}

// Publish events to the registered listeners
func (p *EventPublisher) Publish(event LifecycleEvent) {
	wg := sync.WaitGroup{}
	wg.Add(len(p.listeners))

	for _, listener := range p.listeners {
		go func(li PolicyListener, ev LifecycleEvent) {
			li(event)
			wg.Done()
		}(listener, event)
	}

	wg.Wait()
}

// NewEventPublisher constructure for EventPublisher
func NewEventPublisher() *EventPublisher {
	return &EventPublisher{}
}
