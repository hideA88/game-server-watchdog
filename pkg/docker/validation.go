package docker

import (
	"errors"
	"regexp"
)

var (
	// ErrInvalidServiceName はサービス名が不正な場合のエラー
	ErrInvalidServiceName = errors.New(
		"invalid service name: must contain only alphanumeric characters, hyphens, and underscores")

	// サービス名の検証用正規表現
	serviceNameRegex = regexp.MustCompile("^[a-zA-Z0-9_-]+$")
)

// IsValidServiceName はサービス名が有効かどうかを検証する
func IsValidServiceName(name string) bool {
	if name == "" {
		return false
	}
	return serviceNameRegex.MatchString(name)
}
