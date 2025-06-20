// Package logging は構造化ロギングのための統一的なインターフェースを提供します
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
func (n *nopLogger) Debug(_ context.Context, _ string, _ ...Field) {}

// Info は何もしない
func (n *nopLogger) Info(_ context.Context, _ string, _ ...Field) {}

// Warn は何もしない
func (n *nopLogger) Warn(_ context.Context, _ string, _ ...Field) {}

// Error は何もしない
func (n *nopLogger) Error(_ context.Context, _ string, _ ...Field) {}

// With は自身を返す
func (n *nopLogger) With(_ ...Field) Logger {
	return n
}

// Named は自身を返す
func (n *nopLogger) Named(_ string) Logger {
	return n
}
