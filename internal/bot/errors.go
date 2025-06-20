package bot

import "errors"

var (
	// ErrServiceOperationInProgress はサービス操作が既に進行中の場合のエラー
	ErrServiceOperationInProgress = errors.New("service operation already in progress")
	// ErrGameInfoCommandNotFound はgame-infoコマンドが見つからない場合のエラー
	ErrGameInfoCommandNotFound = errors.New("game-info command not found")
	// ErrInvalidCommandType はコマンドの型が不正な場合のエラー
	ErrInvalidCommandType = errors.New("invalid command type")
)