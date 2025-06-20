package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/hideA88/game-server-watchdog/config"
	"github.com/hideA88/game-server-watchdog/internal/bot/command"
	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/logging"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

type CommandHandler struct {
	Cmd         command.Command
	SendMsgFunc sendMessageFunc
}

type sendMessageFunc func(
	s *discordgo.Session,
	m *discordgo.MessageCreate,
	content string,
	components []discordgo.MessageComponent,
) (*discordgo.Message, error)

func sendMessage(
	s *discordgo.Session,
	m *discordgo.MessageCreate,
	content string,
	components []discordgo.MessageComponent,
) (*discordgo.Message, error) {
	if len(components) > 0 {
		return s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content:    content,
			Components: components,
		})
	}
	return s.ChannelMessageSend(m.ChannelID, content)
}

// Router はメッセージをルーティングして適切なコマンドに振り分ける
type Router struct {
	ctx                 context.Context
	config              *config.Config
	commands            map[string]*CommandHandler
	interactionHandlers []command.InteractionHandler
}

// NewRouter は新しいルーターを作成し、コマンドを登録
func NewRouter(ctx context.Context, cfg *config.Config, monitor system.Monitor, compose docker.ComposeService) *Router {
	r := &Router{
		ctx:                 ctx,
		config:              cfg,
		commands:            make(map[string]*CommandHandler),
		interactionHandlers: []command.InteractionHandler{},
	}

	// コマンドを初期化して登録
	pingCmd := command.NewPingCommand()
	helpCmd := command.NewHelpCommand()
	statusCmd := command.NewStatusCommand(monitor)
	monitorCmd := command.NewMonitorCommand(ctx, compose, monitor, cfg.DockerComposePath)
	containerCmd := command.NewContainerCommand(compose, cfg.DockerComposePath)
	restartCmd := command.NewRestartCommand(compose, cfg.DockerComposePath)
	logsCmd := command.NewLogsCommand(compose, cfg.DockerComposePath)

	r.RegisterCommand(pingCmd, sendMessage)
	r.RegisterCommand(helpCmd, sendMessage)
	r.RegisterCommand(statusCmd, sendMessage)
	r.RegisterCommand(monitorCmd, sendMessage)
	r.RegisterCommand(containerCmd, sendMessage)
	r.RegisterCommand(restartCmd, sendMessage)
	r.RegisterCommand(logsCmd, sendMessage)

	// インタラクションハンドラーを登録
	r.RegisterInteractionHandler(monitorCmd)

	// helpコマンドに利用可能なコマンドを設定
	commands := []command.Command{pingCmd, helpCmd, statusCmd, monitorCmd, containerCmd, restartCmd, logsCmd}
	helpCmd.SetCommands(commands)

	return r
}

// RegisterCommand はコマンドを登録
func (r *Router) RegisterCommand(cmd command.Command, sendMsgFunc sendMessageFunc) {
	r.commands[cmd.Name()] = &CommandHandler{
		Cmd:         cmd,
		SendMsgFunc: sendMsgFunc,
	}
}

// RegisterInteractionHandler はインタラクションハンドラーを登録
func (r *Router) RegisterInteractionHandler(handler command.InteractionHandler) {
	r.interactionHandlers = append(r.interactionHandlers, handler)
}

// ParseCommand はメッセージからメンションを削除してコマンドと引数を抽出
func ParseCommand(content string, mentions []string) (command string, args []string) {
	// メンション部分を削除
	cleanContent := strings.TrimSpace(content)
	for _, userID := range mentions {
		mention := "<@" + userID + ">"
		cleanContent = strings.ReplaceAll(cleanContent, mention, "")
		mention = "<@!" + userID + ">"
		cleanContent = strings.ReplaceAll(cleanContent, mention, "")
	}
	cleanContent = strings.TrimSpace(cleanContent)

	// コマンドと引数を分割
	parts := strings.Fields(cleanContent)
	if len(parts) == 0 {
		return "", nil
	}

	command = strings.ToLower(parts[0])
	args = parts[1:]
	return command, args
}

