package easycache

import (
	"sync"
	"time"
)

func New(ttl time.Duration, fn func(key any) any) func(key any) (any, error) {
	type l struct {
		data any
		ttl  time.Time
		sync.RWMutex
	}
	type cache struct {
		list   map[any]*l
		update func(key any)
		sync.RWMutex
		originLock sync.RWMutex
	}
	var c = cache{}

	c.list = make(map[any]*l)

	c.update = func(key any) {
		var lc l
		lcp, ok := c.list[key]
		if ok {
			lcp.Lock()
			defer lcp.Unlock()
			lcp.ttl = time.Now().Add(ttl)
		}
		c.originLock.Lock()
		lc.data = fn(key)
		c.originLock.Unlock()
		c.Lock()
		lc.ttl = time.Now().Add(ttl)
		c.list[key] = &lc
		c.Unlock()
	}

	return func(key any) (any, error) {
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
}
