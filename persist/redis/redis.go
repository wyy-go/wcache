package redis

import (
	"context"
	"github.com/wyy-go/wcache/persist"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisStore store http response in redis
type RedisStore struct {
	RedisClient *redis.Client
}

// NewRedisStore create a redis memory store with redis client
func NewRedisStore(redisClient *redis.Client) *RedisStore {
	return &RedisStore{
		RedisClient: redisClient,
	}
}

// Set put key value pair to redis, and expire after expireDuration
func (store *RedisStore) Set(key string, value interface{}, expire time.Duration) error {
	ctx := context.TODO()
	return store.RedisClient.Set(ctx, key, value, expire).Err()
}

// Delete remove key in redis, do nothing if key doesn't exist
func (store *RedisStore) Delete(key string) error {
	ctx := context.TODO()
	return store.RedisClient.Del(ctx, key).Err()
}

// Get get key in redis, if key doesn't exist, return ErrCacheMiss
func (store *RedisStore) Get(key string, value interface{}) error {
	ctx := context.TODO()
	err := store.RedisClient.Get(ctx, key).Scan(value)
	if err != nil {
		if err == redis.Nil {
			return persist.ErrCacheMiss
		}
		return err
	}
	return nil
}
