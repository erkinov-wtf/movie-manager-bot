package cache

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database/repository"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/encryption"
	"sync"
)

type Manager struct {
	MovieCache  map[int]*Item
	TVShowCache map[int]*Item
	UserCache   *UserCacheData
	ImageCache  *Image
	mu          sync.RWMutex
}

func NewCacheManager(repos *repository.Manager, encryptor *encryption.KeyEncryptor) *Manager {
	return &Manager{
		MovieCache:  make(map[int]*Item),
		TVShowCache: make(map[int]*Item),
		UserCache:   NewUserCache(repos, encryptor),
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
