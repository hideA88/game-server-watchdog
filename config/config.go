package config

import (
	"fmt"
	"os"

	"github.com/hideA88/game-server-watchdog/pkg/logging"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
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

	return &cfg, nil
}
