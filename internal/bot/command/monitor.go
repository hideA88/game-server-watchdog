package command

import (
	"fmt"
	"strings"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

// MonitorCommand handles the monitor command
type MonitorCommand struct {
	compose     docker.ComposeService
	monitor     system.Monitor
	composePath string
}

// NewMonitorCommand creates a new MonitorCommand
func NewMonitorCommand(compose docker.ComposeService, monitor system.Monitor, composePath string) *MonitorCommand {
	if composePath == "" {
		composePath = "docker-compose.yml"
	}
	return &MonitorCommand{
		compose:     compose,
		monitor:     monitor,
		composePath: composePath,
	}
}

// Name returns the command name
func (c *MonitorCommand) Name() string {
	return "monitor"
}

// Description returns the command description
func (c *MonitorCommand) Description() string {
	return "システムとコンテナのリアルタイム監視ダッシュボードを表示"
}

// Execute runs the command
func (c *MonitorCommand) Execute(args []string) (string, error) {
	var builder strings.Builder
	builder.WriteString("🖥️ **システム監視ダッシュボード**\n\n")

	// ホストシステム情報を取得
	sysInfo, err := c.monitor.GetSystemInfo()
	if err != nil {
		builder.WriteString("⚠️ システム情報の取得に失敗しました\n")
	} else {
		builder.WriteString("📊 **ホストサーバー**\n")

		// CPU使用率
		cpuBar := createProgressBar(sysInfo.CPUUsagePercent, 10)
		builder.WriteString(fmt.Sprintf("CPU: %s %.1f%%\n", cpuBar, sysInfo.CPUUsagePercent))

		// メモリ使用率
		memBar := createProgressBar(sysInfo.MemoryUsedPercent, 10)
		builder.WriteString(fmt.Sprintf("MEM: %s %.1f%% (%.1fGB/%.1fGB)\n",
			memBar, sysInfo.MemoryUsedPercent, sysInfo.MemoryUsedGB, sysInfo.MemoryTotalGB))

		// ディスク使用率
		diskBar := createProgressBar(sysInfo.DiskUsedPercent, 10)
		builder.WriteString(fmt.Sprintf("DISK: %s %.1f%% (%.1fGB free)\n",
			diskBar, sysInfo.DiskUsedPercent, sysInfo.DiskFreeGB))
	}

	builder.WriteString("\n📦 **コンテナ状況**\n")
	builder.WriteString("```\n")
	builder.WriteString("┌─────────────────┬────────┬────────┬────────┬────────┐\n")
	builder.WriteString("│ サービス         │ 状態   │ CPU    │ メモリ │ 稼働   │\n")
	builder.WriteString("├─────────────────┼────────┼────────┼────────┼────────┤\n")

	// コンテナ情報を取得
	containers, err := c.compose.ListContainers(c.composePath)
	if err != nil {
		builder.WriteString("│ エラー: コンテナ情報の取得に失敗しました         │\n")
	} else {
		// コンテナの統計情報を取得
		stats, _ := c.compose.GetAllContainersStats(c.composePath)
		statsMap := make(map[string]*docker.ContainerStats)
		for i := range stats {
			statsMap[stats[i].Name] = &stats[i]
		}

		for _, container := range containers {
			// サービス名（最大17文字）
			serviceName := container.Service
			if len(serviceName) > 15 {
				serviceName = serviceName[:15] + ".."
			}
			serviceName = fmt.Sprintf("%-17s", serviceName)

			// 状態アイコン
			var stateIcon string
			switch strings.ToLower(container.State) {
			case "running":
				stateIcon = "🟢"
			case "stopped", "exited":
				stateIcon = "🔴"
			case "restarting":
				stateIcon = "🟡"
			default:
				stateIcon = "❓"
			}
			state := fmt.Sprintf("%-8s", stateIcon)

			// リソース使用状況
			var cpu, memory string
			if stat, ok := statsMap[container.Name]; ok {
				cpu = fmt.Sprintf("%6.1f%%", stat.CPUPercent)
				// メモリ使用量から数値を抽出
				memParts := strings.Split(stat.MemoryUsage, " / ")
				if len(memParts) > 0 {
					memory = fmt.Sprintf("%-8s", memParts[0])
				} else {
					memory = fmt.Sprintf("%-8s", "-")
				}
			} else {
				cpu = fmt.Sprintf("%-8s", "-")
				memory = fmt.Sprintf("%-8s", "-")
			}

			// 稼働時間
			runningFor := container.RunningFor
			if runningFor == "" {
				runningFor = "-"
			}
			if len(runningFor) > 8 {
				runningFor = runningFor[:7] + "."
			}
			runningFor = fmt.Sprintf("%-8s", runningFor)

			builder.WriteString(fmt.Sprintf("│%s│%s│%s│%s│%s│\n",
				serviceName, state, cpu, memory, runningFor))
		}
	}

	builder.WriteString("└─────────────────┴────────┴────────┴────────┴────────┘\n")
	builder.WriteString("```\n")

	// アラート情報を追加
	builder.WriteString("\n⚠️ **アラート**\n")
	alertCount := 0

	// CPU使用率が高いコンテナをチェック
	if stats, err := c.compose.GetAllContainersStats(c.composePath); err == nil {
		for _, stat := range stats {
			if stat.CPUPercent > 85.0 {
				builder.WriteString(fmt.Sprintf("- %s: CPU使用率が高い (%.1f%%)\n",
					FormatServiceName(getServiceFromContainerName(stat.Name)), stat.CPUPercent))
				alertCount++
			}
			if stat.MemoryPercent > 90.0 {
				builder.WriteString(fmt.Sprintf("- %s: メモリ使用率が高い (%.1f%%)\n",
					FormatServiceName(getServiceFromContainerName(stat.Name)), stat.MemoryPercent))
				alertCount++
			}
		}
	}

	// ホストシステムのアラート
	if sysInfo != nil {
		if sysInfo.CPUUsagePercent > 85.0 {
			builder.WriteString(fmt.Sprintf("- ホストサーバー: CPU使用率が高い (%.1f%%)\n", sysInfo.CPUUsagePercent))
			alertCount++
		}
		if sysInfo.MemoryUsedPercent > 90.0 {
			builder.WriteString(fmt.Sprintf("- ホストサーバー: メモリ使用率が高い (%.1f%%)\n", sysInfo.MemoryUsedPercent))
			alertCount++
		}
		if sysInfo.DiskUsedPercent > 90.0 {
			builder.WriteString(fmt.Sprintf("- ホストサーバー: ディスク使用率が高い (%.1f%%)\n", sysInfo.DiskUsedPercent))
			alertCount++
		}
	}

	if alertCount == 0 {
		builder.WriteString("- 現在アラートはありません\n")
	}

	return builder.String(), nil
}

// createProgressBar creates a text-based progress bar
func createProgressBar(percent float64, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int(percent * float64(width) / 100)
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return bar
}

// getServiceFromContainerName extracts service name from container name
func getServiceFromContainerName(containerName string) string {
	// Docker Composeのコンテナ名は通常 "project_service_1" の形式
	parts := strings.Split(containerName, "_")
	if len(parts) >= 2 {
		return parts[1]
	}
	return containerName
}
