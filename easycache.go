package easycache

import (
	"errors"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

type Cache interface {
	Get(key string, req any) (any, error)
}

type service struct {
	origin func(key any) any
	cache  *cache.Cache
	ttl    map[string]time.Time
	keyTTL time.Duration
	ttlRW  sync.RWMutex
}

func New(ttl time.Duration, fn func(key any) any) Cache {
	c := cache.New(ttl, ttl*2)

	s := &service{
		origin: fn,
		cache:  c,
	}

	s.ttl = make(map[string]time.Time)
	s.keyTTL = ttl

	go s.runner() // race condition panic

	return s
}

func (c *service) runner() {
	ticker := time.NewTicker(c.keyTTL)
	for range ticker.C {
		c.ttlRW.RLock()
		ttlMap := make(map[string]time.Time)
		for k, v := range c.ttl {
			ttlMap[k] = v
		}
		c.ttlRW.RUnlock()

		for i := range ttlMap {

			ttl := ttlMap[i]

			if ttl.Add(c.keyTTL).Before(time.Now()) {
				c.cache.Delete(i)
				c.ttlRW.Lock()
				delete(c.ttl, i)
				c.ttlRW.Unlock()
			}
		}
	}
}

func (c *service) update(key string, data any) {
	c.ttlRW.Lock()
	c.ttl[key] = time.Now().Add(c.keyTTL)
	c.ttlRW.Unlock()

	retData := c.origin(data)
	c.cache.Set(key, retData, cache.NoExpiration)
}

func (c *service) Get(key string, data any) (any, error) {
	cacheData, ok := c.cache.Get(key)

	if ok {
		c.ttlRW.Lock()
		vTTL, ttlOK := c.ttl[key]
		if ttlOK {
			c.ttlRW.Unlock()
			if time.Now().After(vTTL) {
				go c.update(key, data)
			}
		}

		return cacheData, nil
	}

	c.update(key, data)

	cacheData, ok = c.cache.Get(key)
	if !ok {
		return nil, errors.New("error fetching cache")
	}
	return cacheData, nil
}
