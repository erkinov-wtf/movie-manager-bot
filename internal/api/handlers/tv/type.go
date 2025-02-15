package tv

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
)

type TVHandler struct {
	App *app.App
}

func NewTVHandler(app *app.App) interfaces.TVInterface {
	return &TVHandler{
		App: app,
	}
}
