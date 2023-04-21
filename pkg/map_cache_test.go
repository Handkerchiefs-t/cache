package pkg

import (
	"context"
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
