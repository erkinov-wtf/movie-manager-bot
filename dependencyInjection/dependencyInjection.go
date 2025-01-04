package dependencyInjection

import (
	"github.com/erkinov-wtf/movie-manager-bot/handlers/defaults"
	"github.com/erkinov-wtf/movie-manager-bot/handlers/info"
	"github.com/erkinov-wtf/movie-manager-bot/handlers/movie"
	"github.com/erkinov-wtf/movie-manager-bot/handlers/tv"
	"github.com/erkinov-wtf/movie-manager-bot/handlers/watchlist"
	"github.com/erkinov-wtf/movie-manager-bot/interfaces"
)

type Container struct {
	DefaultHandler   interfaces.DefaultInterface
	MovieHandler     interfaces.MovieInterface
	TVHandler        interfaces.TVInterface
	InfoHandler      interfaces.InfoInterface
	WatchlistHandler interfaces.WatchlistInterface
}

func NewContainer() *Container {
	return &Container{
		DefaultHandler:   defaults.NewDefaultHandler(),
		MovieHandler:     movie.NewMovieHandler(),
		TVHandler:        tv.NewTVHandler(),
		InfoHandler:      info.NewInfoHandler(),
		WatchlistHandler: watchlist.NewWatchlistHandler(),
	}
}
