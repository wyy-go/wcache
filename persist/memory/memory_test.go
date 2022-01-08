package memory

import (
	"github.com/wyy-go/wcache/persist"

	"testing"
	"time"
)

var newInMemoryStore = func(t *testing.T, defaultExpiration time.Duration) persist.CacheStore {
	return NewMemoryStore(defaultExpiration)
}

// Test typical cache interactions
func TestInMemoryCache_TypicalGetSet(t *testing.T) {
	var err error
	cache := newInMemoryStore(t, time.Hour)

	value := "foo"
	if err = cache.Set("value", value, time.Hour); err != nil {
		t.Errorf("Error setting a value: %s", err)
	}

	value = ""
	err = cache.Get("value", &value)
	if err != nil {
		t.Errorf("Error getting a value: %s", err)
	}
	if value != "foo" {
		t.Errorf("Expected to get foo back, got %s", value)
	}
}

func TestInMemoryCache_Expiration(t *testing.T) {
	// memcached does not support expiration times less than 1 second.
	var err error
	cache := newInMemoryStore(t, time.Second)
	value := 10

	// Test Set w/ short time
	if err := cache.Set("int", value, time.Second); err != nil {
		t.Errorf("wrong to set cache, but got: %s", err)
	}
	time.Sleep(2 * time.Second)
	err = cache.Get("int", &value)
	if err != persist.ErrCacheMiss {
		t.Errorf("Expected CacheMiss, but got: %s", err)
	}

	// Test Set w/ longer time.
	if err := cache.Set("int", value, time.Hour); err != nil {
		t.Errorf("wrong to set cache, but got: %s", err)
	}
	time.Sleep(2 * time.Second)
	err = cache.Get("int", &value)
	if err != nil {
		t.Errorf("Expected to get the value, but got: %s", err)
	}

	// Test Set w/ forever.
	if err := cache.Set("int", value, -1); err != nil {
		t.Errorf("wrong to set cache, but got: %s", err)
	}
	time.Sleep(2 * time.Second)
	err = cache.Get("int", &value)
	if err != nil {
		t.Errorf("Expected to get the value, but got: %s", err)
	}
}

func TestInMemoryCache_EmptyCache(t *testing.T) {
	var err error
	cache := newInMemoryStore(t, time.Hour)

	err = cache.Get("notexist", 0)
	if err == nil {
		t.Errorf("Error expected for non-existent key")
	}
	if err != persist.ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss for non-existent key: %s", err)
	}

	err = cache.Delete("notexist")
	if err != persist.ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss for non-existent key: %s", err)
	}
}
