package middleware

import (
	"gopkg.in/telebot.v3"
	"log"
	"movie-manager-bot/helpers/messages"
	"movie-manager-bot/helpers/utils"
	"movie-manager-bot/models"
	"movie-manager-bot/storage/database"
	"sync"
)

// UserState tracks users waiting for token input
type UserState struct {
	WaitingForToken bool
	mutex           sync.RWMutex
}

// UserStates manages state for all users
var userStates = make(map[int64]*UserState)
var statesMutex sync.RWMutex

func getUserState(userID int64) *UserState {
	statesMutex.Lock()
	defer statesMutex.Unlock()

	if state, exists := userStates[userID]; exists {
		return state
	}

	state := &UserState{}
	userStates[userID] = state
	return state
}

// HandleText processes text messages
func HandleText(context telebot.Context) error {
	userID := context.Sender().ID
	state := getUserState(userID)

	state.mutex.RLock()
	waiting := state.WaitingForToken
	state.mutex.RUnlock()

	if !waiting {
		return nil
	}

	userToken := context.Text()

	state.mutex.Lock()
	state.WaitingForToken = false
	state.mutex.Unlock()

	if !utils.TestApiToken(userToken) {
		return context.Send(messages.TokenTestFailed)
	}

	if err := database.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("tmdb_api_key", userToken).Error; err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	return context.Send(messages.TokenSaved)
}
