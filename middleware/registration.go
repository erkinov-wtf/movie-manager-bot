package middleware

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"movie-manager-bot/models"
	"movie-manager-bot/storage/cache"
	"movie-manager-bot/storage/database"
	"time"
)

var userCache = cache.NewUserCache()

func IsRegisteredUser(c telebot.Context) bool {
	userID := c.Sender().ID
	if registered, found := userCache.Get(userID); found && registered {
		return true
	}

	var user models.User
	err := database.DB.First(&user, "id = ?", userID).Error
	if err == nil {
		userCache.Set(userID, true, 24*time.Hour)
		return true
	}

	return false
}

func RequireRegistration(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if !IsRegisteredUser(c) {
			return c.Send("You need to register to use this bot. Please type /start to continue")
		}
		fmt.Println(userCache)
		return next(c)
	}
}
