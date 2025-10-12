package storage

import (
	"slices"
	"sync"
	"testing"
)

func TestHashOperations(t *testing.T) {
	store := NewMemoryStorage()

	// Test HSET and HGET
	created, err := store.HSet("user:1", "name", "Alice")
	if err != nil {
		t.Fatalf("HSET failed: %v", err)
	}
	if !created {
		t.Error("HSET should return true for new field")
	}

	value, err := store.HGet("user:1", "name")
	if err != nil {
		t.Fatalf("HGET failed: %v", err)
	}
	if value != "Alice" {
		t.Errorf("Expected 'Alice', got '%s'", value)
	}

	// Test updating existing field
	created, err = store.HSet("user:1", "name", "Bob")
	if err != nil {
		t.Fatalf("HSET failed: %v", err)
	}
	if created {
		t.Error("HSET should return false for existing field")
	}

	value, err = store.HGet("user:1", "name")
	if err != nil {
		t.Fatalf("HGET failed: %v", err)
	}
	if value != "Bob" {
		t.Errorf("Expected 'Bob', got '%s'", value)
	}

	// Test HGET non-existing field
	_, err = store.HGet("user:1", "age")
	if err != ErrFieldNotFound {
		t.Errorf("Expected ErrFieldNotFound, got %v", err)
	}

	// Test HGET non-existing key
	_, err = store.HGet("user:999", "name")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestHashMultipleFields(t *testing.T) {
	store := NewMemoryStorage()

	// Set multiple fields
	store.HSet("user:1", "name", "Alice")
	store.HSet("user:1", "age", "30")
	store.HSet("user:1", "city", "New York")

	// Test HGETALL
	fields, err := store.HGetAll("user:1")
	if err != nil {
		t.Fatalf("HGETALL failed: %v", err)
	}

	expected := map[string]string{
		"name": "Alice",
		"age":  "30",
		"city": "New York",
	}

	if len(fields) != len(expected) {
		t.Errorf("Expected %d fields, got %d", len(expected), len(fields))
	}

	for key, expectedValue := range expected {
		if fields[key] != expectedValue {
			t.Errorf("Field %s: expected '%s', got '%s'", key, expectedValue, fields[key])
		}
	}

	// Test HKEYS
	keys, err := store.HKeys("user:1")
	if err != nil {
		t.Fatalf("HKEYS failed: %v", err)
	}

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Test HLEN
	length, err := store.HLen("user:1")
	if err != nil {
		t.Fatalf("HLEN failed: %v", err)
	}
	if length != 3 {
		t.Errorf("Expected length 3, got %d", length)
	}

	// Test HEXISTS
	exists, err := store.HExists("user:1", "name")
	if err != nil {
		t.Fatalf("HEXISTS failed: %v", err)
	}
	if !exists {
		t.Error("HEXISTS should return true for existing field")
	}

	exists, err = store.HExists("user:1", "country")
	if err != nil {
		t.Fatalf("HEXISTS failed: %v", err)
	}
	if exists {
		t.Error("HEXISTS should return false for non-existing field")
	}

	// Test HDEL
	deleted, err := store.HDel("user:1", "age")
	if err != nil {
		t.Fatalf("HDEL failed: %v", err)
	}
	if !deleted {
		t.Error("HDEL should return true for existing field")
	}

	length, err = store.HLen("user:1")
	if err != nil {
		t.Fatalf("HLEN failed: %v", err)
	}
	if length != 2 {
		t.Errorf("Expected length 2 after deletion, got %d", length)
	}
}

func TestListOperations(t *testing.T) {
	store := NewMemoryStorage()

	// Test LPUSH
	length, err := store.LPush("mylist", "world")
	if err != nil {
		t.Fatalf("LPUSH failed: %v", err)
	}
	if length != 1 {
		t.Errorf("Expected length 1, got %d", length)
	}

	length, err = store.LPush("mylist", "hello")
	if err != nil {
		t.Fatalf("LPUSH failed: %v", err)
	}
	if length != 2 {
		t.Errorf("Expected length 2, got %d", length)
	}

	// Test LPOP
	value, err := store.LPop("mylist")
	if err != nil {
		t.Fatalf("LPOP failed: %v", err)
	}
	if value != "hello" {
		t.Errorf("Expected 'hello', got '%s'", value)
	}

	// Test RPUSH
	length, err = store.RPush("mylist", "!")
	if err != nil {
		t.Fatalf("RPUSH failed: %v", err)
	}
	if length != 2 {
		t.Errorf("Expected length 2, got %d", length)
	}

	// Test RPOP
	value, err = store.RPop("mylist")
	if err != nil {
		t.Fatalf("RPOP failed: %v", err)
	}
	if value != "!" {
		t.Errorf("Expected '!', got '%s'", value)
	}

	// Test LLEN
	length, err = store.LLen("mylist")
	if err != nil {
		t.Fatalf("LLEN failed: %v", err)
	}
	if length != 1 {
		t.Errorf("Expected length 1, got %d", length)
	}

	// Test LRANGE
	elements, err := store.LRange("mylist", 0, -1)
	if err != nil {
		t.Fatalf("LRANGE failed: %v", err)
	}
	if len(elements) != 1 || elements[0] != "world" {
		t.Errorf("Expected ['world'], got %v", elements)
	}
}

func TestSetOperations(t *testing.T) {
	store := NewMemoryStorage()

	// Test SADD
	added, err := store.SAdd("myset", "apple", "banana")
	if err != nil {
		t.Fatalf("SADD failed: %v", err)
	}
	if added != 2 {
		t.Errorf("Expected 2 elements added, got %d", added)
	}

	// Test duplicate addition
	added, err = store.SAdd("myset", "apple")
	if err != nil {
		t.Fatalf("SADD failed: %v", err)
	}
	if added != 0 {
		t.Errorf("Expected 0 elements added for duplicate, got %d", added)
	}

	// Test SISMEMBER
	isMember, err := store.SIsMember("myset", "apple")
	if err != nil {
		t.Fatalf("SISMEMBER failed: %v", err)
	}
	if !isMember {
		t.Error("SISMEMBER should return true for existing member")
	}

	isMember, err = store.SIsMember("myset", "orange")
	if err != nil {
		t.Fatalf("SISMEMBER failed: %v", err)
	}
	if isMember {
		t.Error("SISMEMBER should return false for non-existing member")
	}

	// Test SMEMBERS
	members, err := store.SMembers("myset")
	if err != nil {
		t.Fatalf("SMEMBERS failed: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}

	// Test SCARD
	count, err := store.SCard("myset")
	if err != nil {
		t.Fatalf("SCARD failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Test SREM
	removed, err := store.SRem("myset", "apple")
	if err != nil {
		t.Fatalf("SREM failed: %v", err)
	}
	if removed != 1 {
		t.Errorf("Expected 1 element removed, got %d", removed)
	}

	count, err = store.SCard("myset")
	if err != nil {
		t.Fatalf("SCARD failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1 after removal, got %d", count)
	}
}

func TestTypeChecking(t *testing.T) {
	store := NewMemoryStorage()

	// Create a string key
	store.Set("mystring", "hello")

	// Try to perform hash operation on string key
	_, err := store.HSet("mystring", "field", "value")
	if err != ErrWrongType {
		t.Errorf("Expected ErrWrongType, got %v", err)
	}

	// Create a hash key
	store.HSet("myhash", "field", "value")

	// Try to perform string operation on hash key
	_, err = store.Get("myhash")
	if err != ErrWrongType {
		t.Errorf("Expected ErrWrongType, got %v", err)
	}

	// Test TYPE command
	keyType, err := store.Type("mystring")
	if err != nil {
		t.Fatalf("TYPE failed: %v", err)
	}
	if keyType != StringType {
		t.Errorf("Expected StringType, got %v", keyType)
	}

	keyType, err = store.Type("myhash")
	if err != nil {
		t.Fatalf("TYPE failed: %v", err)
	}
	if keyType != HashType {
		t.Errorf("Expected HashType, got %v", keyType)
	}
}

func TestConcurrentComplexOperations(t *testing.T) {
	store := NewMemoryStorage()
	var wg sync.WaitGroup

	// Concurrent hash operations
	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				store.HSet("concurrent:hash", string(rune('A'+id)), "value")
				store.HGet("concurrent:hash", string(rune('A'+id)))
			}
		}(i)
	}

	// Concurrent list operations
	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for range 10 {
				store.LPush("concurrent:list", string(rune('A'+id)))
				store.LPop("concurrent:list")
			}
		}(i)
	}

	// Concurrent set operations
	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				store.SAdd("concurrent:set", string(rune('A'+id)))
				store.SIsMember("concurrent:set", string(rune('A'+id)))
			}
		}(i)
	}

	wg.Wait()

	// Verify final state
	hashLen, _ := store.HLen("concurrent:hash")
	if hashLen < 0 {
		t.Errorf("Invalid hash length: %d", hashLen)
	}

	listLen, _ := store.LLen("concurrent:list")
	if listLen < 0 {
		t.Errorf("Invalid list length: %d", listLen)
	}

	setCard, _ := store.SCard("concurrent:set")
	if setCard < 0 {
		t.Errorf("Invalid set cardinality: %d", setCard)
	}
}

