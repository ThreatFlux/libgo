package logger

import (
	"os"

	"github.com/threatflux/libgo/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger implements Logger using zap
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new ZapLogger
func NewZapLogger(config config.LoggingConfig) (*ZapLogger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(config.Level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if config.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var output zapcore.WriteSyncer
	if config.FilePath == "" || config.FilePath == "stdout" {
		output = zapcore.AddSync(os.Stdout)
	} else if config.FilePath == "stderr" {
		output = zapcore.AddSync(os.Stderr)
	} else {
		file, err := os.OpenFile(config.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		output = zapcore.AddSync(file)
	}

	core := zapcore.NewCore(encoder, output, zapLevel)
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &ZapLogger{
		logger: zapLogger,
	}, nil
}

// Debug implements Logger.Debug
func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, l.zapFields(fields...)...)
}

// Info implements Logger.Info
func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, l.zapFields(fields...)...)
}

// Warn implements Logger.Warn
func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, l.zapFields(fields...)...)
}

// Error implements Logger.Error
func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, l.zapFields(fields...)...)
}

// Fatal implements Logger.Fatal
func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, l.zapFields(fields...)...)
}

// WithFields implements Logger.WithFields
func (l *ZapLogger) WithFields(fields ...Field) Logger {
	return &ZapLogger{
		logger: l.logger.With(l.zapFields(fields...)...),
	}
}

// WithError implements Logger.WithError
func (l *ZapLogger) WithError(err error) Logger {
	return &ZapLogger{
		logger: l.logger.With(zap.Error(err)),
	}
}

// Sync implements Logger.Sync
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

// zapFields converts our Fields to zap.Fields
func (l *ZapLogger) zapFields(fields ...Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}
