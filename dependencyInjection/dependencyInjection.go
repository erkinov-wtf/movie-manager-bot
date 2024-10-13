package dependencyInjection

import (
	"movie-manager-bot/handlers/movie"
	"movie-manager-bot/handlers/tv"
	"movie-manager-bot/interfaces"
)

type Container struct {
	MovieHandler interfaces.MovieInterface
	TVHandler    interfaces.TVInterface
}

func NewContainer() *Container {
	return &Container{
		MovieHandler: movie.NewMovieHandler(),
		TVHandler:    tv.NewTVHandler(),
	}
}
