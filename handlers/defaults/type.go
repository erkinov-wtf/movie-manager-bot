package defaults

import (
	"movie-manager-bot/interfaces"
)

type defaultHandler struct{}

func NewDefaultHandler() interfaces.DefaultInterface {
	return &defaultHandler{}
}
