package pkg

import (
	"context"
	"sync"
	"time"
)

type MapCache struct {
	lock sync.RWMutex

	data  map[string]item
	close chan struct{}

	onEvict func(key string, val any)

	closeOnce sync.Once
}

type MapCacheOption func(c *MapCache)

func OptionWithOnEvict(f func(key string, val any)) MapCacheOption {
	return func(c *MapCache) {
		c.onEvict = f
	}
}

func NewMapCache(expireInterval time.Duration, opts ...MapCacheOption) *MapCache {
	c := &MapCache{
		lock:      sync.RWMutex{},
		closeOnce: sync.Once{},
		close:     make(chan struct{}),
		data:      map[string]item{},
		onEvict:   func(key string, val any) {},
	}

	for _, op := range opts {
		op(c)
	}

	go func() {
		ticker := time.NewTicker(expireInterval)
		for {
			select {
			case t := <-ticker.C:
				c.lock.Lock()
				count := 0
				for k, v := range c.data {
					if count >= 10000 {
						break
					}
					if !v.deadline.IsZero() && v.expired(t) {
						c.delete(k)
					}
					count++
				}
				c.lock.Unlock()
			case <-c.close:
				return
			}
		}
	}()

	return c
}

type item struct {
	value    any
	deadline time.Time
}

func newItem(val any, deadline time.Time) item {
	return item{
		value:    val,
		deadline: deadline,
	}
}

func (m *MapCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.setWithoutLock(ctx, key, val, expiration)
}

func (m *MapCache) setWithoutLock(ctx context.Context, key string, val any, expiration time.Duration) error {
	m.data[key] = newItem(val, time.Now().Add(expiration))
	return nil
}

func (m *MapCache) Get(ctx context.Context, key string) (any, error) {
	m.lock.RLock()
	v, ok := m.data[key]
	m.lock.RUnlock()
	now := time.Now()
	if !ok {
		return nil, ErrKeyNotFound
	}
	if ok && !v.expired(now) {
		return v.value, nil
	}

	// double check
	m.lock.Lock()
	defer m.lock.Unlock()
	v, ok = m.data[key]
	if !ok {
		return nil, ErrKeyNotFound
	}
	if ok && !v.expired(now) {
		return v.value, nil
	}
	// expired, delete it
	m.delete(key)
	return nil, ErrKeyNotFound
}

func (i *item) expired(now time.Time) bool {
	if i.deadline.IsZero() {
		return false
	}
	if now.Before(i.deadline) {
		return false
	}
	return true
}

func (m *MapCache) Delete(ctx context.Context, key string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.delete(key)
	return nil
}

func (m *MapCache) Close() {
	m.closeOnce.Do(func() {
		close(m.close)
	})
}

func (m *MapCache) delete(key string) {
	v, have := m.data[key]
	if !have {
		return
	}

	delete(m.data, key)
	m.onEvict(key, v.value)
}
