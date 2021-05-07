package kyverno

import "sync"

type PolicyStore struct {
	store map[string]Policy
	rwm   *sync.RWMutex
}

func (s *PolicyStore) Get(id string) (Policy, bool) {
	s.rwm.RLock()
	r, ok := s.store[id]
	s.rwm.RUnlock()

	return r, ok
}

func (s *PolicyStore) List() []Policy {
	s.rwm.RLock()
	list := make([]Policy, 0, len(s.store))

	for _, r := range s.store {
		list = append(list, r)
	}
	s.rwm.RUnlock()

	return list
}

func (s *PolicyStore) Add(r Policy) {
	s.rwm.Lock()
	s.store[r.GetIdentifier()] = r
	s.rwm.Unlock()
}

func (s *PolicyStore) Remove(id string) {
	s.rwm.Lock()
	delete(s.store, id)
	s.rwm.Unlock()
}

func NewPolicyStore() *PolicyStore {
	return &PolicyStore{
		store: map[string]Policy{},
		rwm:   new(sync.RWMutex),
	}
}
