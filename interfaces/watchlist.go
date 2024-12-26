package interfaces

import "gopkg.in/telebot.v3"

type WatchlistInterface interface {
	WatchlistInfo(context telebot.Context) error
	WatchlistCallback(context telebot.Context) error
}
