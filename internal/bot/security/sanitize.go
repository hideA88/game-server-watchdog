// Package security はセキュリティ関連のユーティリティを提供します
package security

import (
	"regexp"
	"strings"
)

var (
	// tokenPatterns はトークンのようなセンシティブな情報を検出する正規表現
	tokenPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(token[:\s]+)([\x60]?)([a-zA-Z0-9.\-_+]{8,})([\x60]?)`),  // \x60 = backtick
		regexp.MustCompile(`(?i)(secret[:\s]+)([\x60]?)([a-zA-Z0-9.\-_+]{6,})([\x60]?)`), // secretは6文字以上
		regexp.MustCompile(`(?i)(password[:\s]+)([\x60]?)([a-zA-Z0-9.\-_+]{4,})([\x60]?)`),
		regexp.MustCompile(`(?i)(key[:\s]+)([\x60]?)([a-zA-Z0-9.\-_+]{8,})([\x60]?)`),
		regexp.MustCompile(`(?i)(api[_\s]*key[:\s]+)([\x60]?)([a-zA-Z0-9.\-_+]{8,})([\x60]?)`),
	}

	// pathPatterns はファイルパスのような情報を検出する正規表現
	pathPatterns = []*regexp.Regexp{
		regexp.MustCompile(`/[a-zA-Z0-9.\-_/]+`),
		regexp.MustCompile(`[a-zA-Z]:\\[a-zA-Z0-9.\-_\\]+`),
	}

	// ipPatterns はIPアドレスを検出する正規表現
	ipPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}(?::\d+)?\b`),           // IPv4（ポート番号含む）
		regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){2,7}[0-9a-fA-F]{0,4}\b`), // IPv6の改良パターン
	}
)

// SanitizeErrorMessage はエラーメッセージからセンシティブな情報を削除します
func SanitizeErrorMessage(message string) string {
	sanitized := message

	// トークンやシークレットを削除
	for _, pattern := range tokenPatterns {
		sanitized = pattern.ReplaceAllString(sanitized, "${1}${2}[REDACTED]${4}")
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
		"/proc", "/sys", "/dev", "/run",
	}

	path = strings.ToLower(path)

	// /homeは特別扱い - /home/username までは保持、それ以降の個人ファイルは隠す
	if strings.HasPrefix(path, "/home/") {
		parts := strings.Split(path, "/")
		// /home/username までは一般的、それ以下の深いパスは非一般的
		return len(parts) <= 3
	}

	for _, common := range commonPaths {
		if strings.HasPrefix(path, common) {
			return true
		}
	}

	return false
}

// isLocalOrPrivateIP はローカルまたはプライベートIPアドレスかどうかを判定します
func isLocalOrPrivateIP(ip string) bool {
	// IPv4の場合（ポート番号を含む可能性がある）
	if strings.Contains(ip, ".") {
		// ポート番号を削除してIPアドレス部分だけを取得
		ipOnly := strings.Split(ip, ":")[0]
		if strings.HasPrefix(ipOnly, "127.") ||
			strings.HasPrefix(ipOnly, "10.") ||
			strings.HasPrefix(ipOnly, "192.168.") ||
			strings.HasPrefix(ipOnly, "172.") {
			return true
		}
	}

	// IPv6の場合（簡易判定）
	if strings.Contains(ip, ":") && !strings.Contains(ip, ".") {
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
