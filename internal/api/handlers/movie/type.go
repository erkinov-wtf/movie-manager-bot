package movie

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
)

type MovieHandler struct {
	App *app.App
}

func NewMovieHandler(app *app.App) interfaces.MovieInterface {
	return &MovieHandler{
		App: app,
	}
}
