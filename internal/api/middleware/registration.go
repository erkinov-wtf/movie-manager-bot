package middleware

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/internal/models"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"gopkg.in/telebot.v3"
	"log"

	"time"
)

func IsRegisteredUser(c telebot.Context, app *app.App) bool {
	userId := c.Sender().ID
	if isActive, userCache := app.Cache.UserCache.Get(userId); isActive && userCache.Value {
		return true
	}
	log.Printf("User ID %d not found in cache, checking database", userId)

	var user models.User
	err := app.Database.First(&user, "id = ?", userId).Error
	if err == nil {
		isTokenWaiting := user.TmdbApiKey == nil
		app.Cache.UserCache.Set(userId, true, 24*time.Hour, isTokenWaiting)
		log.Printf("User ID %d found in database and added to cache", userId)
		return true
	}

	log.Printf("User ID %d not found in database", userId)
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
