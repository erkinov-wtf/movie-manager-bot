package info

import (
	"movie-manager-bot/interfaces"
)

type infoHandler struct{}

func NewInfoHandler() interfaces.InfoInterface {
	return &infoHandler{}
}

type tvStats struct {
	amount    int
	totalTime int64
}

type movieStats struct {
	amount    int
	totalTime int64
}
