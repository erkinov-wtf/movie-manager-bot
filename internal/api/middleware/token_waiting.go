package middleware

import (
	appCfg "github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"gopkg.in/telebot.v3"
)

func RequireTMDBToken(next telebot.HandlerFunc, app *appCfg.App) telebot.HandlerFunc {
	return RequireRegistration(func(c telebot.Context) error {
		const op = "middleware.RequireTMDBToken"
		userId := c.Sender().ID

		app.Logger.Debug(op, c, "Checking if user has TMDB token")
		_, userCache := app.Cache.UserCache.Fetch(userId)

		if userCache.ApiToken.IsTokenWaiting {
			app.Logger.Info(op, c, "TMDB token required for access")
			return c.Send(messages.TokenRequired)
		}

		app.Logger.Debug(op, c, "User has valid TMDB token, proceeding with request")
		return next(c)
	}, app)
}
