package cache

import (
	"strconv"
	"sync"
	"time"
)

type item struct {
	val       string
	count     int64
	expiresAt time.Time
}

type MemoryCache struct {
	mu       sync.RWMutex
	items    map[string]item
	stopCh   chan struct{}
	stopOnce sync.Once
}

func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{
		items:  make(map[string]item),
		stopCh: make(chan struct{}),
	}
	go c.janitor()
	return c
}

func (c *MemoryCache) janitor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			now := time.Now()
			c.mu.Lock()
			for k, it := range c.items {
				if !now.Before(it.expiresAt) {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		}
	}
}

// Stop terminates the background janitor goroutine.
func (c *MemoryCache) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
	})
}

func (c *MemoryCache) Get(key string) (string, bool) {
	c.mu.RLock()
	it, ok := c.items[key]
	c.mu.RUnlock()

	if !ok || !time.Now().Before(it.expiresAt) {
		return "", false
	}

	if it.val == "" && it.count > 0 {
		return strconv.FormatInt(it.count, 10), true
	}

	return it.val, true
}

func (c *MemoryCache) Set(key string, val string, ttl time.Duration) {
	var count int64
	if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
		count = parsed
	}

	c.mu.Lock()
	c.items[key] = item{
		val:       val,
		count:     count,
		expiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}

func (c *MemoryCache) Del(keys ...string) {
	c.mu.Lock()
	for _, k := range keys {
		delete(c.items, k)
	}
	c.mu.Unlock()
}

func (c *MemoryCache) Incr(key string, window time.Duration) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	it, exists := c.items[key]
	if !exists || now.After(it.expiresAt) {
		c.items[key] = item{
			val:       "1",
			count:     1,
			expiresAt: now.Add(window),
		}
		return 1
	}

	it.count++
	it.val = strconv.FormatInt(it.count, 10)
	c.items[key] = it
	return it.count
}
