package security

import (
	"strings"
	"testing"
)

func TestSanitizeErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tokenを含むメッセージ",
			input:    "Discord token: abc123def456ghi789jkl012mno345pqr678",
			expected: "Discord token: [REDACTED]",
		},
		{
			name:     "secretを含むメッセージ",
			input:    "API secret: xyz987uvw654rst321opq098nml765",
			expected: "API secret: [REDACTED]",
		},
		{
			name:     "passwordを含むメッセージ",
			input:    "Database password: mypassword123",
			expected: "Database password: [REDACTED]",
		},
		{
			name:     "keyを含むメッセージ",
			input:    "Private key: abcdef123456789012345678901234567890",
			expected: "Private key: [REDACTED]",
		},
		{
			name:     "api_keyを含むメッセージ",
			input:    "API_KEY: sk-1234567890abcdefghijklmnop",
			expected: "API_KEY: [REDACTED]",
		},
		{
			name:     "ファイルパスを含むメッセージ（非一般的）",
			input:    "Error reading file /home/user/secret/config.json",
			expected: "Error reading file [PATH_REDACTED]",
		},
		{
			name:     "一般的なファイルパス（保持される）",
			input:    "Error reading file /tmp/tempfile.txt",
			expected: "Error reading file /tmp/tempfile.txt",
		},
		{
			name:     "Windowsパスを含むメッセージ",
			input:    "Error reading C:\\Users\\admin\\secrets\\config.ini",
			expected: "Error reading [PATH_REDACTED]",
		},
		{
			name:     "パブリックIPアドレス",
			input:    "Connection failed to 203.0.113.1",
			expected: "Connection failed to [IP_REDACTED]",
		},
		{
			name:     "プライベートIPアドレス（保持される）",
			input:    "Connection to 192.168.1.1 successful",
			expected: "Connection to 192.168.1.1 successful",
		},
		{
			name:     "localhostアドレス（保持される）",
			input:    "Server running on 127.0.0.1:8080",
			expected: "Server running on 127.0.0.1:8080",
		},
		{
			name:     "IPv6プライベートアドレス（保持される）",
			input:    "IPv6 address ::1 is localhost",
			expected: "IPv6 address ::1 is localhost",
		},
		{
			name:     "複数のセンシティブ情報",
			input:    "token: abc123def456 secret: xyz789 at /home/user/app/config",
			expected: "token: [REDACTED] secret: [REDACTED] at [PATH_REDACTED]",
		},
		{
			name:     "センシティブ情報なし",
			input:    "Normal error message without sensitive data",
			expected: "Normal error message without sensitive data",
		},
		{
			name:     "短いトークン（マッチしない）",
			input:    "token: abc",
			expected: "token: abc",
		},
		{
			name:     "大文字小文字の混在",
			input:    "TOKEN: AbC123DeF456GhI789JkL012MnO345PqR678",
			expected: "TOKEN: [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeErrorMessage(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeErrorMessage() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizeForDiscord(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Discord特殊文字のエスケープ",
			input:    "Error with `code` and *bold* and _italic_ and ~strikethrough~",
			expected: "Error with \\`code\\` and \\*bold\\* and \\_italic\\_ and \\~strikethrough\\~",
		},
		{
			name:     "トークンとDiscord特殊文字の組み合わせ",
			input:    "Discord token: `abc123def456ghi789jkl012mno345pqr678`",
			expected: "Discord token: \\`[REDACTED]\\`",
		},
		{
			name:     "特殊文字なし",
			input:    "Normal message",
			expected: "Normal message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeForDiscord(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeForDiscord() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizeForLogging(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ログ用サニタイズ",
			input:    "token: abc123def456ghi789jkl012mno345pqr678",
			expected: "token: [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeForLogging(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeForLogging() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIsCommonPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "一般的なパス - /tmp",
			path:     "/tmp/file.txt",
			expected: true,
		},
		{
			name:     "一般的なパス - /var",
			path:     "/var/log/app.log",
			expected: true,
		},
		{
			name:     "一般的なパス - /usr",
			path:     "/usr/bin/app",
			expected: true,
		},
		{
			name:     "一般的なパス - /etc",
			path:     "/etc/config.conf",
			expected: true,
		},
		{
			name:     "一般的なパス - /home",
			path:     "/home/user",
			expected: true,
		},
		{
			name:     "非一般的なパス",
			path:     "/secret/private/key.pem",
			expected: false,
		},
		{
			name:     "相対パス",
			path:     "./config/secret.json",
			expected: false,
		},
		{
			name:     "大文字小文字の違い",
			path:     "/TMP/file.txt",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCommonPath(tt.path)
			if result != tt.expected {
				t.Errorf("isCommonPath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsLocalOrPrivateIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "localhost IPv4",
			ip:       "127.0.0.1",
			expected: true,
		},
		{
			name:     "プライベートIP - 10.x.x.x",
			ip:       "10.0.0.1",
			expected: true,
		},
		{
			name:     "プライベートIP - 192.168.x.x",
			ip:       "192.168.1.1",
			expected: true,
		},
		{
			name:     "プライベートIP - 172.x.x.x",
			ip:       "172.16.0.1",
			expected: true,
		},
		{
			name:     "パブリックIP",
			ip:       "203.0.113.1",
			expected: false,
		},
		{
			name:     "GoogleのDNS",
			ip:       "8.8.8.8",
			expected: false,
		},
		{
			name:     "IPv6 localhost",
			ip:       "::1",
			expected: true,
		},
		{
			name:     "IPv6 link-local",
			ip:       "fe80::1",
			expected: true,
		},
		{
			name:     "IPv6 unique local",
			ip:       "fc00::1",
			expected: true,
		},
		{
			name:     "IPv6 unique local fd",
			ip:       "fd00::1",
			expected: true,
		},
		{
			name:     "IPv6 パブリック",
			ip:       "2001:db8::1",
			expected: false,
		},
		{
			name:     "不正なIP",
			ip:       "not.an.ip",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLocalOrPrivateIP(tt.ip)
			if result != tt.expected {
				t.Errorf("isLocalOrPrivateIP(%q) = %v, want %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestContainsCredentials(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "tokenを含む",
			message:  "Discord token is invalid",
			expected: true,
		},
		{
			name:     "passwordを含む",
			message:  "Wrong password provided",
			expected: true,
		},
		{
			name:     "secretを含む",
			message:  "API secret key missing",
			expected: true,
		},
		{
			name:     "keyを含む",
			message:  "Private key not found",
			expected: true,
		},
		{
			name:     "authを含む",
			message:  "Authentication failed",
			expected: true,
		},
		{
			name:     "credentialを含む",
			message:  "Invalid credentials",
			expected: true,
		},
		{
			name:     "bearerを含む",
			message:  "Bearer token required",
			expected: true,
		},
		{
			name:     "oauthを含む",
			message:  "OAuth flow completed",
			expected: true,
		},
		{
			name:     "jwtを含む",
			message:  "JWT token expired",
			expected: true,
		},
		{
			name:     "大文字小文字の違い",
			message:  "TOKEN is missing",
			expected: true,
		},
		{
			name:     "認証情報なし",
			message:  "Connection timeout error",
			expected: false,
		},
		{
			name:     "空文字列",
			message:  "",
			expected: false,
		},
		{
			name:     "部分マッチ",
			message:  "tokenize the input",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsCredentials(tt.message)
			if result != tt.expected {
				t.Errorf("ContainsCredentials(%q) = %v, want %v", tt.message, result, tt.expected)
			}
		})
	}
}

// ベンチマークテスト
func BenchmarkSanitizeErrorMessage(b *testing.B) {
	message := "Discord token: abc123def456ghi789jkl012mno345pqr678 at /home/user/secret/config.json with IP 203.0.113.1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeErrorMessage(message)
	}
}

