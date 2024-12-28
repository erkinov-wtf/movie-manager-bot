package defaults

import (
	"errors"
	"fmt"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"log"
	"movie-manager-bot/helpers/messages"
	"movie-manager-bot/models"
	"movie-manager-bot/storage/database"
	"strconv"
	"strings"
)

func (*defaultHandler) Start(context telebot.Context) error {
	log.Print(messages.StartCommand)

	var existingUser models.User
	if err := database.DB.Where("id = ?", context.Sender().ID).First(&existingUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			btn := &telebot.ReplyMarkup{}
			btnRows := []telebot.Row{
				btn.Row(btn.Data("✅ I Agree", "", fmt.Sprintf("default|start|%d", context.Sender().ID))),
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
		err = context.Send(messages.UseHelp)
		if err != nil {
			log.Print(err)
			return context.Send(messages.InternalError)
		}
	}

	return nil
}

func (*defaultHandler) handleStartCallback(context telebot.Context, userId string) error {
	parsedId, err := strconv.Atoi(userId)
	if err != nil {
		log.Printf("cant convert id: %v", err.Error())
		return context.Send(messages.InternalError)
	}

	newUser := models.User{
		ID:        int64(parsedId),
		FirstName: &context.Sender().FirstName,
		LastName:  &context.Sender().LastName,
		Username:  &context.Sender().Username,
		Language:  &context.Sender().LanguageCode,
	}

	if err = database.DB.Create(&newUser).Error; err != nil {
		log.Printf("cant create user: %v", err.Error())
		return context.Send(messages.InternalError)
	}

	err = context.Send(messages.Registered)
	if err != nil {
		log.Printf(err.Error())
		return context.Send(messages.InternalError)
	}

	return nil
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
	data := dataParts[2]

	switch action {
	case "start":
		return h.handleStartCallback(context, data)
	default:
		return context.Respond(&telebot.CallbackResponse{Text: messages.UnknownAction})
	}
}
