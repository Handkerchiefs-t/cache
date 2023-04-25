package pkg

import (
	"context"
	"github.com/Handkerchiefs-t/cache/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRedisCache_Set(t *testing.T) {
	testcases := []struct {
		name string

		mock       func(ctrl *gomock.Controller) redis.Cmdable
		key        string
		val        string
		expiration time.Duration

		wantValue string
		wantErr   error
	}{
		{
			name: "set value",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				resCmd := redis.NewStatusCmd(context.Background())
				resCmd.SetVal("OK")
				cmd := mocks.NewMockCmdable(ctrl)
				cmd.EXPECT().Set(context.Background(), "key1", "value1", time.Second).Return(resCmd)
				return cmd
			},
			key:        "key1",
			val:        "value1",
			expiration: time.Second,
			wantErr:    nil,
		},
		{
			name: "timeout",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				resCmd := redis.NewStatusCmd(context.Background())
				resCmd.SetErr(context.DeadlineExceeded)
				cmd := mocks.NewMockCmdable(ctrl)
				cmd.EXPECT().Set(context.Background(), "key timeout", "value timeout", time.Second).Return(resCmd)
				return cmd
			},
			key:        "key timeout",
			val:        "value timeout",
			expiration: time.Second,
			wantErr:    context.DeadlineExceeded,
		},
		{
			name: "unexpected msg",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				resCmd := redis.NewStatusCmd(context.Background())
				resCmd.SetVal("NOT OK")
				cmd := mocks.NewMockCmdable(ctrl)
				cmd.EXPECT().Set(context.Background(), "key", "value", time.Second).Return(resCmd)
				return cmd
			},
			key:        "key",
			val:        "value",
			expiration: time.Second,
			wantErr:    ErrSetFailed,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := NewRedisCache(tc.mock(ctrl))
			err := c.Set(context.Background(), tc.key, tc.val, tc.expiration)
			if err != nil {
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
		})
	}
}

func TestRedisCache_Get(t *testing.T) {
	testcases := []struct {
		name string

		mock func(ctrl *gomock.Controller) redis.Cmdable
		key  string

		wantVal string
		wantErr error
	}{
		{
			name: "get value",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				redisRes := redis.NewStringResult("value1", nil)
				cmd := mocks.NewMockCmdable(ctrl)
				cmd.EXPECT().Get(context.Background(), "key1").Return(redisRes)
				return cmd
			},
			key:     "key1",
			wantVal: "value1",
			wantErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := NewRedisCache(tc.mock(ctrl))
			val, err := c.Get(context.Background(), tc.key)
			if err != nil {
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
			assert.Equal(t, tc.wantVal, val)
		})
	}
}
