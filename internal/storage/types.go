package storage

import (
	"sync"
	"time"
)

type ValueType int

const (
	StringType ValueType = iota
	HashType
	ListType
	SetType
)

func (vt ValueType) String() string {
	switch vt {
	case StringType:
		return "string"
	case HashType:
		return "hash"
	case ListType:
		return "list"
	case SetType:
		return "set"
	default:
		return "unknown"
	}
}

type StorageValue struct {
	Type      ValueType
	Data      interface{}
	ExpiredAt time.Time
	createdAt time.Time
}

func NewStringValue(data string) *StorageValue {
	return &StorageValue{
		Type:      StringType,
		Data:      data,
		createdAt: time.Now(),
	}
}

func NewHashValue() *StorageValue {
	return &StorageValue{
		Type:      HashType,
		Data:      NewHashData(),
		createdAt: time.Now(),
	}
}

func NewListValue() *StorageValue {
	return &StorageValue{
		Type:      ListType,
		Data:      NewListData(),
		createdAt: time.Now(),
	}
}

func NewSetValue() *StorageValue {
	return &StorageValue{
		Type:      SetType,
		Data:      NewSetData(),
		createdAt: time.Now(),
	}
}

func (sv *StorageValue) IsExpired() bool {
	return !sv.ExpiredAt.IsZero() && time.Now().After(sv.ExpiredAt)
}

type HashData struct {
	fields map[string]string
	mu     sync.RWMutex
}

func NewHashData() *HashData {
	return &HashData{
		fields: make(map[string]string),
	}
}

func (h *HashData) Set(field, value string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.fields[field] = value
}

func (h *HashData) Get(field string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	value, exists := h.fields[field]
	return value, exists
}

func (h *HashData) Delete(field string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, exists := h.fields[field]; exists {
		delete(h.fields, field)
		return true
	}
	return false
}

func (h *HashData) Fields() map[string]string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[string]string, len(h.fields))
	for k, v := range h.fields {
		result[k] = v
	}
	return result
}

func (h *HashData) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.fields)
}

func (h *HashData) Exists(field string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.fields[field]
	return exists
}

type ListData struct {
	elements []string
	mu       sync.RWMutex
}

func NewListData() *ListData {
	return &ListData{
		elements: make([]string, 0),
	}
}

func (l *ListData) PushLeft(element string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.elements = append([]string{element}, l.elements...)
}

func (l *ListData) PushRight(element string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.elements = append(l.elements, element)
}

func (l *ListData) PopLeft() (string, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.elements) == 0 {
		return "", false
	}

	element := l.elements[0]
	l.elements = l.elements[1:]
	return element, true
}

func (l *ListData) PopRight() (string, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.elements) == 0 {
		return "", false
	}

	lastIndex := len(l.elements) - 1
	element := l.elements[lastIndex]
	l.elements = l.elements[:lastIndex]
	return element, true
}

func (l *ListData) Range(start, stop int) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	length := len(l.elements)
	if length == 0 {
		return []string{}
	}

	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop {
		return []string{}
	}

	result := make([]string, stop-start+1)
	copy(result, l.elements[start:stop+1])
	return result
}

func (l *ListData) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.elements)
}

func (l *ListData) GetAll() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]string, len(l.elements))
	copy(result, l.elements)
	return result
}

type SetData struct {
	members map[string]struct{}
	mu      sync.RWMutex
}

func NewSetData() *SetData {
	return &SetData{
		members: make(map[string]struct{}),
	}
}

func (s *SetData) Add(member string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.members[member]; exists {
		return false
	}

	s.members[member] = struct{}{}
	return true
}

func (s *SetData) Remove(member string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.members[member]; exists {
		delete(s.members, member)
		return true
	}
	return false
}

func (s *SetData) IsMember(member string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.members[member]
	return exists
}

func (s *SetData) Members() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	members := make([]string, 0, len(s.members))
	for member := range s.members {
		members = append(members, member)
	}
	return members
}

func (s *SetData) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.members)
}

func (s *SetData) Intersection(other *SetData) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	result := make([]string, 0)
	for member := range s.members {
		if _, exists := other.members[member]; exists {
			result = append(result, member)
		}
	}
	return result
}

func (s *SetData) Union(other *SetData) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	result := make([]string, 0, len(s.members)+len(other.members))

	for member := range s.members {
		result = append(result, member)
	}

	for member := range other.members {
		if _, exists := s.members[member]; !exists {
			result = append(result, member)
		}
	}

	return result
}
