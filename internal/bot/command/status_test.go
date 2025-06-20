package command

import (
	"errors"
	"strings"
	"testing"

	"github.com/hideA88/game-server-watchdog/pkg/system"
)

func TestStatusCommand_Name(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want string
	}{
		{
			name: "コマンド名がstatusであること",
			want: "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockMonitor := &system.MockMonitor{}
			cmd := NewStatusCommand(mockMonitor)
			if got := cmd.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusCommand_Description(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want string
	}{
		{
			name: "説明文が正しいこと",
			want: "サーバーのステータスを表示",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockMonitor := &system.MockMonitor{}
			cmd := NewStatusCommand(mockMonitor)
			if got := cmd.Description(); got != tt.want {
				t.Errorf("Description() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusCommand_Execute(t *testing.T) { //nolint:funlen // テーブル駆動テストのため長い関数を許可
	t.Parallel()
	tests := []struct {
		name         string
		systemInfo   *system.SystemInfo
		mockErr      error
		args         []string
		wantErr      bool
		wantContains []string
	}{
		{
			name: "正常なシステム情報を表示",
			systemInfo: &system.SystemInfo{
				CPUUsagePercent:   25.5,
				MemoryUsedGB:      8.0,
				MemoryTotalGB:     16.0,
				MemoryUsedPercent: 50.0,
				DiskFreeGB:        256.0,
				DiskTotalGB:       512.0,
				DiskUsedPercent:   50.0,
			},
			args:    []string{},
			wantErr: false,
			wantContains: []string{
				"📊 **サーバーステータス**",
				"CPU使用率: 25.5%",
				"メモリ使用量: 8.0GB / 16.0GB (50.0%)",
				"ディスク空き容量: 256.0GB / 512.0GB (50.0%)",
			},
		},
		{
			name:       "システム情報取得エラー",
			systemInfo: nil,
			mockErr:    errors.New("monitor error"),
			args:       []string{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockMonitor := &system.MockMonitor{
				SystemInfo: tt.systemInfo,
				Err:        tt.mockErr,
			}
			cmd := NewStatusCommand(mockMonitor)

			got, err := cmd.Execute(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				for _, want := range tt.wantContains {
					if !strings.Contains(got, want) {
						t.Errorf("Execute() = %v, want to contain %v", got, want)
					}
				}
			}
		})
	}
}
