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
	IsSearchWaiting bool
	IsMovieSearch   bool
	IsTVShowSearch  bool
}

var UserCache UserCacheData

func NewUserCache() {
	UserCache.items = make(map[int64]UserCacheItem)
	log.Print("User cache setup")
}

func (c *UserCacheData) Set(userId int64, value bool, expiration time.Duration, isTokenWaiting bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tokenDb := c.getTokenDb(userId, isTokenWaiting)
	c.items[userId] = UserCacheItem{
		Value:      value,
		ExpireTime: time.Now().Add(expiration),
		ApiToken: ApiToken{
			IsTokenWaiting: isTokenWaiting,
			Token:          tokenDb,
		},

		SearchState: SearchState{
			IsSearchWaiting: false,
			IsMovieSearch:   false,
			IsTVShowSearch:  false,
		},
	}
}

func (c *UserCacheData) Get(userId int64) (isActive bool, data *UserCacheItem) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	userCache, found := c.items[userId]
	if !found || userCache.ExpireTime.Before(time.Now()) {
		return false, nil
	}

	return true, &userCache
}

func (c *UserCacheData) UpdateTokenState(userId int64, isTokenWaiting bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	userCache, found := c.items[userId]
	if !found || userCache.ExpireTime.Before(time.Now()) {
		log.Printf("User ID %d not found in cache or cache expired", userId)
		return
	}

	userCache.ApiToken.IsTokenWaiting = isTokenWaiting
	userCache.ApiToken.Token = c.getTokenDb(userId, isTokenWaiting)

	c.items[userId] = userCache

	log.Printf("Updated token state for user ID %d: TokenWaiting=%v, Token=%s", userId, isTokenWaiting, userCache.ApiToken.Token)
}

func (c *UserCacheData) SetSearchStartTrue(userId int64, isMovieSearch bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if userCache, found := c.items[userId]; found && time.Now().Before(userCache.ExpireTime) {
		userCache.SearchState = SearchState{
			IsSearchWaiting: true,
			IsMovieSearch:   isMovieSearch,
			IsTVShowSearch:  !isMovieSearch,
		}

		c.items[userId] = userCache
		log.Printf("Updated search state to TRUE for user ID %d", userId)
	}
}

func (c *UserCacheData) SetSearchStartFalse(userId int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if userCache, found := c.items[userId]; found {
		userCache.SearchState = SearchState{
			IsSearchWaiting: false,
			IsMovieSearch:   false,
			IsTVShowSearch:  false,
		}

		c.items[userId] = userCache
		log.Printf("Updated search state to FALSE for user ID %d", userId)
	}
}

// Clear removes all items from the cache
func (c *UserCacheData) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[int64]UserCacheItem)
}

func (c *UserCacheData) getTokenDb(userId int64, isTokenWaiting bool) string {
	if isTokenWaiting {
		return ""
	}

	var apiTokenDb string

	if err := database.DB.Model(&models.User{}).Where("id = ?", userId).
		Select("tmdb_api_key").
		Pluck("tmdb_api_key", &apiTokenDb).Error; err != nil {
		log.Print(err)
		return ""
	}

	return apiTokenDb
}
