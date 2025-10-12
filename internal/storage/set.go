package storage

// Set operations

func (s *MemoryStorage) SAdd(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	storageValue, exists := s.data[key]
	if exists && storageValue.IsExpired() {
		delete(s.data, key)
		exists = false
	}

	if !exists {
		storageValue = NewSetValue()
		s.data[key] = storageValue
	}

	if storageValue.Type != SetType {
		return 0, ErrWrongType
	}

	setData := storageValue.Data.(*SetData)
	added := 0
	for _, member := range members {
		if setData.Add(member) {
			added++
		}
	}

	return added, nil
}

func (s *MemoryStorage) SRem(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, exists := s.data[key]
	if !exists {
		return 0, ErrKeyNotFound
	}

	if value.IsExpired() {
		delete(s.data, key)
		return 0, ErrKeyExpired
	}

	if value.Type != SetType {
		return 0, ErrWrongType
	}

	setData := value.Data.(*SetData)
	removed := 0
	for _, member := range members {
		if setData.Remove(member) {
			removed++
		}
	}

	return removed, nil
}

func (s *MemoryStorage) SIsMember(key, member string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return false, ErrKeyNotFound
	}

	if value.IsExpired() {
		return false, ErrKeyExpired
	}

	if value.Type != SetType {
		return false, ErrWrongType
	}

	setData := value.Data.(*SetData)
	return setData.IsMember(member), nil
}

func (s *MemoryStorage) SMembers(key string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	if value.IsExpired() {
		return nil, ErrKeyExpired
	}

	if value.Type != SetType {
		return nil, ErrWrongType
	}

	setData := value.Data.(*SetData)
	return setData.Members(), nil
}

func (s *MemoryStorage) SCard(key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return 0, ErrKeyNotFound
	}

	if value.IsExpired() {
		return 0, ErrKeyExpired
	}

	if value.Type != SetType {
		return 0, ErrWrongType
	}

	setData := value.Data.(*SetData)
	return setData.Len(), nil
}

func (s *MemoryStorage) SInter(keys ...string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(keys) == 0 {
		return []string{}, nil
	}

	firstValue, exists := s.data[keys[0]]
	if !exists {
		return []string{}, nil
	}

	if firstValue.IsExpired() {
		return []string{}, nil
	}

	if firstValue.Type != SetType {
		return nil, ErrWrongType
	}

	firstSet := firstValue.Data.(*SetData)
	result := firstSet.Members()

	for i := 1; i < len(keys); i++ {
		value, exists := s.data[keys[i]]
		if !exists || value.IsExpired() || value.Type != SetType {
			return []string{}, nil
		}

		currentSet := value.Data.(*SetData)
		result = intersectSlices(result, currentSet.Members())
	}

	return result, nil
}
