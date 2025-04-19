package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(env string) *zap.Logger {
	var cfg zap.Config

	switch env {
	case "local":
		cfg = zap.NewDevelopmentConfig()
		cfg.Encoding = "json"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

	case "prod":
		cfg = zap.NewProductionConfig()
		cfg.Encoding = "json"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	default:
		cfg = zap.NewProductionConfig()
		cfg.Encoding = "json"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil
	}

	return logger.With(zap.String("service", "RatingService"))
}
