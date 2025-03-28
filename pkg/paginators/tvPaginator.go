package paginators

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/tv"
)

func PaginateTV(tvCache *cache.Item, page, tvCount int) []tv.TV {
	const itemsPerPage = 3

	start := (page - 1) * itemsPerPage
	end := start + itemsPerPage

	if end > tvCount {
		end = tvCount
	}

	var paginatedTV []tv.TV
	for i := start; i < end; i++ {
		if cachedValue, exists := tvCache.Get(i + 1); exists {
			if tvData, ok := cachedValue.(tv.TV); ok {
				paginatedTV = append(paginatedTV, tvData)
			}
		}
	}

	return paginatedTV
}
