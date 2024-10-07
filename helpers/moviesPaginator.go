package helpers

import "movie-manager-bot/api/media/movie"

func PaginateMovies(movies map[int]movie.Movie, page, movieCount int) []movie.Movie {
	const itemsPerPage = 3

	start := (page - 1) * itemsPerPage
	end := start + itemsPerPage

	if end > movieCount {
		end = movieCount
	}

	var paginatedMovies []movie.Movie
	for i := start; i < end; i++ {
		if mov, exists := movies[i]; exists {
			paginatedMovies = append(paginatedMovies, mov)
		}
	}

	return paginatedMovies
}
