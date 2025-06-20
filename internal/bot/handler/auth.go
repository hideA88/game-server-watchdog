// Package handler はDiscordボットのメッセージハンドラーを提供します
package handler

import (
	"github.com/hideA88/game-server-watchdog/config"
)

// IsAuthorized はユーザーがボットにアクセスする権限があるかチェック
func IsAuthorized(cfg *config.Config, channelID, userID string) bool {
	// 制限が設定されていない場合は全員許可
	if len(cfg.AllowedChannelIDs) == 0 && len(cfg.AllowedUserIDs) == 0 {
		return true
	}

	// チャンネル制限チェック
	if len(cfg.AllowedChannelIDs) > 0 {
		channelAuthorized := false
		for _, allowedChannel := range cfg.AllowedChannelIDs {
			if channelID == allowedChannel {
				channelAuthorized = true
				break
			}
		}
		if !channelAuthorized {
			return false
		}
	}

	// ユーザー制限チェック
	if len(cfg.AllowedUserIDs) > 0 {
		for _, allowedUser := range cfg.AllowedUserIDs {
			if userID == allowedUser {
				return true
			}
		}
		return false
	}

	return true
}
