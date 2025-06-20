package command

import "github.com/bwmarrin/discordgo"

// Command はコマンドのインターフェース
type Command interface {
	// Name はコマンド名を返す
	Name() string
	// Description はコマンドの説明を返す
	Description() string
	// Execute はコマンドを実行する
	Execute(args []string) (string, error)
}

// InteractiveCommand はインタラクティブなコンポーネントを返すコマンドのインターフェース
type InteractiveCommand interface {
	Command
	// GetComponents はコマンド結果に対するDiscordメッセージコンポーネントを返す
	GetComponents(args []string) ([]discordgo.MessageComponent, error)
}

// InteractionHandler はDiscordのインタラクションを処理するインターフェース
type InteractionHandler interface {
	// HandleInteraction はインタラクションを処理する
	HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error
	// CanHandle は指定されたカスタムIDを処理できるかどうかを返す
	CanHandle(customID string) bool
}
