package cache

import (
	"sync"
	"time"
)

func New(ttl time.Duration, fn func(key any) any) func(key any) (any, error) {
	type l struct {
		ilr          any
		ttl          time.Time
		beingUpdated bool
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
		lc.ilr = fn(key)
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
			return val.ilr, nil
		}

		c.update(key)
		return c.list[key].ilr, nil
	}
}