func TestSetIntersection(t *testing.T) {
	store := NewMemoryStorage()

	// Create sets
	store.SAdd("set1", "a", "b", "c", "d")
	store.SAdd("set2", "c", "d", "e", "f")
	store.SAdd("set3", "d", "e", "f", "g")

	// Test intersection of two sets
	intersection, err := store.SInter("set1", "set2")
	if err != nil {
		t.Fatalf("SINTER failed: %v", err)
	}

	expected := []string{"c", "d"}
	if len(intersection) != len(expected) {
		t.Errorf("Expected %d elements in intersection, got %d", len(expected), len(intersection))
	}

	for _, item := range expected {
		found := slices.Contains(intersection, item)
		if !found {
			t.Errorf("Expected element %s not found in intersection", item)
		}
	}

	// Test intersection of three sets
	intersection, err = store.SInter("set1", "set2", "set3")
	if err != nil {
		t.Fatalf("SINTER failed: %v", err)
	}

	if len(intersection) != 1 || intersection[0] != "d" {
		t.Errorf("Expected ['d'], got %v", intersection)
	}

	// Test intersection with non-existing set
	intersection, err = store.SInter("set1", "nonexistent")
	if err != nil {
		t.Fatalf("SINTER failed: %v", err)
	}
	if len(intersection) != 0 {
		t.Errorf("Expected empty intersection with non-existing set, got %v", intersection)
	}
}
