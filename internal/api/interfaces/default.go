package interfaces

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"gopkg.in/telebot.v3"
)

type DefaultInterface interface {
	Start(context telebot.Context) error
	GetToken(context telebot.Context) error
	HandleReplySearch(context telebot.Context, userCache *cache.UserCacheItem) error
	HandleTextInput(context telebot.Context) error
	DefaultCallback(context telebot.Context) error
}
