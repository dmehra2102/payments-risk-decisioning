package store

import "sync"

type StateStore struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

func New() *StateStore {
	return &StateStore{seen: make(map[string]struct{})}
}

func (s *StateStore) Seen(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.seen[id]
	return ok
}

func (s *StateStore) Mark(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen[id] = struct{}{}
}
