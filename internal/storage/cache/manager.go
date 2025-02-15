package cache

import "sync"

type Manager struct {
	MovieCache  map[int]*Cache
	TVShowCache map[int]*Cache
	UserCache   *UserCacheData
	ImageCache  *Image
	mu          sync.RWMutex
}

func NewCacheManager() *Manager {
	return &Manager{
		MovieCache:  make(map[int]*Cache),
		TVShowCache: make(map[int]*Cache),
		UserCache:   NewUserCache(),
		ImageCache:  NewImageCache(),
	}
}

func (cm *Manager) GetMovieCache(userId int) *Cache {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.MovieCache[userId]; !exists {
		cm.MovieCache[userId] = NewCache()
	}
	return cm.MovieCache[userId]
}

func (cm *Manager) GetTVShowCache(userId int) *Cache {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.TVShowCache[userId]; !exists {
		cm.TVShowCache[userId] = NewCache()
	}
	return cm.TVShowCache[userId]
}
