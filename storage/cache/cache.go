package cache

import (
	"sync"
)

type CacheItem struct {
	Value interface{}
}

type Cache struct {
	items map[int]CacheItem
	mu    sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[int]CacheItem),
	}
}

func (c *Cache) Set(key int, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = CacheItem{
		Value: value,
	}
}

func (c *Cache) Get(key int) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	return item.Value, true
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[int]CacheItem)
}
