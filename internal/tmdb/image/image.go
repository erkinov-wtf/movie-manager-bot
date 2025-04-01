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
	"net/http"
	"time"

	"github.com/nfnt/resize"
	"image/jpeg"
)

// GetImage now uses caching mechanism with improved error handling and image compression
func GetImage(app *appCfg.App, imageId string) (*bytes.Buffer, error) {
	const op = "image.GetImage"

	url := fmt.Sprintf("%s%s", app.Cfg.Endpoints.ImageUrl, imageId)
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte(url)))

	app.Logger.Debug(op, nil, "Checking image cache", "image_id", imageId)
	app.Cache.ImageCache.Mu.RLock()
	cachedImg, exists := app.Cache.ImageCache.Cache[cacheKey]
	app.Cache.ImageCache.Mu.RUnlock()

	if exists {
		app.Logger.Info(op, nil, "Image retrieved from cache", "image_id", imageId)
		return cachedImg.Data, nil
	}

	// Creating a client with a longer timeout for image downloads
	app.Logger.Debug(op, nil, "Image not in cache, creating HTTP client for retrieval", "url", url)
	imageClient := app.TMDBClient.NewClientWithCustomTimeout(10 * time.Second)

	app.Logger.Debug(op, nil, "Making image retrieval request", "url", url)
	resp, err := imageClient.Get(url)
	if err != nil {
		app.Logger.Error(op, nil, "Failed to fetch image data", "url", url, "error", err.Error())
		return nil, fmt.Errorf("error fetching image data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		app.Logger.Error(op, nil, "Received non-200 response from image source",
			"url", url, "status_code", resp.StatusCode)
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	app.Logger.Debug(op, nil, "Decoding image data", "image_id", imageId)
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		app.Logger.Error(op, nil, "Failed to decode image data", "image_id", imageId, "error", err.Error())
		return nil, fmt.Errorf("error decoding image: %w", err)
	}

	// Compress and resize the image
	app.Logger.Debug(op, nil, "Compressing and resizing image",
		"image_id", imageId,
		"original_width", img.Bounds().Dx(),
		"original_height", img.Bounds().Dy())
	compressedImg := compressImage(img, 800, 1200)

	var imageBuffer bytes.Buffer
	app.Logger.Debug(op, nil, "Encoding compressed image as JPEG", "image_id", imageId)
	err = jpeg.Encode(&imageBuffer, compressedImg, &jpeg.Options{Quality: 85})
	if err != nil {
		app.Logger.Error(op, nil, "Failed to encode compressed image", "image_id", imageId, "error", err.Error())
		return nil, fmt.Errorf("error encoding compressed image: %w", err)
	}

	// Update the cache
	app.Logger.Debug(op, nil, "Updating image cache", "image_id", imageId, "buffer_size", imageBuffer.Len())
	app.Cache.ImageCache.Mu.Lock()
	defer app.Cache.ImageCache.Mu.Unlock()

	// If cache is full, evict oldest entry
	if len(app.Cache.ImageCache.Cache) >= app.Cache.ImageCache.MaxSize {
		app.Logger.Debug(op, nil, "Cache full, evicting oldest entry",
			"cache_size", len(app.Cache.ImageCache.Cache),
			"max_size", app.Cache.ImageCache.MaxSize)
		var oldestKey string
		var oldestTime time.Time
		for k, v := range app.Cache.ImageCache.Cache {
			if oldestTime.IsZero() || v.Timestamp.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.Timestamp
			}
		}
		delete(app.Cache.ImageCache.Cache, oldestKey)
		app.Logger.Debug(op, nil, "Evicted oldest cache entry", "evicted_key", oldestKey)
	}

	app.Cache.ImageCache.Cache[cacheKey] = &cache.CachedImage{
		Data:      &imageBuffer,
		Timestamp: time.Now(),
	}

	app.Logger.Info(op, nil, "Image successfully retrieved, compressed, and cached",
		"image_id", imageId,
		"compressed_size", imageBuffer.Len(),
		"new_width", compressedImg.Bounds().Dx(),
		"new_height", compressedImg.Bounds().Dy())
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
