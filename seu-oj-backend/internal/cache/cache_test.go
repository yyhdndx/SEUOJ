package cache

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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
