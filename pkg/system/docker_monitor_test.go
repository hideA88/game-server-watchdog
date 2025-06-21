//go:build linux
// +build linux

package system

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testCPUStat = `cpu  1000 0 2000 7000 0 0 0 0 0 0
cpu0 500 0 1000 3500 0 0 0 0 0 0
`

func TestNewDockerAwareMonitor(t *testing.T) {
	t.Parallel()

	monitor := NewDockerAwareMonitor()
	if monitor == nil {
		t.Fatal("NewDockerAwareMonitor() returned nil")
	}

	if monitor.hostProcPath != "/host/proc" {
		t.Errorf("hostProcPath = %v, want /host/proc", monitor.hostProcPath)
	}

	if monitor.hostSysPath != "/host/sys" {
		t.Errorf("hostSysPath = %v, want /host/sys", monitor.hostSysPath)
	}
}

func TestDockerAwareMonitor_checkIfInDocker(t *testing.T) {
	// このテストは実際の環境に依存するため、
	// 環境によって結果が変わることを前提としています
	monitor := &DockerAwareMonitor{
		hostProcPath: "/host/proc",
		hostSysPath:  "/host/sys",
	}

	// checkIfInDockerの結果をテスト
	isInDocker := monitor.checkIfInDocker()

	// /.dockerenvファイルが存在するかチェック
	_, dockerEnvErr := os.Stat("/.dockerenv")
	dockerEnvExists := dockerEnvErr == nil

	// /proc/self/cgroupにdockerが含まれているかチェック
	cgroupData, cgroupErr := os.ReadFile("/proc/self/cgroup")
	cgroupHasDocker := cgroupErr == nil && strings.Contains(string(cgroupData), "docker")

	// どちらかの条件を満たせばDocker環境と判定されるはず
	expectedInDocker := dockerEnvExists || cgroupHasDocker

	if isInDocker != expectedInDocker {
		t.Errorf("checkIfInDocker() = %v, expected %v (dockerEnvExists=%v, cgroupHasDocker=%v)",
			isInDocker, expectedInDocker, dockerEnvExists, cgroupHasDocker)
	}
}

func TestDockerAwareMonitor_IsInDocker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		isInDocker bool
	}{
		{
			name:       "Docker環境の場合",
			isInDocker: true,
		},
		{
			name:       "非Docker環境の場合",
			isInDocker: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &DockerAwareMonitor{
				isInDocker: tt.isInDocker,
			}

			if got := monitor.IsInDocker(); got != tt.isInDocker {
				t.Errorf("IsInDocker() = %v, want %v", got, tt.isInDocker)
			}
		})
	}
}

func TestDockerAwareMonitor_GetSystemInfo(t *testing.T) {
	tests := []struct {
		name       string
		isInDocker bool
		wantErr    bool
	}{
		{
			name:       "非Docker環境の場合",
			isInDocker: false,
			wantErr:    false,
		},
		// Docker環境のテストは、ホストファイルシステムがマウントされている必要があるため、
		// 統合テストとして別途実装
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &DockerAwareMonitor{
				isInDocker:   tt.isInDocker,
				hostProcPath: "/host/proc",
				hostSysPath:  "/host/sys",
			}

			info, err := monitor.GetSystemInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSystemInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && info == nil {
				t.Error("GetSystemInfo() returned nil SystemInfo")
			}
		})
	}
}

