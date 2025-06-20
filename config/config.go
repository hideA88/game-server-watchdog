// Package config はアプリケーションの設定を管理します
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/hideA88/game-server-watchdog/pkg/logging"
)

type Config struct {
	DiscordToken             string        `envconfig:"DISCORD_TOKEN" required:"true"`
	DebugMode                bool          `envconfig:"DEBUG_MODE" default:"false"`
	LogLevel                 logging.Level `envconfig:"-"` // 環境変数から直接読み込まない
	LogLevelStr              string        `envconfig:"LOG_LEVEL" default:""`
	AllowedChannelIDs        []string      `envconfig:"ALLOWED_CHANNEL_IDS" separator:","`
	AllowedUserIDs           []string      `envconfig:"ALLOWED_USER_IDS" separator:","`
	DockerComposePath        string        `envconfig:"DOCKER_COMPOSE_PATH" default:"docker-compose.yml"`
	DockerComposeProjectName string        `envconfig:"DOCKER_COMPOSE_PROJECT_NAME" default:""`
}

// Load は環境変数から設定を読み込みます
func Load() (*Config, error) {
	// .envファイルが存在する場合のみ読み込む
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	// ログレベルの変換と検証（大文字小文字を区別しない）
	if cfg.LogLevelStr != "" {
		// loggingパッケージの大文字小文字を区別しない関数を使用
		level, err := logging.ParseLevelCaseInsensitiveWithWarning(cfg.LogLevelStr)
		if err != nil {
			// 無効な値の場合は警告を出力
			fmt.Fprintf(os.Stderr, "Invalid LOG_LEVEL: %s, using info level\n", cfg.LogLevelStr)
		}
		cfg.LogLevel = level
	} else {
		// 未指定の場合はInfoレベル
		cfg.LogLevel = logging.InfoLevel
	}

	// 設定の検証
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate は設定値の妥当性を検証します
func (c *Config) Validate() error {
	var errs []error

	// Discord tokenの検証
	if err := validateDiscordToken(c.DiscordToken); err != nil {
		errs = append(errs, fmt.Errorf("invalid DISCORD_TOKEN: %w", err))
	}

	// DockerCompose pathの検証（デフォルト値以外の場合のみ）
	if c.DockerComposePath != "" && c.DockerComposePath != "docker-compose.yml" {
		if _, err := os.Stat(c.DockerComposePath); os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("DOCKER_COMPOSE_PATH file not found: %s", c.DockerComposePath))
		}
	}

	// チャンネルIDの検証
	for _, channelID := range c.AllowedChannelIDs {
		if channelID != "" && !isValidDiscordID(channelID) {
			errs = append(errs, fmt.Errorf("invalid channel ID: %s", channelID))
		}
	}

	// ユーザーIDの検証
	for _, userID := range c.AllowedUserIDs {
		if userID != "" && !isValidDiscordID(userID) {
			errs = append(errs, fmt.Errorf("invalid user ID: %s", userID))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// validateDiscordToken はDiscord tokenの基本的な形式を検証します
func validateDiscordToken(token string) error {
	if len(token) < 50 {
		return errors.New("token too short (minimum 50 characters)")
	}

	// Discord botトークンの基本的な形式をチェック
	// 通常、Discord botトークンは3つの部分がピリオドで区切られている
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errors.New("invalid token format (expected 3 parts separated by dots)")
	}

	return nil
}

// isValidDiscordID はDiscord IDの形式を検証します
func isValidDiscordID(id string) bool {
	// Discord IDは17-19桁の数字
	if len(id) < 17 || len(id) > 19 {
		return false
	}

	// 数値であることを確認
	_, err := strconv.ParseUint(id, 10, 64)
	return err == nil
}
