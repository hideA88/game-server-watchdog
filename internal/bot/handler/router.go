package handler

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/hideA88/game-server-watchdog/internal/bot/command"
	"github.com/hideA88/game-server-watchdog/internal/config"
)


// Router はメッセージをルーティングして適切なコマンドに振り分ける
type Router struct {
	config   *config.Config
	commands map[string]command.Command
}

// NewRouter は新しいルーターを作成し、コマンドを登録
func NewRouter(cfg *config.Config) *Router {
	r := &Router{
		config:   cfg,
		commands: make(map[string]command.Command),
	}

	// コマンドを初期化して登録
	pingCmd := command.NewPingCommand()
	helpCmd := command.NewHelpCommand()
	
	r.RegisterCommand(pingCmd)
	r.RegisterCommand(helpCmd)
	
	// helpコマンドに利用可能なコマンドを設定
	helpCmd.SetCommands(r.commands)

	return r
}

// RegisterCommand はコマンドを登録
func (r *Router) RegisterCommand(cmd command.Command) {
	r.commands[cmd.Name()] = cmd
}

// Handle はDiscordのメッセージイベントを処理
func (r *Router) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ボット自身のメッセージは無視
	if m.Author.ID == s.State.User.ID {
		return
	}

	// ボットがメンションされているかチェック
	botMentioned := false
	for _, user := range m.Mentions {
		if user.ID == s.State.User.ID {
			botMentioned = true
			break
		}
	}

	// メンションされていない場合は無視
	if !botMentioned {
		return
	}

	// アクセス権限チェック
	if !IsAuthorized(r.config, m.ChannelID, m.Author.ID) {
		log.Printf("Unauthorized access attempt from user %s in channel %s", m.Author.ID, m.ChannelID)
		return
	}

	// メンション部分を削除してコマンドを取得
	content := strings.TrimSpace(m.Content)
	for _, user := range m.Mentions {
		mention := "<@" + user.ID + ">"
		content = strings.ReplaceAll(content, mention, "")
		mention = "<@!" + user.ID + ">"
		content = strings.ReplaceAll(content, mention, "")
	}
	content = strings.TrimSpace(content)

	// コマンドと引数を分割
	parts := strings.Fields(content)
	if len(parts) == 0 {
		return
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	// コマンドを探して実行
	if cmd, exists := r.commands[command]; exists {
		if err := cmd.Execute(s, m, args); err != nil {
			log.Printf("コマンド実行エラー: %v", err)
			_, _ = s.ChannelMessageSend(m.ChannelID, "コマンドの実行中にエラーが発生しました。")
		}
	} else {
		// 未知のコマンドの場合
		_, err := s.ChannelMessageSend(m.ChannelID, "不明なコマンドです。`@ボット help`でコマンド一覧を確認してください。")
		if err != nil {
			log.Printf("メッセージの送信に失敗しました: %v", err)
		}
	}
}