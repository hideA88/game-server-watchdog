package command

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

func TestMonitorCommand_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "コマンド名がmonitorであること",
			want: "monitor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewMonitorCommand(context.Background(), &docker.MockComposeService{}, &system.MockMonitor{}, "")
			if got := cmd.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonitorCommand_Description(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "説明文が正しいこと",
			want: "システムとゲームサーバーの監視情報を表示（操作ボタン付き）",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewMonitorCommand(context.Background(), &docker.MockComposeService{}, &system.MockMonitor{}, "")
			if got := cmd.Description(); got != tt.want {
				t.Errorf("Description() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonitorCommand_Execute(t *testing.T) { //nolint:funlen // テーブル駆動テストのため長い関数を許可
	tests := []struct {
		name           string
		systemInfo     *system.SystemInfo
		systemError    error
		containers     []docker.ContainerInfo
		containerErr   error
		gameContainers []docker.ContainerInfo
		gameErr        error
		stats          []docker.ContainerStats
		wantContains   []string
		wantErr        bool
		wantNotContain []string
	}{
		{
			name: "正常な監視情報を表示",
			systemInfo: &system.SystemInfo{
				CPUUsagePercent:   65.5,
				MemoryUsedGB:      12.3,
				MemoryTotalGB:     16.0,
				MemoryUsedPercent: 76.9,
				DiskFreeGB:        150.5,
				DiskUsedPercent:   45.2,
			},
			containers: []docker.ContainerInfo{
				{
					ID:         "abc123def456",
					Name:       "minecraft-server",
					Service:    "minecraft",
					State:      "running",
					RunningFor: "2 hours",
					Ports:      []string{"25565->25565/tcp"},
				},
			},
			gameContainers: []docker.ContainerInfo{
				{
					ID:         "abc123def456",
					Name:       "minecraft-server",
					Service:    "minecraft",
					State:      "running",
					RunningFor: "2 hours",
				},
			},
			stats: []docker.ContainerStats{
				{
					Name:          "minecraft-server",
					CPUPercent:    45.2,
					MemoryPercent: 68.7,
					MemoryUsage:   "2.3GiB",
				},
			},
			wantContains: []string{
				"システム監視ダッシュボード",
				"CPU:",
				"65.5%",
				"MEM:",
				"76.9%",
				"12.3GB/16.0GB",
				"DISK:",
				"45.2%",
				"150.5GB free",
				"コンテナ状況",
				"minecraft",
				"45.2%",
				"2.3GiB",
				"2 hours",
				"🟢 ⛏️ **Minecraft**: running (2 hours)",
			},
		},
		{
			name:        "システム情報取得エラー",
			systemError: fmt.Errorf("system error"),
			containers: []docker.ContainerInfo{
				{
					ID:      "abc123",
					Name:    "test-container",
					Service: "test",
					State:   "stopped",
				},
			},
			wantContains: []string{
				"⚠️ システム情報の取得に失敗しました",
				"コンテナ状況",
			},
		},
		{
			name: "コンテナ情報取得エラー",
			systemInfo: &system.SystemInfo{
				CPUUsagePercent: 50.0,
			},
			containerErr: fmt.Errorf("container error"),
			wantContains: []string{
				"⚠️ **コンテナ情報の取得に失敗しました**",
				"エラー: container error",
			},
		},
		{
			name: "高負荷アラート表示",
			systemInfo: &system.SystemInfo{
				CPUUsagePercent:   90.0,
				MemoryUsedPercent: 95.0,
				DiskUsedPercent:   92.0,
			},
			stats: []docker.ContainerStats{
				{
					Name:          "game-server",
					CPUPercent:    88.0,
					MemoryPercent: 93.0,
				},
			},
			wantContains: []string{
				"⚠️ **アラート**",
				"ホストサーバー: CPU使用率が高い",
				"ホストサーバー: メモリ使用率が高い",
				"ホストサーバー: ディスク使用率が高い",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMonitor := &system.MockMonitor{
				SystemInfo: tt.systemInfo,
				Err:        tt.systemError,
			}
			mockCompose := &docker.MockComposeService{
				ListContainersFunc: func(_ string) ([]docker.ContainerInfo, error) {
					return tt.containers, tt.containerErr
				},
				ListGameContainersFunc: func(_ string) ([]docker.ContainerInfo, error) {
					return tt.gameContainers, tt.gameErr
				},
				GetAllContainersStatsFunc: func(_ string) ([]docker.ContainerStats, error) {
					return tt.stats, nil
				},
				GetContainerStatsFunc: func(containerName string) (*docker.ContainerStats, error) {
					for i := range tt.stats {
						if tt.stats[i].Name == containerName {
							return &tt.stats[i], nil
						}
					}
					return nil, nil
				},
			}

			cmd := NewMonitorCommand(context.Background(), mockCompose, mockMonitor, "")
			result, err := cmd.Execute([]string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("Execute() result does not contain %q\nGot: %s", want, result)
				}
			}
		})
	}
}

func TestMonitorCommand_GetComponents(t *testing.T) { //nolint:funlen // テーブル駆動テストのため長い関数を許可
	tests := []struct {
		name        string
		containers  []docker.ContainerInfo
		err         error
		wantButtons int
		wantErr     bool
	}{
		{
			name: "停止中のコンテナに起動ボタン",
			containers: []docker.ContainerInfo{
				{
					Service: "minecraft",
					State:   "stopped",
				},
			},
			wantButtons: 1,
		},
		{
			name: "稼働中のコンテナに停止ボタン",
			containers: []docker.ContainerInfo{
				{
					Service: "terraria",
					State:   "running",
				},
			},
			wantButtons: 1,
		},
		{
			name: "複数のコンテナ",
			containers: []docker.ContainerInfo{
				{
					Service: "minecraft",
					State:   "stopped",
				},
				{
					Service: "terraria",
					State:   "running",
				},
			},
			wantButtons: 2,
		},
		{
			name:        "コンテナ情報取得エラー",
			err:         fmt.Errorf("error"),
			wantButtons: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompose := &docker.MockComposeService{
				ListGameContainersFunc: func(_ string) ([]docker.ContainerInfo, error) {
					return tt.containers, tt.err
				},
			}

			cmd := NewMonitorCommand(context.Background(), mockCompose, &system.MockMonitor{}, "")
			components, err := cmd.GetComponents([]string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("GetComponents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// ボタン数をカウント
			buttonCount := 0
			for _, comp := range components {
				if row, ok := comp.(discordgo.ActionsRow); ok {
					buttonCount += len(row.Components)
				}
			}

			if buttonCount != tt.wantButtons {
				t.Errorf("GetComponents() button count = %v, want %v", buttonCount, tt.wantButtons)
			}
		})
	}
}

func TestMonitorCommand_CanHandle(t *testing.T) {
	tests := []struct {
		name     string
		customID string
		want     bool
	}{
		{
			name:     "起動ボタンのカスタムID",
			customID: "start_service_minecraft",
			want:     true,
		},
		{
			name:     "停止ボタンのカスタムID",
			customID: "stop_service_terraria",
			want:     true,
		},
		{
			name:     "無関係のカスタムID",
			customID: "something_else",
			want:     false,
		},
		{
			name:     "空のカスタムID",
			customID: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewMonitorCommand(context.Background(), &docker.MockComposeService{}, &system.MockMonitor{}, "")
			if got := cmd.CanHandle(tt.customID); got != tt.want {
				t.Errorf("CanHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetGameIcon(t *testing.T) {
	tests := []struct {
		name    string
		service string
		want    string
	}{
		{
			name:    "minecraft",
			service: "minecraft",
			want:    "⛏️",
		},
		{
			name:    "rust",
			service: "rust",
			want:    "🔧",
		},
		{
			name:    "terraria",
			service: "terraria",
			want:    "🌳",
		},
		{
			name:    "valheim",
			service: "valheim",
			want:    "⚔️",
		},
		{
			name:    "ark",
			service: "ark",
			want:    "🦕",
		},
		{
			name:    "unknown",
			service: "unknown-game",
			want:    "📦",
		},
		{
			name:    "大文字",
			service: "MINECRAFT",
			want:    "⛏️",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGameIcon(tt.service); got != tt.want {
				t.Errorf("GetGameIcon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStatusIcon(t *testing.T) {
	tests := []struct {
		name  string
		state string
		want  string
	}{
		{
			name:  "running",
			state: "running",
			want:  "🟢",
		},
		{
			name:  "stopped",
			state: "stopped",
			want:  "🔴",
		},
		{
			name:  "exited",
			state: "exited",
			want:  "🔴",
		},
		{
			name:  "restarting",
			state: "restarting",
			want:  "🟡",
		},
		{
			name:  "paused",
			state: "paused",
			want:  "⏸️",
		},
		{
			name:  "unknown",
			state: "unknown",
			want:  "❓",
		},
		{
			name:  "大文字",
			state: "RUNNING",
			want:  "🟢",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStatusIcon(tt.state); got != tt.want {
				t.Errorf("GetStatusIcon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatServiceName(t *testing.T) {
	tests := []struct {
		name    string
		service string
		want    string
	}{
		{
			name:    "ハイフン区切り",
			service: "minecraft-server",
			want:    "Minecraft Server",
		},
		{
			name:    "アンダースコア区切り",
			service: "game_server",
			want:    "Game Server",
		},
		{
			name:    "単一単語",
			service: "minecraft",
			want:    "Minecraft",
		},
		{
			name:    "空文字列",
			service: "",
			want:    "",
		},
		{
			name:    "複数のハイフンとアンダースコア",
			service: "my-game_server-test",
			want:    "My Game Server Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatServiceName(tt.service); got != tt.want {
				t.Errorf("FormatServiceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHealthIcon(t *testing.T) {
	tests := []struct {
		name   string
		health string
		want   string
	}{
		{
			name:   "healthy",
			health: "healthy",
			want:   "✅",
		},
		{
			name:   "unhealthy",
			health: "unhealthy",
			want:   "❌",
		},
		{
			name:   "starting",
			health: "starting",
			want:   "🔄",
		},
		{
			name:   "unknown",
			health: "unknown",
			want:   "❓",
		},
		{
			name:   "大文字",
			health: "HEALTHY",
			want:   "✅",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHealthIcon(tt.health); got != tt.want {
				t.Errorf("GetHealthIcon() = %v, want %v", got, tt.want)
			}
		})
	}
}
