package storage

import "sync"

type Storage struct {
	URLsStore map[string]string
	mu        sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		URLsStore: make(map[string]string),
	}
}
func (s *Storage) SetURL(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.URLsStore[key] = value
}
func (s *Storage) GetURL(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value := s.URLsStore[key]
	return value, true
}

func (s *Storage) DeleteURL(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.URLsStore, key)
}
