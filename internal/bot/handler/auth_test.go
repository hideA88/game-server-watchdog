package handler

import (
	"testing"

	"github.com/hideA88/game-server-watchdog/internal/config"
)

func TestIsAuthorized(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		config         *config.Config
		channelID      string
		userID         string
		expectedResult bool
	}{
		{
			name: "制限なし - 全員許可",
			config: &config.Config{
				AllowedChannelIDs: []string{},
				AllowedUserIDs:    []string{},
			},
			channelID:      "channel123",
			userID:         "user123",
			expectedResult: true,
		},
		{
			name: "チャンネル制限 - 許可されたチャンネル",
			config: &config.Config{
				AllowedChannelIDs: []string{"channel123", "channel456"},
				AllowedUserIDs:    []string{},
			},
			channelID:      "channel123",
			userID:         "user123",
			expectedResult: true,
		},
		{
			name: "チャンネル制限 - 許可されていないチャンネル",
			config: &config.Config{
				AllowedChannelIDs: []string{"channel123", "channel456"},
				AllowedUserIDs:    []string{},
			},
			channelID:      "channel789",
			userID:         "user123",
			expectedResult: false,
		},
		{
			name: "ユーザー制限 - 許可されたユーザー",
			config: &config.Config{
				AllowedChannelIDs: []string{},
				AllowedUserIDs:    []string{"user123", "user456"},
			},
			channelID:      "channel123",
			userID:         "user123",
			expectedResult: true,
		},
		{
			name: "ユーザー制限 - 許可されていないユーザー",
			config: &config.Config{
				AllowedChannelIDs: []string{},
				AllowedUserIDs:    []string{"user123", "user456"},
			},
			channelID:      "channel123",
			userID:         "user789",
			expectedResult: false,
		},
		{
			name: "チャンネルとユーザー両方制限 - 両方許可",
			config: &config.Config{
				AllowedChannelIDs: []string{"channel123"},
				AllowedUserIDs:    []string{"user123"},
			},
			channelID:      "channel123",
			userID:         "user123",
			expectedResult: true,
		},
		{
			name: "チャンネルとユーザー両方制限 - チャンネルのみ許可",
			config: &config.Config{
				AllowedChannelIDs: []string{"channel123"},
				AllowedUserIDs:    []string{"user123"},
			},
			channelID:      "channel123",
			userID:         "user456",
			expectedResult: false,
		},
		{
			name: "チャンネルとユーザー両方制限 - ユーザーのみ許可",
			config: &config.Config{
				AllowedChannelIDs: []string{"channel123"},
				AllowedUserIDs:    []string{"user123"},
			},
			channelID:      "channel456",
			userID:         "user123",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAuthorized(tt.config, tt.channelID, tt.userID)
			if result != tt.expectedResult {
				t.Errorf("IsAuthorized() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}
