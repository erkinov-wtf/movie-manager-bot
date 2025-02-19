package middleware

import (
	appCfg "github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"gopkg.in/telebot.v3"
)

func RequireTMDBToken(next telebot.HandlerFunc, app *appCfg.App) telebot.HandlerFunc {
	return RequireRegistration(func(c telebot.Context) error {
		_, userCache := app.Cache.UserCache.Fetch(c.Sender().ID)
		if userCache.ApiToken.IsTokenWaiting {
			return c.Send(messages.TokenRequired)
		}
		return next(c)
	}, app)
}
