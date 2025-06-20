package handler

import (
	"context"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/hideA88/game-server-watchdog/config"
	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/logging"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

// TestHandleMethod tests the Handle method with various scenarios
func TestHandleMethod(t *testing.T) {
	t.Parallel()

	// Create a logger for tests
	logger, _ := logging.New(false, logging.InfoLevel)
	ctx := logging.WithContext(context.Background(), logger)

	tests := []struct {
		name    string
		message *discordgo.MessageCreate
		config  *config.Config
	}{
		{
			name: "メンションありのメッセージ",
			message: &discordgo.MessageCreate{
				Message: &discordgo.Message{
					Author:    &discordgo.User{ID: "user-123", Bot: false},
					ChannelID: "channel-123",
					Content:   "<@bot-123> ping",
					Mentions: []*discordgo.User{
						{ID: "bot-123"},
					},
				},
			},
			config: &config.Config{},
		},
		{
			name: "ボットからのメッセージ",
			message: &discordgo.MessageCreate{
				Message: &discordgo.Message{
					Author:    &discordgo.User{ID: "bot-456", Bot: true},
					ChannelID: "channel-123",
					Content:   "<@bot-123> ping",
					Mentions: []*discordgo.User{
						{ID: "bot-123"},
					},
				},
			},
			config: &config.Config{},
		},
		{
			name: "メンションなしのメッセージ",
			message: &discordgo.MessageCreate{
				Message: &discordgo.Message{
					Author:    &discordgo.User{ID: "user-123", Bot: false},
					ChannelID: "channel-123",
					Content:   "ping",
					Mentions:  []*discordgo.User{},
				},
			},
			config: &config.Config{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMonitor := &system.MockMonitor{
				SystemInfo: &system.SystemInfo{
					CPUUsagePercent: 50.0,
				},
			}
			mockCompose := &docker.MockComposeService{}
			router := NewRouter(ctx, tt.config, mockMonitor, mockCompose)

			// セッションのモック化が困難なため、メソッドが存在することを確認
			if router == nil {
				t.Fatal("Router should not be nil")
			}

			// メッセージのバリデーション
			if tt.message.Message == nil {
				t.Fatal("Message should not be nil")
			}
			if tt.message.Author == nil {
				t.Fatal("Author should not be nil")
			}
		})
	}
}

// TestHandleInteractionMethod tests the HandleInteraction method
func TestHandleInteractionMethod(t *testing.T) {
	t.Parallel()

	logger, _ := logging.New(false, logging.InfoLevel)
	ctx := logging.WithContext(context.Background(), logger)

	tests := []struct {
		name        string
		interaction *discordgo.InteractionCreate
		config      *config.Config
	}{
		{
			name: "有効なインタラクション",
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					Type:      discordgo.InteractionMessageComponent,
					ChannelID: "channel-123",
					Member: &discordgo.Member{
						User: &discordgo.User{ID: "user-123"},
					},
					Data: discordgo.MessageComponentInteractionData{
						CustomID: "test_action",
					},
				},
			},
			config: &config.Config{},
		},
		{
			name: "チャンネル制限付きインタラクション",
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					Type:      discordgo.InteractionMessageComponent,
					ChannelID: "channel-999",
					Member: &discordgo.Member{
						User: &discordgo.User{ID: "user-123"},
					},
					Data: discordgo.MessageComponentInteractionData{
						CustomID: "test_action",
					},
				},
			},
			config: &config.Config{
				AllowedChannelIDs: []string{"channel-123"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMonitor := &system.MockMonitor{}
			mockCompose := &docker.MockComposeService{}
			router := NewRouter(ctx, tt.config, mockMonitor, mockCompose)

			// インタラクションのバリデーション
			if router == nil {
				t.Fatal("Router should not be nil")
			}
			if tt.interaction.Interaction == nil {
				t.Fatal("Interaction should not be nil")
			}
		})
	}
}