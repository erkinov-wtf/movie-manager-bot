package helpers

import (
	movieType "movie-manager-bot/api/media/movie"
	"movie-manager-bot/storage"
)

func PaginateMovies(moviesCache *storage.Cache, page, movieCount int) []movieType.Movie {
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
