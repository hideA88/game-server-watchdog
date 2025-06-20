package command

import (
	"fmt"
)

// HelpCommand はhelpコマンドの実装
type HelpCommand struct {
	commands []Command
}

// NewHelpCommand は新しいHelpCommandを作成
func NewHelpCommand() *HelpCommand {
	return &HelpCommand{
		commands: []Command{},
	}
}

// Name はコマンド名を返す
func (c *HelpCommand) Name() string {
	return "help"
}

const helpDescription = "コマンド一覧を表示"

// Description はコマンドの説明を返す
func (c *HelpCommand) Description() string {
	return helpDescription
}

// SetCommands は利用可能なコマンドのリストを設定
func (c *HelpCommand) SetCommands(commands []Command) {
	c.commands = commands
}

// Execute はコマンドを実行する
func (c *HelpCommand) Execute(_ []string) (string, error) {
	helpMessage := "**利用可能なコマンド:**\n"

	for _, cmd := range c.commands {
		helpMessage += fmt.Sprintf("`@ボット %s` - %s\n", cmd.Name(), cmd.Description())
	}

	helpMessage += "\n**使い方:**\nボットをメンションしてコマンドを送信してください。"

	return helpMessage, nil
}