func BenchmarkSanitizeForDiscord(b *testing.B) {
	message := "Error with `code` and *bold* token: abc123def456ghi789jkl012mno345pqr678"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeForDiscord(message)
	}
}

// エッジケースのテスト
func TestSanitizeErrorMessage_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "空文字列",
			input:    "",
			expected: "",
		},
		{
			name:     "非常に長い文字列",
			input:    strings.Repeat("token: abc123def456ghi789jkl012mno345pqr678 ", 1000),
			expected: strings.Repeat("token: [REDACTED] ", 1000),
		},
		{
			name:     "特殊文字を含むトークン",
			input:    "token: abc-123_def.456+ghi789",
			expected: "token: [REDACTED]",
		},
		{
			name:     "複数行にわたるメッセージ",
			input:    "Error occurred:\ntoken: abc123def456ghi789jkl012mno345pqr678\nsecret: xyz789uvw456rst123",
			expected: "Error occurred:\ntoken: [REDACTED]\nsecret: [REDACTED]",
		},
		{
			name:     "日本語を含むメッセージ",
			input:    "エラー: token: abc123def456ghi789jkl012mno345pqr678 が無効です",
			expected: "エラー: token: [REDACTED] が無効です",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeErrorMessage(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeErrorMessage() = %q, want %q", result, tt.expected)
			}
		})
	}
}
