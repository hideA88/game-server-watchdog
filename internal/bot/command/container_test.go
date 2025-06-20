package command

import (
	"errors"
	"strings"
	"testing"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
)

func TestContainerCommand_Name(t *testing.T) {
	cmd := NewContainerCommand(&docker.MockComposeService{}, "")
	if got := cmd.Name(); got != "container" {
		t.Errorf("ContainerCommand.Name() = %v, want %v", got, "container")
	}
}

func TestContainerCommand_Description(t *testing.T) {
	cmd := NewContainerCommand(&docker.MockComposeService{}, "")
	if got := cmd.Description(); got != "個別コンテナの詳細情報を表示" {
		t.Errorf("ContainerCommand.Description() = %v, want %v", got, "個別コンテナの詳細情報を表示")
	}
}

func TestNewContainerCommand(t *testing.T) {
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
			expected:    defaultComposePath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompose := &docker.MockComposeService{}
			cmd := NewContainerCommand(mockCompose, tt.composePath)

			if cmd.composePath != tt.expected {
				t.Errorf("NewContainerCommand() composePath = %v, want %v", cmd.composePath, tt.expected)
			}

			if cmd.compose != mockCompose {
				t.Error("NewContainerCommand() compose service not set correctly")
			}
		})
	}
}

func TestContainerCommand_Execute(t *testing.T) { //nolint:funlen // テーブル駆動テストのため長い関数を許可
	tests := []struct {
		name              string
		args              []string
		containers        []docker.ContainerInfo
		containerStats    *docker.ContainerStats
		containerLogs     string
		listError         error
		statsError        error
		logsError         error
		expectError       bool
		expectedContains  []string
		expectedNotContains []string
	}{
		{
			name:             "引数なし",
			args:             []string{},
			expectedContains: []string{"使用方法: `@bot container <サービス名>`"},
		},
		{
			name: "実行中のコンテナ（正常系）",
			args: []string{"web"},
			containers: []docker.ContainerInfo{
				{
					Service:      "web",
					Name:         "app_web_1",
					ID:           "abc123def456789",
					State:        "running",
					RunningFor:   "2h30m",
					HealthStatus: "healthy",
					Ports:        []string{"80:8080/tcp", "443:8443/tcp"},
				},
			},
			containerStats: &docker.ContainerStats{
				CPUPercent:    45.2,
				MemoryPercent: 67.8,
				MemoryUsage:   "2.1GiB / 4.0GiB",
				NetworkIO:     "1.2MB / 800KB",
				BlockIO:       "500KB / 1.5MB",
			},
			containerLogs: "2023-01-01 12:00:00 INFO Starting application\n" +
				"2023-01-01 12:00:01 INFO Server listening on port 8080",
			expectedContains: []string{
				"📦 **Web の詳細情報**",
				"**基本情報**",
				"- コンテナ名: app_web_1",
				"- コンテナID: abc123def456",
				"- 状態: 🟢 running",
				"- 稼働時間: 2h30m",
				"- ヘルス: ✅ healthy",
				"- ポート: 80:8080/tcp, 443:8443/tcp",
				"**リソース使用状況**",
				"- CPU使用率: 45.2%",
				"- メモリ使用率: 67.8%",
				"- メモリ使用量: 2.1GiB / 4.0GiB",
				"- ネットワークI/O: 1.2MB / 800KB",
				"- ブロックI/O: 500KB / 1.5MB",
				"**最近のログ**",
				"INFO Starting application",
				"**使用可能なコマンド**",
				"`@bot restart web`",
				"`@bot logs web [行数]`",
			},
		},
		{
			name: "停止中のコンテナ",
			args: []string{"db"},
			containers: []docker.ContainerInfo{
				{
					Service: "db",
					Name:    "app_db_1",
					ID:      "def456ghi789abc",
					State:   "stopped",
				},
			},
			containerLogs: "Database stopped",
			expectedContains: []string{
				"📦 **Db の詳細情報**",
				"- 状態: 🔴 stopped",
				"- `@bot monitor` から起動ボタンを使用",
			},
			expectedNotContains: []string{
				"**リソース使用状況**",
				"`@bot restart db`",
			},
		},
		{
			name: "高負荷状態のコンテナ",
			args: []string{"worker"},
			containers: []docker.ContainerInfo{
				{
					Service: "worker",
					Name:    "app_worker_1",
					ID:      "ghi789abc123def",
					State:   "running",
				},
			},
			containerStats: &docker.ContainerStats{
				CPUPercent:    92.5,
				MemoryPercent: 95.3,
				MemoryUsage:   "7.8GiB / 8.0GiB",
			},
			containerLogs: "Worker processing tasks",
			expectedContains: []string{
				"⚠️ **警告**",
				"- CPU使用率が高い状態です (92.5%)",
				"- メモリ使用率が高い状態です (95.3%)",
			},
		},
		{
			name: "ヘルスチェックなしのコンテナ",
			args: []string{"redis"},
			containers: []docker.ContainerInfo{
				{
					Service:      "redis",
					Name:         "app_redis_1",
					ID:           "jkl012mno345pqr",
					State:        "running",
					HealthStatus: "none",
				},
			},
			containerStats: &docker.ContainerStats{
				CPUPercent:    5.0,
				MemoryPercent: 30.0,
				MemoryUsage:   "500MB / 1.5GB",
			},
			containerLogs: "Redis server started",
			expectedNotContains: []string{
				"- ヘルス:",
			},
		},
		{
			name: "ポートなしのコンテナ",
			args: []string{"queue"},
			containers: []docker.ContainerInfo{
				{
					Service: "queue",
					Name:    "app_queue_1",
					ID:      "pqr345stu678vwx",
					State:   "running",
					Ports:   []string{},
				},
			},
			containerStats: &docker.ContainerStats{
				CPUPercent:    10.0,
				MemoryPercent: 25.0,
				MemoryUsage:   "300MB / 1GB",
			},
			containerLogs: "Queue worker started",
			expectedNotContains: []string{
				"- ポート:",
			},
		},
		{
			name:        "存在しないサービス",
			args:        []string{"nonexistent"},
			containers:  []docker.ContainerInfo{},
			expectedContains: []string{
				"❌ サービス 'nonexistent' が見つかりません",
			},
		},
		{
			name:        "コンテナリスト取得エラー",
			args:        []string{"web"},
			listError:   errors.New("docker daemon not running"),
			expectError: true,
		},
		{
			name: "統計情報取得エラー（実行中コンテナ）",
			args: []string{"web"},
			containers: []docker.ContainerInfo{
				{
					Service: "web",
					Name:    "app_web_1",
					ID:      "abc123def456789",
					State:   "running",
				},
			},
			statsError:    errors.New("stats not available"),
			containerLogs: "App running",
			expectedContains: []string{
				"📦 **Web の詳細情報**",
				"**基本情報**",
			},
			expectedNotContains: []string{
				"**リソース使用状況**",
			},
		},
		{
			name: "ログ取得エラー",
			args: []string{"web"},
			containers: []docker.ContainerInfo{
				{
					Service: "web",
					Name:    "app_web_1",
					ID:      "abc123def456789",
					State:   "running",
				},
			},
			containerStats: &docker.ContainerStats{
				CPUPercent:    20.0,
				MemoryPercent: 40.0,
				MemoryUsage:   "1GB / 2GB",
			},
			logsError: errors.New("logs not available"),
			expectedContains: []string{
				"ログの取得に失敗しました",
			},
		},
		{
			name: "長いログ行の切り詰め",
			args: []string{"web"},
			containers: []docker.ContainerInfo{
				{
					Service: "web",
					Name:    "app_web_1",
					ID:      "abc123def456789",
					State:   "running",
				},
			},
			containerStats: &docker.ContainerStats{
				CPUPercent:    15.0,
				MemoryPercent: 35.0,
				MemoryUsage:   "800MB / 2GB",
			},
			containerLogs: strings.Repeat(
				"Very long log line that exceeds the maximum allowed length and should be truncated",
				10,
			),
			expectedContains: []string{
				"Very long log line that exceeds the maximum allowed length and should be trun...",
			},
		},
		{
			name: "ハイフン・アンダースコア付きサービス名",
			args: []string{"web-server_v2"},
			containers: []docker.ContainerInfo{
				{
					Service: "web-server_v2",
					Name:    "app_web-server_v2_1",
					ID:      "xyz789abc123def",
					State:   "running",
				},
			},
			containerStats: &docker.ContainerStats{
				CPUPercent:    30.0,
				MemoryPercent: 50.0,
				MemoryUsage:   "1.5GB / 3GB",
			},
			containerLogs: "Web server v2 started",
			expectedContains: []string{
				"📦 **Web Server V2 の詳細情報**",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompose := &docker.MockComposeService{
				ListContainersFunc: func(string) ([]docker.ContainerInfo, error) {
					return tt.containers, tt.listError
				},
				GetContainerStatsFunc: func(string) (*docker.ContainerStats, error) {
					return tt.containerStats, tt.statsError
				},
				GetContainerLogsFunc: func(string, string, int) (string, error) {
					return tt.containerLogs, tt.logsError
				},
			}

			cmd := NewContainerCommand(mockCompose, "test-compose.yml")
			result, err := cmd.Execute(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("ContainerCommand.Execute() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ContainerCommand.Execute() unexpected error = %v", err)
				}

				// 期待される文字列が含まれているかチェック
				for _, expected := range tt.expectedContains {
					if !strings.Contains(result, expected) {
						t.Errorf("ContainerCommand.Execute() result should contain %q\nActual result:\n%s", expected, result)
					}
				}

				// 含まれてはいけない文字列がないかチェック
				for _, notExpected := range tt.expectedNotContains {
					if strings.Contains(result, notExpected) {
						t.Errorf("ContainerCommand.Execute() result should not contain %q\nActual result:\n%s", notExpected, result)
					}
				}
			}
		})
	}
}

