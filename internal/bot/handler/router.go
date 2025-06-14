package handler

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/hideA88/game-server-watchdog/config"
	"github.com/hideA88/game-server-watchdog/internal/bot/command"
	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

type CommandHandler struct {
	Cmd         command.Command
	SendMsgFunc sendMessageFunc
}

type sendMessageFunc func(s *discordgo.Session, m *discordgo.MessageCreate, content string) (*discordgo.Message, error)

func sendSimpleMessage(s *discordgo.Session, m *discordgo.MessageCreate, content string) (*discordgo.Message, error) {
	return s.ChannelMessageSend(m.ChannelID, content)
}

// Router はメッセージをルーティングして適切なコマンドに振り分ける
type Router struct {
	config   *config.Config
	commands map[string]*CommandHandler
}

// NewRouter は新しいルーターを作成し、コマンドを登録
func NewRouter(cfg *config.Config, monitor system.Monitor, compose docker.ComposeService) *Router {
	r := &Router{
		config:   cfg,
		commands: make(map[string]*CommandHandler),
	}

	// コマンドを初期化して登録
	pingCmd := command.NewPingCommand()
	helpCmd := command.NewHelpCommand()
	statusCmd := command.NewStatusCommand(monitor)
	gameInfoCmd := command.NewGameInfoCommand(compose, cfg.DockerComposePath)

	r.RegisterCommand(pingCmd, sendSimpleMessage)
	r.RegisterCommand(helpCmd, sendSimpleMessage)
	r.RegisterCommand(statusCmd, sendSimpleMessage)
	r.RegisterCommand(gameInfoCmd, sendSimpleMessage)

	// helpコマンドに利用可能なコマンドを設定
	commands := []command.Command{pingCmd, helpCmd, statusCmd, gameInfoCmd}
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

// ParseCommand はメッセージからメンションを削除してコマンドと引数を抽出
func ParseCommand(content string, mentions []string) (string, []string) {
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

	command := strings.ToLower(parts[0])
	args := parts[1:]
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
		log.Printf("Unauthorized access attempt from user %s in channel %s", m.Author.ID, m.ChannelID)
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
		log.Printf("コマンド実行エラー: %v", err)
		_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	// 結果を送信
	if handler, exists := r.commands[command]; exists {
		if _, err := handler.SendMsgFunc(s, m, result); err != nil {
			log.Printf("メッセージの送信に失敗しました: %v", err)
			_, _ = s.ChannelMessageSend(m.ChannelID, "メッセージの送信中にエラーが発生しました。")
		}
	}
}
