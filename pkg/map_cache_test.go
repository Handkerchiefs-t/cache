package pkg

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMapCache_Get(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		cache   func() *MapCache
		wantVal any
		wantErr error
	}{
		{
			name: "no key",
			key:  "not exist key",
			cache: func() *MapCache {
				res := NewMapCache(time.Second * 10)
				return res
			},
			wantVal: nil,
			wantErr: ErrKeyNotFound,
		},
		{
			name: "exit",
			key:  "exit key",
			cache: func() *MapCache {
				res := NewMapCache(time.Second * 10)
				e := res.Set(context.Background(), "exit key", 123, time.Second)
				require.NoError(t, e)
				return res
			},
			wantVal: 123,
			wantErr: nil,
		},
		{
			name: "expired",
			key:  "expired key",
			cache: func() *MapCache {
				res := NewMapCache(time.Second * 10)
				e := res.Set(context.Background(), "expired key", 123, time.Second)
				require.NoError(t, e)
				time.Sleep(time.Second * 2)
				return res
			},
			wantVal: nil,
			wantErr: ErrKeyNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.cache()
			v, e := c.Get(context.Background(), tt.key)
			assert.Equal(t, tt.wantErr, e)
			if e != nil {
				return
			}
			assert.Equal(t, tt.wantVal, v)
		})
	}
}

func TestMapCache_Loop(t *testing.T) {
	cnt := 0
	c := NewMapCache(time.Second, OptionWithOnEvict(func(key string, val any) {
		cnt++
	}))
	e := c.Set(context.Background(), "1", 1, time.Second)
	require.NoError(t, e)
	time.Sleep(time.Second * 3)
	c.lock.RLock()
	defer c.lock.RUnlock()
	_, ok := c.data["1"]
	require.False(t, ok)
	require.Equal(t, 1, cnt)
}
