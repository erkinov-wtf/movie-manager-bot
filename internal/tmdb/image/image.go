package image

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/nfnt/resize"
	"image/jpeg"
)

// GetImage now uses caching mechanism with improved error handling and image compression
func GetImage(imageId string) (*bytes.Buffer, error) {
	url := fmt.Sprintf("%s%s", config.Cfg.Endpoints.ImageUrl, imageId)
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte(url)))

	cache.ImageCache.Mu.RLock()
	if cachedImg, exists := cache.ImageCache.Cache[cacheKey]; exists {
		cache.ImageCache.Mu.RUnlock()
		log.Printf("Image retrieved from cache: %s", imageId)
		return cachedImg.Data, nil
	}
	cache.ImageCache.Mu.RUnlock()

	// Creating a client with a longer timeout for image downloads
	imageClient := tmdb.Client.NewClientWithCustomTimeout(15 * time.Second)

	log.Printf("Making image retrieval request: %s", url)
	resp, err := imageClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching image data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}

	// Compress and resize the image
	compressedImg := compressImage(img, 800, 1200)

	var imageBuffer bytes.Buffer
	err = jpeg.Encode(&imageBuffer, compressedImg, &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, fmt.Errorf("error encoding compressed image: %w", err)
	}

	cache.ImageCache.Mu.Lock()
	defer cache.ImageCache.Mu.Unlock()

	if len(cache.ImageCache.Cache) >= cache.ImageCache.MaxSize {
		var oldestKey string
		var oldestTime time.Time
		for k, v := range cache.ImageCache.Cache {
			if oldestTime.IsZero() || v.Timestamp.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.Timestamp
			}
		}
		delete(cache.ImageCache.Cache, oldestKey)
	}

	cache.ImageCache.Cache[cacheKey] = &cache.CachedImage{
		Data:      &imageBuffer,
		Timestamp: time.Now(),
	}

	log.Printf("Image successfully retrieved, compressed, and cached: %s", imageId)
	return &imageBuffer, nil
}

// compressImage resizes the image while maintaining aspect ratio
func compressImage(img image.Image, maxWidth, maxHeight uint) image.Image {
	// Get original image dimensions
	width := uint(img.Bounds().Dx())
	height := uint(img.Bounds().Dy())

	// Calculate aspect ratio
	ratio := float64(width) / float64(height)

	// Determine new dimensions
	var newWidth, newHeight uint
	if width > height {
		if width > maxWidth {
			newWidth = maxWidth
			newHeight = uint(float64(newWidth) / ratio)
		} else {
			return img
		}
	} else {
		if height > maxHeight {
			newHeight = maxHeight
			newWidth = uint(float64(newHeight) * ratio)
		} else {
			return img
		}
	}

	return resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
}
