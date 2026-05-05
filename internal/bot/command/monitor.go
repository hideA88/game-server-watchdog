package command

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"golang.org/x/sync/errgroup"

	"github.com/hideA88/game-server-watchdog/internal/bot/usermsg"
	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/logging"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

const (
	// CPUAlertThreshold はCPU使用率のアラート閾値
	CPUAlertThreshold = 85.0
	// MemoryAlertThreshold はメモリ使用率のアラート闾値
	MemoryAlertThreshold = 90.0
	// DiskAlertThreshold はディスク使用率のアラート闾値
	DiskAlertThreshold = 90.0

	// DiscordMessageLimit はDiscordメッセージの最大文字数
	DiscordMessageLimit = 2000

	// ServiceOperationTimeout は操作のタイムアウト時間
	ServiceOperationTimeout = 60 * time.Second

	// containerStateStopped は停止中のコンテナの状態
	containerStateStopped = "stopped"
	// containerStateExited は終了したコンテナの状態
	containerStateExited = "exited"
	// gameServiceMinecraft はMinecraftサービス名
	gameServiceMinecraft = "minecraft"
	// statusIconUnknown は不明な状態のアイコン
	statusIconUnknown = "❓"
)

// MonitorCommand handles the monitor command
type MonitorCommand struct {
	compose           docker.ComposeService
	monitor           system.Monitor
	composePath       string
	serviceOperations *sync.Map // サービス名をキーとした操作ロック
	ctx               context.Context
}

// NewMonitorCommand creates a new MonitorCommand
func NewMonitorCommand(
	ctx context.Context,
	compose docker.ComposeService,
	monitor system.Monitor,
	composePath string,
) *MonitorCommand {
	if composePath == "" {
		composePath = defaultComposePath
	}
	return &MonitorCommand{
		compose:           compose,
		monitor:           monitor,
		composePath:       composePath,
		serviceOperations: &sync.Map{},
		ctx:               ctx,
	}
}

// Name returns the command name
func (c *MonitorCommand) Name() string {
	return "monitor"
}

// Description returns the command description
func (c *MonitorCommand) Description() string {
	return "システムとゲームサーバーの監視情報を表示（操作ボタン付き）"
}

// Execute runs the command
func (c *MonitorCommand) Execute(_ []string) (string, error) {
	// データ収集
	data, err := c.collectMonitorData()
	if err != nil {
		return "", fmt.Errorf("監視データの収集に失敗しました: %w", err)
	}

	// レポート生成
	report := c.buildMonitorReport(data)

	// メッセージ長チェック
	if len(report) > DiscordMessageLimit {
		return c.buildSummaryMessage(data), nil
	}

	return report, nil
}

// collectMonitorData は監視データを収集する
func (c *MonitorCommand) collectMonitorData() (*MonitorData, error) {
	data := &MonitorData{}

	// 1分のタイムアウトを設定
	ctx, cancel := context.WithTimeout(c.ctx, 1*time.Minute)
	defer cancel()

	// コンテキスト付きのerrgroupを使用
	g, ctx := errgroup.WithContext(ctx)

	// システム情報を取得
	g.Go(func() error {
		select {
		case <-ctx.Done():
			data.SystemError = ctx.Err()
			return nil
		default:
			data.SystemInfo, data.SystemError = c.monitor.GetSystemInfo()
			return nil
		}
	})

	// コンテナ情報と統計情報を取得
	g.Go(func() error {
		select {
		case <-ctx.Done():
			data.ContainerError = ctx.Err()
			return nil
		default:
			// まずコンテナ情報を取得
			data.Containers, data.ContainerError = c.compose.ListContainers(c.composePath)

			// コンテナ情報が取得できた場合のみ統計情報を取得
			if data.ContainerError == nil {
				// コンテキストの確認
				select {
				case <-ctx.Done():
					return nil
				default:
					data.Stats, data.StatsError = c.compose.GetAllContainersStats(c.composePath)
					if data.StatsError != nil {
						logger := logging.FromContext(c.ctx)
						logger.Warn(c.ctx, "Failed to get container stats",
							logging.ErrorField(data.StatsError))
					}
				}
			}
			return nil
		}
	})

	// ゲームコンテナ情報を取得
	g.Go(func() error {
		select {
		case <-ctx.Done():
			data.GameError = ctx.Err()
			return nil
		default:
			data.GameContainers, data.GameError = c.compose.ListGameContainers(c.composePath)
			return nil
		}
	})

	// すべての並行処理が完了するまで待機
	_ = g.Wait() // エラーは個別に保存しているので無視

	return data, nil
}

