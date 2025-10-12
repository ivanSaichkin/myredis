package storage

// Hash operations

func (s *MemoryStorage) HSet(key, field, value string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	storageValue, exists := s.data[key]
	if exists && storageValue.IsExpired() {
		delete(s.data, key)
		exists = false
	}

	if !exists {
		storageValue = NewHashValue()
		s.data[key] = storageValue
	}

	if storageValue.Type != HashType {
		return false, ErrWrongType
	}

	hashData := storageValue.Data.(*HashData)
	exists = hashData.Exists(field)
	hashData.Set(field, value)

	return !exists, nil
}

func (s *MemoryStorage) HGet(key, field string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return "", ErrKeyNotFound
	}

	if value.IsExpired() {
		return "", ErrKeyExpired
	}

	if value.Type != HashType {
		return "", ErrWrongType
	}

	hashData := value.Data.(*HashData)
	if val, exists := hashData.Get(field); exists {
		return val, nil
	}

	return "", ErrFieldNotFound
}

func (s *MemoryStorage) HDel(key, field string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, exists := s.data[key]
	if !exists {
		return false, ErrKeyNotFound
	}

	if value.IsExpired() {
		delete(s.data, key)
		return false, ErrKeyExpired
	}

	if value.Type != HashType {
		return false, ErrWrongType
	}

	hashData := value.Data.(*HashData)
	return hashData.Delete(field), nil
}

func (s *MemoryStorage) HExists(key, field string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return false, ErrKeyNotFound
	}

	if value.IsExpired() {
		return false, ErrKeyExpired
	}

	if value.Type != HashType {
		return false, ErrWrongType
	}

	hashData := value.Data.(*HashData)
	return hashData.Exists(field), nil
}

func (s *MemoryStorage) HGetAll(key string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	if value.IsExpired() {
		return nil, ErrKeyExpired
	}

	if value.Type != HashType {
		return nil, ErrWrongType
	}

	hashData := value.Data.(*HashData)
	return hashData.Fields(), nil
}

func (s *MemoryStorage) HKeys(key string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	if value.IsExpired() {
		return nil, ErrKeyExpired
	}

	if value.Type != HashType {
		return nil, ErrWrongType
	}

	hashData := value.Data.(*HashData)
	fields := hashData.Fields()
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *MemoryStorage) HLen(key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return 0, ErrKeyNotFound
	}

	if value.IsExpired() {
		return 0, ErrKeyExpired
	}

	if value.Type != HashType {
		return 0, ErrWrongType
	}

	hashData := value.Data.(*HashData)
	return hashData.Len(), nil
}
