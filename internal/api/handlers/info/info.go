package info

import (
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"gopkg.in/telebot.v3"
	"strconv"
	"strings"
	"time"
)

func (h *InfoHandler) Info(ctx telebot.Context) error {
	const op = "info.Info"
	h.app.Logger.Info(op, ctx, "Info command received")

	msg, err := ctx.Bot().Send(ctx.Chat(), messages.Loading)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to send loading message", "error", err.Error())
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
		h.app.Logger.Error(op, ctx, "Failed to edit message with info menu", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "Info command handled successfully")
	return nil
}

func (h *InfoHandler) handleTVDetails(ctx telebot.Context, msgId string) error {
	const op = "info.handleTVDetails"
	h.app.Logger.Info(op, ctx, "Processing TV details request", "message_id", msgId)

	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	h.app.Logger.Debug(op, ctx, "Retrieving user TV shows from database")
	watchedShows, err := h.app.Repository.TVShows.GetUserTVShows(ctxDb, ctx.Sender().ID)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to retrieve user TV shows", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	info := tvStats{}
	h.app.Logger.Debug(op, ctx, "Calculating TV show statistics")
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
		h.app.Logger.Error(op, ctx, "Failed to update message with TV statistics", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "TV statistics displayed successfully")
	return nil
}

func (h *InfoHandler) handleMovieDetails(ctx telebot.Context, msgId string) error {
	const op = "info.handleMovieDetails"
	h.app.Logger.Info(op, ctx, "Processing movie details request", "message_id", msgId)

	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	h.app.Logger.Debug(op, ctx, "Retrieving user movies from database")
	watchedMovies, err := h.app.Repository.Movies.GetUserMovies(ctxDb, ctx.Sender().ID)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to retrieve user movies", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	info := movieStats{}
	h.app.Logger.Debug(op, ctx, "Calculating movie statistics")
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

	h.app.Logger.Debug(op, ctx, "Updating message with movie statistics",
		"movies_count", info.amount, "total_minutes", info.totalTime)
	_, err = ctx.Bot().Edit(msg, text, telebot.ModeMarkdown)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to update message with movie statistics", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "Movie statistics displayed successfully")
	return nil
}

func (h *InfoHandler) handleFullDetails(ctx telebot.Context, data string) error {
	const op = "info.handleFullDetails"
	h.app.Logger.Info(op, ctx, "Processing full details request", "message_id", data)

	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	h.app.Logger.Debug(op, ctx, "Retrieving user movies from database")
	watchedMovies, err := h.app.Repository.Movies.GetUserMovies(ctxDb, ctx.Sender().ID)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to retrieve user movies", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Debug(op, ctx, "Retrieving user TV shows from database")
	watchedShows, err := h.app.Repository.TVShows.GetUserTVShows(ctxDb, ctx.Sender().ID)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to retrieve user TV shows", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Debug(op, ctx, "Calculating movie statistics")
	movieInfo := movieStats{}
	for _, s := range watchedMovies {
		movieInfo.amount++
		movieInfo.totalTime += s.Runtime
	}

	h.app.Logger.Debug(op, ctx, "Calculating TV show statistics")
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

	h.app.Logger.Debug(op, ctx, "Updating message with full statistics",
		"movies_count", movieInfo.amount,
		"tv_count", tvInfo.amount,
		"total_minutes", totalTime)
	_, err = ctx.Bot().Edit(msg, text, telebot.ModeMarkdown)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to update message with full statistics", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "Full statistics displayed successfully")
	return nil
}

func (h *InfoHandler) InfoCallback(ctx telebot.Context) error {
	const op = "info.InfoCallback"
	callback := ctx.Callback()
	trimmed := strings.TrimSpace(callback.Data)
	h.app.Logger.Info(op, ctx, "Processing info callback", "callback_data", trimmed)

	if !strings.HasPrefix(trimmed, "info|") {
		h.app.Logger.Warning(op, ctx, "Invalid callback prefix", "callback_data", trimmed)
		return ctx.Send(messages.InternalError)
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) != 3 {
		h.app.Logger.Warning(op, ctx, "Malformed callback data", "callback_data", callback.Data,
			"parts_count", len(dataParts))
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	action := dataParts[1]
	data := dataParts[2]
	h.app.Logger.Debug(op, ctx, "Processing callback action", "action", action, "data", data)

	switch action {
	case "movie_info":
		return h.handleMovieDetails(ctx, data)

	case "tv_info":
		return h.handleTVDetails(ctx, data)

	case "full_info":
		return h.handleFullDetails(ctx, data)

	default:
		h.app.Logger.Warning(op, ctx, "Unknown callback action", "action", action)
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
