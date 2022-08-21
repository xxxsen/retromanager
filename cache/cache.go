package cache

import (
	"context"
	"time"

	lru "github.com/hnlq715/golang-lru"
)

type Cache struct {
	c *lru.Cache
}

var cacheInst *Cache

func init() {
	cacheInst, _ = New(0)
}

func Default() *Cache {
	return cacheInst
}

func New(size int) (*Cache, error) {
	if size <= 0 {
		size = 10000
	}
	c, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	return &Cache{
		c: c,
	}, nil
}

func (c *Cache) Set(ctx context.Context, key interface{}, val interface{}, timeout time.Duration) error {
	_ = c.c.AddEx(key, val, timeout)
	return nil
}

func (c *Cache) Get(ctx context.Context, key interface{}) (interface{}, bool, error) {
	val, ok := c.c.Get(key)
	return val, ok, nil
}
