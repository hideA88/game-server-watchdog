package docker

import (
	"errors"
	"fmt"
)

// Docker操作に関連するエラー定義
var (
	// ErrDockerConnectionFailed はDockerデーモンへの接続に失敗した際のエラー
	ErrDockerConnectionFailed = errors.New("docker daemon connection failed")

	// ErrContainerNotFound は指定されたコンテナが見つからない際のエラー
	ErrContainerNotFound = errors.New("container not found")

	// ErrServiceNotFound は指定されたサービスが見つからない際のエラー
	ErrServiceNotFound = errors.New("service not found")

	// ErrOperationTimeout は操作がタイムアウトした際のエラー
	ErrOperationTimeout = errors.New("operation timed out")

	// ErrInvalidProjectName は無効なプロジェクト名が指定された際のエラー
	ErrInvalidProjectName = errors.New("invalid project name")

	// ErrDockerComposeFileNotFound はdocker-compose.ymlが見つからない際のエラー
	ErrDockerComposeFileNotFound = errors.New("docker-compose.yml not found")

	// ErrContainerNotRunning はコンテナが稼働していない際のエラー
	ErrContainerNotRunning = errors.New("container is not running")
)

// DockerError はDocker操作に関するエラーの詳細情報を含む構造体
//
//nolint:revive // Docker関連のエラーであることを明示するため
type DockerError struct {
	Operation   string // 失敗した操作名
	Container   string // 対象のコンテナ名（該当する場合）
	Service     string // 対象のサービス名（該当する場合）
	Cause       error  // 元のエラー
	Recoverable bool   // 回復可能かどうか
}

// Error はエラーメッセージを返します
func (e *DockerError) Error() string {
	if e.Container != "" {
		return fmt.Sprintf("docker operation '%s' failed for container '%s': %v",
			e.Operation, e.Container, e.Cause)
	}
	if e.Service != "" {
		return fmt.Sprintf("docker operation '%s' failed for service '%s': %v",
			e.Operation, e.Service, e.Cause)
	}
	return fmt.Sprintf("docker operation '%s' failed: %v", e.Operation, e.Cause)
}

// Unwrap は元のエラーを返します
func (e *DockerError) Unwrap() error {
	return e.Cause
}

// Is はエラーの種類を判定します
func (e *DockerError) Is(target error) bool {
	return errors.Is(e.Cause, target)
}

// NewDockerError は新しいDockerErrorを作成します
func NewDockerError(operation string, cause error) *DockerError {
	return &DockerError{
		Operation:   operation,
		Cause:       cause,
		Recoverable: isRecoverableError(cause),
	}
}

// NewDockerContainerError はコンテナ操作に関するDockerErrorを作成します
func NewDockerContainerError(operation, container string, cause error) *DockerError {
	return &DockerError{
		Operation:   operation,
		Container:   container,
		Cause:       cause,
		Recoverable: isRecoverableError(cause),
	}
}

// NewDockerServiceError はサービス操作に関するDockerErrorを作成します
func NewDockerServiceError(operation, service string, cause error) *DockerError {
	return &DockerError{
		Operation:   operation,
		Service:     service,
		Cause:       cause,
		Recoverable: isRecoverableError(cause),
	}
}

// isRecoverableError はエラーが回復可能かどうかを判定します
func isRecoverableError(err error) bool {
	if err == nil {
		return true
	}

	// タイムアウトエラーは通常回復可能
	if errors.Is(err, ErrOperationTimeout) {
		return true
	}

	// 接続エラーも回復可能
	if errors.Is(err, ErrDockerConnectionFailed) {
		return true
	}

	// コンテナが見つからないエラーは通常回復不可能
	if errors.Is(err, ErrContainerNotFound) {
		return false
	}

	// その他のエラーは基本的に回復可能と仮定
	return true
}

// IsRecoverable はエラーが回復可能かどうかを判定します
func IsRecoverable(err error) bool {
	var dockerErr *DockerError
	if errors.As(err, &dockerErr) {
		return dockerErr.Recoverable
	}
	return isRecoverableError(err)
}
