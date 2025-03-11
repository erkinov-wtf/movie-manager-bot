package cache

import (
	"context"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database/repository"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/encryption"
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
	items     map[int64]UserCacheItem
	mu        sync.RWMutex
	db        *repository.Manager
	encryptor *encryption.KeyEncryptor
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

func NewUserCache(repos *repository.Manager, keyEncryptor *encryption.KeyEncryptor) *UserCacheData {
	userCache := UserCacheData{
		items:     make(map[int64]UserCacheItem),
		db:        repos,
		encryptor: keyEncryptor,
	}

	log.Print("User cache setup")
	return &userCache
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

// Get method retrieves user data from cache
func (c *UserCacheData) Get(userId int64) (isActive bool, data *UserCacheItem) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	userCache, found := c.items[userId]
	if !found || userCache.ExpireTime.Before(time.Now()) {
		return false, nil
	}

	return true, &userCache
}

// Fetch method retrieves user data or creates new one and returns it
func (c *UserCacheData) Fetch(userId int64) (isActive bool, data *UserCacheItem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if userCache, found := c.items[userId]; found {
		if !userCache.ExpireTime.Before(time.Now()) {
			return true, &userCache
		}
		delete(c.items, userId)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	user, err := c.db.Users.GetUser(ctx, userId)
	if err != nil {
		log.Printf("Failed to fetch user %d from database: %v", userId, err)
		return false, nil
	}

	isTokenWaiting := user.TmdbApiKey == nil
	userCache := UserCacheItem{
		Value:      true,
		ExpireTime: time.Now().Add(24 * time.Hour),
		ApiToken: ApiToken{
			IsTokenWaiting: isTokenWaiting,
			Token:          c.getTokenDb(userId, isTokenWaiting),
		},
		SearchState: SearchState{
			IsSearchWaiting: false,
			IsMovieSearch:   false,
			IsTVShowSearch:  false,
		},
	}

	c.items[userId] = userCache
	log.Printf("User Id %d found in database and added to cache", userId)

	return true, &userCache
}

func (c *UserCacheData) UpdateTokenState(userId int64, isTokenWaiting bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	userCache, found := c.items[userId]
	if !found || userCache.ExpireTime.Before(time.Now()) {
		log.Printf("User Id %d not found in cache or cache expired", userId)
		return
	}

	userCache.ApiToken.IsTokenWaiting = isTokenWaiting
	userCache.ApiToken.Token = c.getTokenDb(userId, isTokenWaiting)

	c.items[userId] = userCache

	log.Printf("Updated token state for user Id %d: TokenWaiting=%v, Token=%s", userId, isTokenWaiting, userCache.ApiToken.Token)
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
		log.Printf("Updated search state to TRUE for user Id %d", userId)
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
		log.Printf("Updated search state to FALSE for user Id %d", userId)
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

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	apiTokenDb, err := c.db.Users.GetUserTMDBKey(ctx, userId)
	if err != nil {
		log.Println(err)
		return ""
	}

	decryptedToken, err := c.encryptor.Decrypt(apiTokenDb)
	if err != nil {
		log.Printf("error decrypting token: %v", err)
		return ""
	}

	return decryptedToken
}
