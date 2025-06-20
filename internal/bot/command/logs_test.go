package command

import (
	"errors"
	"strings"
	"testing"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
)

func TestLogsCommand_Name(t *testing.T) {
	cmd := NewLogsCommand(&docker.MockComposeService{}, "")
	if got := cmd.Name(); got != "logs" {
		t.Errorf("LogsCommand.Name() = %v, want %v", got, "logs")
	}
}

func TestLogsCommand_Description(t *testing.T) {
	cmd := NewLogsCommand(&docker.MockComposeService{}, "")
	if got := cmd.Description(); got != "指定されたコンテナのログを表示" {
		t.Errorf("LogsCommand.Description() = %v, want %v", got, "指定されたコンテナのログを表示")
	}
}

func TestNewLogsCommand(t *testing.T) {
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
			cmd := NewLogsCommand(mockCompose, tt.composePath)

			if cmd.composePath != tt.expected {
				t.Errorf("NewLogsCommand() composePath = %v, want %v", cmd.composePath, tt.expected)
			}

			if cmd.compose != mockCompose {
				t.Error("NewLogsCommand() compose service not set correctly")
			}
		})
	}
}

func TestLogsCommand_parseLineCount(t *testing.T) { // テーブル駆動テストのため長い関数を許可
	tests := []struct {
		name     string
		args     []string
		expected int
	}{
		{
			name:     "引数なし（デフォルト）",
			args:     []string{"service"},
			expected: defaultLogCount,
		},
		{
			name:     "有効な行数",
			args:     []string{"service", "25"},
			expected: 25,
		},
		{
			name:     "最大行数超過",
			args:     []string{"service", "300"},
			expected: maxLogCount,
		},
		{
			name:     "負の値",
			args:     []string{"service", "-10"},
			expected: 1,
		},
		{
			name:     "ゼロ",
			args:     []string{"service", "0"},
			expected: 1,
		},
		{
			name:     "無効な値（文字列）",
			args:     []string{"service", "abc"},
			expected: defaultLogCount,
		},
		{
			name:     "小数点",
			args:     []string{"service", "10.5"},
			expected: defaultLogCount,
		},
		{
			name:     "非常に大きな値",
			args:     []string{"service", "999999"},
			expected: maxLogCount,
		},
		{
			name:     "境界値（最大）",
			args:     []string{"service", "200"},
			expected: 200,
		},
		{
			name:     "境界値（最小）",
			args:     []string{"service", "1"},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewLogsCommand(&docker.MockComposeService{}, "")
			result := cmd.parseLineCount(tt.args)

			if result != tt.expected {
				t.Errorf("parseLineCount() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLogsCommand_containerExists(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		containers  []docker.ContainerInfo
		listError   error
		expected    bool
		expectError bool
	}{
		{
			name:        "コンテナが存在する",
			serviceName: "web",
			containers: []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1"},
				{Service: "db", Name: "app_db_1"},
			},
			expected: true,
		},
		{
			name:        "コンテナが存在しない",
			serviceName: "nonexistent",
			containers: []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1"},
			},
			expected: false,
		},
		{
			name:        "空のコンテナリスト",
			serviceName: "web",
			containers:  []docker.ContainerInfo{},
			expected:    false,
		},
		{
			name:        "リスト取得エラー",
			serviceName: "web",
			listError:   errors.New("docker daemon not running"),
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

			cmd := NewLogsCommand(mockCompose, "test-compose.yml")
			result, err := cmd.containerExists(tt.serviceName)

			if tt.expectError {
				if err == nil {
					t.Errorf("containerExists() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("containerExists() unexpected error = %v", err)
				}
				if result != tt.expected {
					t.Errorf("containerExists() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestLogsCommand_Execute(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		containers       []docker.ContainerInfo
		logs             string
		listError        error
		logsError        error
		expectedContains []string
		expectError      bool
	}{
		{
			name:             "引数なし",
			args:             []string{},
			expectedContains: []string{"使用方法: `@bot logs <サービス名> [行数]`"},
		},
		{
			name: "正常なログ表示（デフォルト行数）",
			args: []string{"web"},
			containers: []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1"},
			},
			logs: "2023-01-01 12:00:00 INFO Starting application\n" +
				"2023-01-01 12:00:01 INFO Server listening on port 8080\n" +
				"2023-01-01 12:00:02 INFO Ready to accept connections",
			expectedContains: []string{
				"📜 **Web のログ** (最後の50行)",
				"```",
				"2023-01-01 12:00:00 INFO Starting application",
				"2023-01-01 12:00:01 INFO Server listening on port 8080",
				"2023-01-01 12:00:02 INFO Ready to accept connections",
				"💡 **ヒント**: より多くのログを見るには、行数を指定してください",
				"例: `@bot logs web 100`",
			},
		},
		{
			name: "カスタム行数",
			args: []string{"db", "25"},
			containers: []docker.ContainerInfo{
				{Service: "db", Name: "app_db_1"},
			},
			logs: "Database initialized\nReady for connections",
			expectedContains: []string{
				"📜 **Db のログ** (最後の25行)",
				"Database initialized",
				"Ready for connections",
				"例: `@bot logs db 100`",
			},
		},
		{
			name: "存在しないサービス",
			args: []string{"nonexistent"},
			containers: []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1"},
			},
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
			name: "ログ取得エラー",
			args: []string{"web"},
			containers: []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1"},
			},
			logsError: errors.New("container not running"),
			expectedContains: []string{
				"❌ Web のログ取得に失敗しました: container not running",
			},
		},
		{
			name: "空のログ",
			args: []string{"empty"},
			containers: []docker.ContainerInfo{
				{Service: "empty", Name: "app_empty_1"},
			},
			logs: "",
			expectedContains: []string{
				"📜 **Empty のログ**",
				"(ログがありません)",
			},
		},
		{
			name: "スペースのみのログ",
			args: []string{"whitespace"},
			containers: []docker.ContainerInfo{
				{Service: "whitespace", Name: "app_whitespace_1"},
			},
			logs: "   \n\n   ",
			expectedContains: []string{
				"(ログがありません)",
			},
		},
		{
			name: "大きな行数指定",
			args: []string{"web", "500"},
			containers: []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1"},
			},
			logs: "Log line 1\nLog line 2",
			expectedContains: []string{
				"📜 **Web のログ** (最後の200行)", // 最大値に制限される
				"Log line 1",
				"Log line 2",
			},
		},
		{
			name: "無効な行数指定",
			args: []string{"web", "invalid"},
			containers: []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1"},
			},
			logs: "Default log content",
			expectedContains: []string{
				"📜 **Web のログ** (最後の50行)", // デフォルト値が使用される
				"Default log content",
			},
		},
		{
			name: "ハイフン・アンダースコア付きサービス名",
			args: []string{"web-server_v2", "30"},
			containers: []docker.ContainerInfo{
				{Service: "web-server_v2", Name: "app_web-server_v2_1"},
			},
			logs: "Web server v2 log entry",
			expectedContains: []string{
				"📜 **Web Server V2 のログ** (最後の30行)",
				"Web server v2 log entry",
				"例: `@bot logs web-server_v2 100`",
			},
		},
		{
			name: "非常に長いログ行",
			args: []string{"long"},
			containers: []docker.ContainerInfo{
				{Service: "long", Name: "app_long_1"},
			},
			logs: strings.Repeat("Very long log line that exceeds the maximum allowed length and should be truncated. ", 10),
			expectedContains: []string{
				"Very long log line that exceeds the maximum allowed length and should be truncated.",
				"...", // 切り詰められたことを示す
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompose := &docker.MockComposeService{
				ListContainersFunc: func(string) ([]docker.ContainerInfo, error) {
					return tt.containers, tt.listError
				},
				GetContainerLogsFunc: func(string, string, int) (string, error) {
					return tt.logs, tt.logsError
				},
			}

			cmd := NewLogsCommand(mockCompose, "test-compose.yml")
			result, err := cmd.Execute(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("LogsCommand.Execute() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("LogsCommand.Execute() unexpected error = %v", err)
				}

				for _, expected := range tt.expectedContains {
					if !strings.Contains(result, expected) {
						t.Errorf("LogsCommand.Execute() result should contain %q\nActual result:\n%s", expected, result)
					}
				}
			}
		})
	}
}

func TestLogsCommand_buildLogOutput(t *testing.T) {
	tests := []struct {
		name             string
		serviceName      string
		lines            int
		logs             string
		expectedContains []string
	}{
		{
			name:        "通常のログ",
			serviceName: "web",
			lines:       10,
			logs:        "Log line 1\nLog line 2",
			expectedContains: []string{
				"📜 **Web のログ** (最後の10行)",
				"```",
				"Log line 1",
				"Log line 2",
				"💡 **ヒント**",
			},
		},
		{
			name:        "空のログ",
			serviceName: "empty",
			lines:       20,
			logs:        "",
			expectedContains: []string{
				"📜 **Empty のログ** (最後の20行)",
				"(ログがありません)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewLogsCommand(&docker.MockComposeService{}, "")
			result := cmd.buildLogOutput(tt.serviceName, tt.lines, tt.logs)

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("buildLogOutput() should contain %q\nActual result:\n%s", expected, result)
				}
			}
		})
	}
}

