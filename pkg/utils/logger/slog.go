package logger

import (
	"github.com/erkinov-wtf/movie-manager-bot/pkg/constants"
	"log/slog"
	"os"
)

type Logger struct {
	internal *slog.Logger
	env      string
}

func setupLogger(baseLogger *slog.Logger, env string) *Logger {
	return &Logger{
		internal: baseLogger,
		env:      env,
	}
}

func NewLogger(env string) *Logger {
	var log *slog.Logger

	switch env {
	case constants.LocalEnv:
		opts := PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			},
		}
		handler := opts.NewPrettyHandler(os.Stdout)
		log = slog.New(handler)

	case constants.Prod:
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		})
		log = slog.New(handler)
	}

	return setupLogger(log, env)
}
