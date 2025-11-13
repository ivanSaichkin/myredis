package storage

import (
	"ivanSaichkin/myredis/internal/config"
	"os"
	"testing"
	"time"
)

func TestSimplePersistence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "myredis_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &config.PersistenceConfig{
		Enabled:  true,
		DataDir:  tempDir,
		Filename: "test.bin",
		AutoSave: false,
	}

	store := NewMemoryStorageWithPersistence(config)

	store.Set("string_key", "string_value")
	store.HSet("hash_key", "field1", "value1")
	store.LPush("list_key", "item1", "item2")
	store.SAdd("set_key", "member1", "member2")
	store.SetWithTTL("temp_key", "temp_value", time.Minute)

	if err := store.SaveSnapshot(); err != nil {
		t.Fatalf("Failed to save snapshot: %v", err)
	}

	newStore := NewMemoryStorageWithPersistence(config)
	if err := newStore.StartPersistence(); err != nil {
		t.Fatalf("Failed to start persistence: %v", err)
	}

	// String
	val, err := newStore.Get("string_key")
	if err != nil {
		t.Fatalf("Failed to get string key: %v", err)
	}
	if val.Data != "string_value" {
		t.Errorf("String value mismatch: expected 'string_value', got '%v'", val.Data)
	}

	// Hash
	hashVal, err := newStore.HGet("hash_key", "field1")
	if err != nil {
		t.Fatalf("Failed to get hash field: %v", err)
	}
	if hashVal != "value1" {
		t.Errorf("Hash value mismatch: expected 'value1', got '%s'", hashVal)
	}

	// List
	listLen, err := newStore.LLen("list_key")
	if err != nil {
		t.Fatalf("Failed to get list length: %v", err)
	}
	if listLen != 2 {
		t.Errorf("List length mismatch: expected 2, got %d", listLen)
	}

	// Set
	isMember, err := newStore.SIsMember("set_key", "member1")
	if err != nil {
		t.Fatalf("Failed to check set membership: %v", err)
	}
	if !isMember {
		t.Error("Set member not found after restoration")
	}

	// TTL
	ttl, err := newStore.TTL("temp_key")
	if err != nil {
		t.Fatalf("Failed to get TTL: %v", err)
	}
	if ttl <= 0 {
		t.Errorf("TTL not preserved: expected positive value, got %v", ttl)
	}
}

func TestPersistenceAutoSave(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "myredis_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &config.PersistenceConfig{
		Enabled:      true,
		DataDir:      tempDir,
		Filename:     "test.bin",
		AutoSave:     true,
		SaveInterval: 100 * time.Millisecond,
	}

	store := NewMemoryStorageWithPersistence(config)

	if err := store.StartPersistence(); err != nil {
		t.Fatalf("Failed to start persistence: %v", err)
	}

	store.Set("test_key", "test_value")

	time.Sleep(200 * time.Millisecond)

	if err := store.StopPersistence(); err != nil {
		t.Fatalf("Failed to stop persistence: %v", err)
	}

	filePath := config.DataDir + "/" + config.Filename
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Persistence file was not created")
	}
}
