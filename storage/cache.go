package storage

import (
	movieType "movie-manager-bot/api/media/movie"
	"sync"
)

type CacheItem struct {
	Movie movieType.Movie
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

func (c *Cache) Set(key int, movie movieType.Movie) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = CacheItem{
		Movie: movie,
	}
}

func (c *Cache) Get(key int) (movieType.Movie, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return movieType.Movie{}, false
	}

	return item.Movie, true
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[int]CacheItem)
}
