package easycache

import (
	"sync"
	"time"
)

func New(ttl time.Duration, fn func(key any) any) func(key any) (any, error) {
	type l struct {
		list any
		ttl  time.Time
		sync.RWMutex
	}
	type cache struct {
		list   map[any]*l
		update func(key any)
		sync.RWMutex
	}
	var c = cache{}

	c.list = make(map[any]*l)

	c.update = func(key any) {
		var lc l
		lcp, ok := c.list[key]
		if ok {
			lcp.Lock()
			lcp.ttl = time.Now().Add(ttl)
			lcp.Unlock()
		}
		lc.list = fn(key)
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
			return val.list, nil
		}

		c.update(key)
		return c.list[key].list, nil
	}
}