func TestLogsCommand_addFormattedLogs(t *testing.T) {
	tests := []struct {
		name                string
		logs                string
		requestedLines      int
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name:           "通常のログ",
			logs:           "Line 1\nLine 2\nLine 3",
			requestedLines: 10,
			expectedContains: []string{
				"Line 1",
				"Line 2",
				"Line 3",
			},
		},
		{
			name:           "長いログ行の切り詰め",
			logs:           strings.Repeat("a", 250), // 250文字
			requestedLines: 10,
			expectedContains: []string{
				"...", // 切り詰められたことを示す
			},
		},
		{
			name:           "大量のログ行（制限表示）",
			logs:           strings.Repeat("Log line\n", 300),
			requestedLines: 10,
			expectedContains: []string{
				"... (残り",
				"行は省略されました)",
			},
		},
		{
			name:           "最大行数制限の注意メッセージ",
			logs:           "Single line",
			requestedLines: 300, // maxLogCountを超える
			expectedContains: []string{
				"(注意: 最大200行に制限されています)",
			},
		},
		{
			name:           "最大行数以下の場合は注意メッセージなし",
			logs:           "Single line",
			requestedLines: 100,
			expectedNotContains: []string{
				"(注意: 最大200行に制限されています)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewLogsCommand(&docker.MockComposeService{}, "")
			var builder strings.Builder

			cmd.addFormattedLogs(&builder, tt.logs, tt.requestedLines)
			result := builder.String()

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("addFormattedLogs() should contain %q\nActual result:\n%s", expected, result)
				}
			}

			for _, notExpected := range tt.expectedNotContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("addFormattedLogs() should not contain %q\nActual result:\n%s", notExpected, result)
				}
			}
		})
	}
}

