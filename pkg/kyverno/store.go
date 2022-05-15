package kyverno

import (
	"sync"
)

// PolicyStore persists the last state of a Policy in memory
type PolicyStore struct {
	store map[string]*Policy
	rwm   *sync.RWMutex
}

// Get a Policy from the Store by ID
func (s *PolicyStore) Get(id string) (*Policy, bool) {
	s.rwm.RLock()
	r, ok := s.store[id]
	s.rwm.RUnlock()

	return r, ok
}

// List all stored Policies
func (s *PolicyStore) List() []*Policy {
	s.rwm.RLock()
	list := make([]*Policy, 0, len(s.store))

	for _, r := range s.store {
		list = append(list, r)
	}
	s.rwm.RUnlock()

	return list
}

// Add a Policy to the store
func (s *PolicyStore) Add(r *Policy) {
	s.rwm.Lock()
	s.store[r.UID] = r
	s.rwm.Unlock()
}

// Remove a Policy to the store
func (s *PolicyStore) Remove(id string) {
	s.rwm.Lock()
	delete(s.store, id)
	s.rwm.Unlock()
}

// NewPolicyStore returns a pointer to a new in memory store
func NewPolicyStore() *PolicyStore {
	return &PolicyStore{
		store: map[string]*Policy{},
		rwm:   new(sync.RWMutex),
	}
}
