package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hideA88/game-server-watchdog/config"
	"github.com/hideA88/game-server-watchdog/internal/bot"
	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

func main() {
	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// 依存性の初期化
	monitor := system.NewDefaultMonitor()
	compose := docker.NewDefaultComposeService()

	// ボットの初期化
	discordBot, err := bot.New(cfg, monitor, compose)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	// ボットの起動
	if err := discordBot.Start(); err != nil {
		log.Fatalf("Error starting bot: %v", err)
	}
	defer discordBot.Stop()

	// シグナル待ち
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
