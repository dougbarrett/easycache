package easycache

import (
	"sync"
	"time"
)

type Cache interface {
	Get(key any) (any, error)
}

type l struct {
	data any
	ttl  time.Time
	sync.RWMutex
}
type cache struct {
	list map[any]*l
	sync.RWMutex
	originLock sync.RWMutex
	origin     func(key any) any
	ttl        time.Duration
}

func New(ttl time.Duration, fn func(key any) any) Cache {
	var c = cache{
		origin: fn,
	}

	c.list = make(map[any]*l)

	return &c
}

func (c *cache) update(key any) {
	c.originLock.Lock()
	defer c.originLock.Unlock()
	var lc l
	lcp, ok := c.list[key]
	if ok {
		lcp.Lock()
		defer lcp.Unlock()
		lcp.ttl = time.Now().Add(c.ttl)
	}
	lc.data = c.origin(key)
	c.Lock()
	lc.ttl = time.Now().Add(c.ttl)
	c.list[key] = &lc
	c.Unlock()
}

func (c *cache) Get(key any) (any, error) {
	val, ok := c.list[key]

	if ok {
		if time.Now().After(val.ttl) {
			go c.update(key)
		}
		val.RLock()
		defer val.RUnlock()
		return val.data, nil
	}

	c.update(key)
	val, _ = c.list[key]
	val.RLock()
	defer val.RUnlock()
	return val.data, nil
}
