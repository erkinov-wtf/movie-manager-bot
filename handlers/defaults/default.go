package defaults

import (
	"errors"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/helpers/messages"
	"github.com/erkinov-wtf/movie-manager-bot/helpers/utils"
	"github.com/erkinov-wtf/movie-manager-bot/models"
	"github.com/erkinov-wtf/movie-manager-bot/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/storage/database"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"log"

	"strings"
)

func (*defaultHandler) Start(context telebot.Context) error {
	log.Print(messages.StartCommand)

	var existingUser models.User
	if err := database.DB.Where("id = ?", context.Sender().ID).First(&existingUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			btn := &telebot.ReplyMarkup{}
			btnRows := []telebot.Row{
				btn.Row(btn.Data("✅ I Agree", "", fmt.Sprint("default|start|"))),
			}

			btn.Inline(btnRows...)
			err = context.Send(messages.PrivacyPolicy, &telebot.SendOptions{
				ParseMode:   telebot.ModeMarkdown,
				ReplyMarkup: btn,
			})
			if err != nil {
				log.Printf(err.Error())
				return context.Send(messages.InternalError)
			}
		} else {
			log.Printf("Database error: %v", err)
			return context.Send(messages.InternalError)
		}
	} else {
		err = context.Send(messages.UseHelp, telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return context.Send(messages.InternalError)
		}
	}

	return nil
}

func (*defaultHandler) handleStartCallback(context telebot.Context) error {
	newUser := models.User{
		ID:         context.Sender().ID,
		FirstName:  &context.Sender().FirstName,
		LastName:   &context.Sender().LastName,
		Username:   &context.Sender().Username,
		Language:   &context.Sender().LanguageCode,
		TmdbApiKey: nil,
	}

	if err := database.DB.Create(&newUser).Error; err != nil {
		log.Printf("cant create user: %v", err.Error())
		return context.Send(messages.InternalError)
	}

	return context.Send(messages.Registered, telebot.ModeMarkdown)
}

func (h *defaultHandler) GetToken(context telebot.Context) error {
	userID := context.Sender().ID
	isActive, userCache := cache.UserCache.Get(userID)

	if isActive && !userCache.ApiToken.IsTokenWaiting {
		return context.Send(messages.TokenAlreadyExists, telebot.ModeMarkdown)
	}

	return context.Send(messages.TokenInstructions, telebot.ModeMarkdown)
}

func (h *defaultHandler) HandleTextInput(context telebot.Context) error {
	userID := context.Sender().ID
	inputText := context.Message().Text

	if !utils.TestApiToken(inputText) {
		return context.Send(messages.TokenTestFailed, telebot.ModeMarkdown)
	}

	if err := database.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("tmdb_api_key", inputText).Error; err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	cache.UserCache.UpdateTokenState(userID, false)
	return context.Send(messages.TokenSaved, telebot.ModeMarkdown)
}

func (h *defaultHandler) DefaultCallback(context telebot.Context) error {
	callback := context.Callback()
	trimmed := strings.TrimSpace(callback.Data)

	if !strings.HasPrefix(trimmed, "default|") {
		return context.Send(messages.InternalError)
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) != 3 {
		log.Printf("Received malformed callback data: %s", callback.Data)
		return context.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	action := dataParts[1]
	switch action {
	case "start":
		return h.handleStartCallback(context)

	default:
		return context.Respond(&telebot.CallbackResponse{Text: messages.UnknownAction})
	}
}
