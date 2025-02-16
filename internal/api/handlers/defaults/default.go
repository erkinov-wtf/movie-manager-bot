package defaults

import (
	"errors"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/models"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/keyboards"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/utils"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"log"

	"strings"
)

func (h *DefaultHandler) Start(context telebot.Context) error {
	log.Print(messages.StartCommand)

	var existingUser models.User
	if err := h.app.Database.Where("id = ?", context.Sender().ID).First(&existingUser).Error; err != nil {
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
		menu := keyboards.LoadMenuKeyboards(context.Bot(), h.app)
		err = context.Send(messages.UseHelp, telebot.ModeMarkdown, menu)
		if err != nil {
			log.Print(err)
			return context.Send(messages.InternalError)
		}
	}

	return nil
}

func (h *DefaultHandler) handleStartCallback(context telebot.Context) error {
	newUser := models.User{
		Id:         context.Sender().ID,
		FirstName:  &context.Sender().FirstName,
		LastName:   &context.Sender().LastName,
		Username:   &context.Sender().Username,
		Language:   &context.Sender().LanguageCode,
		TmdbApiKey: nil,
	}

	if err := h.app.Database.Create(&newUser).Error; err != nil {
		log.Printf("cant create user: %v", err.Error())
		return context.Send(messages.InternalError)
	}

	keyboard := keyboards.LoadTokenRegistrationKeyboard(context.Bot(), h, h.app)
	return context.Send(messages.Registered, keyboard, telebot.ModeMarkdown)
}

func (h *DefaultHandler) GetToken(context telebot.Context) error {
	userId := context.Sender().ID
	isActive, userCache := h.app.Cache.UserCache.Get(userId)

	if isActive && !userCache.ApiToken.IsTokenWaiting {
		return context.Send(messages.TokenAlreadyExists, telebot.ModeMarkdown)
	}

	return context.Send(messages.TokenInstructions, telebot.ModeMarkdown)
}

func (h *DefaultHandler) HandleReplySearch(context telebot.Context) error {
	_, uc := h.app.Cache.UserCache.Get(context.Sender().ID)

	if uc.SearchState.IsTVShowSearch {
		return h.handleTVShowSearch(context)
	}
	return h.handleMovieSearch(context)
}

func (h *DefaultHandler) handleTVShowSearch(context telebot.Context) error {
	h.app.Cache.UserCache.SetSearchStartFalse(context.Sender().ID)
	return h.tvHandler.SearchTV(context)
}

func (h *DefaultHandler) handleMovieSearch(context telebot.Context) error {
	h.app.Cache.UserCache.SetSearchStartFalse(context.Sender().ID)
	return h.movieHandler.SearchMovie(context)
}

func (h *DefaultHandler) HandleTextInput(context telebot.Context) error {
	userId := context.Sender().ID
	inputText := context.Message().Text

	if !utils.TestApiToken(h.app, inputText) {
		return context.Send(messages.TokenTestFailed, telebot.ModeMarkdown)
	}

	if err := h.app.Database.Model(&models.User{}).
		Where("id = ?", userId).
		Update("tmdb_api_key", inputText).Error; err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	h.app.Cache.UserCache.UpdateTokenState(userId, false)
	menu := keyboards.LoadMenuKeyboards(context.Bot(), h.app)

	return context.Send(messages.TokenSaved, menu, telebot.ModeMarkdown)
}

func (h *DefaultHandler) DefaultCallback(context telebot.Context) error {
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
