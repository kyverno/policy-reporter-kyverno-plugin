package kyverno

import "sync"

// EventPublisher for LifecycleEvents
type EventPublisher interface {
	// RegisterListener register Handlers called on each PolicyReport watch.Event
	RegisterListener(PolicyListener)
	// GetListener returns a list of all registered Listeners
	GetListener() []PolicyListener
	// Publish events to the registered listeners
	Publish(eventChan <-chan LifecycleEvent)
}

type lifecycleEventPublisher struct {
	listeners []PolicyListener
}

func (p *lifecycleEventPublisher) RegisterListener(listener PolicyListener) {
	p.listeners = append(p.listeners, listener)
}

func (p *lifecycleEventPublisher) GetListener() []PolicyListener {
	return p.listeners
}

func (p *lifecycleEventPublisher) Publish(eventChan <-chan LifecycleEvent) {
	for event := range eventChan {
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
}

// NewEventPublisher constructure for EventPublisher
func NewEventPublisher() EventPublisher {
	return &lifecycleEventPublisher{}
}
