package docker

import (
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
)

func TestDefaultComposeService_SetProjectName(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		want        string
	}{
		{
			name:        "プロジェクト名を設定",
			projectName: "test-project",
			want:        "test-project",
		},
		{
			name:        "空のプロジェクト名",
			projectName: "",
			want:        "",
		},
		{
			name:        "特殊文字を含むプロジェクト名",
			projectName: "test_project-123",
			want:        "test_project-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &DefaultComposeService{}
			service.SetProjectName(tt.projectName)
			
			if service.projectName != tt.want {
				t.Errorf("SetProjectName() projectName = %v, want %v", service.projectName, tt.want)
			}
		})
	}
}

func TestDefaultComposeService_getProjectName(t *testing.T) {
	tests := []struct {
		name            string
		service         *DefaultComposeService
		composePath     string
		expectedProject string
	}{
		{
			name: "設定されたプロジェクト名を使用",
			service: &DefaultComposeService{
				projectName: "custom-project",
			},
			composePath:     "/path/to/docker-compose.yml",
			expectedProject: "custom-project",
		},
		{
			name:            "デフォルトプロジェクト名を使用",
			service:         &DefaultComposeService{},
			composePath:     "/home/user/myapp/docker-compose.yml",
			expectedProject: "myapp",
		},
		{
			name:            "ルートディレクトリの場合",
			service:         &DefaultComposeService{},
			composePath:     "/docker-compose.yml",
			expectedProject: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.service.getProjectName(tt.composePath)
			if result != tt.expectedProject {
				t.Errorf("getProjectName() = %v, want %v", result, tt.expectedProject)
			}
		})
	}
}

