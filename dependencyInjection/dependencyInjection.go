package dependencyInjection

import (
	"movie-manager-bot/handlers/defaults"
	"movie-manager-bot/handlers/info"
	"movie-manager-bot/handlers/movie"
	"movie-manager-bot/handlers/tv"
	"movie-manager-bot/interfaces"
)

type Container struct {
	DefaultHandler interfaces.DefaultInterface
	MovieHandler   interfaces.MovieInterface
	TVHandler      interfaces.TVInterface
	InfoHandler    interfaces.InfoInterface
}

func NewContainer() *Container {
	return &Container{
		DefaultHandler: defaults.NewDefaultHandler(),
		MovieHandler:   movie.NewMovieHandler(),
		TVHandler:      tv.NewTVHandler(),
		InfoHandler:    info.NewInfoHandler(),
	}
}
