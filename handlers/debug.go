package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/storage/cache"
	"gopkg.in/telebot.v3"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type GitHubTag struct {
	Name string `json:"name"`
}

func getLatestBotVersion() string {
	url := "https://api.github.com/repos/erkinov-wtf/movie-manager-bot/tags"

	resp, err := http.Get(url)
	if err != nil {
		log.Println("Failed to fetch tags from GitHub:", err)
		return "Unknown"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("GitHub API returned non-200 status:", resp.StatusCode)
		return "Unknown"
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read response body:", err)
		return "Unknown"
	}

	var tags []GitHubTag
	if err := json.Unmarshal(body, &tags); err != nil {
		log.Println("Failed to parse GitHub tags JSON:", err)
		return "Unknown"
	}

	if len(tags) == 0 {
		return "Unknown"
	}

	return tags[0].Name
}

// DebugMessage function
func DebugMessage(context telebot.Context) error {
	user := context.Sender()
	message := context.Message()

	log.Printf("Debug Info - Timestamp: %v", time.Now())

	botVersion := getLatestBotVersion()

	// Format debug response
	debugMessage := fmt.Sprintf(
		"*🛠 Debug Info*\n"+
			"——————————————\n"+
			"*🔹 Bot Version:* `%s`\n"+
			"*🔹 Timestamp:* `%s`\n"+
			"\n"+
			"*👤 User Info:*\n"+
			"• *ID:* `%d`\n"+
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
	isActive, userCache := cache.UserCache.Get(context.Sender().ID)
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

	// Send message with Markdown formatting
	return context.Send(debugMessage, &telebot.SendOptions{ParseMode: "Markdown"})
}
