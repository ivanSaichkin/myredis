package storage

import (
	"testing"
	"time"
)

type TestStorage struct {
	NewStorage func() Storage
}

func (ts *TestStorage) RunTests(t *testing.T) {
	t.Run("BasicSetGet", ts.testBasicSetGet)
	t.Run("GetNonExistent", ts.testGetNonExistent)
	t.Run("Delete", ts.testDelete)
	t.Run("Exists", ts.testExists)
	t.Run("SetWithTTL", ts.testSetWithTTL)
	t.Run("Expire", ts.testExpire)
	t.Run("TTL", ts.testTTL)
	t.Run("Keys", ts.testKeys)
	t.Run("Size", ts.testSize)
	t.Run("Clear", ts.testClear)
	t.Run("ConcurrentAccess", ts.testConcurrentAccess)
}

func (ts *TestStorage) testBasicSetGet(t *testing.T) {
	store := ts.NewStorage()
	if err := store.Set("key1", "value1"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := store.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if val.Value != "value1" {
		t.Errorf("Expected 'value1' got '%v'", val.Value)
	}

	err = store.Set("key2", 42)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err = store.Get("key2")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val.Value != 42 {
		t.Errorf("Expected 42, got '%v'", val.Value)
	}
}

func (ts *TestStorage) testGetNonExistent(t *testing.T) {
	store := ts.NewStorage()

	val, err := store.Get("nonexistent")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
	if val != nil {
		t.Errorf("Expected nil value, got %v", val)
	}
}

func (ts *TestStorage) testDelete(t *testing.T) {
	store := ts.NewStorage()

	store.Set("key1", "value1")

	deleted := store.Delete("key1")
	if !deleted {
		t.Error("Delete should return true for existing key")
	}

	_, err := store.Get("key1")
	if err != ErrKeyNotFound {
		t.Error("Key should be deleted")
	}

	deleted = store.Delete("nonexistent")
	if deleted {
		t.Error("Delete should return false for non-existing key")
	}
}

func (ts *TestStorage) testExists(t *testing.T) {
	store := ts.NewStorage()

	store.Set("key1", "value1")

	if !store.Exists("key1") {
		t.Error("Exists should return true for existing key")
	}
	if store.Exists("nonexistent") {
		t.Error("Exists should return false for non-existing key")
	}
}

func (ts *TestStorage) testSetWithTTL(t *testing.T) {
	store := ts.NewStorage()

	err := store.SetWithTTL("temp", "data", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("SetWithTTL failed: %v", err)
	}

	if !store.Exists("temp") {
		t.Error("Key should exist before TTL expires")
	}

	time.Sleep(150 * time.Millisecond)

	if store.Exists("temp") {
		t.Error("Key should not exist after TTL expires")
	}
}

func (ts *TestStorage) testExpire(t *testing.T) {
	store := ts.NewStorage()

	store.Set("key1", "value1")

	success := store.Expire("key1", 100*time.Millisecond)
	if !success {
		t.Error("Expire should return true for existing key")
	}

	success = store.Expire("nonexistent", time.Second)
	if success {
		t.Error("Expire should return false for non-existing key")
	}

	time.Sleep(150 * time.Millisecond)
	if store.Exists("key1") {
		t.Error("Key should be expired after Expire")
	}
}

func (ts *TestStorage) testTTL(t *testing.T) {
	store := ts.NewStorage()

	store.Set("key1", "value1")
	ttl, err := store.TTL("key1")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if ttl != -1 {
		t.Errorf("Expected TTL -1 for key without expiration, got %v", ttl)
	}

	store.SetWithTTL("key2", "value2", time.Second)
	ttl, err = store.TTL("key2")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if ttl <= 0 || ttl > time.Second {
		t.Errorf("TTL should be around 1 second, got %v", ttl)
	}

	_, err = store.TTL("nonexistent")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func (ts *TestStorage) testKeys(t *testing.T) {
	store := ts.NewStorage()

	store.Set("key1", "value1")
	store.Set("key2", "value2")
	store.SetWithTTL("key3", "value3", 50*time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	keys := store.Keys()
	expectedCount := 2

	if len(keys) != expectedCount {
		t.Errorf("Expected %d keys, got %d: %v", expectedCount, len(keys), keys)
	}

	keysMap := make(map[string]bool)
	for _, key := range keys {
		keysMap[key] = true
	}

	if !keysMap["key1"] || !keysMap["key2"] {
		t.Error("Expected keys 'key1' and 'key2' in result")
	}
}

func (ts *TestStorage) testSize(t *testing.T) {
	store := ts.NewStorage()

	if store.Size() != 0 {
		t.Errorf("Initial size should be 0, got %d", store.Size())
	}

	store.Set("key1", "value1")
	store.Set("key2", "value2")
	store.SetWithTTL("key3", "value3", 50*time.Millisecond)

	if store.Size() != 3 {
		t.Errorf("Expected size 3, got %d", store.Size())
	}

	time.Sleep(100 * time.Millisecond)
	if store.Size() != 2 {
		t.Errorf("Expected size 2 after expiration, got %d", store.Size())
	}

	store.Clear()
	if store.Size() != 0 {
		t.Errorf("Size should be 0 after clear, got %d", store.Size())
	}
}

func (ts *TestStorage) testClear(t *testing.T) {
	store := ts.NewStorage()

	store.Set("key1", "value1")
	store.Set("key2", "value2")
	store.Set("key3", "value3")

	store.Clear()

	if store.Size() != 0 {
		t.Error("Storage should be empty after clear")
	}

	if store.Exists("key1") || store.Exists("key2") || store.Exists("key3") {
		t.Error("No keys should exist after clear")
	}
}

func (ts *TestStorage) testConcurrentAccess(t *testing.T) {
	store := ts.NewStorage()
	const numGoroutines = 10
	const operationsPerGoroutine = 100

	done := make(chan bool)

	for i := range numGoroutines {
		go func(id int) {
			for j := range operationsPerGoroutine {
				key := string(rune('A' + id))
				value := id*100 + j

				// Mixed operations
				store.Set(key, value)
				store.Get(key)
				store.Exists(key)
				if j%10 == 0 {
					store.Delete(key)
				}
			}
			done <- true
		}(i)
	}

	for range numGoroutines {
		<-done
	}

	size := store.Size()
	if size < 0 {
		t.Errorf("Invalid size after concurrent access: %d", size)
	}
}
