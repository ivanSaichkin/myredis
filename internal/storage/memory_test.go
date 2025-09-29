package storage

import (
	"testing"
	"time"
)

// Run all common tests for MemoryStorage
func TestMemoryStorage(t *testing.T) {
	testStorage := &TestStorage{
		NewStorage: func() Storage {
			return NewMemoryStorage()
		},
	}

	testStorage.RunTests(t)
}

// Tests specific for MemoryStorage implementation
func TestMemoryStorageSpecific(t *testing.T) {
	t.Run("Value is expired", testValueIsExpired)
	t.Run("Expiration Checker", testExpirationChecker)
	t.Run("Lazy Expiration", testLazyExpiration)
	t.Run("Multiple types", testMultipleTypes)
}

func testValueIsExpired(t *testing.T) {
	value := &Value{
		Value:     "test",
		ExpiredAt: time.Now().Add(time.Hour),
	}
	if value.IsExpired() {
		t.Error("Value should not be expired")
	}

	value = &Value{
		Value:     "test",
		ExpiredAt: time.Now().Add(-time.Hour),
	}
	if !value.IsExpired() {
		t.Error("Value should be expired")
	}

	value = &Value{
		Value: "test",
	}
	if value.IsExpired() {
		t.Error("Value without expiration should not be expired")
	}
}

func testExpirationChecker(t *testing.T) {
	store := NewMemoryStorage()

	// Add keys with different TTLs
	store.SetWithTTL("short", "data", 50*time.Millisecond)
	store.SetWithTTL("long", "data", time.Hour)
	store.Set("no_ttl", "data")

	store.StartExpirationChecker(10 * time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	// Verify short key is removed, others remain
	if store.Exists("short") {
		t.Error("Short-lived key should be removed by expiration checker")
	}
	if !store.Exists("long") {
		t.Error("Long-lived key should still exist")
	}
	if !store.Exists("no_ttl") {
		t.Error("Key without TTL should still exist")
	}
}

func testLazyExpiration(t *testing.T) {
	store := NewMemoryStorage()

	memStore := store
	memStore.mu.Lock()
	memStore.data["expired"] = &Value{
		Value:     "old_data",
		ExpiredAt: time.Now().Add(-time.Hour),
	}
	memStore.mu.Unlock()

	_, err := store.Get("expired")
	if err != ErrKeyExpired {
		t.Errorf("Expected ErrKeyExpired, got %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	if store.Exists("expired") {
		t.Error("Expired key should be removed by lazy expiration")
	}
}

func testMultipleTypes(t *testing.T) {
	store := NewMemoryStorage()

	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"string", "str_key", "hello world"},
		{"int", "int_key", 42},
		{"float", "float_key", 3.14159},
		{"bool", "bool_key", true},
		{"slice", "slice_key", []string{"a", "b", "c"}},
		{"map", "map_key", map[string]int{"one": 1, "two": 2}},
		{"nil", "nil_key", nil},
	}

	for _, tt := range tests {
		err := store.Set(tt.key, tt.value)
		if err != nil {
			t.Errorf("Set %s failed: %v", tt.name, err)
		}
	}

	for _, tt := range tests {
		_, err := store.Get(tt.key)
		if err != nil {
			t.Errorf("Get %s failed: %v", tt.name, err)
		}
	}
}
