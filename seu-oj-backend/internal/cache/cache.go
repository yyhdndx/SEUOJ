package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"seu-oj-backend/internal/observability"
)

type Cache struct {
	client *redis.Client
}

func New(client *redis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) GetJSON(ctx context.Context, key string, dst any) (bool, error) {
	if c == nil || c.client == nil {
		return false, nil
	}

	value, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(value, dst); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Cache) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, payload, ttl).Err()
}

func (c *Cache) DeletePrefixes(ctx context.Context, prefixes ...string) {
	if c == nil || c.client == nil {
		return
	}

	for _, prefix := range prefixes {
		iter := c.client.Scan(ctx, 0, prefix+"*", 100).Iterator()
		for iter.Next(ctx) {
			if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
				log.Printf("[cache] delete key=%s failed: %v", iter.Val(), err)
			}
		}
		if err := iter.Err(); err != nil {
			log.Printf("[cache] scan prefix=%s failed: %v", prefix, err)
		}
	}
}

func GetOrSet[T any](ctx context.Context, c *Cache, key string, ttl time.Duration, load func() (*T, error)) (*T, error) {
	observer := observability.FromContext(ctx)

	cacheGetStart := time.Now()
	var cached T
	if ok, err := c.GetJSON(ctx, key, &cached); err == nil && ok {
		if observer != nil {
			observer.AddDesc("cache_get", time.Since(cacheGetStart), "hit")
		}
		return &cached, nil
	} else if err != nil {
		log.Printf("[cache] get key=%s failed: %v", key, err)
	}
	if observer != nil {
		observer.AddDesc("cache_get", time.Since(cacheGetStart), "miss")
	}

	loadStart := time.Now()
	value, err := load()
	if err != nil {
		return nil, err
	}
	if observer != nil {
		observer.Add("db", time.Since(loadStart))
	}

	cacheSetStart := time.Now()
	if err := c.SetJSON(ctx, key, value, ttl); err != nil {
		log.Printf("[cache] set key=%s failed: %v", key, err)
	}
	if observer != nil {
		observer.Add("cache_set", time.Since(cacheSetStart))
	}
	return value, nil
}
