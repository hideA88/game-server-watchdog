package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/hideA88/game-server-watchdog/config"
	"github.com/hideA88/game-server-watchdog/internal/bot/handler"
	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

type Bot struct {
	session *discordgo.Session
	config  *config.Config
}

func New(config *config.Config, monitor system.Monitor, compose docker.ComposeService) (*Bot, error) {
	session, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	bot := &Bot{
		session: session,
		config:  config,
	}

	// ルーターを初期化して登録
	router := handler.NewRouter(config, monitor, compose)
	session.AddHandler(router.Handle)
	session.AddHandler(router.HandleInteraction)

	return bot, nil
}

func (b *Bot) Start() error {
	// セッションを開く
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")
	return nil
}

func (b *Bot) Stop() {
	// Discordセッションを閉じる
	b.session.Close()
}
