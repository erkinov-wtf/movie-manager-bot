package dependencyInjection

import (
	"movie-manager-bot/handlers"
	"movie-manager-bot/interfaces"
)

type Container struct {
	BotHandler interfaces.BotInterface
}

func NewContainer() *Container {
	return &Container{
		BotHandler: handlers.NewBotHandler(),
	}
}
