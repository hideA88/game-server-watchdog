package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hideA88/game-server-watchdog/config"
	"github.com/hideA88/game-server-watchdog/internal/bot"
	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

// ビルド時に埋め込まれるバージョン情報
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	// バージョン情報の表示（--versionオプション）
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("Game Server Watchdog %s\n", version)
		fmt.Printf("  Commit: %s\n", commit)
		fmt.Printf("  Built:  %s\n", date)
		fmt.Printf("  Built by: %s\n", builtBy)
		os.Exit(0)
	}

	log.Printf("Starting Game Server Watchdog %s (commit: %s)", version, commit)

	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// 依存性の初期化
	monitor := system.NewDockerAwareMonitor()

	// Docker Compose サービスを作成
	compose, err := docker.NewDefaultComposeService()
	if err != nil {
		log.Fatalf("Error creating compose service: %v", err)
	}

	// プロジェクト名が設定されている場合は設定
	if cfg.DockerComposeProjectName != "" {
		log.Printf("Setting Docker Compose project name: %s", cfg.DockerComposeProjectName)
		compose.SetProjectName(cfg.DockerComposeProjectName)
	} else {
		log.Printf("No Docker Compose project name configured")
	}

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

	// Docker APIクライアントのクリーンアップ
	defer compose.Close()

	// シグナル待ち
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
