package system

import (
	"testing"
)


func TestDefaultMonitor_GetSystemInfo(t *testing.T) {
	t.Parallel()

	monitor := NewDefaultMonitor()
	info, err := monitor.GetSystemInfo()

	if err != nil {
		t.Fatalf("GetSystemInfo() error = %v", err)
	}

	if info == nil {
		t.Fatal("GetSystemInfo() returned nil")
	}

	// 値の妥当性をチェック（0以上100以下）
	if info.CPUUsagePercent < 0 || info.CPUUsagePercent > 100 {
		t.Errorf("CPUUsagePercent = %v, want 0-100", info.CPUUsagePercent)
	}

	if info.MemoryUsedPercent < 0 || info.MemoryUsedPercent > 100 {
		t.Errorf("MemoryUsedPercent = %v, want 0-100", info.MemoryUsedPercent)
	}

	if info.DiskUsedPercent < 0 || info.DiskUsedPercent > 100 {
		t.Errorf("DiskUsedPercent = %v, want 0-100", info.DiskUsedPercent)
	}

	// メモリとディスクのサイズが正の値であることを確認
	if info.MemoryTotalGB <= 0 {
		t.Errorf("MemoryTotalGB = %v, want > 0", info.MemoryTotalGB)
	}

	if info.DiskTotalGB <= 0 {
		t.Errorf("DiskTotalGB = %v, want > 0", info.DiskTotalGB)
	}
}