package defaults

import "github.com/erkinov-wtf/movie-manager-bot/interfaces"

type defaultHandler struct{}

func NewDefaultHandler() interfaces.DefaultInterface {
	return &defaultHandler{}
}
