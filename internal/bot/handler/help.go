package handler

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type HelpHandler struct{}

func NewHelpHandler() *HelpHandler {
	return &HelpHandler{}
}

func (h *HelpHandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ボット自身のメッセージは無視
	if m.Author.ID == s.State.User.ID {
		return
	}

	// コマンドの処理
	if m.Content == "!help" {
		helpMessage := "**利用可能なコマンド:**\n" +
			"`!ping` - ボットの応答を確認\n" +
			"`!help` - このヘルプメッセージを表示"
		_, err := s.ChannelMessageSend(m.ChannelID, helpMessage)
		if err != nil {
			log.Printf("メッセージの送信に失敗しました: %v", err)
		}
	}
}