func TestContainerCommand_findContainer(t *testing.T) { //nolint:funlen // テーブル駆動テストのため長い関数を許可
	tests := []struct {
		name         string
		serviceName  string
		containers   []docker.ContainerInfo
		listError    error
		expectedName string
		expectError  bool
	}{
		{
			name:        "コンテナが見つかる",
			serviceName: "web",
			containers: []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1"},
				{Service: "db", Name: "app_db_1"},
			},
			expectedName: "app_web_1",
		},
		{
			name:        "コンテナが見つからない",
			serviceName: "nonexistent",
			containers: []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1"},
			},
			expectedName: "",
		},
		{
			name:        "リストエラー",
			serviceName: "web",
			listError:   errors.New("docker error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompose := &docker.MockComposeService{
				ListContainersFunc: func(string) ([]docker.ContainerInfo, error) {
					return tt.containers, tt.listError
				},
			}

			cmd := NewContainerCommand(mockCompose, "test-compose.yml")
			result, err := cmd.findContainer(tt.serviceName)

			if tt.expectError {
				if err == nil {
					t.Errorf("findContainer() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("findContainer() unexpected error = %v", err)
				}

				if tt.expectedName == "" {
					if result != nil {
						t.Errorf("findContainer() expected nil, got %v", result)
					}
				} else {
					if result == nil {
						t.Errorf("findContainer() expected container, got nil")
					} else if result.Name != tt.expectedName {
						t.Errorf("findContainer() name = %v, want %v", result.Name, tt.expectedName)
					}
				}
			}
		})
	}
}

