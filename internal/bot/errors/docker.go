// Package errors provides error handling utilities for the bot
package errors

import (
	"strings"
)

// IsDockerPermissionError checks if the error is a Docker permission error
func IsDockerPermissionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "permission denied") &&
		strings.Contains(errStr, "docker.sock")
}

// GetDockerPermissionErrorMessage returns a user-friendly message for Docker permission errors
func GetDockerPermissionErrorMessage() string {
	return `Docker権限エラーが発生しました。以下の手順で解決してください：

1. 現在のセッションを更新:
   ` + "`newgrp docker`" + `
   
2. または再ログイン:
   ` + "`exit`" + ` して再度ログイン

3. それでも解決しない場合:
   - dockerグループに属しているか確認: ` + "`groups`" + `
   - Docker socketの権限を確認: ` + "`ls -la /var/run/docker.sock`" + `

詳細は管理者にお問い合わせください。`
}
