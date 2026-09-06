package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	host    string
	db      int
	expires time.Duration
	client  *redis.Client
}

func NewRedisCache(host string, db int, expires time.Duration) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: host,
		DB:   db,
	})

	return &RedisCache{
		host:    host,
		db:      db,
		expires: expires,
		client:  client,
	}
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, r.expires).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) SetBytes(ctx context.Context, key string, data []byte) error {
	return r.client.Set(ctx, key, data, r.expires).Err()
}

func (r *RedisCache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return r.client.Get(ctx, key).Bytes()
}

func (r *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *RedisCache) GetInt(ctx context.Context, key string) (int64, error) {
	value, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return value, err
}
