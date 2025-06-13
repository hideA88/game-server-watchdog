package system

// SystemInfo はシステム情報を表す構造体
type SystemInfo struct {
	CPUUsagePercent   float64
	MemoryUsedGB      float64
	MemoryTotalGB     float64
	MemoryUsedPercent float64
	DiskFreeGB        float64
	DiskTotalGB       float64
	DiskUsedPercent   float64
}

// Monitor はシステム情報を取得するインターフェース
type Monitor interface {
	GetSystemInfo() (*SystemInfo, error)
}
