package cache

import (
	"github.com/erkinov-wtf/movie-manager-bot/models"
	"github.com/erkinov-wtf/movie-manager-bot/storage/database"
	"log"
	"sync"
	"time"
)

type UserCacheItem struct {
	Value       bool
	ExpireTime  time.Time
	ApiToken    ApiToken
	SearchState SearchState
}

type UserCacheData struct {
	items map[int64]UserCacheItem
	mu    sync.RWMutex
}

type ApiToken struct {
	IsTokenWaiting bool
	Token          string
}

type SearchState struct {
	IsMovieSearch  bool
	IsTVShowSearch bool
}

var UserCache UserCacheData

func NewUserCache() {
	UserCache.items = make(map[int64]UserCacheItem)
	log.Print("User cache setup")
}

// Set adds a user to the cache with an optional expiration time
func (c *UserCacheData) Set(userID int64, value bool, expiration time.Duration, isTokenWaiting bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tokenDb := c.getTokenDb(userID, isTokenWaiting)
	c.items[userID] = UserCacheItem{
		Value:      value,
		ExpireTime: time.Now().Add(expiration),
		ApiToken: ApiToken{
			IsTokenWaiting: isTokenWaiting,
			Token:          tokenDb,
		},

		SearchState: SearchState{
			IsMovieSearch:  false,
			IsTVShowSearch: false,
		},
	}
}

func (c *UserCacheData) UpdateTokenState(userID int64, isTokenWaiting bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tokenDb := c.getTokenDb(userID, isTokenWaiting)
	c.items[userID] = UserCacheItem{
		ApiToken: ApiToken{IsTokenWaiting: isTokenWaiting, Token: tokenDb},
	}
}

// Get retrieves a user from the cache
func (c *UserCacheData) Get(userID int64) (isActive bool, data *UserCacheItem) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	userCache, found := c.items[userID]
	if !found || userCache.ExpireTime.Before(time.Now()) {
		return false, &userCache
	}

	return true, &userCache
}

// Clear removes all items from the cache
func (c *UserCacheData) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[int64]UserCacheItem)
}

func (c *UserCacheData) getTokenDb(userId int64, isTokenWaiting bool) string {
	var apiTokenDb string

	if !isTokenWaiting {
		err := database.DB.Model(&models.User{}).Where("id = ?", userId).
			Select("tmdb_api_key").
			Pluck("tmdb_api_key", &apiTokenDb).Error

		if err != nil {
			log.Print(err)
			return ""
		}
	}

	if apiTokenDb == "" {
		return ""
	}

	return apiTokenDb
}
