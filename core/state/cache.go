package state

import (
	"frizo-blockchain/common"
	"sync"
)

// Simple LRU cache implementation
type lruCache struct {
	cache map[common.Hash]interface{}
	size  int
	mu    sync.RWMutex
}

func newLRUCache(size int) *lruCache {
	return &lruCache{
		cache: make(map[common.Hash]interface{}),
		size:  size,
	}
}

// Get get data from cache
func (c *lruCache) Get(key common.Hash) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.cache[key]
	return val, ok
}

// Add add data into cache
func (c *lruCache) Add(key common.Hash, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.cache) >= c.size {
		// remove random entry
		for hash := range c.cache {
			delete(c.cache, hash)
			break
		}
	}
	c.cache[key] = value
}