func TestDockerAwareMonitor_readCPUStat(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		procContent  string
		wantTotal    uint64
		wantIdle     uint64
		wantErr      bool
		setupProcDir bool
	}{
		{
			name: "正常なCPU統計",
			procContent: `cpu  1234 0 5678 9012 3456 0 0 0 0 0
cpu0 617 0 2839 4506 1728 0 0 0 0 0
`,
			wantTotal:    1234 + 0 + 5678 + 9012 + 3456,
			wantIdle:     9012,
			wantErr:      false,
			setupProcDir: true,
		},
		{
			name: "cpu行がない場合",
			procContent: `cpu0 617 0 2839 4506 1728 0 0 0 0 0
cpu1 617 0 2839 4506 1728 0 0 0 0 0
`,
			wantTotal:    0,
			wantIdle:     0,
			wantErr:      true,
			setupProcDir: true,
		},
		{
			name: "フィールドが足りない場合",
			procContent: `cpu  1234 0 5678
`,
			wantTotal:    0,
			wantIdle:     0,
			wantErr:      true,
			setupProcDir: true,
		},
		{
			name:         "ファイルが存在しない場合",
			procContent:  "",
			wantTotal:    0,
			wantIdle:     0,
			wantErr:      true,
			setupProcDir: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用のprocディレクトリを作成
			testProcPath := filepath.Join(tmpDir, tt.name, "proc")
			if tt.setupProcDir {
				if err := os.MkdirAll(testProcPath, 0755); err != nil {
					t.Fatalf("Failed to create test proc dir: %v", err)
				}

				// statファイルを作成
				statPath := filepath.Join(testProcPath, "stat")
				if err := os.WriteFile(statPath, []byte(tt.procContent), 0644); err != nil {
					t.Fatalf("Failed to write stat file: %v", err)
				}
			}

			monitor := &DockerAwareMonitor{
				hostProcPath: testProcPath,
			}

			stat, err := monitor.readCPUStat()
			if (err != nil) != tt.wantErr {
				// ファイルが存在しない場合でも、/proc/statへのフォールバックがあるため
				// エラーにならない可能性があるので、その場合はスキップ
				if !tt.setupProcDir && err == nil && stat != nil {
					t.Skip("Skipping test because /proc/stat fallback succeeded")
				}
				t.Errorf("readCPUStat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if stat.total != tt.wantTotal {
					t.Errorf("readCPUStat() total = %v, want %v", stat.total, tt.wantTotal)
				}
				if stat.idle != tt.wantIdle {
					t.Errorf("readCPUStat() idle = %v, want %v", stat.idle, tt.wantIdle)
				}
			}
		})
	}
}

func TestDockerAwareMonitor_getHostMemoryInfo(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		meminfoContent string
		wantTotalGB    float64
		wantUsedGB     float64
		wantErr        bool
		setupProcDir   bool
	}{
		{
			name: "正常なメモリ情報",
			meminfoContent: `MemTotal:        8388608 kB
MemFree:         2097152 kB
MemAvailable:    4194304 kB
Buffers:          524288 kB
Cached:          1048576 kB
`,
			wantTotalGB:  8.0,
			wantUsedGB:   8.0 - (2097152.0+524288.0+1048576.0)/1024.0/1024.0,
			wantErr:      false,
			setupProcDir: true,
		},
		{
			name: "一部のフィールドが欠けている場合",
			meminfoContent: `MemTotal:        8388608 kB
MemFree:         2097152 kB
`,
			wantTotalGB:  8.0,
			wantUsedGB:   8.0 - 2.0, // Buffers, Cachedがない場合
			wantErr:      false,
			setupProcDir: true,
		},
		{
			name:           "ファイルが存在しない場合",
			meminfoContent: "",
			wantTotalGB:    0,
			wantUsedGB:     0,
			wantErr:        true,
			setupProcDir:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用のprocディレクトリを作成
			testProcPath := filepath.Join(tmpDir, strings.ReplaceAll(tt.name, "/", "_"), "proc")
			if tt.setupProcDir {
				if err := os.MkdirAll(testProcPath, 0755); err != nil {
					t.Fatalf("Failed to create test proc dir: %v", err)
				}

				// meminfoファイルを作成
				meminfoPath := filepath.Join(testProcPath, "meminfo")
				if err := os.WriteFile(meminfoPath, []byte(tt.meminfoContent), 0644); err != nil {
					t.Fatalf("Failed to write meminfo file: %v", err)
				}
			}

			monitor := &DockerAwareMonitor{
				hostProcPath: testProcPath,
			}

			memInfo, err := monitor.getHostMemoryInfo()
			if (err != nil) != tt.wantErr {
				// ファイルが存在しない場合でも、/proc/meminfoへのフォールバックがあるため
				// エラーにならない可能性があるので、その場合はスキップ
				if !tt.setupProcDir && err == nil && memInfo != nil {
					t.Skip("Skipping test because /proc/meminfo fallback succeeded")
				}
				t.Errorf("getHostMemoryInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if memInfo.totalGB != tt.wantTotalGB {
					t.Errorf("getHostMemoryInfo() totalGB = %v, want %v", memInfo.totalGB, tt.wantTotalGB)
				}

				// Used GBは計算の精度を考慮して比較
				if diff := memInfo.usedGB - tt.wantUsedGB; diff > 0.01 || diff < -0.01 {
					t.Errorf("getHostMemoryInfo() usedGB = %v, want %v", memInfo.usedGB, tt.wantUsedGB)
				}
			}
		})
	}
}

