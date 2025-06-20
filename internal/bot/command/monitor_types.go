package command

import (
	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

// MonitorData は監視データを集約する構造体
type MonitorData struct {
	SystemInfo     *system.SystemInfo
	SystemError    error
	Containers     []docker.ContainerInfo
	ContainerError error
	Stats          []docker.ContainerStats
	StatsError     error
	GameContainers []docker.ContainerInfo
	GameError      error
}

// Alert はアラート情報を表す構造体
type Alert struct {
	Component string
	Message   string
	Value     float64
}

// ProgressBar はプログレスバーを表す構造体
type ProgressBar struct {
	Value  float64
	Width  int
	Filled rune
	Empty  rune
}

// NewProgressBar は新しいProgressBarを作成する
func NewProgressBar(value float64, width int) ProgressBar {
	return ProgressBar{
		Value:  value,
		Width:  width,
		Filled: '█',
		Empty:  '░',
	}
}

// String はプログレスバーを文字列として描画する
func (p ProgressBar) String() string {
	if p.Value < 0 {
		p.Value = 0
	}
	if p.Value > 100 {
		p.Value = 100
	}

	filled := int(p.Value * float64(p.Width) / 100)
	empty := p.Width - filled

	bar := ""
	for i := 0; i < filled; i++ {
		bar += string(p.Filled)
	}
	for i := 0; i < empty; i++ {
		bar += string(p.Empty)
	}
	return bar
}
