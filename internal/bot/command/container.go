package command

import (
	"fmt"
	"strings"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
)

// ContainerCommand handles the container command
type ContainerCommand struct {
	compose     docker.ComposeService
	composePath string
}

// NewContainerCommand creates a new ContainerCommand
func NewContainerCommand(compose docker.ComposeService, composePath string) *ContainerCommand {
	if composePath == "" {
		composePath = "docker-compose.yml"
	}
	return &ContainerCommand{
		compose:     compose,
		composePath: composePath,
	}
}

// Name returns the command name
func (c *ContainerCommand) Name() string {
	return "container"
}

// Description returns the command description
func (c *ContainerCommand) Description() string {
	return "個別コンテナの詳細情報を表示"
}

// Execute runs the command
func (c *ContainerCommand) Execute(args []string) (string, error) {
	if len(args) == 0 {
		return "使用方法: `@bot container <サービス名>`", nil
	}

	serviceName := args[0]

	// コンテナ一覧を取得して対象のコンテナを探す
	containers, err := c.compose.ListContainers(c.composePath)
	if err != nil {
		return "", fmt.Errorf("コンテナ情報の取得に失敗しました: %w", err)
	}

	var targetContainer *docker.ContainerInfo
	for i := range containers {
		if containers[i].Service == serviceName {
			targetContainer = &containers[i]
			break
		}
	}

	if targetContainer == nil {
		return fmt.Sprintf("❌ サービス '%s' が見つかりません", serviceName), nil
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📦 **%s の詳細情報**\n\n", FormatServiceName(serviceName)))

	// 基本情報
	builder.WriteString("**基本情報**\n")
	builder.WriteString(fmt.Sprintf("- コンテナ名: %s\n", targetContainer.Name))
	builder.WriteString(fmt.Sprintf("- コンテナID: %s\n", targetContainer.ID[:12]))

	// 状態
	statusIcon := getStatusIcon(targetContainer.State)
	builder.WriteString(fmt.Sprintf("- 状態: %s %s\n", statusIcon, targetContainer.State))
	if targetContainer.RunningFor != "" {
		builder.WriteString(fmt.Sprintf("- 稼働時間: %s\n", targetContainer.RunningFor))
	}

	// ヘルスチェック
	if targetContainer.HealthStatus != "" && targetContainer.HealthStatus != "none" {
		healthIcon := getHealthIcon(targetContainer.HealthStatus)
		builder.WriteString(fmt.Sprintf("- ヘルス: %s %s\n", healthIcon, targetContainer.HealthStatus))
	}

	// ポート
	if len(targetContainer.Ports) > 0 {
		builder.WriteString(fmt.Sprintf("- ポート: %s\n", strings.Join(targetContainer.Ports, ", ")))
	}

	// 実行中の場合はリソース使用状況を表示
	if strings.ToLower(targetContainer.State) == "running" {
		stats, err := c.compose.GetContainerStats(targetContainer.Name)
		if err == nil {
			builder.WriteString("\n**リソース使用状況**\n")
			builder.WriteString(fmt.Sprintf("- CPU使用率: %.1f%%\n", stats.CPUPercent))
			builder.WriteString(fmt.Sprintf("- メモリ使用率: %.1f%%\n", stats.MemoryPercent))
			builder.WriteString(fmt.Sprintf("- メモリ使用量: %s\n", stats.MemoryUsage))
			if stats.NetworkIO != "" && stats.NetworkIO != "0B / 0B" {
				builder.WriteString(fmt.Sprintf("- ネットワークI/O: %s\n", stats.NetworkIO))
			}
			if stats.BlockIO != "" && stats.BlockIO != "0B / 0B" {
				builder.WriteString(fmt.Sprintf("- ブロックI/O: %s\n", stats.BlockIO))
			}

			// 高負荷警告
			if stats.CPUPercent > 85.0 || stats.MemoryPercent > 90.0 {
				builder.WriteString("\n⚠️ **警告**\n")
				if stats.CPUPercent > 85.0 {
					builder.WriteString(fmt.Sprintf("- CPU使用率が高い状態です (%.1f%%)\n", stats.CPUPercent))
				}
				if stats.MemoryPercent > 90.0 {
					builder.WriteString(fmt.Sprintf("- メモリ使用率が高い状態です (%.1f%%)\n", stats.MemoryPercent))
				}
			}
		}
	}

	// 最近のログ（最後の10行）
	builder.WriteString("\n**最近のログ** (最後の10行)\n")
	builder.WriteString("```\n")

	logs, err := c.compose.GetContainerLogs(c.composePath, serviceName, 10)
	if err != nil {
		builder.WriteString("ログの取得に失敗しました\n")
	} else {
		// ログが長すぎる場合は切り詰める
		logLines := strings.Split(strings.TrimSpace(logs), "\n")
		for i, line := range logLines {
			if i >= 10 {
				break
			}
			// 各行を最大80文字に制限
			if len(line) > 80 {
				line = line[:77] + "..."
			}
			builder.WriteString(line + "\n")
		}
	}

	builder.WriteString("```\n")

	// 使用可能なコマンド
	builder.WriteString("\n**使用可能なコマンド**\n")
	if strings.ToLower(targetContainer.State) == "running" {
		builder.WriteString("- `@bot restart " + serviceName + "` - コンテナを再起動\n")
		builder.WriteString("- `@bot logs " + serviceName + " [行数]` - より多くのログを表示\n")
	} else {
		builder.WriteString("- `@bot game-info` から起動ボタンを使用してコンテナを起動\n")
	}

	return builder.String(), nil
}
