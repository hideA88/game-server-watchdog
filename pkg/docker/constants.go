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
	// ListOperationTimeout はリスト操作のタイムアウト時間
	ListOperationTimeout = 10 * time.Second
	// QueryOperationTimeout はクエリ操作のタイムアウト時間
	QueryOperationTimeout = 5 * time.Second

	// LabelDockerComposeProject はDocker Composeプロジェクト名のラベル
	LabelDockerComposeProject = "com.docker.compose.project"
	// LabelDockerComposeService はDocker Composeサービス名のラベル
	LabelDockerComposeService = "com.docker.compose.service"
	// LabelGameType はゲームコンテナを識別するためのラベル
	LabelGameType = "game.type"

	// Container states
	// containerStateRunning は実行中のコンテナの状態
	containerStateRunning = "running"
)
