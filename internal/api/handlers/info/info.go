package info

import (
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"gopkg.in/telebot.v3"
	"log"
	"time"

	"strconv"
	"strings"
)

func (h *InfoHandler) Info(ctx telebot.Context) error {
	log.Print(messages.InfoCommand)

	msg, err := ctx.Bot().Send(ctx.Chat(), messages.Loading)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	btn := &telebot.ReplyMarkup{}
	btnRows := []telebot.Row{
		btn.Row(btn.Data("📺 TV Shows", "", fmt.Sprintf("info|tv_info|%d", msg.ID))),
		btn.Row(btn.Data("🎥 Movies", "", fmt.Sprintf("info|movie_info|%d", msg.ID))),
		btn.Row(btn.Data("🍿 Full Info", "", fmt.Sprintf("info|full_info|%d", msg.ID))),
	}

	btn.Inline(btnRows...)

	_, err = ctx.Bot().Edit(msg, messages.InfoFirstMessage, btn)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	return nil
}

func (h *InfoHandler) handleTVDetails(ctx telebot.Context, msgId string) error {
	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	watchedShows, err := h.app.Repository.TVShows.GetUserTVShows(ctxDb, ctx.Sender().ID)
	if err != nil {
		log.Printf("cant get all tv shows: %v", err.Error())
		return ctx.Send(messages.InternalError)
	}

	info := tvStats{}

	for _, s := range watchedShows {
		info.amount++
		info.totalTime += s.Runtime
	}

	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: ctx.Chat()}

	formattedTime := formatDuration(info.totalTime)
	text := fmt.Sprintf(`📺 *TV Shows - Total Info*

📊 *Statistics:*
└ 📝 Shows Watched: *%d*
└ 🕙 Total Time Wasted: *%d* minutes
└ ⌛️ Time Breakdown: *%s*

🎯 *Achievement:* You've spent *%d* hours watching TV shows! Keep ruining your precious time! 👍`,
		info.amount,
		info.totalTime,
		formattedTime,
		info.totalTime/60,
	)

	_, err = ctx.Bot().Edit(msg, text, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	return nil
}

func (h *InfoHandler) handleMovieDetails(ctx telebot.Context, msgId string) error {
	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	watchedMovies, err := h.app.Repository.Movies.GetUserMovies(ctxDb, ctx.Sender().ID)
	if err != nil {
		log.Printf("cant get all tv movies: %v", err.Error())
		return ctx.Send(messages.InternalError)
	}

	info := movieStats{}

	for _, s := range watchedMovies {
		info.amount++
		info.totalTime += s.Runtime
	}

	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: ctx.Chat()}

	formattedTime := formatDuration(info.totalTime)
	text := fmt.Sprintf(`📺 *Movies - Total Info*

📊 *Statistics:*
└ 📝 Movies Watched: *%d*
└ 🕙 Total Time Wasted: *%d* minutes
└ ⌛️ Time Breakdown: *%s*

🎯 *Achievement:* You've spent *%d* hours watching movies! Keep ruining your precious time! 👍`,
		info.amount,
		info.totalTime,
		formattedTime,
		info.totalTime/60,
	)

	_, err = ctx.Bot().Edit(msg, text, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	return nil
}

func (h *InfoHandler) handleFullDetails(ctx telebot.Context, data string) error {
	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	watchedMovies, err := h.app.Repository.Movies.GetUserMovies(ctxDb, ctx.Sender().ID)
	if err != nil {
		log.Printf("cant get all movies: %v", err.Error())
		return ctx.Send(messages.InternalError)
	}

	watchedShows, err := h.app.Repository.TVShows.GetUserTVShows(ctxDb, ctx.Sender().ID)
	if err != nil {
		log.Printf("cant get all tv shows: %v", err.Error())
		return ctx.Send(messages.InternalError)
	}

	movieInfo := movieStats{}
	for _, s := range watchedMovies {
		movieInfo.amount++
		movieInfo.totalTime += s.Runtime
	}

	// TV show stats
	tvInfo := tvStats{}
	for _, s := range watchedShows {
		tvInfo.amount++
		tvInfo.totalTime += s.Runtime
	}

	// Formatting the data
	movieFormattedTime := formatDuration(movieInfo.totalTime)
	tvFormattedTime := formatDuration(tvInfo.totalTime)
	totalTime := movieInfo.totalTime + tvInfo.totalTime
	totalFormattedTime := formatDuration(totalTime)

	// Create the message
	msgID, _ := strconv.Atoi(data)
	msg := &telebot.Message{ID: msgID, Chat: ctx.Chat()}

	text := fmt.Sprintf(`📺 *Full Info - Total Details*

🎥 *Movies - Total Info*
📊 *Statistics:*
└ 📝 Movies Watched: *%d*
└ 🕙 Total Time Wasted: *%d* minutes
└ ⌛️ Time Breakdown: *%s*

📺 *TV Shows - Total Info*
📊 *Statistics:*
└ 📝 Shows Watched: *%d*
└ 🕙 Total Time Wasted: *%d* minutes
└ ⌛️ Time Breakdown: *%s*

🎯 *Total Info:*
└ 📝 Total Movies + TV Shows Watched: *%d*
└ 🕙 Total Time Wasted: *%d* minutes
└ ⌛️ Total Time Breakdown: *%s*

🎯 *Achievement:* You've spent *%d* hours watching movies and TV shows! Keep ruining your precious time! 👍`,
		movieInfo.amount,
		movieInfo.totalTime,
		movieFormattedTime,
		tvInfo.amount,
		tvInfo.totalTime,
		tvFormattedTime,
		movieInfo.amount+tvInfo.amount,
		totalTime,
		totalFormattedTime,
		totalTime/60,
	)

	_, err = ctx.Bot().Edit(msg, text, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	return nil
}

func (h *InfoHandler) InfoCallback(ctx telebot.Context) error {
	callback := ctx.Callback()
	trimmed := strings.TrimSpace(callback.Data)

	if !strings.HasPrefix(trimmed, "info|") {
		return ctx.Send(messages.InternalError)
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) != 3 {
		log.Printf("Received malformed callback data: %s", callback.Data)
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	action := dataParts[1]
	data := dataParts[2]

	switch action {
	case "movie_info":
		return h.handleMovieDetails(ctx, data)

	case "tv_info":
		return h.handleTVDetails(ctx, data)

	case "full_info":
		return h.handleFullDetails(ctx, data)

	default:
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.UnknownAction})
	}
}

func formatDuration(minutes int32) string {
	days := minutes / (24 * 60)
	remainingMinutes := minutes % (24 * 60)
	hours := remainingMinutes / 60
	mins := remainingMinutes % 60

	return fmt.Sprintf("%d days - %d hours - %d minutes", days, hours, mins)
}
