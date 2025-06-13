package command

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
func (c *PingCommand) Execute(_ []string) (string, error) {
	return "pong!!", nil
}
