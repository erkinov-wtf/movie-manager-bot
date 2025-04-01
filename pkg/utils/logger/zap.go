package logger

import (
	"github.com/erkinov-wtf/movie-manager-bot/pkg/constants"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// internalLogger interface abstracts the actual logging implementation
type internalLogger interface {
	info(msg string, args ...interface{})
	error(msg string, args ...interface{})
	debug(msg string, args ...interface{})
	warn(msg string, args ...interface{})
	sync() error
}

// zapLogger implements internalLogger using Zap
type zapLogger struct {
	logger        *zap.Logger
	sugar         *zap.SugaredLogger
	useStructured bool
}

// newZapLogger creates a new Zap-based logger
func newZapLogger(env string) *zapLogger {
	var zapLog *zap.Logger
	var useStructured bool

	// Common encoder config customizations
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        zapcore.OmitKey, // Omit logger name
		CallerKey:      zapcore.OmitKey, // Omit caller information
		FunctionKey:    zapcore.OmitKey, // Omit function name
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	switch env {
	case constants.LocalEnv:
		// Development config with console encoder for better readability
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		config := zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
			Development:      true,
			Encoding:         "console",
			EncoderConfig:    encoderConfig,
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}

		zapLog, _ = config.Build()
		useStructured = false // Use sugared logger in development

	case constants.Prod:
		// Production config with JSON encoder
		config := zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			Development:      false,
			Encoding:         "json",
			EncoderConfig:    encoderConfig,
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}

		zapLog, _ = config.Build()
		useStructured = true // Use structured logger in production
	default:
		// Fallback to development config
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.CallerKey = zapcore.OmitKey
		zapLog, _ = config.Build()
		useStructured = false
	}

	return &zapLogger{
		logger:        zapLog,
		sugar:         zapLog.Sugar(),
		useStructured: useStructured,
	}
}

// info logs an info level message
func (z *zapLogger) info(msg string, args ...interface{}) {
	if z.useStructured {
		fields := convertToZapFields(args)
		z.logger.Info(msg, fields...)
	} else {
		z.sugar.Infow(msg, args...)
	}
}

// error logs an error level message
func (z *zapLogger) error(msg string, args ...interface{}) {
	if z.useStructured {
		fields := convertToZapFields(args)
		z.logger.Error(msg, fields...)
	} else {
		z.sugar.Errorw(msg, args...)
	}
}

// debug logs a debug level message
func (z *zapLogger) debug(msg string, args ...interface{}) {
	if z.useStructured {
		fields := convertToZapFields(args)
		z.logger.Debug(msg, fields...)
	} else {
		z.sugar.Debugw(msg, args...)
	}
}

// warn logs a warning level message
func (z *zapLogger) warn(msg string, args ...interface{}) {
	if z.useStructured {
		fields := convertToZapFields(args)
		z.logger.Warn(msg, fields...)
	} else {
		z.sugar.Warnw(msg, args...)
	}
}

// sync flushes buffered logs
func (z *zapLogger) sync() error {
	return z.logger.Sync()
}

// convertToZapFields converts args to zap.Field slice
func convertToZapFields(args []interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(args)/2)

	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			if key, ok := args[i].(string); ok {
				fields = append(fields, zap.Any(key, args[i+1]))
			}
		}
	}

	return fields
}
