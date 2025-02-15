package defaults

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
)

type DefaultHandler struct {
	app          *app.App
	movieHandler interfaces.MovieInterface
	tvHandler    interfaces.TVInterface
}

func NewDefaultHandler(app *app.App, movieHandler interfaces.MovieInterface, tvHandler interfaces.TVInterface) *DefaultHandler {
	return &DefaultHandler{
		app:          app,
		movieHandler: movieHandler,
		tvHandler:    tvHandler,
	}
}