// buildMonitorReport は監視レポートを生成する
func (c *MonitorCommand) buildMonitorReport(data *MonitorData) string {
	var builder strings.Builder
	builder.Grow(4096)
	builder.WriteString("🖥️ **システム監視ダッシュボード**\n\n")

	// システム情報
	systemInfo := c.buildSystemInfo(data.SystemInfo)
	builder.WriteString(systemInfo)

	// コンテナテーブル
	if data.ContainerError != nil {
		builder.WriteString("\n⚠️ **コンテナ情報の取得に失敗しました**\n")
		if docker.IsPermissionDenied(data.ContainerError) {
			builder.WriteString(usermsg.DockerPermissionMessage())
		} else {
			builder.WriteString(fmt.Sprintf("エラー: %v\n", data.ContainerError))
		}
	} else {
		statsMap := make(map[string]*docker.ContainerStats)
		for i := range data.Stats {
			statsMap[data.Stats[i].Name] = &data.Stats[i]
		}
		containerTable := c.buildContainerTable(data.Containers, statsMap)
		builder.WriteString(containerTable)
	}

	// アラート
	alerts := c.checkAlerts(data.SystemInfo, data.Stats)
	alertSection := c.buildAlertSection(alerts)
	builder.WriteString(alertSection)

	// ゲームサーバー情報
	if data.GameError != nil {
		builder.WriteString("\n⚠️ **ゲームサーバー情報の取得に失敗しました**\n")
		builder.WriteString(fmt.Sprintf("エラー: %v\n", data.GameError))
	} else {
		gameServerInfo := c.buildGameServerInfo(data.GameContainers)
		builder.WriteString(gameServerInfo)
	}

	return builder.String()
}

