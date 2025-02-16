package defaults

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/keyboards"
)

type DefaultHandler struct {
	app          *app.App
	movieHandler interfaces.MovieInterface
	tvHandler    interfaces.TVInterface
	keyboards    *keyboards.KeyboardFactory
}

func NewDefaultHandler(app *app.App, movieHandler interfaces.MovieInterface, tvHandler interfaces.TVInterface, keyboard *keyboards.KeyboardFactory) *DefaultHandler {
	return &DefaultHandler{
		app:          app,
		movieHandler: movieHandler,
		tvHandler:    tvHandler,
		keyboards:    keyboard,
	}
}
