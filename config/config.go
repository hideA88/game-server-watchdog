package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DiscordToken      string   `envconfig:"DISCORD_TOKEN" required:"true"`
	DebugMode         bool     `envconfig:"DEBUG_MODE" default:"false"`
	AllowedChannelIDs []string `envconfig:"ALLOWED_CHANNEL_IDS" separator:","`
	AllowedUserIDs    []string `envconfig:"ALLOWED_USER_IDS" separator:","`
	DockerComposePath string   `envconfig:"DOCKER_COMPOSE_PATH" default:"docker-compose.yml"`
}

func Load() (*Config, error) {
	// .envファイルが存在する場合のみ読み込む
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
