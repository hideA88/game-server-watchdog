package logging

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger はzapを使ったLogger実装
type zapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// newZapLogger はzapを使った新しいLoggerを作成（内部使用）
func newZapLogger(config *Config) (Logger, error) {
	zapConfig := zap.NewProductionConfig()
	if config.Development {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// ログレベルの設定
	zapConfig.Level = zap.NewAtomicLevelAt(convertToZapLevel(config.Level))

	// エンコーディングの設定
	if config.Format == "console" {
		zapConfig.Encoding = "console"
	} else {
		zapConfig.Encoding = "json"
	}

	// 出力先の設定
	if len(config.OutputPaths) > 0 {
		zapConfig.OutputPaths = config.OutputPaths
	}
	if len(config.ErrorOutputPaths) > 0 {
		zapConfig.ErrorOutputPaths = config.ErrorOutputPaths
	}

	// タイムスタンプフォーマットの設定
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build zap logger: %w", err)
	}

	return &zapLogger{
		logger: logger,
		sugar:  logger.Sugar(),
	}, nil
}

func convertToZapLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func convertFieldsToZap(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		switch v := f.Value.(type) {
		case string:
			zapFields = append(zapFields, zap.String(f.Key, v))
		case int:
			zapFields = append(zapFields, zap.Int(f.Key, v))
		case int64:
			zapFields = append(zapFields, zap.Int64(f.Key, v))
		case float64:
			zapFields = append(zapFields, zap.Float64(f.Key, v))
		case bool:
			zapFields = append(zapFields, zap.Bool(f.Key, v))
		case error:
			zapFields = append(zapFields, zap.Error(v))
		default:
			zapFields = append(zapFields, zap.Any(f.Key, v))
		}
	}
	return zapFields
}

// ログレベル実装（コンテキスト付き）
func (l *zapLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	// 将来的にはコンテキストからトレースIDなどを抽出可能
	l.logger.Debug(msg, convertFieldsToZap(fields)...)
}

func (l *zapLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.logger.Info(msg, convertFieldsToZap(fields)...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.logger.Warn(msg, convertFieldsToZap(fields)...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.logger.Error(msg, convertFieldsToZap(fields)...)
}

// With はフィールドを追加した新しいロガーを作成
func (l *zapLogger) With(fields ...Field) Logger {
	zapFields := convertFieldsToZap(fields)
	return &zapLogger{
		logger: l.logger.With(zapFields...),
		sugar:  l.logger.With(zapFields...).Sugar(),
	}
}

// Named は子ロガーを作成
func (l *zapLogger) Named(name string) Logger {
	return &zapLogger{
		logger: l.logger.Named(name),
		sugar:  l.logger.Named(name).Sugar(),
	}
}
