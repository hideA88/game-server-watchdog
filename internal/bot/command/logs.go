package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
)

// LogsCommand handles the logs command
type LogsCommand struct {
	compose     docker.ComposeService
	composePath string
}

// NewLogsCommand creates a new LogsCommand
func NewLogsCommand(compose docker.ComposeService, composePath string) *LogsCommand {
	if composePath == "" {
		composePath = "docker-compose.yml"
	}
	return &LogsCommand{
		compose:     compose,
		composePath: composePath,
	}
}

// Name returns the command name
func (c *LogsCommand) Name() string {
	return "logs"
}

// Description returns the command description
func (c *LogsCommand) Description() string {
	return "指定されたコンテナのログを表示"
}

// Execute runs the command
func (c *LogsCommand) Execute(args []string) (string, error) {
	if len(args) == 0 {
		return "使用方法: `@bot logs <サービス名> [行数]`\n例: `@bot logs minecraft 50`", nil
	}

	serviceName := args[0]
	lines := 50 // デフォルト行数

	// 行数が指定されている場合
	if len(args) > 1 {
		if l, err := strconv.Atoi(args[1]); err == nil {
			lines = l
		}
	}

	// 行数の制限
	if lines < 1 {
		lines = 1
	} else if lines > 200 {
		lines = 200
		// 制限を超えた場合の警告
		defer func() {
			// この警告は結果の最後に追加される
		}()
	}

	// コンテナの存在確認
	containers, err := c.compose.ListContainers(c.composePath)
	if err != nil {
		return "", fmt.Errorf("コンテナ情報の取得に失敗しました: %w", err)
	}

	found := false
	for _, container := range containers {
		if container.Service == serviceName {
			found = true
			break
		}
	}

	if !found {
		return fmt.Sprintf("❌ サービス '%s' が見つかりません", serviceName), nil
	}

	// ログを取得
	logs, err := c.compose.GetContainerLogs(c.composePath, serviceName, lines)
	if err != nil {
		return fmt.Sprintf("❌ %s のログ取得に失敗しました: %v", FormatServiceName(serviceName), err), nil
	}

	// 結果を構築
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📜 **%s のログ** (最後の%d行)\n", FormatServiceName(serviceName), lines))
	builder.WriteString("```\n")

	// ログが空の場合
	if strings.TrimSpace(logs) == "" {
		builder.WriteString("(ログがありません)\n")
	} else {
		// ログを行に分割
		logLines := strings.Split(strings.TrimSpace(logs), "\n")

		// Discord のメッセージ制限を考慮（約2000文字）
		totalLength := 0
		maxLength := 1800 // ヘッダーとフッターのための余裕を持たせる
		truncated := false

		for i, line := range logLines {
			// 各行を最大200文字に制限
			if len(line) > 200 {
				line = line[:197] + "..."
			}

			// 全体の長さをチェック
			if totalLength+len(line)+1 > maxLength {
				builder.WriteString(fmt.Sprintf("\n... (残り %d 行は省略されました)", len(logLines)-i))
				truncated = true
				break
			}

			builder.WriteString(line + "\n")
			totalLength += len(line) + 1
		}

		if !truncated && lines > 200 {
			builder.WriteString("\n(注意: 最大200行に制限されています)")
		}
	}

	builder.WriteString("```\n")

	// ヒント
	builder.WriteString("\n💡 **ヒント**: より多くのログを見るには、行数を指定してください\n")
	builder.WriteString(fmt.Sprintf("例: `@bot logs %s 100`", serviceName))

	return builder.String(), nil
}
