package info

import (
	"github.com/erkinov-wtf/movie-manager-bot/interfaces"
)

type InfoHandler struct{}

func NewInfoHandler() interfaces.InfoInterface {
	return &InfoHandler{}
}

type tvStats struct {
	amount    int
	totalTime int64
}

type movieStats struct {
	amount    int
	totalTime int64
}
