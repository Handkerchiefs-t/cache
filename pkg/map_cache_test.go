package pkg

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMapCache(t *testing.T) {
	cache := NewMapCache(time.Second)
	emptyCtx := context.Background()
	testCases := []struct {
		k      string
		v      string
		expire time.Duration
	}{
		{k: "1", v: "1", expire: time.Second},
		{k: "2", v: "2", expire: time.Second},
		{k: "3", v: "3", expire: time.Second},
		{k: "4", v: "4", expire: time.Second},
		{k: "5", v: "5", expire: time.Second},
	}

	for _, c := range testCases {
		e := cache.Set(emptyCtx, c.k, c.v, c.expire)
		if e != nil {
			t.Fatal(e)
		}
	}

	for _, c := range testCases {
		v, e := cache.Get(emptyCtx, c.k)
		if e != nil {
			t.Fatal(e)
		}
		t.Logf("v: %s", v.(string))
	}

	time.Sleep(time.Second)
	for _, c := range testCases {
		v, e := cache.Get(emptyCtx, c.k)
		t.Logf("v: %v, e: %s", v, e.Error())
	}
}

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
