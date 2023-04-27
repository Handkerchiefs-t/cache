package pkg

import (
	"context"
	"time"
)

type WriteThroughCache struct {
	cache     Cache
	StoreFunc func(ctx context.Context, key string, val any) error
	LogFunc   func(key string, err error)
}

func (w *WriteThroughCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	if err := w.StoreFunc(ctx, key, val); err != nil {
		return err
	}
	return w.cache.Set(ctx, key, val, expiration)
}

func (w *WriteThroughCache) SetWithAsync(ctx context.Context, key string, val any, expiration time.Duration) error {
	if err := w.StoreFunc(ctx, key, val); err != nil {
		return err
	}

	// asynchronous cache refresh
	go func() {
		er := w.cache.Set(ctx, key, val, expiration)
		if er != nil {
			w.LogFunc(key, er)
		}
	}()

	return nil
}
