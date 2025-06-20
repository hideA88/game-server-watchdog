// Package command はDiscordボット用のコマンドハンドラーを提供します
package command

import (
	"fmt"
	"strings"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
)

const (
	// defaultComposePath はデフォルトのdocker-compose.ymlのパス
	defaultComposePath = "docker-compose.yml"
	// containerStateRunning は実行中のコンテナの状態
	containerStateRunning = "running"
	// containerMemoryZero はメモリ使用量が0の場合の表示
	containerMemoryZero = "0B / 0B"
	// defaultLogLines はデフォルトのログ行数
	defaultLogLines = 10
	// cpuHighThreshold はCPU使用率の高負荷閾値
	cpuHighThreshold = 85.0
	// memoryHighThreshold はメモリ使用率の高負荷閾値
	memoryHighThreshold = 90.0
	// maxLogLineLength は1行の最大文字数
	maxLogLineLength = 80
)

// ContainerCommand handles the container command
type ContainerCommand struct {
	compose     docker.ComposeService
	composePath string
}

// NewContainerCommand creates a new ContainerCommand
func NewContainerCommand(compose docker.ComposeService, composePath string) *ContainerCommand {
	if composePath == "" {
		composePath = defaultComposePath
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
	targetContainer, err := c.findContainer(serviceName)
	if err != nil {
		return "", err
	}
	if targetContainer == nil {
		return fmt.Sprintf("❌ サービス '%s' が見つかりません", serviceName), nil
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📦 **%s の詳細情報**\n\n", FormatServiceName(serviceName)))

	// 基本情報を追加
	c.addBasicInfo(&builder, targetContainer)

	// 実行中の場合はリソース使用状況を表示
	if strings.EqualFold(targetContainer.State, containerStateRunning) {
		c.addResourceInfo(&builder, targetContainer)
	}

	// 最近のログを追加
	c.addRecentLogs(&builder, serviceName)

	// 使用可能なコマンドを追加
	c.addAvailableCommands(&builder, serviceName, targetContainer.State)

	return builder.String(), nil
}

// findContainer は指定されたサービス名のコンテナを検索する
func (c *ContainerCommand) findContainer(serviceName string) (*docker.ContainerInfo, error) {
	containers, err := c.compose.ListContainers(c.composePath)
	if err != nil {
		return nil, fmt.Errorf("コンテナ情報の取得に失敗しました: %w", err)
	}

	for i := range containers {
		if containers[i].Service == serviceName {
			return &containers[i], nil
		}
	}
	return nil, nil
}

// addBasicInfo は基本情報を追加する
func (c *ContainerCommand) addBasicInfo(builder *strings.Builder, container *docker.ContainerInfo) {
	builder.WriteString("**基本情報**\n")
	fmt.Fprintf(builder, "- コンテナ名: %s\n", container.Name)
	fmt.Fprintf(builder, "- コンテナID: %s\n", container.ID[:12])

	// 状態
	statusIcon := GetStatusIcon(container.State)
	fmt.Fprintf(builder, "- 状態: %s %s\n", statusIcon, container.State)
	if container.RunningFor != "" {
		fmt.Fprintf(builder, "- 稼働時間: %s\n", container.RunningFor)
	}

	// ヘルスチェック
	if container.HealthStatus != "" && container.HealthStatus != "none" {
		healthIcon := GetHealthIcon(container.HealthStatus)
		fmt.Fprintf(builder, "- ヘルス: %s %s\n", healthIcon, container.HealthStatus)
	}

	// ポート
	if len(container.Ports) > 0 {
		fmt.Fprintf(builder, "- ポート: %s\n", strings.Join(container.Ports, ", "))
	}
}

// addResourceInfo はリソース使用状況を追加する
func (c *ContainerCommand) addResourceInfo(builder *strings.Builder, container *docker.ContainerInfo) {
	stats, err := c.compose.GetContainerStats(container.Name)
	if err != nil {
		return
	}

	builder.WriteString("\n**リソース使用状況**\n")
	fmt.Fprintf(builder, "- CPU使用率: %.1f%%\n", stats.CPUPercent)
	fmt.Fprintf(builder, "- メモリ使用率: %.1f%%\n", stats.MemoryPercent)
	fmt.Fprintf(builder, "- メモリ使用量: %s\n", stats.MemoryUsage)

	if stats.NetworkIO != "" && stats.NetworkIO != containerMemoryZero {
		fmt.Fprintf(builder, "- ネットワークI/O: %s\n", stats.NetworkIO)
	}
	if stats.BlockIO != "" && stats.BlockIO != containerMemoryZero {
		fmt.Fprintf(builder, "- ブロックI/O: %s\n", stats.BlockIO)
	}

	// 高負荷警告
	if stats.CPUPercent > cpuHighThreshold || stats.MemoryPercent > memoryHighThreshold {
		builder.WriteString("\n⚠️ **警告**\n")
		if stats.CPUPercent > cpuHighThreshold {
			fmt.Fprintf(builder, "- CPU使用率が高い状態です (%.1f%%)\n", stats.CPUPercent)
		}
		if stats.MemoryPercent > memoryHighThreshold {
			fmt.Fprintf(builder, "- メモリ使用率が高い状態です (%.1f%%)\n", stats.MemoryPercent)
		}
	}
}

// addRecentLogs は最近のログを追加する
func (c *ContainerCommand) addRecentLogs(builder *strings.Builder, serviceName string) {
	builder.WriteString("\n**最近のログ** (最後の10行)\n")
	builder.WriteString("```\n")

	logs, err := c.compose.GetContainerLogs(c.composePath, serviceName, defaultLogLines)
	if err != nil {
		builder.WriteString("ログの取得に失敗しました\n")
	} else {
		// ログが長すぎる場合は切り詰める
		logLines := strings.Split(strings.TrimSpace(logs), "\n")
		for i, line := range logLines {
			if i >= defaultLogLines {
				break
			}
			// 各行を最大80文字に制限
			if len(line) > maxLogLineLength {
				line = line[:maxLogLineLength-3] + "..."
			}
			builder.WriteString(line + "\n")
		}
	}

	builder.WriteString("```\n")
}

// addAvailableCommands は使用可能なコマンドを追加する
func (c *ContainerCommand) addAvailableCommands(builder *strings.Builder, serviceName, state string) {
	builder.WriteString("\n**使用可能なコマンド**\n")
	if strings.EqualFold(state, containerStateRunning) {
		builder.WriteString("- `@bot restart " + serviceName + "` - コンテナを再起動\n")
		builder.WriteString("- `@bot logs " + serviceName + " [行数]` - より多くのログを表示\n")
	} else {
		builder.WriteString("- `@bot monitor` から起動ボタンを使用してコンテナを起動\n")
	}
}
