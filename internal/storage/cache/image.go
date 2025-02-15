package cache

import (
	"bytes"
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

func NewImageCache() *Image {
	return &Image{
		Cache:   make(map[string]*CachedImage),
		MaxSize: 50, // cache capacity - amount of photos
	}
}
