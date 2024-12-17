package cache

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

type Image struct {
	Mu      sync.RWMutex
	Cache   map[string]*CachedImage
	MaxSize int
}

type CachedImage struct {
	Data      *bytes.Buffer
	Timestamp time.Time
}

var (
	ImageCache = &Image{
		Cache:   make(map[string]*CachedImage),
		MaxSize: 50, // cache capacity - amount of photos
	}
)

// GenerateCacheKey Generate a unique cache key
func GenerateCacheKey(url string) string {
	hash := sha256.Sum256([]byte(url))
	return fmt.Sprintf("%x", hash)
}
