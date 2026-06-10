package cache

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"seu-oj-backend/internal/observability"
)

type cacheTestValue struct {
	Name string `json:"name"`
}

func TestNilCacheIsNoop(t *testing.T) {
	var c *Cache
	var dst cacheTestValue

	ok, err := c.GetJSON(context.Background(), "key", &dst)
	if err != nil {
		t.Fatalf("get json: %v", err)
	}
	if ok {
		t.Fatal("expected cache miss for nil cache")
	}
	if err := c.SetJSON(context.Background(), "key", cacheTestValue{Name: "value"}, time.Minute); err != nil {
		t.Fatalf("set json: %v", err)
	}
	c.DeletePrefixes(context.Background(), "prefix:")
}

func TestGetOrSetLoadsOnCacheMissAndRecordsTiming(t *testing.T) {
	observer := &observability.TimingObserver{}
	ctx := observability.WithTimingObserver(context.Background(), observer)
	calls := 0

	value, err := GetOrSet(ctx, New(nil), "key", time.Minute, func() (*cacheTestValue, error) {
		calls++
		return &cacheTestValue{Name: "loaded"}, nil
	})
	if err != nil {
		t.Fatalf("get or set: %v", err)
	}
	if value.Name != "loaded" || calls != 1 {
		t.Fatalf("unexpected load result value=%+v calls=%d", value, calls)
	}
	header := observer.HeaderValue()
	for _, segment := range []string{"cache_get", "db", "cache_set"} {
		if !strings.Contains(header, segment) {
			t.Fatalf("expected timing segment %q in %q", segment, header)
		}
	}
}

func TestGetOrSetReturnsLoadError(t *testing.T) {
	wantErr := errors.New("load failed")
	_, err := GetOrSet(context.Background(), New(nil), "key", time.Minute, func() (*cacheTestValue, error) {
		return nil, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected load error, got %v", err)
	}
}

func TestCacheJSONWithRedis(t *testing.T) {
	mr := miniredis.RunT(t)
	c := New(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	ctx := context.Background()

	var miss cacheTestValue
	ok, err := c.GetJSON(ctx, "missing", &miss)
	if err != nil {
		t.Fatalf("get missing: %v", err)
	}
	if ok {
		t.Fatal("expected cache miss")
	}

	if err := c.SetJSON(ctx, "user:1", cacheTestValue{Name: "alice"}, time.Minute); err != nil {
		t.Fatalf("set json: %v", err)
	}

	var hit cacheTestValue
	ok, err = c.GetJSON(ctx, "user:1", &hit)
	if err != nil {
		t.Fatalf("get hit: %v", err)
	}
	if !ok || hit.Name != "alice" {
		t.Fatalf("unexpected cached value ok=%t value=%+v", ok, hit)
	}

	if err := c.SetJSON(ctx, "prefix:a", cacheTestValue{Name: "a"}, time.Minute); err != nil {
		t.Fatalf("set prefix a: %v", err)
	}
	if err := c.SetJSON(ctx, "prefix:b", cacheTestValue{Name: "b"}, time.Minute); err != nil {
		t.Fatalf("set prefix b: %v", err)
	}
	c.DeletePrefixes(ctx, "prefix:")
	ok, err = c.GetJSON(ctx, "prefix:a", &hit)
	if err != nil {
		t.Fatalf("get after delete: %v", err)
	}
	if ok {
		t.Fatal("expected keys under prefix to be deleted")
	}
}

func TestGetOrSetUsesCacheHit(t *testing.T) {
	mr := miniredis.RunT(t)
	c := New(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	ctx := context.Background()
	calls := 0

	first, err := GetOrSet(ctx, c, "cached", time.Minute, func() (*cacheTestValue, error) {
		calls++
		return &cacheTestValue{Name: "loaded"}, nil
	})
	if err != nil || first.Name != "loaded" || calls != 1 {
		t.Fatalf("first load: value=%+v calls=%d err=%v", first, calls, err)
	}

	second, err := GetOrSet(ctx, c, "cached", time.Minute, func() (*cacheTestValue, error) {
		calls++
		return &cacheTestValue{Name: "reloaded"}, nil
	})
	if err != nil || second.Name != "loaded" || calls != 1 {
		t.Fatalf("cache hit: value=%+v calls=%d err=%v", second, calls, err)
	}
}

func TestGetJSONInvalidPayload(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	c := New(client)
	ctx := context.Background()
	if err := client.Set(ctx, "bad", "not-json", time.Minute).Err(); err != nil {
		t.Fatalf("seed bad json: %v", err)
	}
	var dst cacheTestValue
	ok, err := c.GetJSON(ctx, "bad", &dst)
	if err == nil || ok {
		t.Fatalf("expected unmarshal error, ok=%t err=%v", ok, err)
	}
}
