//go:build e2e

package pkg

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRedisCache_e2e_Set(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	rdb.Ping(context.Background())

	cache := NewRedisCache(rdb)

	testCases := []struct {
		name string

		ctx        func() context.Context
		key        string
		val        string
		expiration time.Duration

		//before func(t *testing.T)
		after func(t *testing.T)

		wantErr error
	}{
		{
			name: "set value",

			ctx: func() context.Context {
				r, _ := context.WithTimeout(context.Background(), time.Second*3)
				return r
			},
			key:        "key1",
			val:        "value1",
			expiration: time.Second * 5,

			//before: func(t *testing.T) {},
			after: func(t *testing.T) {
				val, err := cache.Get(context.Background(), "key1")
				require.NoError(t, err)
				require.Equal(t, "value1", val)
			},

			wantErr: nil,
		},
		{
			name: "timeout",

			// return an expired ctx
			ctx: func() context.Context {
				r, _ := context.WithTimeout(context.Background(), time.Second)
				time.Sleep(time.Second * 2)
				return r
			},
			key:        "key timeout",
			val:        "value timeout",
			expiration: time.Second * 5,

			wantErr: context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//tc.before(t)
			err := cache.Set(tc.ctx(), tc.key, tc.val, tc.expiration)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			tc.after(t)
		})
	}

}

func TestRedisCache_e2e_Get(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	rdb.Ping(context.Background())

	cache := NewRedisCache(rdb)

	testCases := []struct {
		name string

		ctx func() context.Context
		key string

		before func(t *testing.T)
		after  func(t *testing.T)

		wangVal string
		wantErr error
	}{
		{
			name: "get value",

			ctx: func() context.Context { return context.Background() },
			key: "key1",

			before: func(t *testing.T) {
				err := cache.Set(context.Background(), "key1", "value1", time.Second*5)
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				err := cache.Delete(context.Background(), "key1")
				assert.NoError(t, err)
			},

			wangVal: "value1",
			wantErr: nil,
		},
		{
			name: "not found",

			ctx: func() context.Context { return context.Background() },
			key: "key not found",

			before: func(t *testing.T) {},
			after:  func(t *testing.T) {},

			wangVal: "",
			wantErr: redis.Nil,
		},
		{
			name: "ctx timeout",

			// return an expired ctx
			ctx: func() context.Context {
				r, _ := context.WithTimeout(context.Background(), time.Second)
				time.Sleep(time.Second * 2)
				return r
			},
			key: "",

			before: func(t *testing.T) {},
			after:  func(t *testing.T) {},

			wangVal: "",
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			val, err := cache.Get(tc.ctx(), tc.key)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wangVal, val)
			tc.after(t)
		})
	}
}
