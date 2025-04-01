package logger

import (
	"github.com/erkinov-wtf/movie-manager-bot/pkg/constants"
	"gopkg.in/telebot.v3"
	"strconv"
)

// customUser holds user information for logging
type customUser struct {
	ID        string
	FullName  string
	Username  string
	Language  string
	IsPremium bool
}

// parseCtxSender extracts user information from telebot context
func parseCtxSender(user *telebot.User) *customUser {
	if user == nil {
		return nil
	}

	fullName := user.FirstName
	if user.LastName != "" {
		fullName += " " + user.LastName
	}

	return &customUser{
		ID:        strconv.FormatInt(user.ID, 10),
		FullName:  fullName,
		Username:  user.Username,
		Language:  user.LanguageCode,
		IsPremium: user.IsPremium,
	}
}

// buildLogArgs builds log arguments with context
func buildLogArgs(op string, env string, ctx telebot.Context, args ...interface{}) []interface{} {
	logArgs := []interface{}{
		"op", op,
	}

	// Only include user info in production environment
	if env == constants.Prod && ctx != nil && ctx.Sender() != nil {
		user := parseCtxSender(ctx.Sender())
		logArgs = append(logArgs, "user", user)
	}

	// Process additional arguments
	if len(args) > 0 {
		others := make(map[string]interface{})

		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				if key, ok := args[i].(string); ok {
					others[key] = args[i+1]
				}
			}
		}

		if len(others) > 0 {
			logArgs = append(logArgs, "others", others)
		}
	}

	return logArgs
}

// buildWorkerArgs builds arguments for worker logs
func buildWorkerArgs(op string, args ...interface{}) []interface{} {
	allArgs := []interface{}{
		"op", op,
		"source", "worker",
	}

	if len(args) > 0 {
		others := make(map[string]interface{})

		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				if key, ok := args[i].(string); ok {
					others[key] = args[i+1]
				}
			}
		}

		if len(others) > 0 {
			allArgs = append(allArgs, "others", others)
		}
	}

	return allArgs
}
