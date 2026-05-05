package errors

import (
	"errors"
	"testing"
)

func TestIsDockerPermissionError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "Docker権限エラー",
			err:  errors.New("permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock"),
			want: true,
		},
		{
			name: "別のDocker権限エラー",
			err:  errors.New("dial unix /var/run/docker.sock: connect: permission denied"),
			want: true,
		},
		{
			name: "通常のpermission deniedエラー",
			err:  errors.New("permission denied"),
			want: false,
		},
		{
			name: "docker.sockを含むが権限エラーではない",
			err:  errors.New("cannot connect to docker.sock"),
			want: false,
		},
		{
			name: "nilエラー",
			err:  nil,
			want: false,
		},
		{
			name: "別のエラー",
			err:  errors.New("container not found"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDockerPermissionError(tt.err); got != tt.want {
				t.Errorf("IsDockerPermissionError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDockerPermissionErrorMessage(t *testing.T) {
	msg := GetDockerPermissionErrorMessage()

	// メッセージに必要な情報が含まれているか確認
	expectedPhrases := []string{
		"Docker権限エラー",
		"newgrp docker",
		"groups",
		"ls -la /var/run/docker.sock",
		"管理者",
	}

	for _, phrase := range expectedPhrases {
		if !contains(msg, phrase) {
			t.Errorf("GetDockerPermissionErrorMessage() should contain %q", phrase)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
