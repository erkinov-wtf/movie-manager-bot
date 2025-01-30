package middleware

import (
	"github.com/erkinov-wtf/movie-manager-bot/models"
	"github.com/erkinov-wtf/movie-manager-bot/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/storage/database"
	"gopkg.in/telebot.v3"

	"time"
)

func IsRegisteredUser(c telebot.Context) bool {
	userID := c.Sender().ID
	if isActive, userCache := cache.UserCache.Get(userID); isActive && userCache.Value {
		return true
	}

	var user models.User
	err := database.DB.First(&user, "id = ?", userID).Error
	if err == nil {
		isTokenWaiting := true
		if user.TmdbApiKey != nil {
			isTokenWaiting = false
		}
		cache.UserCache.Set(userID, true, 24*time.Hour, isTokenWaiting)
		return true
	}

	return false
}

func RequireRegistration(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if !IsRegisteredUser(c) {
			return c.Send("You need to register to use this bot. Please type /start to continue")
		}
		return next(c)
	}
}
