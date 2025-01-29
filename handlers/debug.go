package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/storage/cache"
	"gopkg.in/telebot.v3"
	"log"
	"net/http"
	"time"
)

type VersionResponse struct {
	Version string `json:"Version"`
	Time    string `json:"Time"`
}

func getLatestBotVersion() VersionResponse {
	url := "https://proxy.golang.org/github.com/erkinov-wtf/movie-manager-bot/@latest"

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Failed to fetch latest version:", err)
		return VersionResponse{
			Version: "Unknown",
			Time:    time.Now().String(),
		}
	}
	defer resp.Body.Close()

	var responseData VersionResponse
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return VersionResponse{
			Version: "Unknown",
			Time:    time.Now().String(),
		}
	}

	return responseData
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
			"*🔹 Bot Deployed Time:* `%s`\n"+
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
		botVersion.Version, botVersion.Time, time.Now().Format("2006-01-02 15:04:05"),
		user.ID, user.Username, user.FirstName, user.LastName,
		message.Text, message.Payload, message.Time().Format("2006-01-02 15:04:05"),
	)

	// Retrieve cache data
	cacheValue, cacheExpired, token := cache.UserCache.Get(context.Sender().ID)
	debugMessage += fmt.Sprintf(
		"*📦 Cache Info:*\n"+
			"• *Cache Value:* `%v`\n"+
			"• *Cache Active:* `%v`\n"+
			"• *Token Waiting:* `%v`\n"+
			"• *Token:* `%v`\n"+
			"——————————————",
		cacheValue, cacheExpired, token.IsTokenWaiting, token.Token,
	)

	// Send message with Markdown formatting
	return context.Send(debugMessage, &telebot.SendOptions{ParseMode: "Markdown"})
}
