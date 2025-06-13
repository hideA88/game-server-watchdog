package command

import (
	"strings"
	"testing"
)

func TestHelpCommand_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "コマンド名がhelpであること",
			want: "help",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewHelpCommand()
			if got := cmd.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHelpCommand_Description(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "説明文が正しいこと",
			want: "コマンド一覧を表示",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewHelpCommand()
			if got := cmd.Description(); got != tt.want {
				t.Errorf("Description() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHelpCommand_SetCommands(t *testing.T) {
	tests := []struct {
		name          string
		commands      []Command
		wantCount     int
		wantCommands  []string
	}{
		{
			name: "2つのコマンドを設定",
			commands: []Command{
				NewPingCommand(),
				NewHelpCommand(),
			},
			wantCount:    2,
			wantCommands: []string{"ping", "help"},
		},
		{
			name:          "空のコマンドスライスを設定",
			commands:      []Command{},
			wantCount:     0,
			wantCommands:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewHelpCommand()
			cmd.SetCommands(tt.commands)

			// コマンド数の確認
			if len(cmd.commands) != tt.wantCount {
				t.Errorf("Expected %d commands, got %d", tt.wantCount, len(cmd.commands))
			}

			// 各コマンドの存在確認
			for i, wantCmd := range tt.wantCommands {
				if i < len(cmd.commands) && cmd.commands[i].Name() != wantCmd {
					t.Errorf("Expected command %s at index %d, got %s", wantCmd, i, cmd.commands[i].Name())
				}
			}
		})
	}
}

func TestHelpCommand_Execute(t *testing.T) {
	tests := []struct {
		name     string
		commands []Command
		args     []string
		wantErr  bool
	}{
		{
			name: "コマンド一覧を表示",
			commands: []Command{
				NewPingCommand(),
				NewHelpCommand(),
			},
			args:    []string{},
			wantErr: false,
		},
		{
			name:     "コマンドなしの場合",
			commands: []Command{},
			args:     []string{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewHelpCommand()
			cmd.SetCommands(tt.commands)
			
			got, err := cmd.Execute(tt.args)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// マップの順序が不定なので、含まれているかチェック
			if tt.name == "コマンド一覧を表示" && len(tt.commands) > 0 {
				expectedParts := []string{
					"**利用可能なコマンド:**",
					"`@ボット ping` - ボットの応答を確認",
					"`@ボット help` - コマンド一覧を表示",
					"**使い方:**\nボットをメンションしてコマンドを送信してください。",
				}
				for _, part := range expectedParts {
					if !strings.Contains(got, part) {
						t.Errorf("Execute() missing expected part: %v", part)
					}
				}
			}
		})
	}
}