package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func InitLogger(debug bool) {
	var level zapcore.Level
	if debug {
		level = zapcore.DebugLevel
	} else {
		level = zapcore.InfoLevel
	}

	cfg := zap.Config{
		Encoding:         "console",
		Level:            zap.NewAtomicLevelAt(level),
		OutputPaths:      []string{"stdout", "pkg/logger/app.log"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    zap.NewProductionEncoderConfig(),
	}

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	logger, _ := cfg.Build()
	Log = logger
}

/*

 logger.Error("Ошибка бла бла бла", zap.Error(err))
 logger.Info("бла бла бла", zap.String("бла", "бла")) // стринги
 logger.Warn("бла бла бла", zap.String("бла", "бла"))

*/
