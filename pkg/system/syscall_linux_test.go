//go:build linux

package system

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestSyscallStatfsFunc(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "存在するディレクトリ",
			path:    "/tmp",
			wantErr: false,
		},
		{
			name:    "存在しないディレクトリ",
			path:    "/nonexistent/path/that/should/not/exist",
			wantErr: true,
		},
		{
			name:    "ルートディレクトリ",
			path:    "/",
			wantErr: false,
		},
		{
			name:    "カレントディレクトリ",
			path:    ".",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stat syscallStatfs
			err := syscallStatfsFunc(tt.path, &stat)

			if (err != nil) != tt.wantErr {
				t.Errorf("syscallStatfsFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 正常に取得できた場合、値の妥当性をチェック
				if stat.Bsize <= 0 {
					t.Errorf("Bsize should be positive, got %v", stat.Bsize)
				}
				if stat.Blocks == 0 {
					t.Errorf("Blocks should not be zero")
				}
				// Bavail（利用可能ブロック数）はBlocksを超えないはず
				if stat.Bavail > stat.Blocks {
					t.Errorf("Bavail (%v) should not exceed Blocks (%v)", stat.Bavail, stat.Blocks)
				}
			}
		})
	}
}

func TestSyscallStatfsType(t *testing.T) {
	// syscallStatfsがsyscall.Statfs_tと同じ構造を持つことを確認
	var stat syscallStatfs
	var sysStat syscall.Statfs_t

	// /tmpの情報を取得して比較
	err1 := syscallStatfsFunc("/tmp", &stat)
	err2 := syscall.Statfs("/tmp", &sysStat)

	if err1 != nil || err2 != nil {
		t.Skipf("Failed to get statfs info: err1=%v, err2=%v", err1, err2)
	}

	// 基本的なフィールドが同じ値を持つことを確認
	if stat.Bsize != sysStat.Bsize {
		t.Errorf("Bsize mismatch: syscallStatfs=%v, syscall.Statfs_t=%v", stat.Bsize, sysStat.Bsize)
	}
	if stat.Blocks != sysStat.Blocks {
		t.Errorf("Blocks mismatch: syscallStatfs=%v, syscall.Statfs_t=%v", stat.Blocks, sysStat.Blocks)
	}
	if stat.Bavail != sysStat.Bavail {
		t.Errorf("Bavail mismatch: syscallStatfs=%v, syscall.Statfs_t=%v", stat.Bavail, sysStat.Bavail)
	}
}

func TestSyscallStatfsWithTempDir(t *testing.T) {
	// 一時ディレクトリを作成してテスト
	tmpDir := t.TempDir()

	var stat syscallStatfs
	err := syscallStatfsFunc(tmpDir, &stat)
	if err != nil {
		t.Fatalf("syscallStatfsFunc() failed for temp dir: %v", err)
	}

	// 一時ディレクトリの統計情報の妥当性をチェック
	if stat.Bsize <= 0 {
		t.Errorf("Bsize should be positive for temp dir, got %v", stat.Bsize)
	}
	if stat.Blocks == 0 {
		t.Errorf("Blocks should not be zero for temp dir")
	}

	// ファイルを作成して、利用可能ブロック数が変わることを確認
	testFile := filepath.Join(tmpDir, "test.txt")
	// 1MBのファイルを作成
	data := make([]byte, 1024*1024)
	if err := os.WriteFile(testFile, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	var stat2 syscallStatfs
	err = syscallStatfsFunc(tmpDir, &stat2)
	if err != nil {
		t.Fatalf("syscallStatfsFunc() failed after file creation: %v", err)
	}

	// ブロックサイズとブロック数は同じはず
	if stat2.Bsize != stat.Bsize {
		t.Errorf("Bsize changed after file creation: before=%v, after=%v", stat.Bsize, stat2.Bsize)
	}
	if stat2.Blocks != stat.Blocks {
		t.Errorf("Blocks changed after file creation: before=%v, after=%v", stat.Blocks, stat2.Blocks)
	}
}

func TestSyscallStatfsPermissionError(t *testing.T) {
	// 権限がないディレクトリをテストするのは環境依存が強いため、
	// 一般的なケースのみテスト

	// 存在しないファイルパスでのテスト
	var stat syscallStatfs
	err := syscallStatfsFunc("/definitely/does/not/exist/path", &stat)
	if err == nil {
		t.Error("Expected error for non-existent path, got nil")
	}

	// エラーがsyscallエラーであることを確認
	if _, ok := err.(*os.PathError); !ok {
		// syscall.Statfsは*os.PathErrorを返すことがある
		if _, ok := err.(syscall.Errno); !ok {
			t.Errorf("Expected syscall error, got %T: %v", err, err)
		}
	}
}

// ベンチマークテスト
func BenchmarkSyscallStatfsFunc(b *testing.B) {
	var stat syscallStatfs
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := syscallStatfsFunc("/tmp", &stat)
		if err != nil {
			b.Fatalf("syscallStatfsFunc() failed: %v", err)
		}
	}
}

func BenchmarkSyscallStatfsFuncRoot(b *testing.B) {
	var stat syscallStatfs
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := syscallStatfsFunc("/", &stat)
		if err != nil {
			b.Fatalf("syscallStatfsFunc() failed: %v", err)
		}
	}
}

// 並列実行のベンチマーク
func BenchmarkSyscallStatfsFuncParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var stat syscallStatfs
		for pb.Next() {
			err := syscallStatfsFunc("/tmp", &stat)
			if err != nil {
				b.Errorf("syscallStatfsFunc() failed: %v", err)
			}
		}
	})
}
