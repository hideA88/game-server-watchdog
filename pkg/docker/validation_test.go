package docker

import (
	"testing"
)

func TestIsValidServiceName(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{
			name: "有効な名前（英数字）",
			arg:  "minecraft",
			want: true,
		},
		{
			name: "有効な名前（アンダースコア）",
			arg:  "minecraft_server",
			want: true,
		},
		{
			name: "有効な名前（ハイフン）",
			arg:  "minecraft-server",
			want: true,
		},
		{
			name: "有効な名前（数字）",
			arg:  "server123",
			want: true,
		},
		{
			name: "有効な名前（大文字）",
			arg:  "MinecraftServer",
			want: true,
		},
		{
			name: "無効な名前（空文字）",
			arg:  "",
			want: false,
		},
		{
			name: "無効な名前（スペース）",
			arg:  "minecraft server",
			want: false,
		},
		{
			name: "無効な名前（セミコロン）",
			arg:  "minecraft;echo",
			want: false,
		},
		{
			name: "無効な名前（パイプ）",
			arg:  "minecraft|echo",
			want: false,
		},
		{
			name: "無効な名前（アンパサンド）",
			arg:  "minecraft&echo",
			want: false,
		},
		{
			name: "無効な名前（バッククォート）",
			arg:  "minecraft`echo`",
			want: false,
		},
		{
			name: "無効な名前（ドル記号）",
			arg:  "minecraft$VAR",
			want: false,
		},
		{
			name: "無効な名前（括弧）",
			arg:  "minecraft(echo)",
			want: false,
		},
		{
			name: "無効な名前（リダイレクト）",
			arg:  "minecraft>file",
			want: false,
		},
		{
			name: "無効な名前（パス区切り）",
			arg:  "minecraft/server",
			want: false,
		},
		{
			name: "無効な名前（ドット）",
			arg:  "minecraft.server",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidServiceName(tt.arg); got != tt.want {
				t.Errorf("IsValidServiceName() = %v, want %v", got, tt.want)
			}
		})
	}
}
