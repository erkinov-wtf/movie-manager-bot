package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"gopkg.in/telebot.v3"
	"io"
	"net/http"
	"time"
)

type GitHubTag struct {
	Name string `json:"name"`
}

func getLatestBotVersion(app *app.App, url string) string {
	const op = "handlers.getLatestBotVersion"
	app.Logger.Debug(op, nil, "Fetching tags from GitHub API", "url", url)

	resp, err := http.Get(url)
	if err != nil {
		app.Logger.Error(op, nil, "Failed to fetch tags from GitHub", "error", err.Error())
		return "Unknown"
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		app.Logger.Error(op, nil, "GitHub API returned non-200 status", "status_code", resp.StatusCode)
		return "Unknown"
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Logger.Error(op, nil, "Failed to read response body", "error", err.Error())
		return "Unknown"
	}

	var tags []GitHubTag
	if err := json.Unmarshal(body, &tags); err != nil {
		app.Logger.Error(op, nil, "Failed to parse GitHub tags JSON", "error", err.Error())
		return "Unknown"
	}

	if len(tags) == 0 {
		app.Logger.Warning(op, nil, "No tags found in GitHub repository")
		return "Unknown"
	}

	app.Logger.Info(op, nil, "Successfully retrieved latest bot version", "version", tags[0].Name)
	return tags[0].Name
}

// DebugMessage function
func DebugMessage(context telebot.Context, app *app.App) error {
	const op = "handlers.DebugMessage"
	app.Logger.Info(op, context, "Debug command received")

	user := context.Sender()
	message := context.Message()

	app.Logger.Debug(op, context, "Fetching bot version information")
	botVersion := getLatestBotVersion(app, app.Cfg.VersionsUrl)

	// Format debug response
	app.Logger.Debug(op, context, "Generating debug message")
	debugMessage := fmt.Sprintf(
		"*🛠 Debug Info*\n"+
			"——————————————\n"+
			"*🔹 Bot Version:* `%s`\n"+
			"*🔹 Timestamp:* `%s`\n"+
			"\n"+
			"*👤 User Info:*\n"+
			"• *Id:* `%d`\n"+
			"• *Username:* `@%s`\n"+
			"• *First Name:* `%s`\n"+
			"• *Last Name:* `%s`\n"+
			"\n"+
			"*💬 Message Info:*\n"+
			"• *Text:* `%s`\n"+
			"• *Payload:* `%s`\n"+
			"• *Date:* `%s`\n"+
			"——————————————\n",
		botVersion, time.Now().Format("2006-01-02 15:04:05"),
		user.ID, user.Username, user.FirstName, user.LastName,
		message.Text, message.Payload, message.Time().Format("2006-01-02 15:04:05"),
	)

	// Retrieve cache data
	app.Logger.Debug(op, context, "Retrieving user cache data", "user_id", user.ID)
	isActive, userCache := app.Cache.UserCache.Get(context.Sender().ID)
	if isActive {
		app.Logger.Debug(op, context, "User cache is active",
			"search_state_movie", userCache.SearchState.IsMovieSearch,
			"search_state_tv", userCache.SearchState.IsTVShowSearch)
		debugMessage += fmt.Sprintf(
			"*📦 Cache Info:*\n"+
				"• *Cache Active:* `%v`\n"+
				"• *Cache Value:* `%v`\n"+
				"• *Token Waiting:* `%v`\n"+
				"• *Token:* `%v`\n"+
				"• *Movie Search:* `%v`\n"+
				"• *TV Show Search:* `%v`\n"+
				"——————————————",
			isActive, userCache.Value, userCache.ApiToken.IsTokenWaiting, userCache.ApiToken.Token, userCache.SearchState.IsMovieSearch, userCache.SearchState.IsTVShowSearch,
		)
	} else {
		app.Logger.Debug(op, context, "No user cache found")
		debugMessage += fmt.Sprintf(
			"*📦 Cache Info:*\n" +
				"*No cache data found*\n" +
				"——————————————",
		)
	}

	// Send message with Markdown formatting
	app.Logger.Debug(op, context, "Sending debug message to user")
	err := context.Send(debugMessage, &telebot.SendOptions{ParseMode: "Markdown"})
	if err != nil {
		app.Logger.Error(op, context, "Failed to send debug message", "error", err.Error())
	} else {
		app.Logger.Info(op, context, "Debug message sent successfully")
	}

	return err
}
