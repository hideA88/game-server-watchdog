package command

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
)

func TestRestartCommand_Name(t *testing.T) {
	cmd := NewRestartCommand(&docker.MockComposeService{}, "")
	if got := cmd.Name(); got != "restart" {
		t.Errorf("RestartCommand.Name() = %v, want %v", got, "restart")
	}
}

func TestRestartCommand_Description(t *testing.T) {
	cmd := NewRestartCommand(&docker.MockComposeService{}, "")
	if got := cmd.Description(); got != "指定されたコンテナを再起動" {
		t.Errorf("RestartCommand.Description() = %v, want %v", got, "指定されたコンテナを再起動")
	}
}

func TestNewRestartCommand(t *testing.T) {
	tests := []struct {
		name        string
		composePath string
		expected    string
	}{
		{
			name:        "カスタムパス",
			composePath: "/path/to/docker-compose.yml",
			expected:    "/path/to/docker-compose.yml",
		},
		{
			name:        "空のパス（デフォルト）",
			composePath: "",
			expected:    "docker-compose.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompose := &docker.MockComposeService{}
			cmd := NewRestartCommand(mockCompose, tt.composePath)

			if cmd.composePath != tt.expected {
				t.Errorf("NewRestartCommand() composePath = %v, want %v", cmd.composePath, tt.expected)
			}

			if cmd.compose != mockCompose {
				t.Error("NewRestartCommand() compose service not set correctly")
			}

			if cmd.serviceOperations == nil {
				t.Error("NewRestartCommand() serviceOperations map not initialized")
			}
		})
	}
}

func TestRestartCommand_Execute(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		containers   []docker.ContainerInfo
		listError    error
		restartError error
		expected     string
		expectError  bool
	}{
		{
			name:        "引数なし",
			args:        []string{},
			expected:    "使用方法: `@bot restart <サービス名>`",
			expectError: false,
		},
		{
			name: "正常な再起動",
			args: []string{"web"},
			containers: []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1", State: "running"},
				{Service: "db", Name: "app_db_1", State: "running"},
			},
			expected:    "🔄 Web を再起動しました！",
			expectError: false,
		},
		{
			name: "存在しないサービス",
			args: []string{"nonexistent"},
			containers: []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1", State: "running"},
			},
			expected:    "❌ サービス 'nonexistent' が見つかりません",
			expectError: false,
		},
		{
			name:        "コンテナリスト取得エラー",
			args:        []string{"web"},
			listError:   errors.New("docker daemon not running"),
			expectError: true,
		},
		{
			name: "再起動エラー",
			args: []string{"web"},
			containers: []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1", State: "running"},
			},
			restartError: errors.New("container not found"),
			expected:     "❌ Web の再起動に失敗しました: container not found",
			expectError:  false,
		},
		{
			name: "複数のコンテナがある場合",
			args: []string{"db"},
			containers: []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1", State: "running"},
				{Service: "db", Name: "app_db_1", State: "running"},
				{Service: "redis", Name: "app_redis_1", State: "stopped"},
			},
			expected:    "🔄 Db を再起動しました！",
			expectError: false,
		},
		{
			name: "停止中のコンテナを再起動",
			args: []string{"redis"},
			containers: []docker.ContainerInfo{
				{Service: "redis", Name: "app_redis_1", State: "stopped"},
			},
			expected:    "🔄 Redis を再起動しました！",
			expectError: false,
		},
		{
			name: "ハイフン付きサービス名",
			args: []string{"web-server"},
			containers: []docker.ContainerInfo{
				{Service: "web-server", Name: "app_web-server_1", State: "running"},
			},
			expected:    "🔄 Web Server を再起動しました！",
			expectError: false,
		},
		{
			name: "アンダースコア付きサービス名",
			args: []string{"background_worker"},
			containers: []docker.ContainerInfo{
				{Service: "background_worker", Name: "app_background_worker_1", State: "running"},
			},
			expected:    "🔄 Background Worker を再起動しました！",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompose := &docker.MockComposeService{
				ListContainersFunc: func(string) ([]docker.ContainerInfo, error) {
					return tt.containers, tt.listError
				},
				RestartContainerFunc: func(string, string) error {
					return tt.restartError
				},
			}

			cmd := NewRestartCommand(mockCompose, "test-compose.yml")
			result, err := cmd.Execute(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("RestartCommand.Execute() expected error, got nil")
				}
				if !strings.Contains(err.Error(), "コンテナ情報の取得に失敗しました") {
					t.Errorf("RestartCommand.Execute() error = %v, expected container list error", err)
				}
			} else {
				if err != nil {
					t.Errorf("RestartCommand.Execute() unexpected error = %v", err)
				}
				if result != tt.expected {
					t.Errorf("RestartCommand.Execute() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestRestartCommand_Execute_ConcurrentOperations(t *testing.T) {
	mockCompose := &docker.MockComposeService{
		ListContainersFunc: func(string) ([]docker.ContainerInfo, error) {
			return []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1", State: "running"},
			}, nil
		},
		RestartContainerFunc: func(string, string) error {
			// 実際の操作をシミュレートするために短時間待機
			time.Sleep(50 * time.Millisecond)
			return nil
		},
	}

	cmd := NewRestartCommand(mockCompose, "test-compose.yml")

	// 同時実行テスト
	var wg sync.WaitGroup
	results := make(chan string, 2)

	// 同じサービスに対して同時に2つの操作を実行
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := cmd.Execute([]string{"web"})
			if err != nil {
				results <- "ERROR: " + err.Error()
			} else {
				results <- result
			}
		}()
	}

	wg.Wait()
	close(results)

	// 結果を収集
	var resultSlice []string
	for result := range results {
		resultSlice = append(resultSlice, result)
	}

	if len(resultSlice) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(resultSlice))
	}

	// 1つは成功、1つは操作中メッセージであることを確認
	successCount := 0
	operationInProgressCount := 0

	for _, result := range resultSlice {
		if strings.Contains(result, "再起動しました") {
			successCount++
		} else if strings.Contains(result, "現在操作中です") {
			operationInProgressCount++
		}
	}

	if successCount != 1 {
		t.Errorf("Expected 1 successful operation, got %d", successCount)
	}
	if operationInProgressCount != 1 {
		t.Errorf("Expected 1 operation-in-progress message, got %d", operationInProgressCount)
	}
}

