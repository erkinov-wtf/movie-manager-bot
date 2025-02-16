package api

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/defaults"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/info"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/movie"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/tv"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/watchlist"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/keyboards"
)

type Resolver struct {
	DefaultHandler   interfaces.DefaultInterface
	MovieHandler     interfaces.MovieInterface
	TVHandler        interfaces.TVInterface
	InfoHandler      interfaces.InfoInterface
	WatchlistHandler interfaces.WatchlistInterface

	KeyboardFactory *keyboards.KeyboardFactory
}

func NewResolver(app *app.App) *Resolver {
	movieHandler := movie.NewMovieHandler(app)
	tvHandler := tv.NewTVHandler(app)
	infoHandler := info.NewInfoHandler(app)
	watchlistHandler := watchlist.NewWatchlistHandler(app)
	keys := keyboards.NewKeyboardFactory(app, watchlistHandler, infoHandler)

	return &Resolver{
		DefaultHandler:   defaults.NewDefaultHandler(app, movieHandler, tvHandler, keys),
		MovieHandler:     movieHandler,
		TVHandler:        tvHandler,
		InfoHandler:      infoHandler,
		WatchlistHandler: watchlistHandler,
		KeyboardFactory:  keys,
	}
}
