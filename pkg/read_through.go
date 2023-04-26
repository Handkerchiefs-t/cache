package pkg

import (
	"context"
	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"
	"time"
)

type ReadThroughCache struct {
	Cache

	LoadFunc   func(ctx context.Context, key string) (any, error)
	Expiration time.Duration

	LogFunc func(key string, err error)

	sf *singleflight.Group
}

func NewReadThroughCache(
	c Cache,
	loadFunc func(ctx context.Context, key string) (any, error),
	logFunc func(key string, err error),
	expiration time.Duration) *ReadThroughCache {
	return &ReadThroughCache{
		Cache:      c,
		LoadFunc:   loadFunc,
		LogFunc:    logFunc,
		Expiration: expiration,
		sf:         &singleflight.Group{},
	}
}

func (r *ReadThroughCache) Get(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if errors.Is(err, ErrKeyNotFound) {
		val, err = r.LoadFunc(ctx, key)
		if err == nil {
			if er := r.Set(ctx, key, val, r.Expiration); er != nil {
				r.LogFunc(key, er)
			}
		}
	}
	return val, err
}

func (r *ReadThroughCache) GetWithAsync(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if errors.Is(err, ErrKeyNotFound) {
		val, err = r.LoadFunc(ctx, key)
		if err == nil {
			// asynchronous cache refresh
			go func() {
				if er := r.Set(ctx, key, val, r.Expiration); er != nil {
					r.LogFunc(key, er)
				}
			}()
		}
	}
	return val, err
}

func (r *ReadThroughCache) GetWithSingleFlight(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if errors.Is(err, ErrKeyNotFound) {
		val, err, _ = r.sf.Do(key, func() (interface{}, error) {
			v, e := r.LoadFunc(ctx, key)
			if e == nil {
				if er := r.Set(ctx, key, v, r.Expiration); er != nil {
					r.LogFunc(key, er)
				}
			}
			return v, e
		})
	}
	return val, err
}
