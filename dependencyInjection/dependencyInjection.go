package dependencyInjection

import (
	"movie-manager-bot/handlers/info"
	"movie-manager-bot/handlers/movie"
	"movie-manager-bot/handlers/tv"
	"movie-manager-bot/interfaces"
)

type Container struct {
	MovieHandler interfaces.MovieInterface
	TVHandler    interfaces.TVInterface
	InfoHandler  interfaces.InfoInterface
}

func NewContainer() *Container {
	return &Container{
		MovieHandler: movie.NewMovieHandler(),
		TVHandler:    tv.NewTVHandler(),
		InfoHandler:  info.NewInfoHandler(),
	}
}