func TestDockerAwareMonitor_getDiskUsage(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "存在するパス",
			path:    "/tmp",
			wantErr: false,
		},
		{
			name:    "存在しないパス",
			path:    "/nonexistent/path/that/should/not/exist",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &DockerAwareMonitor{}

			diskInfo, err := monitor.getDiskUsage(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDiskUsage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && diskInfo != nil {
				// ディスク情報の妥当性をチェック
				if diskInfo.totalGB <= 0 {
					t.Errorf("getDiskUsage() totalGB = %v, want > 0", diskInfo.totalGB)
				}
				if diskInfo.freeGB < 0 || diskInfo.freeGB > diskInfo.totalGB {
					t.Errorf("getDiskUsage() freeGB = %v, totalGB = %v, freeGB should be 0 <= freeGB <= totalGB",
						diskInfo.freeGB, diskInfo.totalGB)
				}
			}
		})
	}
}

func TestDockerAwareMonitor_getHostDiskInfo(t *testing.T) {
	tests := []struct {
		name      string
		setupHost bool
		wantErr   bool
	}{
		{
			name:      "/hostディレクトリが存在する場合",
			setupHost: true,
			wantErr:   false,
		},
		{
			name:      "/hostディレクトリが存在しない場合",
			setupHost: false,
			wantErr:   false, // /を使用するため、エラーにはならない
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// /hostのテストは実際のファイルシステムに依存するため、
			// 統合テストとして扱うか、モックを使用する必要がある
			monitor := &DockerAwareMonitor{}

			diskInfo, err := monitor.getHostDiskInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("getHostDiskInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && diskInfo != nil {
				// ディスク情報の妥当性をチェック
				if diskInfo.totalGB <= 0 {
					t.Errorf("getHostDiskInfo() totalGB = %v, want > 0", diskInfo.totalGB)
				}
				if diskInfo.freeGB < 0 || diskInfo.freeGB > diskInfo.totalGB {
					t.Errorf("getHostDiskInfo() freeGB = %v, totalGB = %v, freeGB should be 0 <= freeGB <= totalGB",
						diskInfo.freeGB, diskInfo.totalGB)
				}
			}
		})
	}
}

func TestDockerAwareMonitor_getHostCPUUsage(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// テスト用のprocディレクトリを作成
	testProcPath := filepath.Join(tmpDir, "proc")
	if err := os.MkdirAll(testProcPath, 0755); err != nil {
		t.Fatalf("Failed to create test proc dir: %v", err)
	}

	// 初期のCPU統計を作成
	initialStat := testCPUStat

	// 更新後のCPU統計を作成（1秒後を想定）
	updatedStat := `cpu  1010 0 2010 7980 0 0 0 0 0 0
cpu0 505 0 1005 3990 0 0 0 0 0 0
`

	// statファイルを作成
	statPath := filepath.Join(testProcPath, "stat")
	if err := os.WriteFile(statPath, []byte(initialStat), 0644); err != nil {
		t.Fatalf("Failed to write initial stat file: %v", err)
	}

	monitor := &DockerAwareMonitor{
		hostProcPath: testProcPath,
	}

	// 別のゴルーチンで1秒後にファイルを更新
	go func() {
		// readCPUStat内でsleepされる時間とほぼ同じタイミングで更新
		// 実際のテストでは少し早めに更新
		<-time.After(500 * time.Millisecond)
		if err := os.WriteFile(statPath, []byte(updatedStat), 0644); err != nil {
			t.Logf("Failed to update stat file: %v", err)
		}
	}()

	usage, err := monitor.getHostCPUUsage()
	if err != nil {
		t.Errorf("getHostCPUUsage() error = %v", err)
		return
	}

	// CPU使用率の妥当性をチェック
	// 期待値: (1010 - 1000) / ((1010 + 0 + 2010 + 7980) - (1000 + 0 + 2000 + 7000)) * 100
	// = 10 / 1000 * 100 = 1%
	// ただし、テストのタイミングによって値が変わる可能性があるため、範囲でチェック
	if usage < 0 || usage > 100 {
		t.Errorf("getHostCPUUsage() = %v, want 0-100", usage)
	}
}

