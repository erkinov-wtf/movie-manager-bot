package defaults

import (
	"errors"
	"fmt"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"log"
	"movie-manager-bot/models"
	"movie-manager-bot/storage/database"
	"strconv"
	"strings"
)

var (
	err error
)

func (*defaultHandler) Start(context telebot.Context) error {
	log.Print("/start command received")

	var existingUser models.User
	if err = database.DB.Where("id = ?", context.Sender().ID).First(&existingUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			btn := &telebot.ReplyMarkup{}
			btnRows := []telebot.Row{
				btn.Row(btn.Data("✅ I Agree", "", fmt.Sprintf("default|start|%d", context.Sender().ID))),
			}

			btn.Inline(btnRows...)
			message := "By using this bot, you agree to our [Privacy Policy](https://example.com/privacy-policy)"
			err := context.Send(message, &telebot.SendOptions{
				ParseMode:   telebot.ModeMarkdown,
				ReplyMarkup: btn,
			})
			if err != nil {
				log.Printf(err.Error())
				return err
			}
		} else {
			log.Printf("Database error: %v", err)
			return err
		}
	} else {
		err = context.Send("use /help for assistance")
		if err != nil {
			log.Print(err)
			return err
		}
	}

	return nil
}

func (*defaultHandler) handleStartCallback(context telebot.Context, userId string) error {
	parsedId, err := strconv.Atoi(userId)
	if err != nil {
		log.Printf("cant convert id: %v", err.Error())
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
		return err
	}

	err = context.Send("You have been successfully registered.\n Now you can use this bot. Use /help for assistance)")
	if err != nil {
		log.Printf(err.Error())
		return err
	}

	return nil
}

func (h *defaultHandler) DefaultCallback(context telebot.Context) error {
	callback := context.Callback()
	trimmed := strings.TrimSpace(callback.Data)

	if !strings.HasPrefix(trimmed, "default|") {
		return nil
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) != 3 {
		log.Printf("Received malformed callback data: %s", callback.Data)
		return context.Respond(&telebot.CallbackResponse{Text: "Malformed data received"})
	}

	action := dataParts[1]
	data := dataParts[2]

	switch action {
	case "start":
		return h.handleStartCallback(context, data)
	default:
		return context.Respond(&telebot.CallbackResponse{Text: "Unknown action"})
	}
}
