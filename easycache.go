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
	_ttl time.Time
	sync.RWMutex
}

func (l *l) getTTL() time.Time {
	l.RLock()
	defer l.RUnlock()
	return l._ttl
}

type cache struct {
	list map[any]*l
	sync.RWMutex
	originLock sync.RWMutex
	origin     func(key any) any
	_ttl       time.Duration
}

func (c *cache) getTTL() time.Duration {
	c.RLock()
	defer c.RUnlock()
	return c._ttl
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
	_, ok := c.list[key]
	if ok {
		c.Lock()
		c.list[key]._ttl = time.Now().Add(c.getTTL())
		c.Unlock()

	}
	lc.data = c.origin(key)
	lc._ttl = time.Now().Add(c.getTTL())
	c.Lock()
	defer c.Unlock()
	lc.Lock()
	defer lc.Unlock()
	c.list[key] = &lc

}

func (c *cache) Get(key any) (any, error) {
	c.RLock()
	val, ok := c.list[key]
	c.RUnlock()

	if ok {
		if time.Now().After(val.getTTL()) {
			go c.update(key)
		}
		c.RLock()
		defer c.RUnlock()
		return val.data, nil
	}

	c.update(key)
	val, _ = c.list[key]
	c.RLock()
	defer c.RUnlock()
	return val.data, nil
}