func TestCalculateCPUPercent(t *testing.T) { //nolint:funlen // テーブル駆動テストのため長い関数を許可
	tests := []struct {
		name     string
		stats    *container.StatsResponse
		expected float64
	}{
		{
			name: "正常なCPU使用率計算",
			stats: &container.StatsResponse{
				CPUStats: container.CPUStats{
					CPUUsage: container.CPUUsage{
						TotalUsage:   2000000000, // 2秒
						PercpuUsage:  []uint64{1000000000, 1000000000}, // 2コア
					},
					SystemUsage: 4000000000, // 4秒（システム全体）
				},
				PreCPUStats: container.CPUStats{
					CPUUsage: container.CPUUsage{
						TotalUsage:   1000000000, // 1秒
						PercpuUsage:  []uint64{500000000, 500000000}, // 2コア
					},
					SystemUsage: 2000000000, // 2秒（システム全体）
				},
			},
			expected: 100.0, // (1000000000 / 2000000000) * 2 * 100 = 100%
		},
		{
			name: "CPU使用率0%",
			stats: &container.StatsResponse{
				CPUStats: container.CPUStats{
					CPUUsage: container.CPUUsage{
						TotalUsage:   1000000000,
						PercpuUsage:  []uint64{500000000, 500000000},
					},
					SystemUsage: 2000000000,
				},
				PreCPUStats: container.CPUStats{
					CPUUsage: container.CPUUsage{
						TotalUsage:   1000000000, // 変化なし
						PercpuUsage:  []uint64{500000000, 500000000},
					},
					SystemUsage: 2000000000, // 変化なし
				},
			},
			expected: 0.0,
		},
		{
			name: "システム使用率が0の場合",
			stats: &container.StatsResponse{
				CPUStats: container.CPUStats{
					CPUUsage: container.CPUUsage{
						TotalUsage:   2000000000,
						PercpuUsage:  []uint64{1000000000, 1000000000},
					},
					SystemUsage: 2000000000,
				},
				PreCPUStats: container.CPUStats{
					CPUUsage: container.CPUUsage{
						TotalUsage:   1000000000,
						PercpuUsage:  []uint64{500000000, 500000000},
					},
					SystemUsage: 2000000000, // 変化なし
				},
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateCPUPercent(tt.stats)
			if result != tt.expected {
				t.Errorf("calculateCPUPercent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "秒単位",
			duration: 30 * time.Second,
			expected: "30s",
		},
		{
			name:     "分単位",
			duration: 5*time.Minute + 30*time.Second,
			expected: "5m30s",
		},
		{
			name:     "時間単位",
			duration: 2*time.Hour + 30*time.Minute,
			expected: "2h30m",
		},
		{
			name:     "日単位",
			duration: 3*24*time.Hour + 5*time.Hour,
			expected: "3d5h",
		},
		{
			name:     "1秒未満",
			duration: 500 * time.Millisecond,
			expected: "0s",
		},
		{
			name:     "ちょうど1分",
			duration: 1 * time.Minute,
			expected: "1m0s",
		},
		{
			name:     "ちょうど1時間",
			duration: 1 * time.Hour,
			expected: "1h0m",
		},
		{
			name:     "ちょうど1日",
			duration: 24 * time.Hour,
			expected: "1d0h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    uint64
		expected string
	}{
		{
			name:     "バイト単位",
			bytes:    512,
			expected: "512B",
		},
		{
			name:     "キロバイト単位",
			bytes:    1536, // 1.5KB
			expected: "1.5KB",
		},
		{
			name:     "メガバイト単位",
			bytes:    2097152, // 2MB
			expected: "2.0MB",
		},
		{
			name:     "ギガバイト単位",
			bytes:    3221225472, // 3GB
			expected: "3.0GB",
		},
		{
			name:     "テラバイト単位",
			bytes:    1099511627776, // 1TB
			expected: "1.0TB",
		},
		{
			name:     "0バイト",
			bytes:    0,
			expected: "0B",
		},
		{
			name:     "1バイト",
			bytes:    1,
			expected: "1B",
		},
		{
			name:     "1023バイト（1KB未満）",
			bytes:    1023,
			expected: "1023B",
		},
		{
			name:     "1024バイト（ちょうど1KB）",
			bytes:    1024,
			expected: "1.0KB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCleanDockerLogs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Dockerヘッダー付きログ",
			input:    "\x01\x00\x00\x00\x00\x00\x00\x0cHello World\n\x01\x00\x00\x00\x00\x00\x00\x09Test log",
			expected: "Hello World\nTest log",
		},
		{
			name:     "ヘッダーなしログ",
			input:    "Simple log line\nAnother line",
			expected: "Simple log line\nAnother line",
		},
		{
			name:     "空のログ",
			input:    "",
			expected: "",
		},
		{
			name:     "短い行（8文字未満）",
			input:    "short\nvery\n",
			expected: "short\nvery",
		},
		{
			name:     "混在ログ",
			input:    "\x01\x00\x00\x00\x00\x00\x00\x05Test\nNormal line\n\x01\x00\x00\x00\x00\x00\x00\x07Another",
			expected: "Test\nNormal line\nAnother",
		},
		{
			name:     "空行を含むログ",
			input:    "\x01\x00\x00\x00\x00\x00\x00\x05Test\n\n\x01\x00\x00\x00\x00\x00\x00\x07Another",
			expected: "Test\nAnother",
		},
		{
			name:     "改行のみ",
			input:    "\n\n\n",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanDockerLogs(tt.input)
			if result != tt.expected {
				t.Errorf("cleanDockerLogs() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// DefaultComposeServiceのメソッドのテスト（calculateMemoryPercentなど）
func TestDefaultComposeService_calculateMemoryPercent(t *testing.T) {
	tests := []struct {
		name     string
		stats    *container.StatsResponse
		expected float64
	}{
		{
			name: "正常なメモリ使用率",
			stats: &container.StatsResponse{
				MemoryStats: container.MemoryStats{
					Usage: 536870912,  // 512MB
					Limit: 1073741824, // 1GB
				},
			},
			expected: 50.0, // 50%
		},
		{
			name: "100%使用",
			stats: &container.StatsResponse{
				MemoryStats: container.MemoryStats{
					Usage: 1073741824, // 1GB
					Limit: 1073741824, // 1GB
				},
			},
			expected: 100.0,
		},
		{
			name: "0%使用",
			stats: &container.StatsResponse{
				MemoryStats: container.MemoryStats{
					Usage: 0,
					Limit: 1073741824, // 1GB
				},
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &DefaultComposeService{}
			result := service.calculateMemoryPercent(tt.stats)
			if result != tt.expected {
				t.Errorf("calculateMemoryPercent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDefaultComposeService_formatMemoryUsage(t *testing.T) {
	tests := []struct {
		name     string
		stats    *container.StatsResponse
		expected string
	}{
		{
			name: "正常なメモリ使用量",
			stats: &container.StatsResponse{
				MemoryStats: container.MemoryStats{
					Usage: 536870912,  // 512MB = 0.5GiB
					Limit: 1073741824, // 1GB = 1GiB
				},
			},
			expected: "0.50GiB / 1.00GiB",
		},
		{
			name: "大きなメモリ使用量",
			stats: &container.StatsResponse{
				MemoryStats: container.MemoryStats{
					Usage: 8589934592,  // 8GB
					Limit: 17179869184, // 16GB
				},
			},
			expected: "8.00GiB / 16.00GiB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &DefaultComposeService{}
			result := service.formatMemoryUsage(tt.stats)
			if result != tt.expected {
				t.Errorf("formatMemoryUsage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDefaultComposeService_calculateNetworkIO(t *testing.T) {
	tests := []struct {
		name     string
		stats    *container.StatsResponse
		expected string
	}{
		{
			name: "正常なネットワークI/O",
			stats: &container.StatsResponse{
				Networks: map[string]container.NetworkStats{
					"eth0": {
						RxBytes: 1048576, // 1MB受信
						TxBytes: 2097152, // 2MB送信
					},
					"eth1": {
						RxBytes: 512000,  // 500KB受信
						TxBytes: 1024000, // 1000KB送信
					},
				},
			},
			expected: "1.5MB / 3.0MB", // 合計値
		},
		{
			name: "ネットワークなし",
			stats: &container.StatsResponse{
				Networks: map[string]container.NetworkStats{},
			},
			expected: "0B / 0B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &DefaultComposeService{}
			result := service.calculateNetworkIO(tt.stats)
			if result != tt.expected {
				t.Errorf("calculateNetworkIO() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDefaultComposeService_calculateBlockIO(t *testing.T) {
	tests := []struct {
		name     string
		stats    *container.StatsResponse
		expected string
	}{
		{
			name: "正常なブロックI/O",
			stats: &container.StatsResponse{
				BlkioStats: container.BlkioStats{
					IoServiceBytesRecursive: []container.BlkioStatEntry{
						{Op: "read", Value: 1048576},  // 1MB読み取り
						{Op: "write", Value: 2097152}, // 2MB書き込み
						{Op: "read", Value: 512000},   // 500KB読み取り（追加）
						{Op: "write", Value: 1024000}, // 1MB書き込み（追加）
						{Op: "other", Value: 1000000}, // 他の操作（無視される）
					},
				},
			},
			expected: "1.5MB / 3.0MB", // 合計値
		},
		{
			name: "ブロックI/Oなし",
			stats: &container.StatsResponse{
				BlkioStats: container.BlkioStats{
					IoServiceBytesRecursive: []container.BlkioStatEntry{},
				},
			},
			expected: "0B / 0B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &DefaultComposeService{}
			result := service.calculateBlockIO(tt.stats)
			if result != tt.expected {
				t.Errorf("calculateBlockIO() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// ベンチマークテスト
func BenchmarkCalculateCPUPercent(b *testing.B) {
	stats := &container.StatsResponse{
		CPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{
				TotalUsage:  2000000000,
				PercpuUsage: []uint64{1000000000, 1000000000},
			},
			SystemUsage: 4000000000,
		},
		PreCPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{
				TotalUsage:  1000000000,
				PercpuUsage: []uint64{500000000, 500000000},
			},
			SystemUsage: 2000000000,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateCPUPercent(stats)
	}
}

func BenchmarkFormatBytes(b *testing.B) {
	bytes := uint64(1073741824) // 1GB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatBytes(bytes)
	}
}

func BenchmarkCleanDockerLogs(b *testing.B) {
	logs := "\x01\x00\x00\x00\x00\x00\x00\x0cHello World\n\x01\x00\x00\x00\x00\x00\x00\x09Test log\n"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cleanDockerLogs(logs)
	}
}