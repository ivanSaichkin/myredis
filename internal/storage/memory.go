package storage

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string]*StorageValue
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]*StorageValue),
	}
}

func (s *MemoryStorage) Get(key string) (*StorageValue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	if value.IsExpired() {
		return nil, ErrKeyExpired
	}

	return value, nil
}

func (s *MemoryStorage) Set(key string, value interface{}) error {
	return s.setInternal(key, value, 0)
}

func (s *MemoryStorage) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	return s.setInternal(key, value, ttl)
}

func (s *MemoryStorage) setInternal(key string, value interface{}, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var storageValue *StorageValue

	switch v := value.(type) {
	case string:
		storageValue = NewStringValue(v)
	case *StorageValue:
		storageValue = v
	default:
		storageValue = NewStringValue(toString(v))
	}

	if ttl > 0 {
		storageValue.ExpiredAt = time.Now().Add(ttl)
	}

	s.data[key] = storageValue
	return nil
}

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

func (s *MemoryStorage) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return false
	}

	if value.IsExpired() {
		return false
	}

	return true
}

func (s *MemoryStorage) Type(key string) (ValueType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return StringType, ErrKeyNotFound
	}

	if value.IsExpired() {
		return StringType, ErrKeyExpired
	}

	return value.Type, nil
}

func (s *MemoryStorage) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))

	for key, value := range s.data {
		if !value.IsExpired() {
			keys = append(keys, key)
		}
	}

	return keys
}

func (s *MemoryStorage) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, value := range s.data {
		if !value.IsExpired() {
			count++
		}
	}

	return count
}

func (s *MemoryStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]*StorageValue)
}

// TTL operations

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

func (s *MemoryStorage) TTL(key string) (time.Duration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return 0, ErrKeyNotFound
	}

	if value.IsExpired() {
		return 0, ErrKeyExpired
	}

	if value.ExpiredAt.IsZero() {
		return -1, nil
	}

	return time.Until(value.ExpiredAt), nil
}

// Cleanup

func (s *MemoryStorage) StartExpirationChecker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			s.CleanupExpired()
		}
	}()
}

func (s *MemoryStorage) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, value := range s.data {
		if !value.ExpiredAt.IsZero() && now.After(value.ExpiredAt) {
			delete(s.data, key)
		}
	}
}

// Helper functions

func toString(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func intersectSlices(a, b []string) []string {
	set := make(map[string]bool)
	for _, item := range a {
		set[item] = true
	}

	result := make([]string, 0)
	for _, item := range b {
		if set[item] {
			result = append(result, item)
		}
	}
	return result
}
