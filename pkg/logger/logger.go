package logger

import (
	"context"

	"go.uber.org/zap"
)

type ContextLoggerKey string

var (
	// Feel free to override these variables in your application.
	LoggerContextKey ContextLoggerKey = "LoggerKey"
)

// Create a new logger with the given options.
func NewLogger(isDebug bool, opts ...zap.Option) *zap.SugaredLogger {
	var l *zap.Logger
	if isDebug {
		l, _ = zap.NewDevelopment(opts...)
	} else {
		l, _ = zap.NewProduction(opts...)
	}

	zap.ReplaceGlobals(l)
	return zap.S()
}

func LoggerFromContextOrDefault(ctx context.Context) *zap.SugaredLogger {
	l, _ := ctx.Value(LoggerContextKey).(*zap.SugaredLogger)
	if l == nil {
		l = zap.S()
	}
	return l
}

func NewContextWithLogger(parent context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(parent, LoggerContextKey, logger)
}