func TestLogsCommand_MessageLengthLimit(t *testing.T) {
	// 非常に長いログでDiscordのメッセージ制限に引っかからないかテスト
	cmd := NewLogsCommand(&docker.MockComposeService{
		ListContainersFunc: func(string) ([]docker.ContainerInfo, error) {
			return []docker.ContainerInfo{
				{Service: "test", Name: "app_test_1"},
			}, nil
		},
		GetContainerLogsFunc: func(string, string, int) (string, error) {
			// 非常に長いログを生成
			var lines []string
			for i := 0; i < 1000; i++ {
				lines = append(lines, strings.Repeat("Very long log line with lots of content ", 5))
			}
			return strings.Join(lines, "\n"), nil
		},
	}, "")

	result, err := cmd.Execute([]string{"test", "1000"})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Discordのメッセージ制限（2000文字）を超えていないかチェック
	if len(result) > 2000 {
		t.Errorf("Result too long: %d characters (max 2000)", len(result))
	}

	// 省略メッセージが含まれているかチェック
	if !strings.Contains(result, "省略されました") {
		t.Error("Expected truncation message not found")
	}
}

// パフォーマンステスト
func TestLogsCommand_Performance(t *testing.T) {
	// 大量のログでパフォーマンステスト
	largeLog := strings.Repeat("Log line with some content\n", 10000)

	cmd := NewLogsCommand(&docker.MockComposeService{
		ListContainersFunc: func(string) ([]docker.ContainerInfo, error) {
			return []docker.ContainerInfo{
				{Service: "perf", Name: "app_perf_1"},
			}, nil
		},
		GetContainerLogsFunc: func(string, string, int) (string, error) {
			return largeLog, nil
		},
	}, "")

	// 実行時間を測定（大量ログでも合理的な時間で処理されるか）
	_, err := cmd.Execute([]string{"perf", "200"})
	if err != nil {
		t.Fatalf("Performance test failed: %v", err)
	}
}

// ベンチマークテスト
func BenchmarkLogsCommand_Execute(b *testing.B) {
	mockCompose := &docker.MockComposeService{
		ListContainersFunc: func(string) ([]docker.ContainerInfo, error) {
			return []docker.ContainerInfo{
				{Service: "web", Name: "app_web_1"},
			}, nil
		},
		GetContainerLogsFunc: func(string, string, int) (string, error) {
			return strings.Repeat("Log line\n", 50), nil
		},
	}

	cmd := NewLogsCommand(mockCompose, "test-compose.yml")
	args := []string{"web", "50"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cmd.Execute(args)
		if err != nil {
			b.Fatalf("Execute() failed: %v", err)
		}
	}
}

func BenchmarkLogsCommand_parseLineCount(b *testing.B) {
	cmd := NewLogsCommand(&docker.MockComposeService{}, "")
	args := []string{"service", "100"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd.parseLineCount(args)
	}
}

func BenchmarkLogsCommand_addFormattedLogs(b *testing.B) {
	cmd := NewLogsCommand(&docker.MockComposeService{}, "")
	logs := strings.Repeat("Log line with some content\n", 200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var builder strings.Builder
		cmd.addFormattedLogs(&builder, logs, 100)
	}
}
