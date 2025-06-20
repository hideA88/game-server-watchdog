package logging

import (
	"context"
)

// nopLogger は何もしないロガー
type nopLogger struct{}

// newNopLogger は新しいnopLoggerを作成
func newNopLogger() Logger {
	return &nopLogger{}
}

// Debug は何もしない
func (n *nopLogger) Debug(ctx context.Context, msg string, fields ...Field) {}

// Info は何もしない
func (n *nopLogger) Info(ctx context.Context, msg string, fields ...Field) {}

// Warn は何もしない
func (n *nopLogger) Warn(ctx context.Context, msg string, fields ...Field) {}

// Error は何もしない
func (n *nopLogger) Error(ctx context.Context, msg string, fields ...Field) {}

// With は自身を返す
func (n *nopLogger) With(fields ...Field) Logger {
	return n
}

// Named は自身を返す
func (n *nopLogger) Named(name string) Logger {
	return n
}
