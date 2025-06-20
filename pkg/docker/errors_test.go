package docker

import (
	"errors"
	"fmt"
	"testing"
)

func TestDockerError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      DockerError
		expected string
	}{
		{
			name: "コンテナ名付きエラー",
			err: DockerError{
				Operation: "start",
				Container: "web-server",
				Cause:     errors.New("connection failed"),
			},
			expected: "docker operation 'start' failed for container 'web-server': connection failed",
		},
		{
			name: "サービス名付きエラー",
			err: DockerError{
				Operation: "stop",
				Service:   "database",
				Cause:     errors.New("timeout"),
			},
			expected: "docker operation 'stop' failed for service 'database': timeout",
		},
		{
			name: "基本的なエラー",
			err: DockerError{
				Operation: "ps",
				Cause:     errors.New("permission denied"),
			},
			expected: "docker operation 'ps' failed: permission denied",
		},
		{
			name: "Causeがnil",
			err: DockerError{
				Operation: "build",
				Service:   "app",
				Cause:     nil,
			},
			expected: "docker operation 'build' failed for service 'app': <nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("DockerError.Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDockerError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	dockerErr := DockerError{
		Operation: "test",
		Cause:     originalErr,
	}

	unwrapped := dockerErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("DockerError.Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestDockerError_Is(t *testing.T) {
	originalErr := errors.New("original error")
	dockerErr := DockerError{
		Operation: "test",
		Cause:     originalErr,
	}

	// errors.Is()の動作確認
	if !errors.Is(&dockerErr, originalErr) {
		t.Error("errors.Is() should return true for wrapped error")
	}

	otherErr := errors.New("other error")
	if errors.Is(&dockerErr, otherErr) {
		t.Error("errors.Is() should return false for different error")
	}
}

func TestNewDockerError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		cause     error
		expected  DockerError
	}{
		{
			name:      "基本的なエラー作成",
			operation: "start",
			cause:     errors.New("test error"),
			expected: DockerError{
				Operation:   "start",
				Cause:       errors.New("test error"),
				Recoverable: true,
			},
		},
		{
			name:      "コンテナが見つからないエラー",
			operation: "inspect",
			cause:     ErrContainerNotFound,
			expected: DockerError{
				Operation:   "inspect",
				Cause:       ErrContainerNotFound,
				Recoverable: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewDockerError(tt.operation, tt.cause)
			
			if result.Operation != tt.expected.Operation {
				t.Errorf("Operation = %v, want %v", result.Operation, tt.expected.Operation)
			}
			if result.Recoverable != tt.expected.Recoverable {
				t.Errorf("Recoverable = %v, want %v", result.Recoverable, tt.expected.Recoverable)
			}
			if result.Cause.Error() != tt.expected.Cause.Error() {
				t.Errorf("Cause = %v, want %v", result.Cause, tt.expected.Cause)
			}
		})
	}
}

func TestNewDockerContainerError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		container string
		cause     error
		want      string
	}{
		{
			name:      "通常のコンテナエラー",
			operation: "start",
			container: "web-app",
			cause:     errors.New("failed"),
			want:      "web-app",
		},
		{
			name:      "空のコンテナ名",
			operation: "stop",
			container: "",
			cause:     errors.New("failed"),
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewDockerContainerError(tt.operation, tt.container, tt.cause)
			
			if err.Container != tt.want {
				t.Errorf("NewDockerContainerError() Container = %v, want %v", err.Container, tt.want)
			}
		})
	}
}

func TestNewDockerServiceError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		service   string
		cause     error
		want      string
	}{
		{
			name:      "通常のサービスエラー",
			operation: "restart",
			service:   "database",
			cause:     errors.New("failed"),
			want:      "database",
		},
		{
			name:      "空のサービス名",
			operation: "stop",
			service:   "",
			cause:     errors.New("failed"),
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewDockerServiceError(tt.operation, tt.service, tt.cause)
			
			if err.Service != tt.want {
				t.Errorf("NewDockerServiceError() Service = %v, want %v", err.Service, tt.want)
			}
		})
	}
}

func TestIsRecoverable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "DockerError（回復可能）",
			err:  NewDockerError("start", ErrOperationTimeout),
			want: true,
		},
		{
			name: "DockerError（回復不可能）",
			err:  NewDockerError("inspect", ErrContainerNotFound),
			want: false,
		},
		{
			name: "通常のエラー",
			err:  errors.New("general error"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRecoverable(tt.err); got != tt.want {
				t.Errorf("IsRecoverable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRecoverableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil エラー",
			err:  nil,
			want: true,
		},
		{
			name: "タイムアウトエラー（回復可能）",
			err:  ErrOperationTimeout,
			want: true,
		},
		{
			name: "接続エラー（回復可能）",
			err:  ErrDockerConnectionFailed,
			want: true,
		},
		{
			name: "コンテナ未発見エラー（回復不可能）",
			err:  ErrContainerNotFound,
			want: false,
		},
		{
			name: "その他のエラー（回復可能）",
			err:  errors.New("unknown error"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRecoverableError(tt.err); got != tt.want {
				t.Errorf("isRecoverableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 実用的な使用例のテスト
func TestDockerError_RealWorldUsage(t *testing.T) {
	// Docker Composeでのサービス起動失敗をシミュレート
	err := NewDockerServiceError(
		"start",
		"web",
		errors.New("container exited with code 1"),
	)

	// エラーメッセージの確認
	expected := "docker operation 'start' failed for service 'web': container exited with code 1"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}

	// 回復可能性の確認
	if !err.Recoverable {
		t.Error("Error should be recoverable")
	}

	// 元のエラーの確認（Unwrapできるかチェック）
	unwrappedErr := errors.Unwrap(err)
	if unwrappedErr == nil {
		t.Error("Should be able to unwrap error")
	}
	if unwrappedErr.Error() != "container exited with code 1" {
		t.Errorf("Unwrapped error = %q, want %q", unwrappedErr.Error(), "container exited with code 1")
	}
}

// ベンチマークテスト
func BenchmarkDockerError_Error(b *testing.B) {
	err := DockerError{
		Operation: "start",
		Container: "web-server",
		Service:   "web",
		Cause:     errors.New("connection failed"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkNewDockerError(b *testing.B) {
	cause := errors.New("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewDockerError("start", cause)
	}
}

// エラーラッピングのテスト
func TestDockerError_ErrorWrapping(t *testing.T) {
	originalErr := fmt.Errorf("network timeout")
	dockerErr := NewDockerServiceError("connect", "database", originalErr)

	// fmt.Errorf でラップ
	wrappedErr := fmt.Errorf("failed to connect to database: %w", dockerErr)

	// 元のエラーが見つけられるか確認
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("Should be able to find original error through multiple wrapping levels")
	}

	// DockerErrorが見つけられるか確認
	var dockerErrPtr *DockerError
	if !errors.As(wrappedErr, &dockerErrPtr) {
		t.Error("Should be able to extract DockerError from wrapped error")
	}

	if dockerErrPtr.Operation != "connect" {
		t.Errorf("Extracted DockerError.Operation = %q, want %q", dockerErrPtr.Operation, "connect")
	}
}