package tv

import "github.com/erkinov-wtf/movie-manager-bot/interfaces"

type TVHandler struct{}

func NewTVHandler() interfaces.TVInterface {
	return &TVHandler{}
}
