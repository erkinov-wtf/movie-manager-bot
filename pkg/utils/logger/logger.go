package logger

import (
	"github.com/erkinov-wtf/movie-manager-bot/pkg/constants"
	"go.uber.org/zap/zapcore"
	"gopkg.in/telebot.v3"
)

// Logger is a wrapper around zap.Logger/SugaredLogger
type Logger struct {
	internal    internalLogger
	betterStack *BetterStackLogger
	env         string
}

// NewLogger creates a new Logger instance based on environment
func NewLogger(env, host, token string) *Logger {
	var logLevel zapcore.Level

	if env == constants.Prod {
		logLevel = zapcore.InfoLevel
	} else {
		logLevel = zapcore.DebugLevel
	}

	baseLogger := newZapLogger(env)

	logger := &Logger{
		internal: baseLogger,
		env:      env,
	}

	// Only enable BetterStack in production
	if env == constants.Prod {
		logger.betterStack = newBetterStackLogger(baseLogger, logLevel, host, token)
	}

	return logger
}

// getLogger returns the appropriate logger based on configuration
func (l *Logger) getLogger() internalLogger {
	if l.env == constants.Prod && l.betterStack != nil {
		return l.betterStack
	}
	return l.internal
}

// Info logs an info message
func (l *Logger) Info(op string, ctx telebot.Context, msg string, args ...interface{}) {
	logArgs := buildLogArgs(op, l.env, ctx, args...)
	l.getLogger().info(msg, logArgs...)
}

// Error logs an error message
func (l *Logger) Error(op string, ctx telebot.Context, msg string, args ...interface{}) {
	logArgs := buildLogArgs(op, l.env, ctx, args...)
	l.getLogger().error(msg, logArgs...)
}

// Debug logs a debug message
func (l *Logger) Debug(op string, ctx telebot.Context, msg string, args ...interface{}) {
	logArgs := buildLogArgs(op, l.env, ctx, args...)
	l.getLogger().debug(msg, logArgs...)
}

// Warning logs a warning message
func (l *Logger) Warning(op string, ctx telebot.Context, msg string, args ...interface{}) {
	logArgs := buildLogArgs(op, l.env, ctx, args...)
	l.getLogger().warn(msg, logArgs...)
}

// WorkerInfo logs worker info message
func (l *Logger) WorkerInfo(op string, msg string, args ...interface{}) {
	prefixedMsg := "[WORKER] " + msg
	workerArgs := buildWorkerArgs(op, args...)
	l.getLogger().info(prefixedMsg, workerArgs...)
}

// WorkerError logs worker error message
func (l *Logger) WorkerError(op string, msg string, args ...interface{}) {
	prefixedMsg := "[WORKER] " + msg
	workerArgs := buildWorkerArgs(op, args...)
	l.getLogger().error(prefixedMsg, workerArgs...)
}

// WorkerDebug logs worker debug message
func (l *Logger) WorkerDebug(op string, msg string, args ...interface{}) {
	prefixedMsg := "[WORKER] " + msg
	workerArgs := buildWorkerArgs(op, args...)
	l.getLogger().debug(prefixedMsg, workerArgs...)
}

// WorkerWarning logs worker warning message
func (l *Logger) WorkerWarning(op string, msg string, args ...interface{}) {
	prefixedMsg := "[WORKER] " + msg
	workerArgs := buildWorkerArgs(op, args...)
	l.getLogger().warn(prefixedMsg, workerArgs...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.getLogger().sync()
}

// Stop gracefully shuts down the logger and flushes logs
func (l *Logger) Stop() {
	if l.env == constants.Prod && l.betterStack != nil {
		l.betterStack.Stop()
	}
	l.Sync()
}