func TestContainerCommand_addBasicInfo(t *testing.T) {
	tests := []struct {
		name         string
		container    *docker.ContainerInfo
		expectedContains []string
	}{
		{
			name: "完全な情報",
			container: &docker.ContainerInfo{
				Name:         "app_web_1",
				ID:           "abc123def456789012345",
				State:        "running",
				RunningFor:   "2h30m",
				HealthStatus: "healthy",
				Ports:        []string{"80:8080/tcp", "443:8443/tcp"},
			},
			expectedContains: []string{
				"**基本情報**",
				"- コンテナ名: app_web_1",
				"- コンテナID: abc123def456",
				"- 状態: 🟢 running",
				"- 稼働時間: 2h30m",
				"- ヘルス: ✅ healthy",
				"- ポート: 80:8080/tcp, 443:8443/tcp",
			},
		},
		{
			name: "最小限の情報",
			container: &docker.ContainerInfo{
				Name:         "app_minimal_1",
				ID:           "def456ghi789abc",
				State:        "stopped",
				HealthStatus: "none",
				Ports:        []string{},
			},
			expectedContains: []string{
				"- コンテナ名: app_minimal_1",
				"- コンテナID: def456ghi789",
				"- 状態: 🔴 stopped",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewContainerCommand(&docker.MockComposeService{}, "")
			var builder strings.Builder
			
			cmd.addBasicInfo(&builder, tt.container)
			result := builder.String()

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("addBasicInfo() should contain %q\nActual result:\n%s", expected, result)
				}
			}
		})
	}
}

