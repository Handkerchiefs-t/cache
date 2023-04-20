package pkg

import (
	"context"
	"sync"
	"time"
)

type MapCache struct {
	lock  sync.RWMutex
	close chan struct{}
	data  map[string]any
}

func (m *MapCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.data[key] = val
	return nil
}

func (m *MapCache) Get(ctx context.Context, key string) (any, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if v, ok := m.data[key]; ok {
		return v, nil
	}
	return nil, ErrKeyNotFound
}

func (m *MapCache) Delete(ctx context.Context, key string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.data, key)
	return nil
}
