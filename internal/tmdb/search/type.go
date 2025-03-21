package search

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/movie"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/tv"
)

type MovieSearch struct {
	Results      []movie.Movie `json:"results"`
	Page         int64         `json:"page"`
	TotalPages   int64         `json:"total_pages"`
	TotalResults int64         `json:"total_results"`
}

type TVSearch struct {
	Results      []tv.TV `json:"results"`
	Page         int64   `json:"page"`
	TotalPages   int64   `json:"total_pages"`
	TotalResults int64   `json:"total_results"`
}
