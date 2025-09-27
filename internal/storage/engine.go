package storage

import "time"

type Value struct {
	Value     interface{}
	ExpiredAt time.Time
}

type Storage interface {
	// Basic operations
	Get(key string) (*Value, error)
	Set(key string, value interface{}) error
	SetWithTTL(key string, value interface{}, ttl time.Duration) error
	Delete(key string) bool
	Exists(key string) bool
}
