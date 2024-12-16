package info

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"movie-manager-bot/models"
	"movie-manager-bot/storage/database"
	"strconv"
	"strings"
)

func (*infoHandler) Info(context telebot.Context) error {
	log.Print("/info command received")

	msg, err := context.Bot().Send(context.Chat(), "Loading...")
	if err != nil {
		log.Print(err)
		return err
	}

	btn := &telebot.ReplyMarkup{}
	btnRows := []telebot.Row{
		btn.Row(btn.Data("📺 TV Shows", "", fmt.Sprintf("info|tv_info|%d", msg.ID))),
		btn.Row(btn.Data("🎥 Movies", "", fmt.Sprintf("info|movie_info|%d", msg.ID))),
	}

	btn.Inline(btnRows...)

	_, err = context.Bot().Edit(msg, "What you want info about?", btn)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (i *infoHandler) handleTVDetails(context telebot.Context, msgId string) error {
	var watchedShows []models.TVShows

	if err := database.DB.Find(&watchedShows).Error; err != nil {
		log.Printf("cant get all tv shows: %v", err.Error())
		return err
	}

	info := tvStats{}

	for _, s := range watchedShows {
		info.amount++
		info.totalTime += s.Runtime
	}

	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: context.Chat()}

	formattedTime := formatDuration(info.totalTime)
	text := fmt.Sprintf(`📺 *TV Shows - Total Info*

📊 *Statistics:*
└ 📝 Shows Watched: *%d*
└ 🕙 Total Time Wasted: *%d*
└ ⌛️ Time Breakdown: *%s*

🎯 *Achievement:* You've spent *%d* minutes watching TV shows! Keep ruining your precious time! 👍`,
		info.amount,
		info.totalTime,
		formattedTime,
		info.totalTime,
	)

	_, err := context.Bot().Edit(msg, text, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (i *infoHandler) handleMovieDetails(context telebot.Context, msgId string) error {
	var watchedMovies []models.Movie

	if err := database.DB.Find(&watchedMovies).Error; err != nil {
		log.Printf("cant get all tv movies: %v", err.Error())
		return err
	}

	info := movieStats{}

	for _, s := range watchedMovies {
		info.amount++
		info.totalTime += s.Runtime
	}

	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: context.Chat()}

	formattedTime := formatDuration(info.totalTime)
	text := fmt.Sprintf(`📺 *Movies - Total Info*

📊 *Statistics:*
└ 📝 Movies Watched: *%d*
└ 🕙 Total Time Wasted: *%d*
└ ⌛️ Time Breakdown: *%s*

🎯 *Achievement:* You've spent *%d* minutes watching movies! Keep ruining your precious time! 👍`,
		info.amount,
		info.totalTime,
		formattedTime,
		info.totalTime,
	)

	_, err := context.Bot().Edit(msg, text, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (i *infoHandler) InfoCallback(context telebot.Context) error {
	callback := context.Callback()
	trimmed := strings.TrimSpace(callback.Data)

	if !strings.HasPrefix(trimmed, "info|") {
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
	case "movie_info":
		return i.handleMovieDetails(context, data)
	case "tv_info":
		return i.handleTVDetails(context, data)
	default:
		return context.Respond(&telebot.CallbackResponse{Text: "Unknown action"})
	}
}

func formatDuration(minutes int64) string {
	days := minutes / (24 * 60)
	remainingMinutes := minutes % (24 * 60)
	hours := remainingMinutes / 60
	mins := remainingMinutes % 60

	return fmt.Sprintf("%d days - %d hours - %d minutes", days, hours, mins)
}
