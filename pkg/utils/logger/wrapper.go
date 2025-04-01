package logger

import (
	"github.com/erkinov-wtf/movie-manager-bot/pkg/constants"
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

// Common helper function to build log args
func (l *Logger) buildLogArgs(op string, ctx telebot.Context, args ...interface{}) []interface{} {
	logArgs := []interface{}{
		"op", op,
	}

	// Only include user info in production environment
	if l.env == constants.Prod && ctx != nil && ctx.Sender() != nil {
		user := parseCtxSender(ctx.Sender())
		logArgs = append(logArgs, "user", user)
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

	return logArgs
}

func (l *Logger) Info(op string, ctx telebot.Context, msg string, args ...interface{}) {
	logArgs := l.buildLogArgs(op, ctx, args...)
	l.internal.Info(msg, logArgs...)
}

func (l *Logger) Error(op string, ctx telebot.Context, msg string, args ...interface{}) {
	logArgs := l.buildLogArgs(op, ctx, args...)
	l.internal.Error(msg, logArgs...)
}

func (l *Logger) Debug(op string, ctx telebot.Context, msg string, args ...interface{}) {
	logArgs := l.buildLogArgs(op, ctx, args...)
	l.internal.Debug(msg, logArgs...)
}

func (l *Logger) Warning(op string, ctx telebot.Context, msg string, args ...interface{}) {
	logArgs := l.buildLogArgs(op, ctx, args...)
	l.internal.Warn(msg, logArgs...)
}

// Worker-specific logging methods that don't need telebot.Context
func parseArgs(op string, args ...interface{}) []interface{} {
	allArgs := []interface{}{"op", op, "source", "worker"}

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

func (l *Logger) WorkerInfo(op string, msg string, args ...interface{}) {
	prefixedMsg := "[WORKER] " + msg
	allArgs := parseArgs(op, args...)
	l.internal.Info(prefixedMsg, allArgs...)
}

func (l *Logger) WorkerError(op string, msg string, args ...interface{}) {
	prefixedMsg := "[WORKER] " + msg
	allArgs := parseArgs(op, args...)
	l.internal.Error(prefixedMsg, allArgs...)
}

func (l *Logger) WorkerDebug(op string, msg string, args ...interface{}) {
	prefixedMsg := "[WORKER] " + msg
	allArgs := parseArgs(op, args...)
	l.internal.Debug(prefixedMsg, allArgs...)
}

func (l *Logger) WorkerWarning(op string, msg string, args ...interface{}) {
	prefixedMsg := "[WORKER] " + msg
	allArgs := parseArgs(op, args...)
	l.internal.Warn(prefixedMsg, allArgs...)
}