// CheckMention はメッセージ内に特定のユーザーがメンションされているかチェック
func CheckMention(mentions []string, targetUserID string) bool {
	for _, userID := range mentions {
		if userID == targetUserID {
			return true
		}
	}
	return false
}

// ExecuteCommand はコマンドを実行して結果を返す
func (r *Router) ExecuteCommand(commandName string, args []string) (string, error) {
	handler, exists := r.commands[commandName]
	if !exists {
		return "", fmt.Errorf("不明なコマンドです。`@ボット help`でコマンド一覧を確認してください。")
	}
	return handler.Cmd.Execute(args)
}

// Handle はDiscordのメッセージイベントを処理
func (r *Router) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	logger := logging.FromContext(r.ctx)

	// ボット自身のメッセージは無視
	if m.Author.ID == s.State.User.ID {
		return
	}

	// メンションされたユーザーのIDリストを作成
	mentionIDs := make([]string, len(m.Mentions))
	for i, user := range m.Mentions {
		mentionIDs[i] = user.ID
	}

	// ボットがメンションされているかチェック
	if !CheckMention(mentionIDs, s.State.User.ID) {
		return
	}

	// アクセス権限チェック
	if !IsAuthorized(r.config, m.ChannelID, m.Author.ID) {
		logger.Warn(r.ctx, "Unauthorized access attempt",
			logging.String("user_id", m.Author.ID),
			logging.String("channel_id", m.ChannelID))
		return
	}

	// コマンドをパース
	command, args := ParseCommand(m.Content, mentionIDs)
	if command == "" {
		return
	}

	// コマンドを実行
	result, err := r.ExecuteCommand(command, args)
	if err != nil {
		logger.Error(r.ctx, "コマンド実行エラー", logging.ErrorField(err))
		_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	// 結果を送信
	if handler, exists := r.commands[command]; exists {
		// インタラクティブコマンドの場合はコンポーネントも送信
		var components []discordgo.MessageComponent
		if interactiveCmd, ok := handler.Cmd.(interface {
			GetComponents(args []string) ([]discordgo.MessageComponent, error)
		}); ok {
			if comps, err := interactiveCmd.GetComponents(args); err == nil {
				components = comps
			}
		}

		if _, err := handler.SendMsgFunc(s, m, result, components); err != nil {
			logger.Error(r.ctx, "メッセージの送信に失敗しました", logging.ErrorField(err))
			_, _ = s.ChannelMessageSend(m.ChannelID, "メッセージの送信中にエラーが発生しました。")
		}
	}
}

// HandleInteraction はDiscordのインタラクションイベントを処理
func (r *Router) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	logger := logging.FromContext(r.ctx)

	// メッセージコンポーネントのインタラクションのみ処理
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	// アクセス権限チェック
	if !IsAuthorized(r.config, i.ChannelID, i.Member.User.ID) {
		logger.Warn(r.ctx, "Unauthorized interaction",
			logging.String("user_id", i.Member.User.ID),
			logging.String("channel_id", i.ChannelID))
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "このアクションを実行する権限がありません。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.Error(r.ctx, "Failed to respond to unauthorized interaction", logging.ErrorField(err))
		}
		return
	}

	data := i.MessageComponentData()

	// 登録されたハンドラーから適切なものを探す
	for _, handler := range r.interactionHandlers {
		if handler.CanHandle(data.CustomID) {
			if err := handler.HandleInteraction(s, i); err != nil {
				logger.Error(r.ctx, "Failed to handle interaction", logging.ErrorField(err))
				// エラー応答を試みる
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "処理中にエラーが発生しました。",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
			}
			return
		}
	}

	// 未知のインタラクション
	logger.Warn(r.ctx, "Unknown interaction custom ID",
		logging.String("custom_id", data.CustomID))
}
