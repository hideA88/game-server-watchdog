package system

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

// DefaultMonitor はシステム情報を取得するデフォルト実装
type DefaultMonitor struct{}

// NewDefaultMonitor は新しいDefaultMonitorを作成
func NewDefaultMonitor() *DefaultMonitor {
	return &DefaultMonitor{}
}

// GetSystemInfo はシステム情報を取得
func (m *DefaultMonitor) GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}

	// CPU使用率を取得（1秒間のサンプリング）
	cpuPercent, err := cpu.Percent(1*time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		info.CPUUsagePercent = cpuPercent[0]
	}

	// メモリ情報を取得
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		info.MemoryTotalGB = float64(vmStat.Total) / 1024 / 1024 / 1024
		info.MemoryUsedGB = float64(vmStat.Used) / 1024 / 1024 / 1024
		info.MemoryUsedPercent = vmStat.UsedPercent
	}

	// ディスク情報を取得（ルートパーティション）
	diskStat, err := disk.Usage("/")
	if err == nil {
		info.DiskTotalGB = float64(diskStat.Total) / 1024 / 1024 / 1024
		info.DiskFreeGB = float64(diskStat.Free) / 1024 / 1024 / 1024
		info.DiskUsedPercent = diskStat.UsedPercent
	}

	return info, nil
}