// GetGameIcon returns an icon based on the game service name
func GetGameIcon(service string) string {
	lowerService := strings.ToLower(service)
	switch lowerService {
	case gameServiceMinecraft:
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

// GetStatusIcon returns an icon based on the container state
func GetStatusIcon(state string) string {
	lowerState := strings.ToLower(state)
	switch lowerState {
	case containerStateRunning:
		return "🟢"
	case containerStateStopped, containerStateExited:
		return "🔴"
	case "restarting":
		return "🟡"
	case "paused":
		return "⏸️"
	default:
		return statusIconUnknown
	}
}

// FormatServiceName formats the service name for display
func FormatServiceName(service string) string {
	if service == "" {
		return ""
	}

	// Replace hyphens and underscores with spaces
	formatted := strings.ReplaceAll(service, "-", " ")
	formatted = strings.ReplaceAll(formatted, "_", " ")

	// Split into words and capitalize each
	words := strings.Fields(formatted)
	for i, word := range words {
		if word != "" {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

// GetHealthIcon returns an icon based on the health status
func GetHealthIcon(health string) string {
	lowerHealth := strings.ToLower(health)
	switch lowerHealth {
	case "healthy":
		return "✅"
	case "unhealthy":
		return "❌"
	case "starting":
		return "🔄"
	default:
		return statusIconUnknown
	}
}

// getServiceFromContainerName extracts service name from container name
func getServiceFromContainerName(containerName string) string {
	// Docker Composeのコンテナ名は通常 "project_service_1" の形式
	parts := strings.Split(containerName, "_")
	if len(parts) >= 2 && parts[1] != "" {
		return parts[1]
	}
	return containerName
}

// GetComponents returns Discord message components for the monitor command
func (c *MonitorCommand) GetComponents(_ []string) ([]discordgo.MessageComponent, error) {
	containers, err := c.compose.ListGameContainers(c.composePath)
	if err != nil {
		return nil, fmt.Errorf("コンテナ情報の取得に失敗しました: %w", err)
	}

	var components []discordgo.MessageComponent
	var buttons []discordgo.MessageComponent

	// ボタン数のカウント（最大値を超えないように）
	buttonCount := 0
	for i := range containers {
		// 最大ボタン数に達したら終了
		if buttonCount >= docker.MaxTotalButtons {
			break
		}
		// 停止中のコンテナに対しては起動ボタンを追加
		if strings.EqualFold(containers[i].State, containerStateStopped) ||
			strings.EqualFold(containers[i].State, containerStateExited) {
			button := discordgo.Button{
				Label:    fmt.Sprintf("🚀 %s を起動", FormatServiceName(containers[i].Service)),
				Style:    discordgo.SuccessButton,
				CustomID: fmt.Sprintf("start_service_%s", containers[i].Service),
			}
			buttons = append(buttons, button)
			buttonCount++
		} else if strings.EqualFold(containers[i].State, containerStateRunning) {
			// 稼働中のコンテナに対しては停止ボタンを追加
			button := discordgo.Button{
				Label:    fmt.Sprintf("🛑 %s を停止", FormatServiceName(containers[i].Service)),
				Style:    discordgo.DangerButton,
				CustomID: fmt.Sprintf("stop_service_%s", containers[i].Service),
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

// CanHandle は指定されたカスタムIDを処理できるかどうかを返す
func (c *MonitorCommand) CanHandle(customID string) bool {
	return strings.HasPrefix(customID, "start_service_") || strings.HasPrefix(customID, "stop_service_")
}

// HandleInteraction はサービスの起動/停止インタラクションを処理する
func (c *MonitorCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Type != discordgo.InteractionMessageComponent {
		return fmt.Errorf("unexpected interaction type: %v", i.Type)
	}

	data := i.MessageComponentData()

	// サービス名と操作を判定
	var serviceName string
	var isStart bool

	switch {
	case strings.HasPrefix(data.CustomID, "start_service_"):
		serviceName = strings.TrimPrefix(data.CustomID, "start_service_")
		isStart = true
	case strings.HasPrefix(data.CustomID, "stop_service_"):
		serviceName = strings.TrimPrefix(data.CustomID, "stop_service_")
		isStart = false
	default:
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
func (c *MonitorCommand) handleServiceOperation(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	serviceName string,
	isStart bool,
) {
	// 初期化と検証
	ctx, cancel, logger := c.setupServiceOperation()
	defer cancel()
	defer c.serviceOperations.Delete(serviceName)

	// パニックリカバリーを設定
	defer c.handlePanicRecovery(ctx, s, i, serviceName, isStart, logger)

	// 親コンテキストのキャンセルをチェック
	if c.isParentContextCanceled(serviceName, logger) {
		return
	}

	// サービス操作を実行
	result := c.executeServiceOperation(ctx, serviceName, isStart)

	// 結果を処理してメッセージを送信
	c.handleOperationResult(s, i, serviceName, isStart, result, logger)
}

// setupServiceOperation は初期化とコンテキストの設定を行う
func (c *MonitorCommand) setupServiceOperation() (context.Context, context.CancelFunc, logging.Logger) {
	logger := logging.FromContext(c.ctx)
	ctx, cancel := context.WithTimeout(c.ctx, ServiceOperationTimeout)
	return ctx, cancel, logger
}

// isParentContextCanceled は親コンテキストがキャンセルされているかチェックする
func (c *MonitorCommand) isParentContextCanceled(serviceName string, logger logging.Logger) bool {
	select {
	case <-c.ctx.Done():
		logger.Warn(c.ctx, "Parent context canceled, aborting operation",
			logging.String("service", serviceName))
		return true
	default:
		return false
	}
}

// handlePanicRecovery はパニックが発生した場合のリカバリー処理を行う
func (c *MonitorCommand) handlePanicRecovery(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	serviceName string,
	isStart bool,
	logger logging.Logger,
) {
	if r := recover(); r != nil {
		logger.Error(ctx, "Panic in handleServiceOperation",
			logging.String("panic", fmt.Sprintf("%v", r)),
			logging.String("service", serviceName),
			logging.Bool("start", isStart))

		// エラーメッセージを送信
		message := fmt.Sprintf("❌ 予期しないエラーが発生しました: %v", r)
		_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: message,
		})
	}
}

// ServiceOperationResult はサービス操作の実行結果を表す
type ServiceOperationResult struct {
	Err            error
	SuccessMessage string
	ErrorPrefix    string
}

// executeServiceOperation はサービス操作を実行し、結果を返す
func (c *MonitorCommand) executeServiceOperation(
	ctx context.Context,
	serviceName string,
	isStart bool,
) ServiceOperationResult {
	var result ServiceOperationResult

	// タイムアウトチャンネルとの競合を処理
	done := make(chan struct{})
	go func() {
		defer close(done)
		if isStart {
			result.Err = c.compose.StartService(c.composePath, serviceName)
			formattedName := FormatServiceName(serviceName)
			result.SuccessMessage = fmt.Sprintf("✅ %s を起動しました！", formattedName)
			result.ErrorPrefix = "起動"
		} else {
			result.Err = c.compose.StopService(c.composePath, serviceName)
			formattedName := FormatServiceName(serviceName)
			result.SuccessMessage = fmt.Sprintf("🛑 %s を停止しました。", formattedName)
			result.ErrorPrefix = "停止"
		}
	}()

	// タイムアウトまたは完了を待つ
	select {
	case <-done:
		// 正常完了
	case <-ctx.Done():
		// タイムアウト
		result.Err = fmt.Errorf("操作がタイムアウトしました (%v)", ServiceOperationTimeout)
		result.ErrorPrefix = "操作"
	}

	return result
}

// handleOperationResult は操作結果を処理してDiscordにメッセージを送信する
func (c *MonitorCommand) handleOperationResult(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	serviceName string,
	isStart bool,
	result ServiceOperationResult,
	logger logging.Logger,
) {
	// 結果に応じてメッセージを作成
	content := c.createResponseMessage(serviceName, result)

	// ログを出力
	c.logOperationResult(serviceName, isStart, result.Err, logger)

	// フォローアップメッセージを送信
	c.sendFollowupMessage(s, i, content, logger)
}

// createResponseMessage は操作結果に基づいてレスポンスメッセージを作成する
func (c *MonitorCommand) createResponseMessage(serviceName string, result ServiceOperationResult) string {
	if result.Err != nil {
		formattedName := FormatServiceName(serviceName)
		errorMsg := fmt.Sprintf("❌ %s の%sに失敗しました: %v", formattedName, result.ErrorPrefix, result.Err)
		return errorMsg
	}
	return result.SuccessMessage
}

// logOperationResult は操作結果をログに記録する
func (c *MonitorCommand) logOperationResult(serviceName string, isStart bool, err error, logger logging.Logger) {
	if err != nil {
		logger.Error(c.ctx, "Service operation failed",
			logging.String("service", serviceName),
			logging.Bool("start", isStart),
			logging.ErrorField(err))
	} else {
		logger.Info(c.ctx, "Service operation succeeded",
			logging.String("service", serviceName),
			logging.Bool("start", isStart))
	}
}

// sendFollowupMessage はDiscordにフォローアップメッセージを送信する
func (c *MonitorCommand) sendFollowupMessage(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	content string,
	logger logging.Logger,
) {
	_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
	if err != nil {
		logger.Error(c.ctx, "Failed to send followup message", logging.ErrorField(err))
	}
}
