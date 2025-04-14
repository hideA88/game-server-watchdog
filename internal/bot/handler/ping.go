package handler

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type PingHandler struct{}

func NewPingHandler() *PingHandler {
	return &PingHandler{}
}

func (h *PingHandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ボット自身のメッセージは無視
	if m.Author.ID == s.State.User.ID {
		return
	}

	// コマンドの処理
	if m.Content == "!ping" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
		if err != nil {
			log.Printf("メッセージの送信に失敗しました: %v", err)
		}
	}
}
