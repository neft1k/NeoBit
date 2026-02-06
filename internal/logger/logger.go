package logger

import (
	"context"
	"fmt"

	"NeoBIT/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field struct {
	Key   string
	Value any
}

func FieldAny(key string, value any) Field {
	return Field{Key: key, Value: value}
}

func Nop() Logger {
	return &zapLogger{logger: zap.NewNop()}
}

func New(cfg config.LoggerConfig) (Logger, error) {
	zapCfg := zap.NewProductionConfig()
	if cfg.Development {
		zapCfg = zap.NewDevelopmentConfig()
	}

	if cfg.Level != "" {
		if err := zapCfg.Level.UnmarshalText([]byte(cfg.Level)); err != nil {
			return nil, fmt.Errorf("invalid log level: %w", err)
		}
	}

	if cfg.Encoding != "" {
		zapCfg.Encoding = cfg.Encoding
	}
	zapCfg.EncoderConfig.TimeKey = "ts"
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if len(cfg.OutputPaths) > 0 {
		zapCfg.OutputPaths = cfg.OutputPaths
	}
	if len(cfg.ErrorOutputPaths) > 0 {
		zapCfg.ErrorOutputPaths = cfg.ErrorOutputPaths
	}

	l, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	return &zapLogger{logger: l}, nil
}

type zapLogger struct {
	logger *zap.Logger
}

func (l *zapLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	zap.S().Info()
	l.logger.WithOptions(zap.AddCallerSkip(1)).Debug(msg, toZapFields(fields)...)
}

func (l *zapLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.logger.WithOptions(zap.AddCallerSkip(1)).Info(msg, toZapFields(fields)...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.logger.WithOptions(zap.AddCallerSkip(1)).Warn(msg, toZapFields(fields)...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.logger.WithOptions(zap.AddCallerSkip(1)).Error(msg, toZapFields(fields)...)
}

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{logger: l.logger.With(toZapFields(fields)...)}
}

func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

func toZapFields(fields []Field) []zap.Field {
	zFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zFields = append(zFields, zap.Any(f.Key, f.Value))
	}
	return zFields
}
