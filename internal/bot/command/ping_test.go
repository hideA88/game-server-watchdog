package command

import (
	"testing"
)

func TestPingCommand_Name(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want string
	}{
		{
			name: "コマンド名がpingであること",
			want: "ping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewPingCommand()
			if got := cmd.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPingCommand_Description(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want string
	}{
		{
			name: "説明文が正しいこと",
			want: "ボットの応答を確認",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewPingCommand()
			if got := cmd.Description(); got != tt.want {
				t.Errorf("Description() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPingCommand_Execute(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name:    "正常にpong!!を返す",
			args:    []string{},
			want:    "pong!!",
			wantErr: false,
		},
		{
			name:    "引数があってもpong!!を返す",
			args:    []string{"test", "args"},
			want:    "pong!!",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewPingCommand()
			got, err := cmd.Execute(tt.args)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if got != tt.want {
				t.Errorf("Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}