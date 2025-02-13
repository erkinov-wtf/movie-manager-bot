package movie

import "github.com/erkinov-wtf/movie-manager-bot/interfaces"

type MovieHandler struct{}

func NewMovieHandler() interfaces.MovieInterface {
	return &MovieHandler{}
}
