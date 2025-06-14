package handler

import (
	"reflect"
	"testing"

	"github.com/hideA88/game-server-watchdog/config"
	"github.com/hideA88/game-server-watchdog/internal/bot/command"
	"github.com/hideA88/game-server-watchdog/pkg/docker"
	"github.com/hideA88/game-server-watchdog/pkg/system"
)

func TestNewRouter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		config           *config.Config
		wantCommands     []string
		wantCommandCount int
	}{
		{
			name: "デフォルトコマンドが登録される",
			config: &config.Config{
				AllowedChannelIDs: []string{},
				AllowedUserIDs:    []string{},
			},
			wantCommands:     []string{"ping", "help", "status", "game-info"},
			wantCommandCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMonitor := &system.MockMonitor{}
			mockCompose := &docker.MockComposeService{}
			router := NewRouter(tt.config, mockMonitor, mockCompose)

			// ルーターが正しく初期化されているか確認
			if router == nil {
				t.Fatal("NewRouter() returned nil")
			}

			if router.config != tt.config {
				t.Error("Router config not set correctly")
			}

			// コマンド数の確認
			if len(router.commands) != tt.wantCommandCount {
				t.Errorf("Expected %d commands, got %d", tt.wantCommandCount, len(router.commands))
			}

			// 各コマンドが登録されているか確認
			for _, cmd := range tt.wantCommands {
				if _, exists := router.commands[cmd]; !exists {
					t.Errorf("%s command not registered", cmd)
				}
			}
		})
	}
}

func TestRouter_RegisterCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		command     command.Command
		wantCmdName string
	}{
		{
			name:        "pingコマンドを登録",
			command:     command.NewPingCommand(),
			wantCmdName: "ping",
		},
		{
			name:        "helpコマンドを登録",
			command:     command.NewHelpCommand(),
			wantCmdName: "help",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			router := &Router{
				config:   cfg,
				commands: make(map[string]*CommandHandler),
			}

			// コマンドを登録
			router.RegisterCommand(tt.command, sendSimpleMessage)

			// コマンドが正しく登録されているか確認
			if registeredCmd, exists := router.commands[tt.wantCmdName]; !exists {
				t.Error("Command not registered")
			} else if registeredCmd.Cmd != tt.command {
				t.Error("Wrong command registered")
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		content     string
		mentions    []string
		wantCommand string
		wantArgs    []string
	}{
		{
			name:        "メンション付きpingコマンド",
			content:     "<@123456> ping",
			mentions:    []string{"123456"},
			wantCommand: "ping",
			wantArgs:    []string{},
		},
		{
			name:        "メンション付きhelpコマンド",
			content:     "<@123456> help",
			mentions:    []string{"123456"},
			wantCommand: "help",
			wantArgs:    []string{},
		},
		{
			name:        "複数メンション付きコマンド",
			content:     "<@123456> <@789012> ping test",
			mentions:    []string{"123456", "789012"},
			wantCommand: "ping",
			wantArgs:    []string{"test"},
		},
		{
			name:        "ニックネーム形式のメンション",
			content:     "<@!123456> ping",
			mentions:    []string{"123456"},
			wantCommand: "ping",
			wantArgs:    []string{},
		},
		{
			name:        "引数付きコマンド",
			content:     "<@123456> test arg1 arg2",
			mentions:    []string{"123456"},
			wantCommand: "test",
			wantArgs:    []string{"arg1", "arg2"},
		},
		{
			name:        "空のコンテンツ",
			content:     "",
			mentions:    []string{},
			wantCommand: "",
			wantArgs:    nil,
		},
		{
			name:        "メンションのみ",
			content:     "<@123456>",
			mentions:    []string{"123456"},
			wantCommand: "",
			wantArgs:    nil,
		},
		{
			name:        "スペースのみ",
			content:     "   ",
			mentions:    []string{},
			wantCommand: "",
			wantArgs:    nil,
		},
		{
			name:        "大文字のコマンド",
			content:     "<@123456> PING",
			mentions:    []string{"123456"},
			wantCommand: "ping",
			wantArgs:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCommand, gotArgs := ParseCommand(tt.content, tt.mentions)

			if gotCommand != tt.wantCommand {
				t.Errorf("ParseCommand() command = %v, want %v", gotCommand, tt.wantCommand)
			}

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("ParseCommand() args = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestCheckMention(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		mentions     []string
		targetUserID string
		want         bool
	}{
		{
			name:         "メンションあり",
			mentions:     []string{"123456", "789012"},
			targetUserID: "123456",
			want:         true,
		},
		{
			name:         "メンションなし",
			mentions:     []string{"789012", "345678"},
			targetUserID: "123456",
			want:         false,
		},
		{
			name:         "空のメンションリスト",
			mentions:     []string{},
			targetUserID: "123456",
			want:         false,
		},
		{
			name:         "単一メンション一致",
			mentions:     []string{"123456"},
			targetUserID: "123456",
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckMention(tt.mentions, tt.targetUserID)
			if got != tt.want {
				t.Errorf("CheckMention() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouter_ExecuteCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		commandName string
		args        []string
		wantResult  string
		wantErr     bool
	}{
		{
			name:        "pingコマンド実行",
			commandName: "ping",
			args:        []string{},
			wantResult:  "pong!!",
			wantErr:     false,
		},
		{
			name:        "helpコマンド実行",
			commandName: "help",
			args:        []string{},
			wantResult:  "", // 実際の内容はテストしない
			wantErr:     false,
		},
		{
			name:        "存在しないコマンド",
			commandName: "unknown",
			args:        []string{},
			wantResult:  "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMonitor := &system.MockMonitor{
				SystemInfo: &system.SystemInfo{
					CPUUsagePercent:   50.0,
					MemoryUsedGB:      8.0,
					MemoryTotalGB:     16.0,
					MemoryUsedPercent: 50.0,
					DiskFreeGB:        100.0,
					DiskTotalGB:       200.0,
					DiskUsedPercent:   50.0,
				},
			}
			mockCompose := &docker.MockComposeService{}
			router := NewRouter(&config.Config{}, mockMonitor, mockCompose)

			gotResult, err := router.ExecuteCommand(tt.commandName, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.commandName == "ping" && gotResult != tt.wantResult {
				t.Errorf("ExecuteCommand() result = %v, want %v", gotResult, tt.wantResult)
			}

			if tt.commandName == "help" && !tt.wantErr && gotResult == "" {
				t.Error("ExecuteCommand() help command should return non-empty result")
			}
		})
	}
}
