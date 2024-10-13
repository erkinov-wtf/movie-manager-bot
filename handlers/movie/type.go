package movie

import (
	"movie-manager-bot/interfaces"
)

type movieHandler struct{}

func NewMovieHandler() interfaces.MovieInterface {
	return &movieHandler{}
}
