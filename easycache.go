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
	done chan bool
}

type cache struct {
	list   map[any]*l
	origin func(key any) any
	_ttl   time.Duration
	sync.RWMutex
}

func (c *cache) getTTL() time.Duration {
	c.RLock()
	defer c.RUnlock()
	return c._ttl
}

func New(ttl time.Duration, fn func(key any) any) Cache {
	var c = cache{
		origin: fn,
		_ttl:   ttl,
	}

	c.list = make(map[any]*l)

	go c.runner() // race condition panic

	return &c
}

func (c *cache) runner() {

	cacheTTL := c._ttl
	ticker := time.NewTicker(cacheTTL)
	for range ticker.C {
		count := len(c.list)
		for i := 0; i < count; i++ { // rance condition panic
			v, ok := c.list[i]

			if !ok {
				continue
			}
			v.RLock()
			ttl := v._ttl
			v.RUnlock()
			if ttl.Before(time.Now().Add(-(5 * cacheTTL))) {
				c.Lock()
				delete(c.list, i)
				c.Unlock()
			}
		}
	}
}

func (c *cache) update(key any) {
	chc, ok := c.list[key]
	if !ok {
		chc = &l{}
		c.Lock()
		c.list[key] = chc
		c.Unlock()
	}
	chc.Lock()
	chc.done = make(chan bool)
	chc._ttl = time.Now().Add(c.getTTL())
	chc.Unlock()

	chc.Lock()
	defer chc.Unlock()
	chc.data = c.origin(key)
	chc._ttl = time.Now().Add(c.getTTL())
}

func (c *cache) Get(key any) (any, error) {
	c.RLock()
	val, ok := c.list[key]
	c.RUnlock()

	if ok {
		val.RLock()
		vTTL := val._ttl
		val.RUnlock()
		if time.Now().After(vTTL) {
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
