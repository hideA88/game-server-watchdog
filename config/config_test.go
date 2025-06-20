package config

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hideA88/game-server-watchdog/pkg/logging"
)

// テーブル駆動テストのため長い関数を許可
func TestLoad(t *testing.T) {
	// Temporarily rename .env file if it exists
	if _, err := os.Stat(".env"); err == nil {
		if err := os.Rename(".env", ".env.test.bak"); err != nil {
			t.Fatalf("Failed to rename .env file: %v", err)
		}
		defer func() {
			if err := os.Rename(".env.test.bak", ".env"); err != nil {
				t.Errorf("Failed to restore .env file: %v", err)
			}
		}()
	}
	tests := []struct {
		name        string
		envVars     map[string]string
		want        *Config
		wantErr     bool
		setupFunc   func()
		cleanupFunc func()
	}{
		{
			name: "すべての環境変数が設定されている",
			envVars: map[string]string{
				"DISCORD_TOKEN":       "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				"DEBUG_MODE":          "true",
				"LOG_LEVEL":           "debug",
				"ALLOWED_CHANNEL_IDS": "123456789012345678,123456789012345679",
				"ALLOWED_USER_IDS":    "987654321098765432,987654321098765433",
			},
			want: &Config{
				DiscordToken:             "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				DebugMode:                true,
				LogLevel:                 logging.DebugLevel,
				LogLevelStr:              "debug",
				AllowedChannelIDs:        []string{"123456789012345678", "123456789012345679"},
				AllowedUserIDs:           []string{"987654321098765432", "987654321098765433"},
				DockerComposePath:        "docker-compose.yml",
				DockerComposeProjectName: "",
			},
			wantErr: false,
		},
		{
			name: "必須項目のみ設定",
			envVars: map[string]string{
				"DISCORD_TOKEN": "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
			},
			want: &Config{
				DiscordToken:             "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				DebugMode:                false,
				LogLevel:                 logging.InfoLevel,
				LogLevelStr:              "",
				AllowedChannelIDs:        nil,
				AllowedUserIDs:           nil,
				DockerComposePath:        "docker-compose.yml",
				DockerComposeProjectName: "",
			},
			wantErr: false,
		},
		{
			name: "DEBUG_MODEが無効な値",
			envVars: map[string]string{
				"DISCORD_TOKEN": "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				"DEBUG_MODE":    "invalid",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "空のチャンネルIDとユーザーID",
			envVars: map[string]string{
				"DISCORD_TOKEN":       "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				"ALLOWED_CHANNEL_IDS": "",
				"ALLOWED_USER_IDS":    "",
			},
			want: &Config{
				DiscordToken:             "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				DebugMode:                false,
				LogLevel:                 logging.InfoLevel,
				LogLevelStr:              "",
				AllowedChannelIDs:        []string{},
				AllowedUserIDs:           []string{},
				DockerComposePath:        "docker-compose.yml",
				DockerComposeProjectName: "",
			},
			wantErr: false,
		},
		{
			name: ".envファイルが存在しない場合でも環境変数から読み込む",
			envVars: map[string]string{
				"DISCORD_TOKEN": "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
			},
			want: &Config{
				DiscordToken:             "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				DebugMode:                false,
				LogLevel:                 logging.InfoLevel,
				LogLevelStr:              "",
				AllowedChannelIDs:        nil,
				AllowedUserIDs:           nil,
				DockerComposePath:        "docker-compose.yml",
				DockerComposeProjectName: "",
			},
			wantErr: false,
			setupFunc: func() {
				// .envファイルが存在しないことを確認
				_ = os.Remove(".env")
			},
		},
		{
			name:    "必須項目が不足",
			envVars: map[string]string{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "有効なログレベル（info）",
			envVars: map[string]string{
				"DISCORD_TOKEN": "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				"LOG_LEVEL":     "info",
			},
			want: &Config{
				DiscordToken:             "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				DebugMode:                false,
				LogLevel:                 logging.InfoLevel,
				LogLevelStr:              "info",
				AllowedChannelIDs:        nil,
				AllowedUserIDs:           nil,
				DockerComposePath:        "docker-compose.yml",
				DockerComposeProjectName: "",
			},
			wantErr: false,
		},
		{
			name: "有効なログレベル（warn）",
			envVars: map[string]string{
				"DISCORD_TOKEN": "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				"LOG_LEVEL":     "warn",
			},
			want: &Config{
				DiscordToken:             "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				DebugMode:                false,
				LogLevel:                 logging.WarnLevel,
				LogLevelStr:              "warn",
				AllowedChannelIDs:        nil,
				AllowedUserIDs:           nil,
				DockerComposePath:        "docker-compose.yml",
				DockerComposeProjectName: "",
			},
			wantErr: false,
		},
		{
			name: "有効なログレベル（error）",
			envVars: map[string]string{
				"DISCORD_TOKEN": "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				"LOG_LEVEL":     "error",
			},
			want: &Config{
				DiscordToken:             "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				DebugMode:                false,
				LogLevel:                 logging.ErrorLevel,
				LogLevelStr:              "error",
				AllowedChannelIDs:        nil,
				AllowedUserIDs:           nil,
				DockerComposePath:        "docker-compose.yml",
				DockerComposeProjectName: "",
			},
			wantErr: false,
		},
		{
			name: "無効なログレベル",
			envVars: map[string]string{
				"DISCORD_TOKEN": "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				"LOG_LEVEL":     "invalid",
			},
			want: &Config{
				DiscordToken:             "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				DebugMode:                false,
				LogLevel:                 logging.InfoLevel, // デフォルト値
				LogLevelStr:              "invalid",
				AllowedChannelIDs:        nil,
				AllowedUserIDs:           nil,
				DockerComposePath:        "docker-compose.yml",
				DockerComposeProjectName: "",
			},
			wantErr: false,
		},
		{
			name: "大文字のログレベル（DEBUG）",
			envVars: map[string]string{
				"DISCORD_TOKEN": "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				"LOG_LEVEL":     "DEBUG",
			},
			want: &Config{
				DiscordToken:             "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				DebugMode:                false,
				LogLevel:                 logging.DebugLevel,
				LogLevelStr:              "DEBUG",
				AllowedChannelIDs:        nil,
				AllowedUserIDs:           nil,
				DockerComposePath:        "docker-compose.yml",
				DockerComposeProjectName: "",
			},
			wantErr: false,
		},
		{
			name: "混在大文字小文字のログレベル（Info）",
			envVars: map[string]string{
				"DISCORD_TOKEN": "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				"LOG_LEVEL":     "Info",
			},
			want: &Config{
				DiscordToken:             "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				DebugMode:                false,
				LogLevel:                 logging.InfoLevel,
				LogLevelStr:              "Info",
				AllowedChannelIDs:        nil,
				AllowedUserIDs:           nil,
				DockerComposePath:        "docker-compose.yml",
				DockerComposeProjectName: "",
			},
			wantErr: false,
		},
		{
			name: "大文字のログレベル（WARN）",
			envVars: map[string]string{
				"DISCORD_TOKEN": "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				"LOG_LEVEL":     "WARN",
			},
			want: &Config{
				DiscordToken:             "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				DebugMode:                false,
				LogLevel:                 logging.WarnLevel,
				LogLevelStr:              "WARN",
				AllowedChannelIDs:        nil,
				AllowedUserIDs:           nil,
				DockerComposePath:        "docker-compose.yml",
				DockerComposeProjectName: "",
			},
			wantErr: false,
		},
		{
			name: "検証エラー（短いトークン）",
			envVars: map[string]string{
				"DISCORD_TOKEN": "short_token",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "検証エラー（無効なチャンネルID）",
			envVars: map[string]string{
				"DISCORD_TOKEN":       "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				"ALLOWED_CHANNEL_IDS": "invalid_channel_id",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "検証エラー（無効なユーザーID）",
			envVars: map[string]string{
				"DISCORD_TOKEN":    "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				"ALLOWED_USER_IDS": "invalid_user_id",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 既存の環境変数を保存
			originalEnv := make(map[string]string)
			envKeys := []string{
				"DISCORD_TOKEN", "DEBUG_MODE", "LOG_LEVEL",
				"ALLOWED_CHANNEL_IDS", "ALLOWED_USER_IDS",
				"DOCKER_COMPOSE_PATH", "DOCKER_COMPOSE_PROJECT_NAME",
			}
			for _, key := range envKeys {
				originalEnv[key] = os.Getenv(key)
				_ = os.Unsetenv(key)
			}

			// setup
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			// テスト用の環境変数を設定
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			// テスト実行
			got, err := Load()

			// クリーンアップ
			for key, value := range originalEnv {
				if value == "" {
					_ = os.Unsetenv(key)
				} else {
					_ = os.Setenv(key, value)
				}
			}

			if tt.cleanupFunc != nil {
				tt.cleanupFunc()
			}

			// アサーション
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// テスト用の.envファイルを作成するヘルパー関数
func createTestEnvFile(t *testing.T, content string) func() {
	t.Helper()

	// 既存の.envファイルをバックアップ
	data, err := os.ReadFile(".env")
	hasBackup := err == nil

	// テスト用の.envファイルを作成
	if err := os.WriteFile(".env", []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// クリーンアップ関数を返す
	return func() {
		if hasBackup {
			// バックアップを復元
			_ = os.WriteFile(".env", data, 0600)
		} else {
			// .envファイルを削除
			_ = os.Remove(".env")
		}
	}
}

func TestLoad_WithEnvFile(t *testing.T) { // テーブル駆動テストのため長い関数を許可
	// 既存の.envファイルを一時的にリネーム
	if _, err := os.Stat(".env"); err == nil {
		if err := os.Rename(".env", ".env.test.bak"); err != nil {
			t.Fatalf("Failed to rename .env file: %v", err)
		}
		defer func() {
			if err := os.Rename(".env.test.bak", ".env"); err != nil {
				t.Errorf("Failed to restore .env file: %v", err)
			}
		}()
	}

	tests := []struct {
		name    string
		envFile string
		want    *Config
		wantErr bool
	}{
		{
			name: ".envファイルから読み込み",
			envFile: `DISCORD_TOKEN=MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example
DEBUG_MODE=true
ALLOWED_CHANNEL_IDS=123456789012345678,123456789012345679
ALLOWED_USER_IDS=987654321098765432,987654321098765433`,
			want: &Config{
				DiscordToken:             "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				DebugMode:                true,
				LogLevel:                 logging.InfoLevel,
				LogLevelStr:              "",
				AllowedChannelIDs:        []string{"123456789012345678", "123456789012345679"},
				AllowedUserIDs:           []string{"987654321098765432", "987654321098765433"},
				DockerComposePath:        "docker-compose.yml",
				DockerComposeProjectName: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数をクリア
			envKeys := []string{
				"DISCORD_TOKEN", "DEBUG_MODE", "LOG_LEVEL",
				"ALLOWED_CHANNEL_IDS", "ALLOWED_USER_IDS",
				"DOCKER_COMPOSE_PATH", "DOCKER_COMPOSE_PROJECT_NAME",
			}
			originalEnv := make(map[string]string)
			for _, key := range envKeys {
				originalEnv[key] = os.Getenv(key)
				_ = os.Unsetenv(key)
			}

			// テスト用の.envファイルを作成
			cleanup := createTestEnvFile(t, tt.envFile)
			defer cleanup()

			// テスト実行
			got, err := Load()

			// 環境変数を復元
			for key, value := range originalEnv {
				if value == "" {
					_ = os.Unsetenv(key)
				} else {
					_ = os.Setenv(key, value)
				}
			}

			// アサーション
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "有効な設定",
			config: Config{
				DiscordToken:      "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				AllowedChannelIDs: []string{"123456789012345678"},
				AllowedUserIDs:    []string{"987654321098765432"},
				DockerComposePath: "",
			},
			wantErr: false,
		},
		{
			name: "無効なDiscordトークン（短すぎる）",
			config: Config{
				DiscordToken: "short",
			},
			wantErr: true,
			errMsg:  "token too short",
		},
		{
			name: "無効なDiscordトークン（ドットで区切られていない）",
			config: Config{
				DiscordToken: "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_test",
			},
			wantErr: true,
			errMsg:  "invalid token format",
		},
		{
			name: "無効なDockerComposePath（ファイルが存在しない）",
			config: Config{
				DiscordToken:      "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				DockerComposePath: "/non/existent/path/docker-compose.yml",
			},
			wantErr: true,
			errMsg:  "DOCKER_COMPOSE_PATH file not found",
		},
		{
			name: "無効なチャンネルID（短すぎる）",
			config: Config{
				DiscordToken:      "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				AllowedChannelIDs: []string{"12345"},
			},
			wantErr: true,
			errMsg:  "invalid channel ID",
		},
		{
			name: "無効なチャンネルID（長すぎる）",
			config: Config{
				DiscordToken:      "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				AllowedChannelIDs: []string{"12345678901234567890"},
			},
			wantErr: true,
			errMsg:  "invalid channel ID",
		},
		{
			name: "無効なチャンネルID（数字以外を含む）",
			config: Config{
				DiscordToken:      "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				AllowedChannelIDs: []string{"12345678901234567a"},
			},
			wantErr: true,
			errMsg:  "invalid channel ID",
		},
		{
			name: "無効なユーザーID（短すぎる）",
			config: Config{
				DiscordToken:   "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				AllowedUserIDs: []string{"12345"},
			},
			wantErr: true,
			errMsg:  "invalid user ID",
		},
		{
			name: "無効なユーザーID（数字以外を含む）",
			config: Config{
				DiscordToken:   "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				AllowedUserIDs: []string{"98765432109876543X"},
			},
			wantErr: true,
			errMsg:  "invalid user ID",
		},
		{
			name: "複数のエラー",
			config: Config{
				DiscordToken:      "short",
				AllowedChannelIDs: []string{"invalid"},
				AllowedUserIDs:    []string{"invalid"},
				DockerComposePath: "/non/existent/path.yml",
			},
			wantErr: true,
			errMsg:  "invalid DISCORD_TOKEN",
		},
		{
			name: "空のチャンネルIDとユーザーID（エラーなし）",
			config: Config{
				DiscordToken:      "MTIzNDU2Nzg5MDEyMzQ1Njc4OS5GdUNrLkluc1AvdXVzZWNyZXRzaGg_.test.example",
				AllowedChannelIDs: []string{""},
				AllowedUserIDs:    []string{""},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
