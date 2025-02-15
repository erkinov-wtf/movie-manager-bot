package api

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/defaults"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/info"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/movie"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/tv"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/watchlist"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
)

type Resolver struct {
	DefaultHandler   interfaces.DefaultInterface
	MovieHandler     interfaces.MovieInterface
	TVHandler        interfaces.TVInterface
	InfoHandler      interfaces.InfoInterface
	WatchlistHandler interfaces.WatchlistInterface
}

func NewResolver(app *app.App) *Resolver {
	return &Resolver{
		DefaultHandler:   defaults.NewDefaultHandler(app, movie.NewMovieHandler(app), tv.NewTVHandler(app)),
		MovieHandler:     movie.NewMovieHandler(app),
		TVHandler:        tv.NewTVHandler(app),
		InfoHandler:      info.NewInfoHandler(app),
		WatchlistHandler: watchlist.NewWatchlistHandler(app),
	}
}
