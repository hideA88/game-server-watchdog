package command

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

func TestMonitorCommand_ExecuteWithLongMessage(t *testing.T) {
	// 2000文字を超えるレスポンスを生成するテスト
	mockMonitor := &system.MockMonitor{
		SystemInfo: &system.SystemInfo{
			CPUUsagePercent:   65.5,
			MemoryUsedPercent: 76.9,
			DiskUsedPercent:   45.2,
			MemoryUsedGB:      12.3,
			MemoryTotalGB:     16.0,
			DiskFreeGB:        150.5,
		},
	}

	// 多数のコンテナを生成
	var containers []docker.ContainerInfo
	var stats []docker.ContainerStats
	for i := 0; i < 50; i++ {
		containers = append(containers, docker.ContainerInfo{
			ID:         fmt.Sprintf("container%d", i),
			Name:       fmt.Sprintf("very-long-container-name-for-testing-%d", i),
			Service:    fmt.Sprintf("service%d", i),
			State:      "running",
			RunningFor: "2 hours",
			Ports:      []string{"8080->8080/tcp", "9090->9090/tcp"},
		})
		stats = append(stats, docker.ContainerStats{
			Name:          fmt.Sprintf("very-long-container-name-for-testing-%d", i),
			CPUPercent:    float64(i % 100),
			MemoryPercent: float64((i * 2) % 100),
			MemoryUsage:   fmt.Sprintf("%d.%dGiB", i/10, i%10),
		})
	}

	mockCompose := &docker.MockComposeService{
		ListContainersFunc: func(_ string) ([]docker.ContainerInfo, error) {
			return containers, nil
		},
		ListGameContainersFunc: func(_ string) ([]docker.ContainerInfo, error) {
			return containers[:10], nil // ゲームコンテナは10個
		},
		GetAllContainersStatsFunc: func(_ string) ([]docker.ContainerStats, error) {
			return stats, nil
		},
	}

	cmd := NewMonitorCommand(context.Background(), mockCompose, mockMonitor, "")
	result, err := cmd.Execute([]string{})

	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}

	// 要約版が返されることを確認
	if !strings.Contains(result, "(要約版)") {
		t.Error("Expected summary version, but got full version")
	}

	if !strings.Contains(result, "詳細情報が多すぎるため要約版を表示しています") {
		t.Error("Expected summary message explanation")
	}

	// メッセージ長が2000文字以下であることを確認
	if len(result) > DiscordMessageLimit {
		t.Errorf("Message length %d exceeds Discord limit %d", len(result), DiscordMessageLimit)
	}
}

func TestMonitorCommand_collectMonitorDataWithErrors(t *testing.T) {
	tests := []struct {
		name               string
		systemError        error
		containerError     error
		gameError          error
		wantSystemError    bool
		wantContainerError bool
		wantGameError      bool
	}{
		{
			name:               "すべてのエラー",
			systemError:        fmt.Errorf("system error"),
			containerError:     fmt.Errorf("container error"),
			gameError:          fmt.Errorf("game error"),
			wantSystemError:    true,
			wantContainerError: true,
			wantGameError:      true,
		},
		{
			name:            "システムエラーのみ",
			systemError:     fmt.Errorf("system error"),
			wantSystemError: true,
		},
		{
			name:               "コンテナエラーのみ",
			containerError:     fmt.Errorf("container error"),
			wantContainerError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMonitor := &system.MockMonitor{
				SystemInfo: &system.SystemInfo{CPUUsagePercent: 50.0},
				Err:        tt.systemError,
			}

			mockCompose := &docker.MockComposeService{
				ListContainersFunc: func(_ string) ([]docker.ContainerInfo, error) {
					return nil, tt.containerError
				},
				ListGameContainersFunc: func(_ string) ([]docker.ContainerInfo, error) {
					return nil, tt.gameError
				},
			}

			cmd := NewMonitorCommand(context.Background(), mockCompose, mockMonitor, "")
			data, err := cmd.collectMonitorData()

			if err != nil {
				t.Errorf("collectMonitorData() returned unexpected error: %v", err)
			}

			// エラーの確認
			if (data.SystemError != nil) != tt.wantSystemError {
				t.Errorf("SystemError = %v, want error: %v", data.SystemError, tt.wantSystemError)
			}
			if (data.ContainerError != nil) != tt.wantContainerError {
				t.Errorf("ContainerError = %v, want error: %v", data.ContainerError, tt.wantContainerError)
			}
			if (data.GameError != nil) != tt.wantGameError {
				t.Errorf("GameError = %v, want error: %v", data.GameError, tt.wantGameError)
			}

			// コンテナエラーがある場合、統計情報が取得されないことを確認
			if tt.containerError != nil && data.Stats != nil {
				t.Error("Expected Stats to be nil when container error occurs")
			}
		})
	}
}

func TestMonitorCommand_formatContainerRowEdgeCases(t *testing.T) {
	cmd := &MonitorCommand{}

	tests := []struct {
		name      string
		container docker.ContainerInfo
		stats     map[string]*docker.ContainerStats
		want      []string
	}{
		{
			name: "長いサービス名",
			container: docker.ContainerInfo{
				Service: "very-long-service-name-that-exceeds-limit",
				State:   "running",
			},
			stats: map[string]*docker.ContainerStats{},
			want:  []string{"very-long-servi.."},
		},
		{
			name: "統計情報なし",
			container: docker.ContainerInfo{
				Service: "test",
				State:   "running",
			},
			stats: map[string]*docker.ContainerStats{},
			want:  []string{"-", "-"}, // CPU、メモリ共に "-"
		},
		{
			name: "メモリ使用量形式異常",
			container: docker.ContainerInfo{
				Name:    "test",
				Service: "test",
				State:   "running",
			},
			stats: map[string]*docker.ContainerStats{
				"test": {
					CPUPercent:  50.0,
					MemoryUsage: "invalid format",
				},
			},
			want: []string{"50.0%", "invalid "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cmd.formatContainerRow(&tt.container, tt.stats)

			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("formatContainerRow() does not contain %q\nGot: %s", want, got)
				}
			}
		})
	}
}
