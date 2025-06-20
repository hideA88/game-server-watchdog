package command

import (
	"strings"
	"testing"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

func TestMonitorCommand_buildSummaryMessage(t *testing.T) {
	tests := []struct {
		name         string
		data         *MonitorData
		wantContains []string
	}{
		{
			name: "正常なデータ",
			data: &MonitorData{
				SystemInfo: &system.SystemInfo{
					CPUUsagePercent:   75.5,
					MemoryUsedPercent: 60.2,
					DiskUsedPercent:   45.0,
				},
				Containers: []docker.ContainerInfo{
					{State: "running"},
					{State: "stopped"},
					{State: "running"},
				},
				Stats: []docker.ContainerStats{
					{CPUPercent: 90.0},
				},
			},
			wantContains: []string{
				"システム監視ダッシュボード** (要約版)",
				"CPU 75.5% | MEM 60.2% | DISK 45.0%",
				"2稼働中 / 3合計",
				"⚠️ **アラート数**: 1件",
				"詳細情報が多すぎるため要約版を表示しています",
			},
		},
		{
			name: "システム情報なし",
			data: &MonitorData{
				SystemInfo: nil,
				Containers: []docker.ContainerInfo{
					{State: "running"},
				},
			},
			wantContains: []string{
				"1稼働中 / 1合計",
				"詳細情報が多すぎるため要約版を表示しています",
			},
		},
		{
			name: "コンテナ情報なし",
			data: &MonitorData{
				SystemInfo: &system.SystemInfo{
					CPUUsagePercent: 50.0,
				},
				Containers: nil,
			},
			wantContains: []string{
				"取得失敗",
				"詳細情報が多すぎるため要約版を表示しています",
			},
		},
		{
			name: "アラートなし",
			data: &MonitorData{
				SystemInfo: &system.SystemInfo{
					CPUUsagePercent:   50.0,
					MemoryUsedPercent: 60.0,
					DiskUsedPercent:   70.0,
				},
				Containers: []docker.ContainerInfo{
					{State: "running"},
				},
				Stats: []docker.ContainerStats{
					{CPUPercent: 50.0, MemoryPercent: 60.0},
				},
			},
			wantContains: []string{
				"1稼働中 / 1合計",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &MonitorCommand{}
			got := cmd.buildSummaryMessage(tt.data)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("buildSummaryMessage() does not contain %q\nGot: %s", want, got)
				}
			}

			// アラートがない場合は、アラート数が含まれていないことを確認
			if tt.name == "アラートなし" {
				if strings.Contains(got, "アラート数") {
					t.Error("buildSummaryMessage() should not contain alert count when there are no alerts")
				}
			}
		})
	}
}
