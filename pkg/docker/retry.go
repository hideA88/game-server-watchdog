package docker

import (
	"context"
	"fmt"
	"time"
)

const (
	// DefaultMaxRetries はデフォルトの最大リトライ回数
	DefaultMaxRetries = 3

	// DefaultRetryDelay はデフォルトのリトライ間隔
	DefaultRetryDelay = time.Second

	// MaxRetryDelay は最大リトライ間隔
	MaxRetryDelay = 30 * time.Second
)

// RetryConfig はリトライの設定を表します
type RetryConfig struct {
	MaxRetries int           // 最大リトライ回数
	BaseDelay  time.Duration // 基本リトライ間隔
	MaxDelay   time.Duration // 最大リトライ間隔
	Backoff    BackoffFunc   // バックオフ戦略
}

// BackoffFunc はリトライ間隔を計算する関数の型
type BackoffFunc func(attempt int, baseDelay time.Duration) time.Duration

// DefaultRetryConfig はデフォルトのリトライ設定を返します
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: DefaultMaxRetries,
		BaseDelay:  DefaultRetryDelay,
		MaxDelay:   MaxRetryDelay,
		Backoff:    ExponentialBackoff,
	}
}

// ExponentialBackoff は指数バックオフを実装します
func ExponentialBackoff(attempt int, baseDelay time.Duration) time.Duration {
	// #nosec G115 - attempt is controlled and bounded by MaxRetries
	delay := baseDelay * time.Duration(1<<uint(attempt))
	if delay > MaxRetryDelay {
		return MaxRetryDelay
	}
	return delay
}

// LinearBackoff は線形バックオフを実装します
func LinearBackoff(attempt int, baseDelay time.Duration) time.Duration {
	delay := baseDelay * time.Duration(attempt+1)
	if delay > MaxRetryDelay {
		return MaxRetryDelay
	}
	return delay
}

// WithRetry は指定された操作をリトライ機能付きで実行します
func WithRetry(ctx context.Context, operation func() error, config RetryConfig) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// 操作を実行
		err := operation()
		if err == nil {
			return nil // 成功
		}

		lastErr = err

		// 回復不可能なエラーの場合はリトライしない
		if !IsRecoverable(err) {
			return fmt.Errorf("operation failed with non-recoverable error: %w", err)
		}

		// 最後の試行の場合はリトライしない
		if attempt == config.MaxRetries {
			break
		}

		// コンテキストがキャンセルされていないかチェック
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation canceled during retry: %w", ctx.Err())
		default:
		}

		// リトライ間隔を計算して待機
		delay := config.Backoff(attempt, config.BaseDelay)

		select {
		case <-ctx.Done():
			return fmt.Errorf("operation canceled during retry delay: %w", ctx.Err())
		case <-time.After(delay):
			// 次のリトライに進む
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", config.MaxRetries, lastErr)
}

// WithRetryDefault はデフォルト設定でリトライを実行します
func WithRetryDefault(ctx context.Context, operation func() error) error {
	return WithRetry(ctx, operation, DefaultRetryConfig())
}

// WithRetrySimple は簡単なリトライを実行します（最大回数のみ指定）
func WithRetrySimple(ctx context.Context, operation func() error, maxRetries int) error {
	config := DefaultRetryConfig()
	config.MaxRetries = maxRetries
	return WithRetry(ctx, operation, config)
}
