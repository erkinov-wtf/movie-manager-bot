package tv

import (
	"movie-manager-bot/interfaces"
)

type tvHandler struct{}

func NewTVHandler() interfaces.TVInterface {
	return &tvHandler{}
}
