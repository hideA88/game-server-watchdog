package command

import (
	"fmt"

	"github.com/hideA88/game-server-watchdog/pkg/system"
)

// StatusCommand はstatusコマンドの実装
type StatusCommand struct {
	monitor system.Monitor
}

// NewStatusCommand は新しいStatusCommandを作成
func NewStatusCommand(monitor system.Monitor) *StatusCommand {
	return &StatusCommand{
		monitor: monitor,
	}
}

// Name はコマンド名を返す
func (c *StatusCommand) Name() string {
	return "status"
}

// Description はコマンドの説明を返す
func (c *StatusCommand) Description() string {
	return "サーバーのステータスを表示"
}

// Execute はコマンドを実行する
func (c *StatusCommand) Execute(_ []string) (string, error) {
	info, err := c.monitor.GetSystemInfo()
	if err != nil {
		return "", fmt.Errorf("システム情報の取得に失敗しました: %w", err)
	}

	message := fmt.Sprintf(
		"📊 **サーバーステータス**\n"+
			"• CPU使用率: %.1f%%\n"+
			"• メモリ使用量: %.1fGB / %.1fGB (%.1f%%)\n"+
			"• ディスク空き容量: %.1fGB / %.1fGB (%.1f%%)",
		info.CPUUsagePercent,
		info.MemoryUsedGB, info.MemoryTotalGB, info.MemoryUsedPercent,
		info.DiskFreeGB, info.DiskTotalGB, info.DiskUsedPercent,
	)

	return message, nil
}
