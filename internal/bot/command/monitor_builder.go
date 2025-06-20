package command

import (
	"fmt"
	"strings"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

// buildSystemInfo はシステム情報の表示を生成する
func (c *MonitorCommand) buildSystemInfo(sysInfo *system.SystemInfo) string {
	if sysInfo == nil {
		return "⚠️ システム情報の取得に失敗しました\n"
	}

	var builder strings.Builder
	builder.WriteString("📊 **ホストサーバー**\n")

	// CPU使用率
	cpuBar := NewProgressBar(sysInfo.CPUUsagePercent, 10)
	builder.WriteString(fmt.Sprintf("CPU: %s %.1f%%\n", cpuBar, sysInfo.CPUUsagePercent))

	// メモリ使用率
	memBar := NewProgressBar(sysInfo.MemoryUsedPercent, 10)
	builder.WriteString(fmt.Sprintf("MEM: %s %.1f%% (%.1fGB/%.1fGB)\n",
		memBar, sysInfo.MemoryUsedPercent, sysInfo.MemoryUsedGB, sysInfo.MemoryTotalGB))

	// ディスク使用率
	diskBar := NewProgressBar(sysInfo.DiskUsedPercent, 10)
	builder.WriteString(fmt.Sprintf("DISK: %s %.1f%% (%.1fGB free)\n",
		diskBar, sysInfo.DiskUsedPercent, sysInfo.DiskFreeGB))

	return builder.String()
}

// buildContainerTable はコンテナテーブルを生成する
func (c *MonitorCommand) buildContainerTable(containers []docker.ContainerInfo, statsMap map[string]*docker.ContainerStats) string {
	var builder strings.Builder
	builder.WriteString("\n📦 **コンテナ状況**\n")
	builder.WriteString("```\n")
	builder.WriteString("┌─────────────────┬────────┬────────┬────────┬────────┐\n")
	builder.WriteString("│ サービス         │ 状態   │ CPU    │ メモリ │ 稼働   │\n")
	builder.WriteString("├─────────────────┼────────┼────────┼────────┼────────┤\n")

	if len(containers) == 0 {
		builder.WriteString("│ 稼働中のコンテナはありません                      │\n")
	} else {
		for i := range containers {
			row := c.formatContainerRow(&containers[i], statsMap)
			builder.WriteString(row)
		}
	}

	builder.WriteString("└─────────────────┴────────┴────────┴────────┴────────┘\n")
	builder.WriteString("```\n")

	return builder.String()
}

// formatContainerRow は1行分のコンテナ情報をフォーマットする
func (c *MonitorCommand) formatContainerRow(container *docker.ContainerInfo, statsMap map[string]*docker.ContainerStats) string {
	// サービス名（最大17文字）
	serviceName := container.Service
	if len(serviceName) > 15 {
		serviceName = serviceName[:15] + ".."
	}
	serviceName = fmt.Sprintf("%-17s", serviceName)

	// 状態アイコン
	stateIcon := GetStatusIcon(container.State)
	state := fmt.Sprintf("%-8s", stateIcon)

	// リソース使用状況
	var cpu, memory string
	if stat, ok := statsMap[container.Name]; ok && stat != nil {
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

	return fmt.Sprintf("│%s│%s│%s│%s│%s│\n",
		serviceName, state, cpu, memory, runningFor)
}

// checkAlerts はアラートをチェックして返す
func (c *MonitorCommand) checkAlerts(sysInfo *system.SystemInfo, stats []docker.ContainerStats) []Alert {
	var alerts []Alert

	// コンテナのアラートチェック
	for i := range stats {
		if stats[i].CPUPercent > CPUAlertThreshold {
			alerts = append(alerts, Alert{
				Component: FormatServiceName(getServiceFromContainerName(stats[i].Name)),
				Message:   "CPU使用率が高い",
				Value:     stats[i].CPUPercent,
			})
		}
		if stats[i].MemoryPercent > MemoryAlertThreshold {
			alerts = append(alerts, Alert{
				Component: FormatServiceName(getServiceFromContainerName(stats[i].Name)),
				Message:   "メモリ使用率が高い",
				Value:     stats[i].MemoryPercent,
			})
		}
	}

	// ホストシステムのアラート
	if sysInfo != nil {
		if sysInfo.CPUUsagePercent > CPUAlertThreshold {
			alerts = append(alerts, Alert{
				Component: "ホストサーバー",
				Message:   "CPU使用率が高い",
				Value:     sysInfo.CPUUsagePercent,
			})
		}
		if sysInfo.MemoryUsedPercent > MemoryAlertThreshold {
			alerts = append(alerts, Alert{
				Component: "ホストサーバー",
				Message:   "メモリ使用率が高い",
				Value:     sysInfo.MemoryUsedPercent,
			})
		}
		if sysInfo.DiskUsedPercent > DiskAlertThreshold {
			alerts = append(alerts, Alert{
				Component: "ホストサーバー",
				Message:   "ディスク使用率が高い",
				Value:     sysInfo.DiskUsedPercent,
			})
		}
	}

	return alerts
}

// buildAlertSection はアラートセクションを生成する
func (c *MonitorCommand) buildAlertSection(alerts []Alert) string {
	var builder strings.Builder
	builder.WriteString("\n⚠️ **アラート**\n")

	if len(alerts) == 0 {
		builder.WriteString("- 現在アラートはありません\n")
	} else {
		for _, alert := range alerts {
			builder.WriteString(fmt.Sprintf("- %s: %s (%.1f%%)\n",
				alert.Component, alert.Message, alert.Value))
		}
	}

	return builder.String()
}

// buildGameServerInfo はゲームサーバー情報を生成する
func (c *MonitorCommand) buildGameServerInfo(gameContainers []docker.ContainerInfo) string {
	var builder strings.Builder
	builder.WriteString("\n🎮 **ゲームサーバー状態**\n")

	if len(gameContainers) == 0 {
		builder.WriteString("- 現在稼働中のゲームサーバーはありません\n")
	} else {
		for i := range gameContainers {
			// Status icon and name
			statusIcon := GetStatusIcon(gameContainers[i].State)
			gameIcon := GetGameIcon(gameContainers[i].Service)
			builder.WriteString(fmt.Sprintf("• %s %s **%s**: %s",
				statusIcon, gameIcon, FormatServiceName(gameContainers[i].Service), gameContainers[i].State))

			if strings.EqualFold(gameContainers[i].State, containerStateRunning) && gameContainers[i].RunningFor != "" {
				builder.WriteString(fmt.Sprintf(" (%s)", gameContainers[i].RunningFor))
			}
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// buildSummaryMessage は要約版メッセージを生成する
func (c *MonitorCommand) buildSummaryMessage(data *MonitorData) string {
	var builder strings.Builder
	builder.Grow(1024)
	builder.WriteString("🖥️ **システム監視ダッシュボード** (要約版)\n\n")

	if data.SystemInfo != nil {
		builder.WriteString(fmt.Sprintf("📊 **ホストサーバー**: CPU %.1f%% | MEM %.1f%% | DISK %.1f%%\n\n",
			data.SystemInfo.CPUUsagePercent, data.SystemInfo.MemoryUsedPercent, data.SystemInfo.DiskUsedPercent))
	}

	builder.WriteString("📦 **コンテナ数**: ")
	if data.Containers != nil {
		runningCount := 0
		for i := range data.Containers {
			if strings.EqualFold(data.Containers[i].State, containerStateRunning) {
				runningCount++
			}
		}
		builder.WriteString(fmt.Sprintf("%d稼働中 / %d合計\n", runningCount, len(data.Containers)))
	} else {
		builder.WriteString("取得失敗\n")
	}

	// アラート数を計算
	alerts := c.checkAlerts(data.SystemInfo, data.Stats)
	if len(alerts) > 0 {
		builder.WriteString(fmt.Sprintf("\n⚠️ **アラート数**: %d件\n", len(alerts)))
	}

	builder.WriteString("\n*詳細情報が多すぎるため要約版を表示しています*")
	return builder.String()
}
