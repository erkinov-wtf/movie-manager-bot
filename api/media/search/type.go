package search

import "movie-manager-bot/api/media/movie"

type MovieSearch struct {
	Results      []movie.Movie `json:"results"`
	Page         int64         `json:"page"`
	TotalPages   int64         `json:"total_pages"`
	TotalResults int64         `json:"total_results"`
}