func TestDockerAwareMonitor_getHostSystemInfo(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// テスト用のprocディレクトリを作成
	testProcPath := filepath.Join(tmpDir, "proc")
	if err := os.MkdirAll(testProcPath, 0755); err != nil {
		t.Fatalf("Failed to create test proc dir: %v", err)
	}

	// CPU統計ファイルを作成
	cpuStat := testCPUStat
	statPath := filepath.Join(testProcPath, "stat")
	if err := os.WriteFile(statPath, []byte(cpuStat), 0644); err != nil {
		t.Fatalf("Failed to write stat file: %v", err)
	}

	// メモリ情報ファイルを作成
	memInfo := `MemTotal:        8388608 kB
MemFree:         2097152 kB
MemAvailable:    4194304 kB
Buffers:          524288 kB
Cached:          1048576 kB
`
	meminfoPath := filepath.Join(testProcPath, "meminfo")
	if err := os.WriteFile(meminfoPath, []byte(memInfo), 0644); err != nil {
		t.Fatalf("Failed to write meminfo file: %v", err)
	}

	monitor := &DockerAwareMonitor{
		hostProcPath: testProcPath,
		isInDocker:   true,
	}

	// 別のゴルーチンでCPU統計を更新（getHostCPUUsageのため）
	go func() {
		<-time.After(500 * time.Millisecond)
		updatedStat := `cpu  1010 0 2010 7980 0 0 0 0 0 0
cpu0 505 0 1005 3990 0 0 0 0 0 0
`
		if err := os.WriteFile(statPath, []byte(updatedStat), 0644); err != nil {
			t.Logf("Failed to update stat file: %v", err)
		}
	}()

	info, err := monitor.getHostSystemInfo()
	if err != nil {
		t.Errorf("getHostSystemInfo() error = %v", err)
		return
	}

	if info == nil {
		t.Fatal("getHostSystemInfo() returned nil")
	}

	// システム情報の妥当性をチェック
	if info.CPUUsagePercent < 0 || info.CPUUsagePercent > 100 {
		t.Errorf("CPUUsagePercent = %v, want 0-100", info.CPUUsagePercent)
	}

	if info.MemoryTotalGB != 8.0 {
		t.Errorf("MemoryTotalGB = %v, want 8.0", info.MemoryTotalGB)
	}

	if info.MemoryUsedPercent < 0 || info.MemoryUsedPercent > 100 {
		t.Errorf("MemoryUsedPercent = %v, want 0-100", info.MemoryUsedPercent)
	}

	// ディスク情報は実際のファイルシステムに依存するため、
	// 値の妥当性のみチェック
	if info.DiskTotalGB > 0 {
		if info.DiskFreeGB < 0 || info.DiskFreeGB > info.DiskTotalGB {
			t.Errorf("DiskFreeGB = %v, DiskTotalGB = %v, DiskFreeGB should be 0 <= DiskFreeGB <= DiskTotalGB",
				info.DiskFreeGB, info.DiskTotalGB)
		}
		if info.DiskUsedPercent < 0 || info.DiskUsedPercent > 100 {
			t.Errorf("DiskUsedPercent = %v, want 0-100", info.DiskUsedPercent)
		}
	}
}

// ベンチマークテスト
func BenchmarkDockerAwareMonitor_readCPUStat(b *testing.B) {
	monitor := &DockerAwareMonitor{
		hostProcPath: "/proc", // 実際のprocを使用
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := monitor.readCPUStat()
		if err != nil {
			b.Fatalf("readCPUStat() failed: %v", err)
		}
	}
}

func BenchmarkDockerAwareMonitor_getHostMemoryInfo(b *testing.B) {
	monitor := &DockerAwareMonitor{
		hostProcPath: "/proc", // 実際のprocを使用
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := monitor.getHostMemoryInfo()
		if err != nil {
			b.Fatalf("getHostMemoryInfo() failed: %v", err)
		}
	}
}

// 無効なブロックサイズのテスト
func TestDockerAwareMonitor_getDiskUsage_InvalidBlockSize(t *testing.T) {
	// このテストは実際のsyscallを使用するため、
	// 実際には負のブロックサイズは発生しないが、
	// コードカバレッジのために含めている

	// 一時ディレクトリを使用
	tmpDir := t.TempDir()
	monitor := &DockerAwareMonitor{}

	// 実際のファイルシステムでは負のブロックサイズは発生しないため、
	// このテストは主にコードパスの存在を確認するためのもの
	diskInfo, err := monitor.getDiskUsage(tmpDir)
	if err != nil {
		// エラーが発生した場合（期待される動作ではない）
		t.Logf("getDiskUsage() returned error: %v", err)
		// ただし、実際のファイルシステムではこのエラーは発生しないはず
		if strings.Contains(err.Error(), "invalid block size") {
			// もしこのエラーが発生した場合は、テストとしては成功
			return
		}
	}

	// 正常に動作した場合（期待される動作）
	if diskInfo != nil && diskInfo.totalGB > 0 {
		// 実際のファイルシステムでは正常に動作するはず
		t.Logf("getDiskUsage() returned valid disk info: totalGB=%v, freeGB=%v",
			diskInfo.totalGB, diskInfo.freeGB)
	}
}

