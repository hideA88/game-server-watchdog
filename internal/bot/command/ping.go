package command

import (
	"github.com/bwmarrin/discordgo"
)

// PingCommand はpingコマンドの実装
type PingCommand struct{}

// NewPingCommand は新しいPingCommandを作成
func NewPingCommand() *PingCommand {
	return &PingCommand{}
}

// Name はコマンド名を返す
func (c *PingCommand) Name() string {
	return "ping"
}

// Description はコマンドの説明を返す
func (c *PingCommand) Description() string {
	return "ボットの応答を確認"
}

// Execute はコマンドを実行する
func (c *PingCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	_, err := s.ChannelMessageSend(m.ChannelID, "pong!!")
	return err
}