package middleware

import (
	"context"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"gopkg.in/telebot.v3"
	"log"

	"time"
)

func IsRegisteredUser(c telebot.Context, app *app.App) bool {

	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	userId := c.Sender().ID
	if isActive, userCache := app.Cache.UserCache.Get(userId); isActive && userCache.Value {
		return true
	}
	log.Printf("User Id %d not found in cache, checking database", userId)

	user, err := app.Repository.Users.GetUser(ctxDb, userId)
	if err == nil {
		isTokenWaiting := user.TmdbApiKey == nil
		app.Cache.UserCache.Set(userId, true, 24*time.Hour, isTokenWaiting)
		log.Printf("User Id %d found in database and added to cache", userId)
		return true
	}
	log.Printf("User Id %d not found in database", userId)
	return false
}

func RequireRegistration(next telebot.HandlerFunc, app *app.App) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if !IsRegisteredUser(c, app) {
			return c.Send(messages.RegistrationRequired)
		}
		return next(c)
	}
}
