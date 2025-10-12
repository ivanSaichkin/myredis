package storage

import (
	"errors"
	"time"
)

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrKeyExpired    = errors.New("key has expired")
	ErrWrongType     = errors.New("operation against a key holding the wrong kind of value")
	ErrInvalidIndex  = errors.New("index out of range")
	ErrFieldNotFound = errors.New("field not found")
)

type Storage interface {
	Get(key string) (*StorageValue, error)
	Set(key string, value interface{}) error
	SetWithTTL(key string, value interface{}, ttl time.Duration) error
	Delete(key string) bool
	Exists(key string) bool

	// Hash operations
	HSet(key, field, value string) (bool, error)
	HGet(key, field string) (string, error)
	HDel(key, field string) (bool, error)
	HExists(key, field string) (bool, error)
	HGetAll(key string) (map[string]string, error)
	HKeys(key string) ([]string, error)
	HLen(key string) (int, error)

	// List operations
	LPush(key string, values ...string) (int, error)
	RPush(key string, values ...string) (int, error)
	LPop(key string) (string, error)
	RPop(key string) (string, error)
	LLen(key string) (int, error)
	LRange(key string, start, stop int) ([]string, error)

	// Set operations
	SAdd(key string, members ...string) (int, error)
	SRem(key string, members ...string) (int, error)
	SIsMember(key, member string) (bool, error)
	SMembers(key string) ([]string, error)
	SCard(key string) (int, error)
	SInter(keys ...string) ([]string, error)

	// Utility methods
	Keys() []string
	Size() int
	Clear()
	Type(key string) (ValueType, error)

	// TTL operations
	Expire(key string, ttl time.Duration) bool
	TTL(key string) (time.Duration, error)

	// Cleanup
	CleanupExpired()
}
