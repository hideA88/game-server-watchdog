package command

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

// HelpCommand はhelpコマンドの実装
type HelpCommand struct {
	commands map[string]Command
}

// NewHelpCommand は新しいHelpCommandを作成
func NewHelpCommand() *HelpCommand {
	return &HelpCommand{
		commands: make(map[string]Command),
	}
}

// Name はコマンド名を返す
func (c *HelpCommand) Name() string {
	return "help"
}

// Description はコマンドの説明を返す
func (c *HelpCommand) Description() string {
	return "コマンド一覧を表示"
}

// SetCommands は利用可能なコマンドのリストを設定
func (c *HelpCommand) SetCommands(commands map[string]Command) {
	c.commands = commands
}

// Execute はコマンドを実行する
func (c *HelpCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	helpMessage := "**利用可能なコマンド:**\n"
	
	for name, cmd := range c.commands {
		helpMessage += fmt.Sprintf("`@ボット %s` - %s\n", name, cmd.Description())
	}
	
	helpMessage += "\n**使い方:**\nボットをメンションしてコマンドを送信してください。"

	_, err := s.ChannelMessageSend(m.ChannelID, helpMessage)
	return err
}