package info

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
)

type InfoHandler struct {
	App *app.App
}

func NewInfoHandler(app *app.App) interfaces.InfoInterface {
	return &InfoHandler{
		App: app,
	}
}

type tvStats struct {
	amount    int
	totalTime int64
}

type movieStats struct {
	amount    int
	totalTime int64
}
