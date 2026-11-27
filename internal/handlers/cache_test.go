package handlers

import (
	"context"
	"time"
)

type mockCache struct {
	setFn           func(context.Context, string, any, time.Duration) error
	getFn           func(context.Context, string, any) (bool, error)
	invalidateFn    func(context.Context, string) error
	invalidateAllFn func(context.Context) error
}

func (m mockCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if m.setFn != nil {
		return m.setFn(ctx, key, value, ttl)
	}
	return nil
}

func (m mockCache) Get(ctx context.Context, key string, dest any) (bool, error) {
	if m.getFn != nil {
		return m.getFn(ctx, key, dest)
	}
	return false, nil
}

func (m mockCache) Invalidate(ctx context.Context, key string) error {
	if m.invalidateFn != nil {
		return m.invalidateFn(ctx, key)
	}
	return nil
}

func (m mockCache) InvalidateAll(ctx context.Context) error {
	if m.invalidateAllFn != nil {
		return m.invalidateAllFn(ctx)
	}
	return nil
}

var testOpts = Options{CacheTTL: time.Minute, MaxListSize: 100}
