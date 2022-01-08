package redis

import (
	"context"
	"github.com/wyy-go/wcache/persist"

	"github.com/go-redis/redis/v8"
	"testing"
	"time"
)

// These tests require redis server running on localhost:6379 (the default)
const redisTestServer = "127.0.0.1:6379"

var newRedisStore = func(t *testing.T, defaultExpiration time.Duration) persist.CacheStore {
	client := redis.NewClient(&redis.Options{
		Addr:     redisTestServer,
		DB:       10,
		Password: "",
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	return NewRedisStore(client)
}

func TestRedisCache_TypicalGetSet(t *testing.T) {
	var err error
	cache := newRedisStore(t, time.Hour)

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

func TestRedisCache_Expiration(t *testing.T) {
	// memcached does not support expiration times less than 1 second.
	var err error
	cache := newRedisStore(t, time.Second)
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
	if err := cache.Set("int", value, 0); err != nil {
		t.Errorf("wrong to set cache, but got: %s", err)
	}
	time.Sleep(2 * time.Second)
	err = cache.Get("int", &value)
	if err != nil {
		t.Errorf("Expected to get the value, but got: %s", err)
	}
}

func TestRedisCache_EmptyCache(t *testing.T) {
	var err error
	cache := newRedisStore(t, time.Hour)

	err = cache.Get("notexist", time.Hour)
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
