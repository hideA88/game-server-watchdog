package command

import (
	"strings"
	"testing"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

func TestMonitorCommand_buildSystemInfo(t *testing.T) {
	tests := []struct {
		name         string
		sysInfo      *system.SystemInfo
		wantContains []string
	}{
		{
			name: "正常なシステム情報",
			sysInfo: &system.SystemInfo{
				CPUUsagePercent:   75.5,
				MemoryUsedPercent: 60.2,
				MemoryUsedGB:      8.5,
				MemoryTotalGB:     16.0,
				DiskUsedPercent:   45.0,
				DiskFreeGB:        110.5,
			},
			wantContains: []string{
				"📊 **ホストサーバー**",
				"CPU:",
				"75.5%",
				"MEM:",
				"60.2%",
				"8.5GB/16.0GB",
				"DISK:",
				"45.0%",
				"110.5GB free",
			},
		},
		{
			name:    "nilシステム情報",
			sysInfo: nil,
			wantContains: []string{
				"⚠️ システム情報の取得に失敗しました",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &MonitorCommand{}
			got := cmd.buildSystemInfo(tt.sysInfo)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("buildSystemInfo() does not contain %q\nGot: %s", want, got)
				}
			}
		})
	}
}

func TestMonitorCommand_checkAlerts(t *testing.T) { //nolint:funlen // テーブル駆動テストのため長い関数を許可
	tests := []struct {
		name       string
		sysInfo    *system.SystemInfo
		stats      []docker.ContainerStats
		wantAlerts int
		wantTypes  []string
	}{
		{
			name: "高負荷アラート",
			sysInfo: &system.SystemInfo{
				CPUUsagePercent:   90.0,
				MemoryUsedPercent: 95.0,
				DiskUsedPercent:   92.0,
			},
			stats: []docker.ContainerStats{
				{
					Name:          "project_minecraft_1",
					CPUPercent:    88.0,
					MemoryPercent: 93.0,
				},
			},
			wantAlerts: 5,
			wantTypes: []string{
				"CPU使用率が高い",
				"メモリ使用率が高い",
				"ディスク使用率が高い",
			},
		},
		{
			name: "アラートなし",
			sysInfo: &system.SystemInfo{
				CPUUsagePercent:   50.0,
				MemoryUsedPercent: 60.0,
				DiskUsedPercent:   70.0,
			},
			stats: []docker.ContainerStats{
				{
					Name:          "project_minecraft_1",
					CPUPercent:    40.0,
					MemoryPercent: 50.0,
				},
			},
			wantAlerts: 0,
		},
		{
			name:    "nilシステム情報",
			sysInfo: nil,
			stats: []docker.ContainerStats{
				{
					Name:          "project_minecraft_1",
					CPUPercent:    88.0,
					MemoryPercent: 50.0,
				},
			},
			wantAlerts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &MonitorCommand{}
			alerts := cmd.checkAlerts(tt.sysInfo, tt.stats)

			if len(alerts) != tt.wantAlerts {
				t.Errorf("checkAlerts() returned %d alerts, want %d", len(alerts), tt.wantAlerts)
			}

			// アラートタイプのチェック
			for _, wantType := range tt.wantTypes {
				found := false
				for _, alert := range alerts {
					if alert.Message == wantType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("checkAlerts() missing alert type %q", wantType)
				}
			}
		})
	}
}

func TestMonitorCommand_buildGameServerInfo(t *testing.T) {
	tests := []struct {
		name           string
		gameContainers []docker.ContainerInfo
		wantContains   []string
	}{
		{
			name: "稼働中のゲームサーバー",
			gameContainers: []docker.ContainerInfo{
				{
					Service:    "minecraft",
					State:      "running",
					RunningFor: "2 hours",
				},
				{
					Service: "terraria",
					State:   "stopped",
				},
			},
			wantContains: []string{
				"🎮 **ゲームサーバー状態**",
				"🟢 ⛏️ **Minecraft**: running (2 hours)",
				"🔴 🌳 **Terraria**: stopped",
			},
		},
		{
			name:           "ゲームサーバーなし",
			gameContainers: []docker.ContainerInfo{},
			wantContains: []string{
				"現在稼働中のゲームサーバーはありません",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &MonitorCommand{}
			got := cmd.buildGameServerInfo(tt.gameContainers)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("buildGameServerInfo() does not contain %q\nGot: %s", want, got)
				}
			}
		})
	}
}

func TestProgressBar(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		width int
		want  string
	}{
		{
			name:  "50%プログレスバー",
			value: 50.0,
			width: 10,
			want:  "█████░░░░░",
		},
		{
			name:  "0%プログレスバー",
			value: 0.0,
			width: 10,
			want:  "░░░░░░░░░░",
		},
		{
			name:  "100%プログレスバー",
			value: 100.0,
			width: 10,
			want:  "██████████",
		},
		{
			name:  "負の値",
			value: -10.0,
			width: 10,
			want:  "░░░░░░░░░░",
		},
		{
			name:  "100%超過",
			value: 150.0,
			width: 10,
			want:  "██████████",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := NewProgressBar(tt.value, tt.width)
			got := bar.String()

			if got != tt.want {
				t.Errorf("ProgressBar.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlert(t *testing.T) {
	alert := Alert{
		Component: "minecraft",
		Message:   "CPU使用率が高い",
		Value:     88.5,
	}

	// 構造体のフィールドが正しく設定されているか確認
	if alert.Component != "minecraft" {
		t.Errorf("Alert.Component = %v, want %v", alert.Component, "minecraft")
	}
	if alert.Message != "CPU使用率が高い" {
		t.Errorf("Alert.Message = %v, want %v", alert.Message, "CPU使用率が高い")
	}
	if alert.Value != 88.5 {
		t.Errorf("Alert.Value = %v, want %v", alert.Value, 88.5)
	}
}
