package command

// Command はコマンドのインターフェース
type Command interface {
	// Name はコマンド名を返す
	Name() string
	// Description はコマンドの説明を返す
	Description() string
	// Execute はコマンドを実行する
	Execute(args []string) (string, error)
}