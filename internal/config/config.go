package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken      string
	DebugMode         bool
	AllowedChannelIDs []string
	AllowedUserIDs    []string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	debugMode, _ := strconv.ParseBool(os.Getenv("DEBUG_MODE"))

	// 許可されたチャンネルIDとユーザーIDを環境変数から読み込み
	allowedChannels := os.Getenv("ALLOWED_CHANNEL_IDS")
	allowedUsers := os.Getenv("ALLOWED_USER_IDS")

	var allowedChannelIDs []string
	var allowedUserIDs []string

	if allowedChannels != "" {
		allowedChannelIDs = splitIDs(allowedChannels)
	}
	if allowedUsers != "" {
		allowedUserIDs = splitIDs(allowedUsers)
	}

	return &Config{
		DiscordToken:      os.Getenv("DISCORD_TOKEN"),
		DebugMode:         debugMode,
		AllowedChannelIDs: allowedChannelIDs,
		AllowedUserIDs:    allowedUserIDs,
	}, nil
}

// splitIDs はカンマ区切りのID文字列を分割してスライスに変換
func splitIDs(ids string) []string {
	parts := strings.Split(ids, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
