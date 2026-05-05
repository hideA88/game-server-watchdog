package usermsg

import (
	"strings"
	"testing"
)

func TestDockerPermissionMessage(t *testing.T) {
	t.Parallel()

	msg := DockerPermissionMessage()

	tests := []struct {
		name   string
		phrase string
	}{
		{name: "見出し", phrase: "Docker権限エラー"},
		{name: "ホスト向け案内", phrase: "newgrp docker"},
		{name: "groups コマンド案内", phrase: "groups"},
		{name: "コンテナ向け案内", phrase: "group_add"},
		{name: "socket マウント案内", phrase: "/var/run/docker.sock"},
		{name: "管理者連絡", phrase: "管理者"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(msg, tt.phrase) {
				t.Errorf("DockerPermissionMessage() に %q が含まれていません", tt.phrase)
			}
		})
	}
}
