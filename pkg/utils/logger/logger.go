package logger

import (
	"gopkg.in/telebot.v3"
)

// Logger is a wrapper around zap.Logger/SugaredLogger
type Logger struct {
	internal internalLogger
	env      string
}

// NewLogger creates a new Logger instance based on environment
func NewLogger(env string) *Logger {
	return &Logger{
		internal: newZapLogger(env),
		env:      env,
	}
}

// Info logs an info message
func (l *Logger) Info(op string, ctx telebot.Context, msg string, args ...interface{}) {
	logArgs := buildLogArgs(op, l.env, ctx, args...)
	l.internal.info(msg, logArgs...)
}

// Error logs an error message
func (l *Logger) Error(op string, ctx telebot.Context, msg string, args ...interface{}) {
	logArgs := buildLogArgs(op, l.env, ctx, args...)
	l.internal.error(msg, logArgs...)
}

// Debug logs a debug message
func (l *Logger) Debug(op string, ctx telebot.Context, msg string, args ...interface{}) {
	logArgs := buildLogArgs(op, l.env, ctx, args...)
	l.internal.debug(msg, logArgs...)
}

// Warning logs a warning message
func (l *Logger) Warning(op string, ctx telebot.Context, msg string, args ...interface{}) {
	logArgs := buildLogArgs(op, l.env, ctx, args...)
	l.internal.warn(msg, logArgs...)
}

// WorkerInfo logs worker info message
func (l *Logger) WorkerInfo(op string, msg string, args ...interface{}) {
	prefixedMsg := "[WORKER] " + msg
	workerArgs := buildWorkerArgs(op, args...)
	l.internal.info(prefixedMsg, workerArgs...)
}

// WorkerError logs worker error message
func (l *Logger) WorkerError(op string, msg string, args ...interface{}) {
	prefixedMsg := "[WORKER] " + msg
	workerArgs := buildWorkerArgs(op, args...)
	l.internal.error(prefixedMsg, workerArgs...)
}

// WorkerDebug logs worker debug message
func (l *Logger) WorkerDebug(op string, msg string, args ...interface{}) {
	prefixedMsg := "[WORKER] " + msg
	workerArgs := buildWorkerArgs(op, args...)
	l.internal.debug(prefixedMsg, workerArgs...)
}

// WorkerWarning logs worker warning message
func (l *Logger) WorkerWarning(op string, msg string, args ...interface{}) {
	prefixedMsg := "[WORKER] " + msg
	workerArgs := buildWorkerArgs(op, args...)
	l.internal.warn(prefixedMsg, workerArgs...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.internal.sync()
}
