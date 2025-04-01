package middleware

import (
	"context"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"gopkg.in/telebot.v3"
	"time"
)

func IsRegisteredUser(c telebot.Context, app *app.App) bool {
	const op = "middleware.IsRegisteredUser"
	userId := c.Sender().ID

	// First check cache
	if isActive, userCache := app.Cache.UserCache.Get(userId); isActive && userCache.Value {
		app.Logger.Debug(op, c, "User found in cache", "user_id", userId)
		return true
	}
	app.Logger.Debug(op, c, "User not found in cache, checking database", "user_id", userId)

	// If not in cache, check database
	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	user, err := app.Repository.Users.GetUser(ctxDb, userId)
	if err == nil {
		isTokenWaiting := user.TmdbApiKey == nil
		app.Logger.Debug(op, c, "User found in database, adding to cache",
			"user_id", userId, "token_waiting", isTokenWaiting)
		app.Cache.UserCache.Set(userId, true, 24*time.Hour, isTokenWaiting)
		return true
	}

	app.Logger.Info(op, c, "User not found in database", "user_id", userId)
	return false
}

func RequireRegistration(next telebot.HandlerFunc, app *app.App) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		const op = "middleware.RequireRegistration"

		if !IsRegisteredUser(c, app) {
			app.Logger.Info(op, c, "Registration required for unregistered user")
			return c.Send(messages.RegistrationRequired)
		}

		app.Logger.Debug(op, c, "User is registered, proceeding with request")
		return next(c)
	}
}
