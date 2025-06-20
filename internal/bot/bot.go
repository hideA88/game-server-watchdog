// Package bot はDiscordボットの実装を提供します
package bot

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/hideA88/game-server-watchdog/config"
	"github.com/hideA88/game-server-watchdog/internal/bot/handler"
	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/logging"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

type Bot struct {
	session *discordgo.Session
	config  *config.Config
}

// New は新しいBotインスタンスを作成します
func New(ctx context.Context, config *config.Config, monitor system.Monitor, compose docker.ComposeService) (*Bot, error) {
	session, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	bot := &Bot{
		session: session,
		config:  config,
	}

	// ルーターを初期化して登録
	router := handler.NewRouter(ctx, config, monitor, compose)
	session.AddHandler(router.Handle)
	session.AddHandler(router.HandleInteraction)

	return bot, nil
}

func (b *Bot) Start(ctx context.Context) error {
	// セッションを開く
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}

	logger := logging.FromContext(ctx)
	logger.Info(ctx, "Bot is now running. Press CTRL-C to exit.")
	return nil
}

func (b *Bot) Stop() {
	// Discordセッションを閉じる
	_ = b.session.Close()
}
