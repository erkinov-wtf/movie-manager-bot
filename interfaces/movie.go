package interfaces

import "gopkg.in/telebot.v3"

type MovieInterface interface {
	SearchMovie(context telebot.Context) error
	MovieCallback(context telebot.Context) error
}
