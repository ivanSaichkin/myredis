package storage

import "time"

type StorageSnapshot struct {
	Timestamp time.Time      `json:"timestamp"`
	KeyCount  int            `json:"key_count"`
	Entries   []StorageEntry `json:"entries"`
}

type StorageEntry struct {
	Key       string    `json:"key"`
	Type      ValueType `json:"type"`
	Data      any       `json:"data"`
	ExpiredAt time.Time `json:"expired_at,omitempty"`
}
