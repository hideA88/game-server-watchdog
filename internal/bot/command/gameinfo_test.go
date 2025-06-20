package command

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
)

func TestGameInfoCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		composePath         string
		mockContainers      []docker.ContainerInfo
		mockError           error
		expectedContains    []string
		notExpectedContains []string
		wantErr             bool
	}{
		{
			name:        "複数のゲームサーバーが稼働中",
			composePath: "docker-compose.yml",
			mockContainers: []docker.ContainerInfo{
				{
					Name:         "minecraft-bedrock-server",
					Service:      "minecraft",
					State:        "running",
					RunningFor:   "2 hours",
					Ports:        []string{"19132:19132/udp", "19133:19133/udp"},
					HealthStatus: "healthy",
				},
				{
					Name:       "rust-server",
					Service:    "rust",
					State:      "running",
					RunningFor: "5 hours",
					Ports:      []string{"28015:28015", "28015:28015/udp", "8080:8080"},
				},
			},
			expectedContains: []string{
				"🎮 **ゲームサーバー情報**",
				"⛏️ **Minecraft**",
				"minecraft-bedrock-server",
				"🟢 running (2 hours)",
				"19132:19132/udp",
				"✅ healthy",
				"🔧 **Rust**",
				"rust-server",
				"🟢 running (5 hours)",
				"28015:28015",
			},
			wantErr: false,
		},
		{
			name:           "稼働中のサーバーがない",
			composePath:    "docker-compose.yml",
			mockContainers: []docker.ContainerInfo{},
			expectedContains: []string{
				"🎮 **ゲームサーバー情報**",
				"現在稼働中のゲームサーバーはありません",
			},
			wantErr: false,
		},
		{
			name:        "停止中のサーバー",
			composePath: "docker-compose.yml",
			mockContainers: []docker.ContainerInfo{
				{
					Name:    "terraria-server",
					Service: "terraria",
					State:   "stopped",
				},
			},
			expectedContains: []string{
				"🌳 **Terraria**",
				"🔴 stopped",
			},
			notExpectedContains: []string{
				"ポート:",
			},
			wantErr: false,
		},
		{
			name:        "再起動中のサーバー",
			composePath: "docker-compose.yml",
			mockContainers: []docker.ContainerInfo{
				{
					Name:    "valheim-server",
					Service: "valheim",
					State:   "restarting",
				},
			},
			expectedContains: []string{
				"⚔️ **Valheim**",
				"🟡 restarting",
			},
			wantErr: false,
		},
		{
			name:        "不明なゲームサーバー",
			composePath: "docker-compose.yml",
			mockContainers: []docker.ContainerInfo{
				{
					Name:    "unknown-game",
					Service: "unknown",
					State:   "running",
				},
			},
			expectedContains: []string{
				"📦 **Unknown**",
			},
			wantErr: false,
		},
		{
			name:        "docker-composeエラー",
			composePath: "docker-compose.yml",
			mockError:   fmt.Errorf("docker-compose not found"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの設定
			mockCompose := &docker.MockComposeService{
				ListGameContainersFunc: func(path string) ([]docker.ContainerInfo, error) {
					if path != tt.composePath {
						t.Errorf("Expected composePath %s, got %s", tt.composePath, path)
					}
					return tt.mockContainers, tt.mockError
				},
			}

			// コマンドの作成と実行
			cmd := NewGameInfoCommand(mockCompose, tt.composePath)
			result, err := cmd.Execute([]string{})

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 期待される文字列が含まれているかチェック
				for _, expected := range tt.expectedContains {
					if !strings.Contains(result, expected) {
						t.Errorf("Expected result to contain %q, but it didn't.\nGot:\n%s", expected, result)
					}
				}

				// 期待されない文字列が含まれていないかチェック
				for _, notExpected := range tt.notExpectedContains {
					if strings.Contains(result, notExpected) {
						t.Errorf("Expected result not to contain %q, but it did.\nGot:\n%s", notExpected, result)
					}
				}
			}
		})
	}
}

func TestGameInfoCommand_Metadata(t *testing.T) {
	t.Parallel()

	cmd := NewGameInfoCommand(nil, "")

	if cmd.Name() != "game-info" {
		t.Errorf("Name() = %v, want %v", cmd.Name(), "game-info")
	}

	if cmd.Description() != "ゲームサーバーの稼働状況を表示" {
		t.Errorf("Description() = %v, want %v", cmd.Description(), "ゲームサーバーの稼働状況を表示")
	}
}

func TestGameInfoCommand_DefaultPath(t *testing.T) {
	t.Parallel()

	mockCompose := &docker.MockComposeService{
		ListGameContainersFunc: func(path string) ([]docker.ContainerInfo, error) {
			// デフォルトパスが使用されているかチェック
			if path != "docker-compose.yml" {
				t.Errorf("Expected default path 'docker-compose.yml', got %s", path)
			}
			return []docker.ContainerInfo{}, nil
		},
	}

	// 空のパスでコマンドを作成
	cmd := NewGameInfoCommand(mockCompose, "")
	_, err := cmd.Execute([]string{})

	if err != nil {
		t.Errorf("Execute() unexpected error = %v", err)
	}
}

func TestGameInfoCommand_CanHandle(t *testing.T) {
	t.Parallel()

	cmd := NewGameInfoCommand(nil, "")

	tests := []struct {
		name     string
		customID string
		want     bool
	}{
		{
			name:     "start button",
			customID: "start_service_minecraft",
			want:     true,
		},
		{
			name:     "stop button",
			customID: "stop_service_rust-server",
			want:     true,
		},
		{
			name:     "unrelated custom ID",
			customID: "something_else",
			want:     false,
		},
		{
			name:     "empty custom ID",
			customID: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cmd.CanHandle(tt.customID); got != tt.want {
				t.Errorf("CanHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGameInfoCommand_HandleInteraction(t *testing.T) {
	// Skip this test for now as it requires proper Discord session mocking
	t.Skip("Skipping HandleInteraction test - requires proper Discord session mocking")
}

func TestFormatServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		want        string
	}{
		{
			name:        "lowercase service",
			serviceName: "minecraft",
			want:        "Minecraft",
		},
		{
			name:        "service with hyphen",
			serviceName: "rust-server",
			want:        "Rust Server",
		},
		{
			name:        "service with underscore",
			serviceName: "terraria_world",
			want:        "Terraria World",
		},
		{
			name:        "service with mixed separators",
			serviceName: "my-game_server",
			want:        "My Game Server",
		},
		{
			name:        "empty service name",
			serviceName: "",
			want:        "",
		},
		{
			name:        "single character",
			serviceName: "a",
			want:        "A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatServiceName(tt.serviceName); got != tt.want {
				t.Errorf("FormatServiceName() = %v, want %v", got, tt.want)
			}
		})
	}
}
