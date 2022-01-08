package memory

import (
	"errors"
	"github.com/wyy-go/wcache/persist"
	"reflect"
	"time"

	"github.com/ReneKroon/ttlcache/v2"
)

// MemoryStore local memory cache store
type MemoryStore struct {
	Cache *ttlcache.Cache
}

// NewMemoryStore allocate a local memory store with default expiration
func NewMemoryStore(defaultExpiration time.Duration) *MemoryStore {
	cacheStore := ttlcache.NewCache()
	_ = cacheStore.SetTTL(defaultExpiration)

	return &MemoryStore{
		Cache: cacheStore,
	}
}

// Set put key value pair to memory store, and expire after expireDuration
func (c *MemoryStore) Set(key string, value interface{}, expireDuration time.Duration) error {
	return c.Cache.SetWithTTL(key, value, expireDuration)
}

// Delete remove key in memory store, do nothing if key doesn't exist
func (c *MemoryStore) Delete(key string) error {
	err := c.Cache.Remove(key)
	if err != nil {
		if errors.Is(err, ttlcache.ErrNotFound) {
			return persist.ErrCacheMiss
		}
		return err
	}

	return c.Cache.Remove(key)
}

// Get get key in memory store, if key doesn't exist, return ErrCacheMiss
func (c *MemoryStore) Get(key string, value interface{}) error {
	val, err := c.Cache.Get(key)
	if err != nil {
		if errors.Is(err, ttlcache.ErrNotFound) {
			return persist.ErrCacheMiss
		}
		return err
	}

	v := reflect.ValueOf(value)
	if v.Type().Kind() == reflect.Ptr && v.Elem().CanSet() {
		v.Elem().Set(reflect.Indirect(reflect.ValueOf(val)))
	}

	return nil
}
