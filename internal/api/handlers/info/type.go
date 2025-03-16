package info

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
)

type InfoHandler struct {
	app *app.App
}

func NewInfoHandler(app *app.App) interfaces.InfoInterface {
	return &InfoHandler{
		app: app,
	}
}

type tvStats struct {
	amount    int
	totalTime int32
}

type movieStats struct {
	amount    int
	totalTime int32
}