func TestContainerCommand_addResourceInfo(t *testing.T) { //nolint:funlen // テーブル駆動テストのため長い関数を許可
	tests := []struct {
		name             string
		container        *docker.ContainerInfo
		stats            *docker.ContainerStats
		statsError       error
		expectedContains []string
		expectedNotContains []string
	}{
		{
			name: "正常なリソース情報",
			container: &docker.ContainerInfo{Name: "app_web_1"},
			stats: &docker.ContainerStats{
				CPUPercent:    45.2,
				MemoryPercent: 67.8,
				MemoryUsage:   "2.1GiB / 4.0GiB",
				NetworkIO:     "1.2MB / 800KB",
				BlockIO:       "500KB / 1.5MB",
			},
			expectedContains: []string{
				"**リソース使用状況**",
				"- CPU使用率: 45.2%",
				"- メモリ使用率: 67.8%",
				"- メモリ使用量: 2.1GiB / 4.0GiB",
				"- ネットワークI/O: 1.2MB / 800KB",
				"- ブロックI/O: 500KB / 1.5MB",
			},
		},
		{
			name: "高負荷警告",
			container: &docker.ContainerInfo{Name: "app_worker_1"},
			stats: &docker.ContainerStats{
				CPUPercent:    90.0,
				MemoryPercent: 95.0,
				MemoryUsage:   "7.5GiB / 8.0GiB",
			},
			expectedContains: []string{
				"⚠️ **警告**",
				"- CPU使用率が高い状態です (90.0%)",
				"- メモリ使用率が高い状態です (95.0%)",
			},
		},
		{
			name:       "統計取得エラー",
			container:  &docker.ContainerInfo{Name: "app_error_1"},
			statsError: errors.New("stats error"),
			expectedNotContains: []string{
				"**リソース使用状況**",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompose := &docker.MockComposeService{
				GetContainerStatsFunc: func(string) (*docker.ContainerStats, error) {
					return tt.stats, tt.statsError
				},
			}

			cmd := NewContainerCommand(mockCompose, "")
			var builder strings.Builder
			
			cmd.addResourceInfo(&builder, tt.container)
			result := builder.String()

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("addResourceInfo() should contain %q\nActual result:\n%s", expected, result)
				}
			}

			for _, notExpected := range tt.expectedNotContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("addResourceInfo() should not contain %q\nActual result:\n%s", notExpected, result)
				}
			}
		})
	}
}

func TestContainerCommand_addRecentLogs(t *testing.T) {
	tests := []struct {
		name             string
		serviceName      string
		logs             string
		logsError        error
		expectedContains []string
	}{
		{
			name:        "正常なログ",
			serviceName: "web",
			logs:        "2023-01-01 12:00:00 INFO Starting\n2023-01-01 12:00:01 INFO Ready",
			expectedContains: []string{
				"**最近のログ** (最後の10行)",
				"```",
				"2023-01-01 12:00:00 INFO Starting",
				"2023-01-01 12:00:01 INFO Ready",
			},
		},
		{
			name:        "ログ取得エラー",
			serviceName: "error",
			logsError:   errors.New("logs error"),
			expectedContains: []string{
				"ログの取得に失敗しました",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompose := &docker.MockComposeService{
				GetContainerLogsFunc: func(string, string, int) (string, error) {
					return tt.logs, tt.logsError
				},
			}

			cmd := NewContainerCommand(mockCompose, "")
			var builder strings.Builder
			
			cmd.addRecentLogs(&builder, tt.serviceName)
			result := builder.String()

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("addRecentLogs() should contain %q\nActual result:\n%s", expected, result)
				}
			}
		})
	}
}

func TestContainerCommand_addAvailableCommands(t *testing.T) {
	tests := []struct {
		name             string
		serviceName      string
		state            string
		expectedContains []string
		expectedNotContains []string
	}{
		{
			name:        "実行中のコンテナ",
			serviceName: "web",
			state:       "running",
			expectedContains: []string{
				"**使用可能なコマンド**",
				"`@bot restart web`",
				"`@bot logs web [行数]`",
			},
		},
		{
			name:        "停止中のコンテナ",
			serviceName: "db",
			state:       "stopped",
			expectedContains: []string{
				"`@bot monitor` から起動ボタンを使用",
			},
			expectedNotContains: []string{
				"`@bot restart db`",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewContainerCommand(&docker.MockComposeService{}, "")
			var builder strings.Builder
			
			cmd.addAvailableCommands(&builder, tt.serviceName, tt.state)
			result := builder.String()

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("addAvailableCommands() should contain %q\nActual result:\n%s", expected, result)
				}
			}

			for _, notExpected := range tt.expectedNotContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("addAvailableCommands() should not contain %q\nActual result:\n%s", notExpected, result)
				}
			}
		})
	}
}

// ベンチマークテスト
func BenchmarkContainerCommand_Execute(b *testing.B) {
	mockCompose := &docker.MockComposeService{
		ListContainersFunc: func(string) ([]docker.ContainerInfo, error) {
			return []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1", ID: "abc123def456", State: "running"},
			}, nil
		},
		GetContainerStatsFunc: func(string) (*docker.ContainerStats, error) {
			return &docker.ContainerStats{
				CPUPercent:    50.0,
				MemoryPercent: 60.0,
				MemoryUsage:   "2GB / 4GB",
			}, nil
		},
		GetContainerLogsFunc: func(string, string, int) (string, error) {
			return "Sample log line", nil
		},
	}

	cmd := NewContainerCommand(mockCompose, "test-compose.yml")
	args := []string{"web"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cmd.Execute(args)
		if err != nil {
			b.Fatalf("Execute() failed: %v", err)
		}
	}
}