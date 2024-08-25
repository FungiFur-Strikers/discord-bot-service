package flow

import (
	"sync"
)

type Store interface {
	Save(id string, data FlowData)
	Get(id string) (FlowData, bool)
}

type InMemoryStore struct {
	data map[string]FlowData
	mu   sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]FlowData),
	}
}

func (s *InMemoryStore) Save(id string, data FlowData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = data
}

func (s *InMemoryStore) Get(id string) (FlowData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.data[id]
	return data, ok
}
