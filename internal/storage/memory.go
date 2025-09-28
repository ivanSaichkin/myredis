package storage

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExpired  = errors.New("key expired")
)

type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string]*Value
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]*Value),
	}
}

// Get method
func (s *MemoryStorage) Get(key string) (*Value, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	if value.IsExpired() {
		go s.deleteKey(key)
		return nil, ErrKeyExpired
	}

	return value, nil
}

func (v *Value) IsExpired() bool {
	return !v.ExpiredAt.IsZero() && time.Now().After(v.ExpiredAt)
}

// Set methods
func (s *MemoryStorage) Set(key string, value interface{}) error {
	return s.setInternal(key, value, 0)
}

func (s *MemoryStorage) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	return s.setInternal(key, value, ttl)
}

func (s *MemoryStorage) setInternal(key string, value interface{}, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	newValue := &Value{
		Value: value,
	}

	if ttl > 0 {
		newValue.ExpiredAt = time.Now().Add(ttl)
	}

	s.data[key] = newValue

	return nil
}

// Delete methods
func (s *MemoryStorage) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.deleteKey(key)
}

func (s *MemoryStorage) deleteKey(key string) bool {
	if _, exists := s.data[key]; exists {
		delete(s.data, key)
		return true
	}
	return false
}

// Exists method
func (s *MemoryStorage) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return false
	}

	if value.IsExpired() {
		go s.deleteKey(key)
		return false
	}

	return true
}

func (s *MemoryStorage) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))

	for key, value := range s.data {
		if value.IsExpired() {
			go s.deleteKey(key)
			continue
		}

		keys = append(keys, key)
	}

	return keys
}

func (s *MemoryStorage) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0

	for key, value := range s.data {
		if value.IsExpired() {
			go s.deleteKey(key)
			continue
		}
		count++
	}

	return count
}

func (s *MemoryStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]*Value)
}

func (s *MemoryStorage) TTL(key string) (time.Duration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]

	if !exists {
		return 0, ErrKeyNotFound
	}

	if value.IsExpired() {
		go s.deleteKey(key)
		return 0, ErrKeyExpired
	}

	if value.ExpiredAt.IsZero() {
		return -1, nil
	}

	return time.Until(value.ExpiredAt), nil
}

func (s *MemoryStorage) Expire(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, exists := s.data[key]
	if !exists || value.IsExpired() {
		return false
	}

	value.ExpiredAt = time.Now().Add(ttl)
	return true
}

// Auto expire checker
func (s *MemoryStorage) checkExpiredKeys() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for key, value := range s.data {
		if !value.ExpiredAt.IsZero() && now.After(value.ExpiredAt) {
			delete(s.data, key)
		}
	}
}

func (s *MemoryStorage) StartExpirationChecker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			s.checkExpiredKeys()
		}
	}()
}
