package docker

import "time"

const (
	// MaxButtonRows はDiscordの最大ボタン行数
	MaxButtonRows = 5
	// MaxButtonsPerRow は1行あたりの最大ボタン数
	MaxButtonsPerRow = 5
	// MaxTotalButtons は総ボタン数の最大値
	MaxTotalButtons = MaxButtonRows * MaxButtonsPerRow
	// ServiceOperationTimeout はDocker操作のタイムアウト時間
	ServiceOperationTimeout = 30 * time.Second
)