// Package logging は構造化ロギングのための統一的なインターフェースを提供します
package logging

import (
	"context"
)

// loggerContextKey はcontextにロガーを格納するためのキーの型
type loggerContextKey struct{}

// loggerKey はcontextにロガーを格納するためのキー
var loggerKey = loggerContextKey{}

// WithContext はcontextにロガーを格納する
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext はcontextからロガーを取得する
// ロガーが存在しない場合はnopロガーを返す
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	// contextにロガーがない場合はnopロガーを返す
	return newNopLogger()
}
