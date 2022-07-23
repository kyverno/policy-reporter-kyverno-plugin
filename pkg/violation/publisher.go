package violation

import (
	"sync"
)

type Listener = func(PolicyViolation)

type Publisher struct {
	listeners []Listener
}

func (p *Publisher) RegisterListener(listener Listener) {
	p.listeners = append(p.listeners, listener)
}

func (p *Publisher) GetListener() []Listener {
	return p.listeners
}

func (p *Publisher) Publish(event PolicyViolation) {
	wg := sync.WaitGroup{}
	wg.Add(len(p.listeners))

	for _, listener := range p.listeners {
		go func(li Listener, ev PolicyViolation) {
			li(event)
			wg.Done()
		}(listener, event)
	}

	wg.Wait()
}

// NewPublisher constructure for EventPublisher
func NewPublisher() *Publisher {
	return &Publisher{}
}
