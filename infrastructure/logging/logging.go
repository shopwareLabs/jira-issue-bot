package logging

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey struct{}

var fallbackLogger *zap.SugaredLogger

func WithLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func NewLeveledLogger(logger *zap.SugaredLogger) *LeveledLogger {
	return &LeveledLogger{logger: logger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar()}
}

func FromContext(ctx context.Context) *zap.SugaredLogger {
	if logger, ok := ctx.Value(loggerKey{}).(*zap.SugaredLogger); ok {
		return logger
	}

	if fallbackLogger == nil {
		loggerCfg := zap.NewProductionConfig()
		logger, _ := loggerCfg.Build()

		fallbackLogger = logger.Sugar()
	}

	return fallbackLogger
}

// LeveledLogger interface implements the basic methods that a logger library needs.
type LeveledLogger struct {
	logger *zap.SugaredLogger
}

func (l *LeveledLogger) Error(msg string, keysAndVals ...interface{}) {
	l.logger.Errorw(msg, keysAndVals...)
}

func (l *LeveledLogger) Info(msg string, keysAndVals ...interface{}) {
	l.logger.Infow(msg, keysAndVals...)
}

func (l *LeveledLogger) Debug(msg string, keysAndVals ...interface{}) {
	l.logger.Debugw(msg, keysAndVals...)
}

func (l *LeveledLogger) Warn(msg string, keysAndVals ...interface{}) {
	l.logger.Warnw(msg, keysAndVals...)
}

func (l *LeveledLogger) Log(msg string) {
	l.logger.Infow(msg)
}
