package helpers

import (
	"movie-manager-bot/api/media/movie"
	"movie-manager-bot/storage"
)

func PaginateMovies(moviesCache *storage.Cache, page, movieCount int) []movie.Movie {
	const itemsPerPage = 3

	start := (page - 1) * itemsPerPage
	end := start + itemsPerPage

	if end > movieCount {
		end = movieCount
	}

	var paginatedMovies []movie.Movie
	for i := start; i < end; i++ {
		if mov, exists := moviesCache.Get(i + 1); exists {
			paginatedMovies = append(paginatedMovies, mov)
		}
	}

	return paginatedMovies
}
