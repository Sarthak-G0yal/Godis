package storage

import (
	"errors"
	"strconv"
	"sync"
	"testing"
)

func TestMemoryStorageSetGet(t *testing.T) {
	store := NewMemoryStorage()
	if err := store.Set("name", "mini-redis"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	value, err := store.Get("name")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if value != "mini-redis" {
		t.Fatalf("Expected 'mini-redis', got '%s'", value)
	}
}

func TestMemoryStorageGetNotFound(t *testing.T) {
	store := NewMemoryStorage()
	_, err := store.Get("nonexistent")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestMemoryStorageDelete(t *testing.T) {
	store := NewMemoryStorage()
	if err := store.Set("key", "value"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if err := store.Delete("key"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if store.Exists("key") {
		t.Fatalf("Expected key to be deleted")
	}
}

func TestMemoryStorageEmptyKey(t *testing.T) {
	store := NewMemoryStorage()
	if err := store.Set("", "value"); !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("Expected ErrEmptyKey, got %v", err)
	}
	if _, err := store.Get(""); !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("Expected ErrEmptyKey, got %v", err)
	}
	if err := store.Delete(""); !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("Expected ErrEmptyKey, got %v", err)
	}
	if store.Exists("") {
		t.Fatalf("Expected empty key to not exist")
	}
}
func TestMemoryStoreConcurrentAccess(t *testing.T) {
	store := NewMemoryStorage()
	const writers = 100
	const readers = 100

	var wg sync.WaitGroup

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key" + strconv.Itoa(i)
			value := "value" + strconv.Itoa(i)
			if err := store.Set(key, value); err != nil {
				t.Errorf("Set failed: %v", err)
			}
		}(i)
	}
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key" + strconv.Itoa(i)
			_, _ = store.Get(key)
		}(i)
	}
	wg.Wait()
}
