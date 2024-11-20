package config

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(cfg Config) *zap.SugaredLogger {
	loggerCfg := zap.NewProductionConfig()
	loggerCfg.EncoderConfig.EncodeTime = timeEncoder()

	loggerCfg.EncoderConfig.MessageKey = "message"
	loggerCfg.EncoderConfig.TimeKey = "timestamp"
	loggerCfg.EncoderConfig.EncodeDuration = zapcore.NanosDurationEncoder
	loggerCfg.DisableStacktrace = !cfg.Debug

	if cfg.Debug {
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	logger, err := loggerCfg.Build()
	if err != nil {
		logger = zap.NewNop()
	}

	return logger.Sugar()
}

// timeEncoder encodes the time as RFC3339 nano.
func timeEncoder() zapcore.TimeEncoder {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339Nano))
	}
}
