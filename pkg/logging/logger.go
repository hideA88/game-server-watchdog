// Package logging は構造化ロギングのための統一的なインターフェースを提供します
package logging

import (
	"context"
	"fmt"
	"strings"
)

// Logger は構造化ログのインターフェース
type Logger interface {
	// ログレベル（コンテキスト付き）
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)

	// フィールドを追加した新しいロガーを作成
	With(fields ...Field) Logger

	// 子ロガーを作成（名前付き）
	Named(name string) Logger
}

// Field は構造化ログのフィールドを表す
type Field struct {
	Key   string
	Value interface{}
}

// String はstring型のフィールドを作成します
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int はint型のフィールドを作成します
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 はint64型のフィールドを作成します
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 はfloat64型のフィールドを作成します
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool はbool型のフィールドを作成します
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// ErrorField はerror型のフィールドを作成します
func ErrorField(err error) Field {
	return Field{Key: "error", Value: err}
}

// Any は任意の型のフィールドを作成します
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Level はログレベルを表す型です
type Level int

const (
	// DebugLevel はデバッグレベルを表します
	DebugLevel Level = iota
	// InfoLevel は情報レベルを表します
	InfoLevel
	// WarnLevel は警告レベルを表します
	WarnLevel
	// ErrorLevel はエラーレベルを表します
	ErrorLevel
)

// ログレベル文字列定数
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return LogLevelDebug
	case InfoLevel:
		return LogLevelInfo
	case WarnLevel:
		return LogLevelWarn
	case ErrorLevel:
		return LogLevelError
	default:
		return "unknown"
	}
}

// ParseLevel は文字列からLevelに変換
// 無効な値の場合はInfoLevelを返す
func ParseLevel(level string) Level {
	switch level {
	case LogLevelDebug:
		return DebugLevel
	case LogLevelInfo:
		return InfoLevel
	case LogLevelWarn:
		return WarnLevel
	case LogLevelError:
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// ParseLevelCaseInsensitive は文字列からLevelに変換（大文字小文字を区別しない）
// 無効な値の場合はInfoLevelを返す
func ParseLevelCaseInsensitive(level string) Level {
	return ParseLevel(strings.ToLower(level))
}

// ParseLevelWithWarning は文字列からLevelに変換し、無効な値の場合はエラーを返す
func ParseLevelWithWarning(level string) (Level, error) {
	switch level {
	case LogLevelDebug:
		return DebugLevel, nil
	case LogLevelInfo:
		return InfoLevel, nil
	case LogLevelWarn:
		return WarnLevel, nil
	case LogLevelError:
		return ErrorLevel, nil
	case "":
		// 空文字列は有効（デフォルト値を使用）
		return InfoLevel, nil
	default:
		return InfoLevel, fmt.Errorf("invalid log level: %s", level)
	}
}

// ParseLevelCaseInsensitiveWithWarning は文字列からLevelに変換（大文字小文字を区別しない）
// 無効な値の場合はエラーを返す
func ParseLevelCaseInsensitiveWithWarning(level string) (Level, error) {
	return ParseLevelWithWarning(strings.ToLower(level))
}

// Config はロガーの設定
type Config struct {
	// Level はログレベル
	Level Level

	// Development は開発モード（より詳細なログ）
	Development bool

	// Format はログフォーマット（json, console）
	Format string

	// OutputPaths は出力先のパス（stdout, stderr, ファイルパス）
	OutputPaths []string

	// ErrorOutputPaths はエラー出力先のパス
	ErrorOutputPaths []string
}

// DefaultConfig はデフォルト設定を返す
func DefaultConfig() *Config {
	return &Config{
		Level:            InfoLevel,
		Development:      false,
		Format:           "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}
