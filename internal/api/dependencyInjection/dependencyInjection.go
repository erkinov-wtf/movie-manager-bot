package dependencyInjection

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/defaults"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/info"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/movie"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/tv"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/watchlist"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
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
