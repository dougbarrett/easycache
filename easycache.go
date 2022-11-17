package easycache

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

type Cache interface {
	Get(key any) (any, error)
}

type service struct {
	origin func(key any) any
	cache  *cache.Cache
	ttl    map[any]time.Time
	keyTTL time.Duration
	ttlRW  sync.RWMutex
}

func New(ttl time.Duration, fn func(key any) any) Cache {
	c := cache.New(ttl, ttl*2)

	s := &service{
		origin: fn,
		cache:  c,
	}

	s.ttl = make(map[any]time.Time)
	s.keyTTL = ttl

	go s.runner() // race condition panic

	return s
}

func (c *service) runner() {
	ticker := time.NewTicker(c.keyTTL)
	for range ticker.C {
		c.ttlRW.RLock()
		ttlMap := make(map[any]time.Time)
		for k, v := range c.ttl {
			ttlMap[k] = v
		}
		c.ttlRW.RUnlock()

		for i := range ttlMap {

			ttl := ttlMap[i]

			if ttl.Add(c.keyTTL).Before(time.Now()) {
				c.cache.Delete(fmt.Sprintf("%v", i))
			}
		}
	}
}

func (c *service) update(key any) {
	c.ttlRW.Lock()
	c.ttl[key] = time.Now().Add(c.keyTTL)
	c.ttlRW.Unlock()

	data := c.origin(key)
	c.cache.Set(fmt.Sprintf("%v", key), data, cache.NoExpiration)
}

func (c *service) Get(key any) (any, error) {
	data, ok := c.cache.Get(fmt.Sprintf("%v", key))

	if ok {
		c.ttlRW.Lock()
		vTTL := c.ttl[key]
		c.ttlRW.Unlock()
		if time.Now().After(vTTL) {
			go c.update(key)
		}
		return data, nil
	}

	c.update(key)

	data, ok = c.cache.Get(fmt.Sprintf("%v", key))
	if !ok {
		return nil, errors.New("error fetching cache")
	}
	return data, nil
}
