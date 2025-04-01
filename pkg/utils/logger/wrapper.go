package logger

import (
	"gopkg.in/telebot.v3"
	"strconv"
)

type customUser struct {
	ID        string
	FullName  string
	Username  string
	Language  string
	IsPremium bool
}

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

func (l *Logger) Info(op string, ctx telebot.Context, msg string, args ...interface{}) {
	user := parseCtxSender(ctx.Sender())

	logArgs := []interface{}{
		"op", op,
		"user", user,
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

		logArgs = append(logArgs, "others", others)
	}

	l.internal.Info(msg, logArgs...)
}

func (l *Logger) Error(op string, ctx telebot.Context, msg string, args ...interface{}) {
	user := parseCtxSender(ctx.Sender())

	logArgs := []interface{}{
		"op", op,
		"user", user,
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

		logArgs = append(logArgs, "others", others)
	}

	l.internal.Error(msg, logArgs...)
}

func (l *Logger) Debug(op string, ctx telebot.Context, msg string, args ...interface{}) {
	user := parseCtxSender(ctx.Sender())

	logArgs := []interface{}{
		"op", op,
		"user", user,
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

		logArgs = append(logArgs, "others", others)
	}

	l.internal.Debug(msg, logArgs...)
}

func (l *Logger) Warning(op string, ctx telebot.Context, msg string, args ...interface{}) {
	user := parseCtxSender(ctx.Sender())

	logArgs := []interface{}{
		"op", op,
		"user", user,
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

		logArgs = append(logArgs, "others", others)
	}

	l.internal.Warn(msg, logArgs...)
}
