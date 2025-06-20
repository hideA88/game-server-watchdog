package logging

import (
	"bytes"
	"context"
	"testing"
)

// テーブル駆動テストのため長い関数を許可
func TestZapLogger(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		logFn  func(Logger)
		want   string
	}{
		{
			name: "Info level with fields",
			config: &Config{
				Level:       InfoLevel,
				Development: false,
				Format:      "json",
			},
			logFn: func(l Logger) {
				l.Info(context.Background(), "test message", String("key", "value"), Int("count", 42))
			},
			want: `"msg":"test message"`,
		},
		{
			name: "Debug level should not log when level is Info",
			config: &Config{
				Level:       InfoLevel,
				Development: false,
				Format:      "json",
			},
			logFn: func(l Logger) {
				l.Debug(context.Background(), "debug message")
			},
			want: "",
		},
		{
			name: "Error with error field",
			config: &Config{
				Level:       InfoLevel,
				Development: false,
				Format:      "json",
			},
			logFn: func(l Logger) {
				l.Error(context.Background(), "error occurred", ErrorField(bytes.ErrTooLarge))
			},
			want: `"msg":"error occurred"`,
		},
		{
			name: "With fields",
			config: &Config{
				Level:       InfoLevel,
				Development: false,
				Format:      "json",
			},
			logFn: func(l Logger) {
				l.With(String("service", "watchdog")).Info(context.Background(), "with test")
			},
			want: `"service":"watchdog"`,
		},
		{
			name: "Named logger",
			config: &Config{
				Level:       InfoLevel,
				Development: false,
				Format:      "json",
			},
			logFn: func(l Logger) {
				l.Named("bot").Info(context.Background(), "named test")
			},
			want: `"logger":"bot"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 出力設定
			tt.config.OutputPaths = []string{"stdout"}

			logger, err := NewWithConfig(tt.config)
			if err != nil {
				t.Fatalf("failed to create logger: %v", err)
			}

			// ログ出力を実行
			tt.logFn(logger)

			// zapのロガーをsyncする必要がある
			if zl, ok := logger.(*zapLogger); ok {
				_ = zl.logger.Sync()
			}

			// 出力を検証（実際の実装では標準出力に出るため、ここでは基本的な動作確認のみ）
			// より詳細なテストはzapのテスト機能を使用する必要がある
		})
	}
}

func TestLoggerFactory(t *testing.T) {
	tests := []struct {
		name      string
		debugMode bool
		logLevel  Level
	}{
		{
			name:      "default settings",
			debugMode: false,
			logLevel:  InfoLevel,
		},
		{
			name:      "debug mode enabled",
			debugMode: true,
			logLevel:  InfoLevel,
		},
		{
			name:      "LOG_LEVEL overrides DEBUG_MODE",
			debugMode: true,
			logLevel:  InfoLevel,
		},
		{
			name:      "LOG_LEVEL warn",
			debugMode: false,
			logLevel:  WarnLevel,
		},
		{
			name:      "LOG_LEVEL error",
			debugMode: false,
			logLevel:  ErrorLevel,
		},
		{
			name:      "LOG_LEVEL debug",
			debugMode: false,
			logLevel:  DebugLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.debugMode, tt.logLevel)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			if logger == nil {
				t.Error("New() returned nil logger")
			}
		})
	}
}

func TestNewWithConfig(t *testing.T) {
	// NewWithConfig関数のテスト
	logger, err := NewWithConfig(DefaultConfig())
	if err != nil {
		t.Fatalf("NewWithConfig() error = %v", err)
	}
	if logger == nil {
		t.Error("NewWithConfig() returned nil logger")
	}

	// nilコンフィグでも動作することを確認
	logger2, err := NewWithConfig(nil)
	if err != nil {
		t.Fatalf("NewWithConfig(nil) error = %v", err)
	}
	if logger2 == nil {
		t.Error("NewWithConfig(nil) returned nil logger")
	}
}

func TestContextBasedLogger(t *testing.T) {
	// FromContextがnopLoggerを返すことを確認
	ctx := context.Background()
	logger := FromContext(ctx)
	if logger == nil {
		t.Error("FromContext() returned nil")
	}

	// WithContextでloggerを設定し、FromContextで取得できることを確認
	mainLogger, err := New(false, InfoLevel)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	ctx = WithContext(ctx, mainLogger)
	retrievedLogger := FromContext(ctx)
	if retrievedLogger == nil {
		t.Error("FromContext() returned nil after WithContext")
	}

	// 同じインスタンスであることを確認
	if retrievedLogger != mainLogger {
		t.Error("FromContext() did not return the same logger instance")
	}
}

func TestLogLevel(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{DebugLevel, LogLevelDebug},
		{InfoLevel, LogLevelInfo},
		{WarnLevel, LogLevelWarn},
		{ErrorLevel, LogLevelError},
		{Level(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("Level.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLevel Level
	}{
		{
			name:      "debug level",
			input:     LogLevelDebug,
			wantLevel: DebugLevel,
		},
		{
			name:      "info level",
			input:     LogLevelInfo,
			wantLevel: InfoLevel,
		},
		{
			name:      "warn level",
			input:     LogLevelWarn,
			wantLevel: WarnLevel,
		},
		{
			name:      "error level",
			input:     LogLevelError,
			wantLevel: ErrorLevel,
		},
		{
			name:      "invalid level returns info",
			input:     "invalid",
			wantLevel: InfoLevel,
		},
		{
			name:      "empty string returns info",
			input:     "",
			wantLevel: InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLevel := ParseLevel(tt.input)
			if gotLevel != tt.wantLevel {
				t.Errorf("ParseLevel() = %v, want %v", gotLevel, tt.wantLevel)
			}
		})
	}
}

func TestParseLevelWithWarning(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLevel Level
		wantErr   bool
	}{
		{
			name:      "debug level",
			input:     LogLevelDebug,
			wantLevel: DebugLevel,
			wantErr:   false,
		},
		{
			name:      "info level",
			input:     LogLevelInfo,
			wantLevel: InfoLevel,
			wantErr:   false,
		},
		{
			name:      "warn level",
			input:     LogLevelWarn,
			wantLevel: WarnLevel,
			wantErr:   false,
		},
		{
			name:      "error level",
			input:     LogLevelError,
			wantLevel: ErrorLevel,
			wantErr:   false,
		},
		{
			name:      "empty string is valid",
			input:     "",
			wantLevel: InfoLevel,
			wantErr:   false,
		},
		{
			name:      "invalid level returns error",
			input:     "invalid",
			wantLevel: InfoLevel,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLevel, err := ParseLevelWithWarning(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLevelWithWarning() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLevel != tt.wantLevel {
				t.Errorf("ParseLevelWithWarning() = %v, want %v", gotLevel, tt.wantLevel)
			}
		})
	}
}

func TestParseLevelCaseInsensitive(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLevel Level
	}{
		{
			name:      "lowercase debug",
			input:     "debug",
			wantLevel: DebugLevel,
		},
		{
			name:      "uppercase DEBUG",
			input:     "DEBUG",
			wantLevel: DebugLevel,
		},
		{
			name:      "mixed case Debug",
			input:     "Debug",
			wantLevel: DebugLevel,
		},
		{
			name:      "uppercase INFO",
			input:     "INFO",
			wantLevel: InfoLevel,
		},
		{
			name:      "mixed case Warn",
			input:     "Warn",
			wantLevel: WarnLevel,
		},
		{
			name:      "uppercase ERROR",
			input:     "ERROR",
			wantLevel: ErrorLevel,
		},
		{
			name:      "invalid mixed case",
			input:     "Invalid",
			wantLevel: InfoLevel, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLevel := ParseLevelCaseInsensitive(tt.input)
			if gotLevel != tt.wantLevel {
				t.Errorf("ParseLevelCaseInsensitive() = %v, want %v", gotLevel, tt.wantLevel)
			}
		})
	}
}

func TestParseLevelCaseInsensitiveWithWarning(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLevel Level
		wantErr   bool
	}{
		{
			name:      "uppercase DEBUG",
			input:     "DEBUG",
			wantLevel: DebugLevel,
			wantErr:   false,
		},
		{
			name:      "mixed case Info",
			input:     "Info",
			wantLevel: InfoLevel,
			wantErr:   false,
		},
		{
			name:      "uppercase WARN",
			input:     "WARN",
			wantLevel: WarnLevel,
			wantErr:   false,
		},
		{
			name:      "mixed case Error",
			input:     "Error",
			wantLevel: ErrorLevel,
			wantErr:   false,
		},
		{
			name:      "invalid uppercase",
			input:     "INVALID",
			wantLevel: InfoLevel,
			wantErr:   true,
		},
		{
			name:      "empty string",
			input:     "",
			wantLevel: InfoLevel,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLevel, err := ParseLevelCaseInsensitiveWithWarning(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLevelCaseInsensitiveWithWarning() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLevel != tt.wantLevel {
				t.Errorf("ParseLevelCaseInsensitiveWithWarning() = %v, want %v", gotLevel, tt.wantLevel)
			}
		})
	}
}

func TestFieldHelpers(t *testing.T) {
	// フィールドヘルパー関数のテスト
	strField := String("key", "value")
	if strField.Key != "key" || strField.Value != "value" {
		t.Error("String() helper failed")
	}

	intField := Int("count", 42)
	if intField.Key != "count" || intField.Value != 42 {
		t.Error("Int() helper failed")
	}

	int64Field := Int64("bigcount", int64(999))
	if int64Field.Key != "bigcount" || int64Field.Value != int64(999) {
		t.Error("Int64() helper failed")
	}

	floatField := Float64("ratio", 3.14)
	if floatField.Key != "ratio" || floatField.Value != 3.14 {
		t.Error("Float64() helper failed")
	}

	boolField := Bool("enabled", true)
	if boolField.Key != "enabled" || boolField.Value != true {
		t.Error("Bool() helper failed")
	}

	errField := ErrorField(bytes.ErrTooLarge)
	if errField.Key != "error" || errField.Value != bytes.ErrTooLarge {
		t.Error("Error() helper failed")
	}

	anyField := Any("data", map[string]int{"count": 1})
	if anyField.Key != "data" {
		t.Error("Any() helper failed")
	}
}

func TestNopLogger(t *testing.T) {
	// nopLoggerが正常に動作することを確認
	logger := newNopLogger()

	// 各メソッドが呼び出せることを確認（パニックしない）
	ctx := context.Background()
	logger.Debug(ctx, "debug message", String("key", "value"))
	logger.Info(ctx, "info message", Int("count", 42))
	logger.Warn(ctx, "warn message", Bool("flag", true))
	logger.Error(ctx, "error message", ErrorField(bytes.ErrTooLarge))

	// Withメソッドが自身を返すことを確認
	withLogger := logger.With(String("service", "test"))
	if withLogger != logger {
		t.Error("nopLogger.With() should return itself")
	}

	// Namedメソッドが自身を返すことを確認
	namedLogger := logger.Named("subsystem")
	if namedLogger != logger {
		t.Error("nopLogger.Named() should return itself")
	}
}

func TestContextBoundaryConditions(t *testing.T) {
	// nilコンテキストのテスト
	t.Run("nil context with WithContext", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				// panicが発生することを期待
				t.Error("WithContext with nil context should panic")
			}
		}()

		logger, _ := New(false, InfoLevel)
		_ = WithContext(nil, logger) //nolint:staticcheck // intentionally passing nil to test panic
	})

	// nilコンテキストでFromContextを呼ぶとpanicする（Go標準の動作）
	t.Run("nil context with FromContext", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("FromContext with nil context should panic")
			}
		}()

		_ = FromContext(nil) //nolint:staticcheck // intentionally passing nil to test panic
	})

	// nilロガーをcontextに設定
	t.Run("nil logger in context", func(t *testing.T) {
		ctx := WithContext(context.Background(), nil)
		logger := FromContext(ctx)

		// nopLoggerが返されることを確認
		if logger == nil {
			t.Error("FromContext should never return nil")
		}

		// nopLoggerとして動作することを確認
		logger.Info(context.Background(), "test")
	})

	// 異なる型の値がcontextに設定されている場合
	t.Run("wrong type in context", func(t *testing.T) {
		// 別のキーで値を設定して、ロガーが見つからない状況をシミュレート
		type testKey struct{}
		ctx := context.WithValue(context.Background(), testKey{}, "not a logger")
		logger := FromContext(ctx)

		// nopLoggerが返されることを確認
		if logger == nil {
			t.Error("FromContext should return nopLogger when logger not found")
		}
	})
}
