package interfaces

import "gopkg.in/telebot.v3"

type InfoInterface interface {
	Info(context telebot.Context) error
	InfoCallback(context telebot.Context) error
}
