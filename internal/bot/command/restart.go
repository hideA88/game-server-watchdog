package command

import (
	"fmt"
	"sync"

	"github.com/hideA88/game-server-watchdog/pkg/docker"
)

// RestartCommand handles the restart command
type RestartCommand struct {
	compose           docker.ComposeService
	composePath       string
	serviceOperations *sync.Map // サービス名をキーとした操作ロック
}

// NewRestartCommand creates a new RestartCommand
func NewRestartCommand(compose docker.ComposeService, composePath string) *RestartCommand {
	if composePath == "" {
		composePath = "docker-compose.yml"
	}
	return &RestartCommand{
		compose:           compose,
		composePath:       composePath,
		serviceOperations: &sync.Map{},
	}
}

// Name returns the command name
func (c *RestartCommand) Name() string {
	return "restart"
}

// Description returns the command description
func (c *RestartCommand) Description() string {
	return "指定されたコンテナを再起動"
}

// Execute runs the command
func (c *RestartCommand) Execute(args []string) (string, error) {
	if len(args) == 0 {
		return "使用方法: `@bot restart <サービス名>`", nil
	}

	serviceName := args[0]

	// 操作ロックをチェック
	if _, loaded := c.serviceOperations.LoadOrStore(serviceName, true); loaded {
		return fmt.Sprintf("⚠️ %s は現在操作中です。しばらくお待ちください。", FormatServiceName(serviceName)), nil
	}
	defer c.serviceOperations.Delete(serviceName)

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

	// 再起動を実行
	err = c.compose.RestartContainer(c.composePath, serviceName)
	if err != nil {
		return fmt.Sprintf("❌ %s の再起動に失敗しました: %v", FormatServiceName(serviceName), err), nil
	}

	return fmt.Sprintf("🔄 %s を再起動しました！", FormatServiceName(serviceName)), nil
}
