package handlers

import "movie-manager-bot/interfaces"

type botHandler struct{}

func NewBotHandler() interfaces.BotInterface {
	return &botHandler{}
}
