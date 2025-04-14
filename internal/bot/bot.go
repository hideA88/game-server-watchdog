package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/hideA88/game-server-watchdog/internal/bot/handler"
	"github.com/hideA88/game-server-watchdog/internal/config"
)

type Bot struct {
	session *discordgo.Session
	config  *config.Config
}

func New(config *config.Config) (*Bot, error) {
	session, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	bot := &Bot{
		session: session,
		config:  config,
	}

	// ハンドラーの登録
	session.AddHandler(handler.NewPingHandler().Handle)
	session.AddHandler(handler.NewHelpHandler().Handle)

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
	b.session.Close()
}
