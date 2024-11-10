package cache

import (
	"sync"
	"time"
)

type UserCacheItem struct {
	Value      bool
	ExpireTime time.Time
}

type UserCache struct {
	items map[int64]UserCacheItem
	mu    sync.RWMutex
}

func NewUserCache() *UserCache {
	return &UserCache{
		items: make(map[int64]UserCacheItem),
	}
}

// Set adds a user to the cache with an optional expiration time
func (c *UserCache) Set(userID int64, value bool, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[userID] = UserCacheItem{
		Value:      value,
		ExpireTime: time.Now().Add(expiration),
	}
}

// Get retrieves a user from the cache
func (c *UserCache) Get(userID int64) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[userID]
	if !found || item.ExpireTime.Before(time.Now()) {
		return false, false
	}

	return item.Value, true
}

// Clear removes all items from the cache
func (c *UserCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[int64]UserCacheItem)
}
