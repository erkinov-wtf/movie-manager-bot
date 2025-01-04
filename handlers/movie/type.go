package movie

import "github.com/erkinov-wtf/movie-manager-bot/interfaces"

type movieHandler struct{}

func NewMovieHandler() interfaces.MovieInterface {
	return &movieHandler{}
}
