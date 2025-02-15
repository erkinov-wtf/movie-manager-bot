package image

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	appCfg "github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
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
func GetImage(app *appCfg.App, imageId string) (*bytes.Buffer, error) {
	url := fmt.Sprintf("%s%s", app.Cfg.Endpoints.ImageUrl, imageId)
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte(url)))

	app.Cache.ImageCache.Mu.Lock()
	if cachedImg, exists := app.Cache.ImageCache.Cache[cacheKey]; exists {
		app.Cache.ImageCache.Mu.RUnlock()
		log.Printf("Image retrieved from cache: %s", imageId)
		return cachedImg.Data, nil
	}
	app.Cache.ImageCache.Mu.RUnlock()

	// Creating a client with a longer timeout for image downloads
	imageClient := app.TMDBClient.NewClientWithCustomTimeout(10 * time.Second)

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

	app.Cache.ImageCache.Mu.Lock()
	defer app.Cache.ImageCache.Mu.Unlock()

	if len(app.Cache.ImageCache.Cache) >= app.Cache.ImageCache.MaxSize {
		var oldestKey string
		var oldestTime time.Time
		for k, v := range app.Cache.ImageCache.Cache {
			if oldestTime.IsZero() || v.Timestamp.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.Timestamp
			}
		}
		delete(app.Cache.ImageCache.Cache, oldestKey)
	}

	app.Cache.ImageCache.Cache[cacheKey] = &cache.CachedImage{
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
