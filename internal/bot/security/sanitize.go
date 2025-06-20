// Package security はセキュリティ関連のユーティリティを提供します
package security

import (
	"regexp"
	"strings"
)

var (
	// tokenPatterns はトークンのようなセンシティブな情報を検出する正規表現
	tokenPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)token[:\s]*[a-zA-Z0-9.\-_]{20,}`),
		regexp.MustCompile(`(?i)secret[:\s]*[a-zA-Z0-9.\-_]{20,}`),
		regexp.MustCompile(`(?i)password[:\s]*[a-zA-Z0-9.\-_]{8,}`),
		regexp.MustCompile(`(?i)key[:\s]*[a-zA-Z0-9.\-_]{20,}`),
		regexp.MustCompile(`(?i)api[_\s]*key[:\s]*[a-zA-Z0-9.\-_]{20,}`),
	}

	// pathPatterns はファイルパスのような情報を検出する正規表現
	pathPatterns = []*regexp.Regexp{
		regexp.MustCompile(`/[a-zA-Z0-9.\-_/]+`),
		regexp.MustCompile(`[a-zA-Z]:\\[a-zA-Z0-9.\-_\\]+`),
	}

	// ipPatterns はIPアドレスを検出する正規表現
	ipPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
		regexp.MustCompile(`\b[0-9a-fA-F:]+:[0-9a-fA-F:]+\b`), // IPv6の簡易パターン
	}
)

// SanitizeErrorMessage はエラーメッセージからセンシティブな情報を削除します
func SanitizeErrorMessage(message string) string {
	sanitized := message

	// トークンやシークレットを削除
	for _, pattern := range tokenPatterns {
		sanitized = pattern.ReplaceAllStringFunc(sanitized, func(match string) string {
			parts := strings.SplitN(match, ":", 2)
			if len(parts) == 2 {
				return parts[0] + ": [REDACTED]"
			}
			return "[REDACTED]"
		})
	}

	// ファイルパスを削除（ただし、一般的でないパスのみ）
	for _, pattern := range pathPatterns {
		sanitized = pattern.ReplaceAllStringFunc(sanitized, func(match string) string {
			// 一般的なパス（/tmp, /var, /usr等）は残す
			if isCommonPath(match) {
				return match
			}
			return "[PATH_REDACTED]"
		})
	}

	// IPアドレスを削除
	for _, pattern := range ipPatterns {
		sanitized = pattern.ReplaceAllStringFunc(sanitized, func(match string) string {
			// localhostやプライベートIPは残す
			if isLocalOrPrivateIP(match) {
				return match
			}
			return "[IP_REDACTED]"
		})
	}

	return sanitized
}

// SanitizeForLogging はログ出力用にメッセージをサニタイズします
func SanitizeForLogging(message string) string {
	return SanitizeErrorMessage(message)
}

// SanitizeForDiscord はDiscord出力用にメッセージをサニタイズします
func SanitizeForDiscord(message string) string {
	sanitized := SanitizeErrorMessage(message)

	// Discord特有のエスケープ
	sanitized = strings.ReplaceAll(sanitized, "`", "\\`")
	sanitized = strings.ReplaceAll(sanitized, "*", "\\*")
	sanitized = strings.ReplaceAll(sanitized, "_", "\\_")
	sanitized = strings.ReplaceAll(sanitized, "~", "\\~")

	return sanitized
}

// isCommonPath は一般的なシステムパスかどうかを判定します
func isCommonPath(path string) bool {
	commonPaths := []string{
		"/tmp", "/var", "/usr", "/etc", "/opt", "/bin", "/sbin",
		"/proc", "/sys", "/dev", "/run", "/home",
	}

	path = strings.ToLower(path)
	for _, common := range commonPaths {
		if strings.HasPrefix(path, common) {
			return true
		}
	}

	return false
}

// isLocalOrPrivateIP はローカルまたはプライベートIPアドレスかどうかを判定します
func isLocalOrPrivateIP(ip string) bool {
	// IPv4の場合
	if strings.Contains(ip, ".") {
		if strings.HasPrefix(ip, "127.") ||
			strings.HasPrefix(ip, "10.") ||
			strings.HasPrefix(ip, "192.168.") ||
			strings.HasPrefix(ip, "172.") {
			return true
		}
	}

	// IPv6の場合（簡易判定）
	if strings.Contains(ip, ":") {
		if strings.HasPrefix(ip, "::1") ||
			strings.HasPrefix(ip, "fe80:") ||
			strings.HasPrefix(ip, "fc00:") ||
			strings.HasPrefix(ip, "fd00:") {
			return true
		}
	}

	return false
}

// ContainsCredentials はメッセージに認証情報が含まれているかをチェックします
func ContainsCredentials(message string) bool {
	message = strings.ToLower(message)

	keywords := []string{
		"token", "password", "secret", "key", "auth",
		"credential", "bearer", "oauth", "jwt",
	}

	for _, keyword := range keywords {
		if strings.Contains(message, keyword) {
			return true
		}
	}

	return false
}
