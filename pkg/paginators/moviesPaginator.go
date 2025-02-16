package paginators

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	movieType "github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/movie"
)

func PaginateMovies(moviesCache *cache.Item, page, movieCount int) []movieType.Movie {
	const itemsPerPage = 3

	start := (page - 1) * itemsPerPage
	end := start + itemsPerPage

	if end > movieCount {
		end = movieCount
	}

	var paginatedMovies []movieType.Movie
	for i := start; i < end; i++ {
		if cachedValue, exists := moviesCache.Get(i + 1); exists {
			if movie, ok := cachedValue.(movieType.Movie); ok {
				paginatedMovies = append(paginatedMovies, movie)
			}
		}
	}

	return paginatedMovies
}
