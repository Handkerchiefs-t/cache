package pkg

import (
	"context"
	"sync/atomic"
	"time"
)

type MaxCountCache struct {
	*MapCache
	maxCount uint64
	count    uint64
}

func NewMaxCountCache(c *MapCache, max uint64) *MaxCountCache {
	res := &MaxCountCache{
		MapCache: c,
		maxCount: max,
	}

	origin := c.onEvict
	res.onEvict = func(key string, val any) {
		atomic.AddUint64(&res.count, -1)
		if origin != nil {
			origin(key, val)
		}
	}

	return res
}

func (c *MaxCountCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	_, ok := c.MapCache.data[key]
	if !ok {
		if c.count+1 > c.maxCount {
			return ErrOverCapacity
		}
		c.count++
	}

	return c.MapCache.setWithoutLock(ctx, key, val, expiration)
}
