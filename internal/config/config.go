package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken string
	DebugMode    bool
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	debugMode, _ := strconv.ParseBool(os.Getenv("DEBUG_MODE"))

	return &Config{
		DiscordToken: os.Getenv("DISCORD_TOKEN"),
		DebugMode:    debugMode,
	}, nil
}
