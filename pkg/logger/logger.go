package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger
type Logger struct {
	*zap.Logger
}

func NewLogger(debug bool) *Logger {
	var level zapcore.Level
	if debug {
		level = zapcore.DebugLevel
	} else {
		level = zapcore.InfoLevel
	}

	cfg := zap.Config{
		Encoding:         "console",
		Level:            zap.NewAtomicLevelAt(level),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    zap.NewProductionEncoderConfig(),
	}

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	logger, _ := cfg.Build()

	return &Logger{Logger: logger}
}

func (l *Logger) InfoS(msg string, keysAndValues ...interface{}) {
	l.Logger.Info(msg, l.fieldsFromArgs(keysAndValues...)...)
}

func (l *Logger) ErrorS(msg string, keysAndValues ...interface{}) {
	l.Logger.Error(msg, l.fieldsFromArgs(keysAndValues...)...)
}

func (l *Logger) WarnS(msg string, keysAndValues ...interface{}) {
	l.Logger.Warn(msg, l.fieldsFromArgs(keysAndValues...)...)
}

func (l *Logger) DebugS(msg string, keysAndValues ...interface{}) {
	l.Logger.Debug(msg, l.fieldsFromArgs(keysAndValues...)...)
}

func (l *Logger) FatalS(msg string, keysAndValues ...interface{}) {
	l.Logger.Fatal(msg, l.fieldsFromArgs(keysAndValues...)...)
}

func (l *Logger) fieldsFromArgs(keysAndValues ...interface{}) []zap.Field {
	if len(keysAndValues)%2 != 0 {
		l.Warn("Odd number of arguments passed to logging method")
		return []zap.Field{}
	}

	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			l.Warn("Non-string key passed to logging method", zap.Any("key", keysAndValues[i]))
			continue
		}
		fields = append(fields, zap.Any(key, keysAndValues[i+1]))
	}

	return fields
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{Logger: l.Logger.With(zap.Any(key, value))}
}

func (l *Logger) WithFields(keysAndValues ...interface{}) *Logger {
	return &Logger{Logger: l.Logger.With(l.fieldsFromArgs(keysAndValues...)...)}
}
