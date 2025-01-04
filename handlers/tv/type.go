package tv

import "github.com/erkinov-wtf/movie-manager-bot/interfaces"

type tvHandler struct{}

func NewTVHandler() interfaces.TVInterface {
	return &tvHandler{}
}