func TestRestartCommand_Execute_MultipleServices(t *testing.T) {
	mockCompose := &docker.MockComposeService{
		ListContainersFunc: func(string) ([]docker.ContainerInfo, error) {
			return []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1", State: "running"},
				{Service: "db", Name: "app_db_1", State: "running"},
				{Service: "redis", Name: "app_redis_1", State: "stopped"},
			}, nil
		},
		RestartContainerFunc: func(_, _ string) error {
			// 異なるサービスの操作は成功
			return nil
		},
	}

	cmd := NewRestartCommand(mockCompose, "test-compose.yml")

	// 異なるサービスに対する並行操作
	var wg sync.WaitGroup
	services := []string{"web", "db", "redis"}
	results := make(chan string, len(services))

	for _, service := range services {
		wg.Add(1)
		go func(svc string) {
			defer wg.Done()
			result, err := cmd.Execute([]string{svc})
			if err != nil {
				results <- "ERROR: " + err.Error()
			} else {
				results <- result
			}
		}(service)
	}

	wg.Wait()
	close(results)

	// すべての操作が成功することを確認
	successCount := 0
	for result := range results {
		if strings.Contains(result, "再起動しました") {
			successCount++
		}
	}

	if successCount != len(services) {
		t.Errorf("Expected %d successful operations, got %d", len(services), successCount)
	}
}

func TestRestartCommand_Execute_InvalidServiceName(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		expectError bool
	}{
		{
			name:        "有効なサービス名",
			serviceName: "web-server",
			expectError: false,
		},
		{
			name:        "有効なサービス名（アンダースコア）",
			serviceName: "background_worker",
			expectError: false,
		},
		{
			name:        "有効なサービス名（数字）",
			serviceName: "worker123",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompose := &docker.MockComposeService{
				ListContainersFunc: func(string) ([]docker.ContainerInfo, error) {
					return []docker.ContainerInfo{
						{Service: tt.serviceName, Name: "app_" + tt.serviceName + "_1", State: "running"},
					}, nil
				},
				RestartContainerFunc: func(string, string) error {
					return nil
				},
			}

			cmd := NewRestartCommand(mockCompose, "test-compose.yml")
			result, err := cmd.Execute([]string{tt.serviceName})

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for invalid service name %s", tt.serviceName)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for valid service name %s: %v", tt.serviceName, err)
				}
				if !strings.Contains(result, "再起動しました") {
					t.Errorf("Expected success message, got: %s", result)
				}
			}
		})
	}
}

// ベンチマークテスト
func BenchmarkRestartCommand_Execute(b *testing.B) {
	mockCompose := &docker.MockComposeService{
		ListContainersFunc: func(string) ([]docker.ContainerInfo, error) {
			return []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1", State: "running"},
			}, nil
		},
		RestartContainerFunc: func(string, string) error {
			return nil
		},
	}

	cmd := NewRestartCommand(mockCompose, "test-compose.yml")
	args := []string{"web"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cmd.Execute(args)
		if err != nil {
			b.Fatalf("Execute() failed: %v", err)
		}
	}
}
