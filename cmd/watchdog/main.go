// Package main は Game Server Watchdog のエントリポイントです
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hideA88/game-server-watchdog/config"
	"github.com/hideA88/game-server-watchdog/internal/bot"
	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/logging"
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

	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// ロガーの初期化
	logger, err := logging.New(cfg.DebugMode, cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// contextにloggerを設定
	ctx := context.Background()
	ctx = logging.WithContext(ctx, logger)

	logger.Info(ctx, "Starting Game Server Watchdog",
		logging.String("version", version),
		logging.String("commit", commit),
		logging.Bool("debug_mode", cfg.DebugMode))

	// 依存性の初期化
	monitor := system.NewDockerAwareMonitor()

	// Docker Compose サービスを作成
	compose, err := docker.NewDefaultComposeService()
	if err != nil {
		logger.Error(ctx, "Error creating compose service", logging.ErrorField(err))
		os.Exit(1)
	}

	// プロジェクト名が設定されている場合は設定
	if cfg.DockerComposeProjectName != "" {
		logger.Info(ctx, "Setting Docker Compose project name",
			logging.String("project_name", cfg.DockerComposeProjectName))
		compose.SetProjectName(cfg.DockerComposeProjectName)
	} else {
		logger.Info(ctx, "No Docker Compose project name configured")
	}

	// ボットの初期化
	discordBot, err := bot.New(ctx, cfg, monitor, compose)
	if err != nil {
		logger.Error(ctx, "Error creating bot", logging.ErrorField(err))
		os.Exit(1)
	}

	// ボットの起動
	if err := discordBot.Start(ctx); err != nil {
		logger.Error(ctx, "Error starting bot", logging.ErrorField(err))
		os.Exit(1)
	}
	defer discordBot.Stop()

	// Docker APIクライアントのクリーンアップ
	defer compose.Close()

	// シグナル待ち
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
