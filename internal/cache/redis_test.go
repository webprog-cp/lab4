package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

type cachedValue struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func newTestRedis(t *testing.T, prefix string) (*Redis, *miniredis.Miniredis) {
	t.Helper()

	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	return NewRedis(client, prefix), server
}

func TestRedisSetGetAndNamespace(t *testing.T) {
	cache, server := newTestRedis(t, "cat")
	ctx := context.Background()
	want := cachedValue{ID: 1, Name: "Murka"}

	if err := cache.Set(ctx, "1", want, time.Minute); err != nil {
		t.Fatal(err)
	}
	if !server.Exists("cat:1") {
		t.Fatal("expected namespaced key cat:1")
	}

	var got cachedValue
	found, err := cache.Get(ctx, "1", &got)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("expected cache hit")
	}
	if got != want {
		t.Fatalf("value = %#v, want %#v", got, want)
	}
}

func TestRedisGetMissAndExpiration(t *testing.T) {
	cache, server := newTestRedis(t, "cat")
	ctx := context.Background()

	var got cachedValue
	found, err := cache.Get(ctx, "missing", &got)
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("missing key reported as cache hit")
	}

	if err := cache.Set(ctx, "1", cachedValue{ID: 1}, time.Minute); err != nil {
		t.Fatal(err)
	}
	server.FastForward(time.Minute)

	found, err = cache.Get(ctx, "1", &got)
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("expired key reported as cache hit")
	}
}

func TestRedisInvalidate(t *testing.T) {
	cache, _ := newTestRedis(t, "shelter")
	ctx := context.Background()

	if err := cache.Set(ctx, "1", cachedValue{ID: 1}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := cache.Invalidate(ctx, "1"); err != nil {
		t.Fatal(err)
	}

	var got cachedValue
	found, err := cache.Get(ctx, "1", &got)
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("invalidated key reported as cache hit")
	}
}

func TestRedisInvalidateAllOnlyRemovesNamespace(t *testing.T) {
	catCache, server := newTestRedis(t, "cat")
	shelterClient := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = shelterClient.Close() })
	shelterCache := NewRedis(shelterClient, "shelter")
	ctx := context.Background()

	if err := catCache.Set(ctx, "1", cachedValue{ID: 1}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := catCache.Set(ctx, "2", cachedValue{ID: 2}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := shelterCache.Set(ctx, "1", cachedValue{ID: 1}, time.Minute); err != nil {
		t.Fatal(err)
	}

	if err := catCache.InvalidateAll(ctx); err != nil {
		t.Fatal(err)
	}
	if server.Exists("cat:1") || server.Exists("cat:2") {
		t.Fatal("cat cache keys were not invalidated")
	}
	if !server.Exists("shelter:1") {
		t.Fatal("another cache namespace was invalidated")
	}
}

func TestRedisRejectsInvalidValues(t *testing.T) {
	cache, server := newTestRedis(t, "cat")
	ctx := context.Background()

	if err := cache.Set(ctx, "bad", make(chan int), time.Minute); err == nil {
		t.Fatal("expected marshal error")
	}

	server.Set("cat:bad", "not-json")
	var got cachedValue
	if _, err := cache.Get(ctx, "bad", &got); err == nil {
		t.Fatal("expected unmarshal error")
	}
}
