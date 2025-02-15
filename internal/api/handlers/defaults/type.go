package defaults

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
)

type DefaultHandler struct {
	App *app.App
}

func NewDefaultHandler(app *app.App) interfaces.DefaultInterface {
	return &DefaultHandler{
		App: app,
	}
}
