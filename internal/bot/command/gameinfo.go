package command

import (
	"fmt"
	"strings"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
)

// GameInfoCommand handles the game-info command
type GameInfoCommand struct {
	compose     docker.ComposeService
	composePath string
}

// NewGameInfoCommand creates a new GameInfoCommand
func NewGameInfoCommand(compose docker.ComposeService, composePath string) *GameInfoCommand {
	if composePath == "" {
		composePath = "docker-compose.yml"
	}
	return &GameInfoCommand{
		compose:     compose,
		composePath: composePath,
	}
}

// Name returns the command name
func (c *GameInfoCommand) Name() string {
	return "game-info"
}

// Description returns the command description
func (c *GameInfoCommand) Description() string {
	return "ゲームサーバーの稼働状況を表示"
}

// Execute runs the command
func (c *GameInfoCommand) Execute(args []string) (string, error) {
	containers, err := c.compose.ListContainers(c.composePath)
	if err != nil {
		return "", fmt.Errorf("コンテナ情報の取得に失敗しました: %w", err)
	}

	if len(containers) == 0 {
		return "🎮 **ゲームサーバー情報**\n\n現在稼働中のゲームサーバーはありません。", nil
	}

	var builder strings.Builder
	builder.WriteString("🎮 **ゲームサーバー情報**\n\n")

	for _, container := range containers {
		// Service name with icon
		icon := getGameIcon(container.Service)
		builder.WriteString(fmt.Sprintf("%s **%s** (%s)\n", icon,
			formatServiceName(container.Service), container.Service))

		// Container name
		builder.WriteString(fmt.Sprintf("  コンテナ: %s\n", container.Name))

		// Status with icon
		statusIcon := getStatusIcon(container.State)
		builder.WriteString(fmt.Sprintf("  状態: %s %s", statusIcon, container.State))
		if container.RunningFor != "" {
			builder.WriteString(fmt.Sprintf(" (%s)", container.RunningFor))
		}
		builder.WriteString("\n")

		// Ports
		if len(container.Ports) > 0 {
			builder.WriteString(fmt.Sprintf("  ポート: %s\n", strings.Join(container.Ports, ", ")))
		}

		// Health status if available
		if container.HealthStatus != "" && container.HealthStatus != "none" {
			healthIcon := getHealthIcon(container.HealthStatus)
			builder.WriteString(fmt.Sprintf("  ヘルス: %s %s\n", healthIcon, container.HealthStatus))
		}

		builder.WriteString("\n")
	}

	return builder.String(), nil
}

// getGameIcon returns an icon based on the game service name
func getGameIcon(service string) string {
	switch strings.ToLower(service) {
	case "minecraft":
		return "⛏️"
	case "rust":
		return "🔧"
	case "terraria":
		return "🌳"
	case "valheim":
		return "⚔️"
	case "ark":
		return "🦕"
	default:
		return "📦"
	}
}

// getStatusIcon returns an icon based on the container state
func getStatusIcon(state string) string {
	switch strings.ToLower(state) {
	case "running":
		return "🟢"
	case "stopped", "exited":
		return "🔴"
	case "restarting":
		return "🟡"
	case "paused":
		return "⏸️"
	default:
		return "❓"
	}
}

// getHealthIcon returns an icon based on the health status
func getHealthIcon(health string) string {
	switch strings.ToLower(health) {
	case "healthy":
		return "✅"
	case "unhealthy":
		return "❌"
	case "starting":
		return "🔄"
	default:
		return "❓"
	}
}

// formatServiceName formats the service name for display
func formatServiceName(service string) string {
	// Capitalize and format common game names
	switch strings.ToLower(service) {
	case "minecraft":
		return "Minecraft Server"
	case "rust":
		return "Rust Server"
	case "terraria":
		return "Terraria Server"
	case "valheim":
		return "Valheim Server"
	case "ark":
		return "ARK Server"
	default:
		// Capitalize first letter
		if len(service) > 0 {
			return strings.ToUpper(service[:1]) + service[1:]
		}
		return service
	}
}
