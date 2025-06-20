package docker

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != DefaultMaxRetries {
		t.Errorf("DefaultRetryConfig() MaxRetries = %v, want %v", config.MaxRetries, DefaultMaxRetries)
	}
	if config.BaseDelay != DefaultRetryDelay {
		t.Errorf("DefaultRetryConfig() BaseDelay = %v, want %v", config.BaseDelay, DefaultRetryDelay)
	}
	if config.MaxDelay != MaxRetryDelay {
		t.Errorf("DefaultRetryConfig() MaxDelay = %v, want %v", config.MaxDelay, MaxRetryDelay)
	}
	if config.Backoff == nil {
		t.Error("DefaultRetryConfig() Backoff is nil")
	}
}

func TestExponentialBackoff(t *testing.T) {
	tests := []struct {
		name      string
		attempt   int
		baseDelay time.Duration
		want      time.Duration
	}{
		{
			name:      "初回リトライ",
			attempt:   0,
			baseDelay: time.Second,
			want:      time.Second,
		},
		{
			name:      "2回目のリトライ",
			attempt:   1,
			baseDelay: time.Second,
			want:      2 * time.Second,
		},
		{
			name:      "3回目のリトライ",
			attempt:   2,
			baseDelay: time.Second,
			want:      4 * time.Second,
		},
		{
			name:      "最大遅延を超える場合",
			attempt:   10,
			baseDelay: time.Second,
			want:      MaxRetryDelay,
		},
		{
			name:      "小さなベース遅延",
			attempt:   3,
			baseDelay: 100 * time.Millisecond,
			want:      800 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExponentialBackoff(tt.attempt, tt.baseDelay)
			if got != tt.want {
				t.Errorf("ExponentialBackoff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinearBackoff(t *testing.T) {
	tests := []struct {
		name      string
		attempt   int
		baseDelay time.Duration
		want      time.Duration
	}{
		{
			name:      "初回リトライ",
			attempt:   0,
			baseDelay: time.Second,
			want:      time.Second,
		},
		{
			name:      "2回目のリトライ",
			attempt:   1,
			baseDelay: time.Second,
			want:      2 * time.Second,
		},
		{
			name:      "3回目のリトライ",
			attempt:   2,
			baseDelay: time.Second,
			want:      3 * time.Second,
		},
		{
			name:      "最大遅延を超える場合",
			attempt:   100,
			baseDelay: time.Second,
			want:      MaxRetryDelay,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LinearBackoff(tt.attempt, tt.baseDelay)
			if got != tt.want {
				t.Errorf("LinearBackoff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithRetry(t *testing.T) {
	tests := []struct {
		name          string
		operation     func() func() error
		config        RetryConfig
		contextFunc   func() (context.Context, context.CancelFunc)
		wantErr       bool
		wantAttempts  int
		checkErrorMsg string
	}{
		{
			name: "初回で成功",
			operation: func() func() error {
				attempts := 0
				return func() error {
					attempts++
					return nil
				}
			},
			config: RetryConfig{
				MaxRetries: 3,
				BaseDelay:  10 * time.Millisecond,
				MaxDelay:   100 * time.Millisecond,
				Backoff:    ExponentialBackoff,
			},
			contextFunc:  func() (context.Context, context.CancelFunc) { return context.WithTimeout(context.Background(), time.Second) },
			wantErr:      false,
			wantAttempts: 1,
		},
		{
			name: "2回目で成功",
			operation: func() func() error {
				var attempts int
				return func() error {
					attempts++
					if attempts < 2 {
						return errors.New("temporary error")
					}
					return nil
				}
			},
			config: RetryConfig{
				MaxRetries: 3,
				BaseDelay:  10 * time.Millisecond,
				MaxDelay:   100 * time.Millisecond,
				Backoff:    ExponentialBackoff,
			},
			contextFunc:  func() (context.Context, context.CancelFunc) { return context.WithTimeout(context.Background(), time.Second) },
			wantErr:      false,
			wantAttempts: 2,
		},
		{
			name: "すべてのリトライが失敗",
			operation: func() func() error {
				return func() error {
					return errors.New("persistent error")
				}
			},
			config: RetryConfig{
				MaxRetries: 2,
				BaseDelay:  10 * time.Millisecond,
				MaxDelay:   100 * time.Millisecond,
				Backoff:    ExponentialBackoff,
			},
			contextFunc:   func() (context.Context, context.CancelFunc) { return context.WithTimeout(context.Background(), time.Second) },
			wantErr:       true,
			wantAttempts:  3, // 初回 + 2回のリトライ
			checkErrorMsg: "operation failed after 2 retries",
		},
		{
			name: "回復不可能なエラー",
			operation: func() func() error {
				return func() error {
					return &DockerError{
						Operation:   "test",
						Cause:       errors.New("fatal error"),
						Recoverable: false,
					}
				}
			},
			config: RetryConfig{
				MaxRetries: 3,
				BaseDelay:  10 * time.Millisecond,
				MaxDelay:   100 * time.Millisecond,
				Backoff:    ExponentialBackoff,
			},
			contextFunc:   func() (context.Context, context.CancelFunc) { return context.WithTimeout(context.Background(), time.Second) },
			wantErr:       true,
			wantAttempts:  1,
			checkErrorMsg: "non-recoverable error",
		},
		{
			name: "コンテキストがキャンセルされる",
			operation: func() func() error {
				attempts := 0
				return func() error {
					attempts++
					return errors.New("error")
				}
			},
			config: RetryConfig{
				MaxRetries: 3,
				BaseDelay:  100 * time.Millisecond,
				MaxDelay:   100 * time.Millisecond,
				Backoff:    ExponentialBackoff,
			},
			contextFunc: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()
				return ctx, cancel
			},
			wantErr:       true,
			checkErrorMsg: "operation canceled",
		},
		{
			name: "リトライ遅延中にコンテキストがキャンセル",
			operation: func() func() error {
				return func() error {
					return errors.New("error")
				}
			},
			config: RetryConfig{
				MaxRetries: 3,
				BaseDelay:  200 * time.Millisecond,
				MaxDelay:   200 * time.Millisecond,
				Backoff:    LinearBackoff,
			},
			contextFunc: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				return ctx, cancel
			},
			wantErr:       true,
			checkErrorMsg: "operation canceled during retry",
		},
		{
			name: "線形バックオフ",
			operation: func() func() error {
				attempts := 0
				return func() error {
					attempts++
					if attempts < 3 {
						return errors.New("temporary error")
					}
					return nil
				}
			},
			config: RetryConfig{
				MaxRetries: 3,
				BaseDelay:  10 * time.Millisecond,
				MaxDelay:   100 * time.Millisecond,
				Backoff:    LinearBackoff,
			},
			contextFunc:  func() (context.Context, context.CancelFunc) { return context.WithTimeout(context.Background(), time.Second) },
			wantErr:      false,
			wantAttempts: 3,
		},
		{
			name: "操作実行前にコンテキストがキャンセル",
			operation: func() func() error {
				var attempts int
				return func() error {
					attempts++
					time.Sleep(50 * time.Millisecond) // 操作に時間がかかる
					return errors.New("error")
				}
			},
			config: RetryConfig{
				MaxRetries: 3,
				BaseDelay:  10 * time.Millisecond,
				MaxDelay:   100 * time.Millisecond,
				Backoff:    ExponentialBackoff,
			},
			contextFunc: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // 即座にキャンセル
				return ctx, cancel
			},
			wantErr:       true,
			checkErrorMsg: "operation canceled during retry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.contextFunc()
			defer cancel()

			operation := tt.operation()

			err := WithRetry(ctx, operation, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("WithRetry() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkErrorMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.checkErrorMsg) {
					t.Errorf("WithRetry() error message = %v, want to contain %v", err.Error(), tt.checkErrorMsg)
				}
			}
		})
	}
}

func TestWithRetryDefault(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	operation := func() error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary error")
		}
		return nil
	}

	err := WithRetryDefault(ctx, operation)
	if err != nil {
		t.Errorf("WithRetryDefault() error = %v, want nil", err)
	}

	if attempts != 2 {
		t.Errorf("WithRetryDefault() attempts = %v, want 2", attempts)
	}
}

func TestWithRetrySimple(t *testing.T) {
	tests := []struct {
		name       string
		maxRetries int
		operation  func() func() error
		wantErr    bool
	}{
		{
			name:       "カスタムリトライ回数で成功",
			maxRetries: 5,
			operation: func() func() error {
				attempts := 0
				return func() error {
					attempts++
					if attempts < 4 {
						return errors.New("temporary error")
					}
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:       "リトライ回数0",
			maxRetries: 0,
			operation: func() func() error {
				attempts := 0
				return func() error {
					attempts++
					if attempts == 1 {
						return nil
					}
					return errors.New("should not retry")
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := WithRetrySimple(ctx, tt.operation(), tt.maxRetries)

			if (err != nil) != tt.wantErr {
				t.Errorf("WithRetrySimple() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