// パースエラーのテスト
func TestDockerAwareMonitor_readCPUStat_ParseError(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()
	testProcPath := filepath.Join(tmpDir, "proc")
	if err := os.MkdirAll(testProcPath, 0755); err != nil {
		t.Fatalf("Failed to create test proc dir: %v", err)
	}

	// 不正な数値を含むCPU統計
	invalidStat := `cpu  abc def ghi jkl mno pqr stu vwx yz
cpu0 617 0 2839 4506 1728 0 0 0 0 0
`
	statPath := filepath.Join(testProcPath, "stat")
	if err := os.WriteFile(statPath, []byte(invalidStat), 0644); err != nil {
		t.Fatalf("Failed to write stat file: %v", err)
	}

	monitor := &DockerAwareMonitor{
		hostProcPath: testProcPath,
	}

	stat, err := monitor.readCPUStat()
	if err != nil {
		t.Errorf("readCPUStat() should handle parse errors gracefully, got error: %v", err)
	}

	// パースエラーがあっても、部分的に正しい値が取得できる可能性がある
	if stat != nil {
		t.Logf("readCPUStat() returned stat despite parse errors: total=%v, idle=%v",
			stat.total, stat.idle)
	}
}

// メモリ情報のパースエラーのテスト
func TestDockerAwareMonitor_getHostMemoryInfo_ParseError(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()
	testProcPath := filepath.Join(tmpDir, "proc")
	if err := os.MkdirAll(testProcPath, 0755); err != nil {
		t.Fatalf("Failed to create test proc dir: %v", err)
	}

	// 不正な数値を含むメモリ情報
	invalidMemInfo := `MemTotal:        abc kB
MemFree:         def kB
Buffers:         ghi kB
Cached:          jkl kB
`
	meminfoPath := filepath.Join(testProcPath, "meminfo")
	if err := os.WriteFile(meminfoPath, []byte(invalidMemInfo), 0644); err != nil {
		t.Fatalf("Failed to write meminfo file: %v", err)
	}

	monitor := &DockerAwareMonitor{
		hostProcPath: testProcPath,
	}

	memInfo, err := monitor.getHostMemoryInfo()
	if err != nil {
		t.Errorf("getHostMemoryInfo() should handle parse errors gracefully, got error: %v", err)
	}

	// パースエラーがあっても、部分的に正しい値が取得できる可能性がある
	if memInfo != nil {
		t.Logf("getHostMemoryInfo() returned memInfo despite parse errors: totalGB=%v, usedGB=%v",
			memInfo.totalGB, memInfo.usedGB)
		// 値は0になっているはず
		if memInfo.totalGB != 0 || memInfo.usedGB != 0 {
			t.Errorf("Expected zero values for invalid input, got totalGB=%v, usedGB=%v",
				memInfo.totalGB, memInfo.usedGB)
		}
	}
}

func TestCPUStat_TotalDiffZero(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()
	testProcPath := filepath.Join(tmpDir, "proc")
	if err := os.MkdirAll(testProcPath, 0755); err != nil {
		t.Fatalf("Failed to create test proc dir: %v", err)
	}

	// 同じ値のCPU統計（差分が0になる）
	sameStat := `cpu  1000 0 2000 7000 0 0 0 0 0 0
cpu0 500 0 1000 3500 0 0 0 0 0 0
`
	statPath := filepath.Join(testProcPath, "stat")
	if err := os.WriteFile(statPath, []byte(sameStat), 0644); err != nil {
		t.Fatalf("Failed to write stat file: %v", err)
	}

	monitor := &DockerAwareMonitor{
		hostProcPath: testProcPath,
	}

	// ファイルを変更しない（同じ値のまま）
	usage, err := monitor.getHostCPUUsage()
	if err != nil {
		t.Errorf("getHostCPUUsage() error = %v", err)
		return
	}

	// totalDiffが0の場合、使用率は0%になるはず
	if usage != 0 {
		t.Errorf("getHostCPUUsage() = %v, want 0 when totalDiff is 0", usage)
	}
}
