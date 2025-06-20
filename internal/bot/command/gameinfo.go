package command

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/hideA88/game-server-watchdog/pkg/docker"
)

// GameInfoCommand handles the game-info command
type GameInfoCommand struct {
	compose           docker.ComposeService
	composePath       string
	serviceOperations *sync.Map // サービス名をキーとした操作ロック
}

// NewGameInfoCommand creates a new GameInfoCommand
func NewGameInfoCommand(compose docker.ComposeService, composePath string) *GameInfoCommand {
	if composePath == "" {
		composePath = "docker-compose.yml"
	}
	return &GameInfoCommand{
		compose:           compose,
		composePath:       composePath,
		serviceOperations: &sync.Map{},
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
			FormatServiceName(container.Service), container.Service))

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

// FormatServiceName formats the service name for display
func FormatServiceName(service string) string {
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

// GetComponents returns Discord message components for the game info command
func (c *GameInfoCommand) GetComponents(args []string) ([]discordgo.MessageComponent, error) {
	containers, err := c.compose.ListContainers(c.composePath)
	if err != nil {
		return nil, fmt.Errorf("コンテナ情報の取得に失敗しました: %w", err)
	}

	var components []discordgo.MessageComponent
	var buttons []discordgo.MessageComponent

	// ボタン数のカウント（最大値を超えないように）
	buttonCount := 0
	for _, container := range containers {
		// 最大ボタン数に達したら終了
		if buttonCount >= docker.MaxTotalButtons {
			break
		}
		// 停止中のコンテナに対しては起動ボタンを追加
		if strings.ToLower(container.State) == "stopped" || strings.ToLower(container.State) == "exited" {
			button := discordgo.Button{
				Label:    fmt.Sprintf("🚀 %s を起動", FormatServiceName(container.Service)),
				Style:    discordgo.SuccessButton,
				CustomID: fmt.Sprintf("start_service_%s", container.Service),
			}
			buttons = append(buttons, button)
			buttonCount++
		} else if strings.ToLower(container.State) == "running" {
			// 稼働中のコンテナに対しては停止ボタンを追加
			button := discordgo.Button{
				Label:    fmt.Sprintf("🛑 %s を停止", FormatServiceName(container.Service)),
				Style:    discordgo.DangerButton,
				CustomID: fmt.Sprintf("stop_service_%s", container.Service),
			}
			buttons = append(buttons, button)
			buttonCount++
		}
	}

	// ボタンがある場合のみコンポーネントを返す
	if len(buttons) > 0 {
		// MaxButtonsPerRow個ずつのボタンをアクションローに分割
		for i := 0; i < len(buttons); i += docker.MaxButtonsPerRow {
			end := i + docker.MaxButtonsPerRow
			if end > len(buttons) {
				end = len(buttons)
			}
			
			row := discordgo.ActionsRow{
				Components: buttons[i:end],
			}
			components = append(components, row)
			
			// 最大MaxButtonRows行まで
			if len(components) >= docker.MaxButtonRows {
				break
			}
		}
	}

	return components, nil
}

// StartService starts a specific service
func (c *GameInfoCommand) StartService(serviceName string) error {
	return c.compose.StartService(c.composePath, serviceName)
}

// StopService stops a specific service
func (c *GameInfoCommand) StopService(serviceName string) error {
	return c.compose.StopService(c.composePath, serviceName)
}

// CanHandle は指定されたカスタムIDを処理できるかどうかを返す
func (c *GameInfoCommand) CanHandle(customID string) bool {
	return strings.HasPrefix(customID, "start_service_") || strings.HasPrefix(customID, "stop_service_")
}

// HandleInteraction はサービスの起動/停止インタラクションを処理する
func (c *GameInfoCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Type != discordgo.InteractionMessageComponent {
		return fmt.Errorf("unexpected interaction type: %v", i.Type)
	}

	data := i.MessageComponentData()
	
	// サービス名と操作を判定
	var serviceName string
	var isStart bool
	
	if strings.HasPrefix(data.CustomID, "start_service_") {
		serviceName = strings.TrimPrefix(data.CustomID, "start_service_")
		isStart = true
	} else if strings.HasPrefix(data.CustomID, "stop_service_") {
		serviceName = strings.TrimPrefix(data.CustomID, "stop_service_")
		isStart = false
	} else {
		return fmt.Errorf("unknown custom ID: %s", data.CustomID)
	}
	
	// 操作ロックをチェック
	if _, loaded := c.serviceOperations.LoadOrStore(serviceName, true); loaded {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("⚠️ %s は現在操作中です。しばらくお待ちください。", FormatServiceName(serviceName)),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	
	// Defer応答を送信（3秒以内）
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		c.serviceOperations.Delete(serviceName)
		return fmt.Errorf("failed to send defer response: %w", err)
	}

	// サービス操作処理を実行
	go c.handleServiceOperation(s, i, serviceName, isStart)
	
	return nil
}

// handleServiceOperation はサービスの起動/停止処理を行う
func (c *GameInfoCommand) handleServiceOperation(s *discordgo.Session, i *discordgo.InteractionCreate, serviceName string, isStart bool) {
	// 処理完了時にロックを解放
	defer c.serviceOperations.Delete(serviceName)
	
	// サービス操作を実行
	var err error
	var successMessage string
	var errorPrefix string
	
	if isStart {
		err = c.StartService(serviceName)
		successMessage = fmt.Sprintf("✅ %s を起動しました！", FormatServiceName(serviceName))
		errorPrefix = "起動"
	} else {
		err = c.StopService(serviceName)
		successMessage = fmt.Sprintf("🛑 %s を停止しました。", FormatServiceName(serviceName))
		errorPrefix = "停止"
	}
	
	// 結果に応じてメッセージを送信
	var content string
	if err != nil {
		content = fmt.Sprintf("❌ %s の%sに失敗しました: %v", FormatServiceName(serviceName), errorPrefix, err)
		log.Printf("Service operation failed: %v", err)
	} else {
		content = successMessage
	}
	
	// フォローアップメッセージを送信
	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
	if err != nil {
		log.Printf("Failed to send followup message: %v", err)
	}
}
