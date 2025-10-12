package storage

// List operations

func (s *MemoryStorage) LPush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	storageValue, exists := s.data[key]
	if exists && storageValue.IsExpired() {
		delete(s.data, key)
		exists = false
	}

	if !exists {
		storageValue = NewListValue()
		s.data[key] = storageValue
	}

	if storageValue.Type != ListType {
		return 0, ErrWrongType
	}

	listData := storageValue.Data.(*ListData)
	for _, value := range values {
		listData.PushLeft(value)
	}

	return listData.Len(), nil
}

func (s *MemoryStorage) RPush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	storageValue, exists := s.data[key]
	if exists && storageValue.IsExpired() {
		delete(s.data, key)
		exists = false
	}

	if !exists {
		storageValue = NewListValue()
		s.data[key] = storageValue
	}

	if storageValue.Type != ListType {
		return 0, ErrWrongType
	}

	listData := storageValue.Data.(*ListData)
	for _, value := range values {
		listData.PushRight(value)
	}

	return listData.Len(), nil
}

func (s *MemoryStorage) LPop(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, exists := s.data[key]
	if !exists {
		return "", ErrKeyNotFound
	}

	if value.IsExpired() {
		delete(s.data, key)
		return "", ErrKeyExpired
	}

	if value.Type != ListType {
		return "", ErrWrongType
	}

	listData := value.Data.(*ListData)
	if element, ok := listData.PopLeft(); ok {
		return element, nil
	}

	return "", ErrKeyNotFound
}

func (s *MemoryStorage) RPop(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, exists := s.data[key]
	if !exists {
		return "", ErrKeyNotFound
	}

	if value.IsExpired() {
		delete(s.data, key)
		return "", ErrKeyExpired
	}

	if value.Type != ListType {
		return "", ErrWrongType
	}

	listData := value.Data.(*ListData)
	if element, ok := listData.PopRight(); ok {
		return element, nil
	}

	return "", ErrKeyNotFound
}

func (s *MemoryStorage) LLen(key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return 0, ErrKeyNotFound
	}

	if value.IsExpired() {
		return 0, ErrKeyExpired
	}

	if value.Type != ListType {
		return 0, ErrWrongType
	}

	listData := value.Data.(*ListData)
	return listData.Len(), nil
}

func (s *MemoryStorage) LRange(key string, start, stop int) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	if value.IsExpired() {
		return nil, ErrKeyExpired
	}

	if value.Type != ListType {
		return nil, ErrWrongType
	}

	listData := value.Data.(*ListData)
	return listData.Range(start, stop), nil
}
