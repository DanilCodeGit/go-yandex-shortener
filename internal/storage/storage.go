package storage

import "sync"

type ST interface {
	SetURL(key, value string)
	GetURL(key string) (string, bool)
	DeleteURL(key string)
}
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
func (s *Storage) SetBatch(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.URLsStore[key] = value
}

//type BatchStorage struct {
//	BatchesStorage map[string]string
//	mu             sync.RWMutex
//}

//func NewBatchStorage() *BatchStorage {
//	return &BatchStorage{
//		BatchesStorage: make(map[string]string),
//	}
//}
//
//func (s *BatchStorage) SetURL(key, value string) {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	s.BatchesStorage[key] = value
//}
//func (s *BatchStorage) GetURL(key string) (string, bool) {
//	s.mu.RLock()
//	defer s.mu.RUnlock()
//	value := s.BatchesStorage[key]
//	return value, true
//}
//
//func (s *BatchStorage) DeleteURL(key string) {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	delete(s.BatchesStorage, key)
//}
