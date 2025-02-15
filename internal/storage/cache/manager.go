package cache

import (
	"gorm.io/gorm"
	"sync"
)

type Manager struct {
	MovieCache  map[int]*Item
	TVShowCache map[int]*Item
	UserCache   *UserCacheData
	ImageCache  *Image
	mu          sync.RWMutex
}

func NewCacheManager(db *gorm.DB) *Manager {
	return &Manager{
		MovieCache:  make(map[int]*Item),
		TVShowCache: make(map[int]*Item),
		UserCache:   NewUserCache(db),
		ImageCache:  NewImageCache(),
	}
}

func (cm *Manager) GetMovieCache(userId int) *Item {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.MovieCache[userId]; !exists {
		cm.MovieCache[userId] = NewCache()
	}
	return cm.MovieCache[userId]
}

func (cm *Manager) GetTVShowCache(userId int) *Item {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.TVShowCache[userId]; !exists {
		cm.TVShowCache[userId] = NewCache()
	}
	return cm.TVShowCache[userId]
}
