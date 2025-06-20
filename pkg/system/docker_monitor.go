//go:build linux
// +build linux

package system

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DockerAwareMonitor はDocker環境でホストシステム情報を取得するモニター
type DockerAwareMonitor struct {
	hostProcPath string
	hostSysPath  string
	isInDocker   bool
}

// NewDockerAwareMonitor は新しいDockerAwareMonitorを作成
func NewDockerAwareMonitor() *DockerAwareMonitor {
	m := &DockerAwareMonitor{
		hostProcPath: "/host/proc",
		hostSysPath:  "/host/sys",
	}

	// Docker環境かどうかを判定
	m.isInDocker = m.checkIfInDocker()

	return m
}

// IsInDocker はDocker環境で実行されているかを返す
func (m *DockerAwareMonitor) IsInDocker() bool {
	return m.isInDocker
}

// checkIfInDocker はDocker環境で実行されているかを判定
func (m *DockerAwareMonitor) checkIfInDocker() bool {
	// /.dockerenvファイルの存在確認
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// /proc/self/cgroupにdockerが含まれているか確認
	data, err := os.ReadFile("/proc/self/cgroup")
	if err == nil && strings.Contains(string(data), "docker") {
		return true
	}

	return false
}

// GetSystemInfo はシステム情報を取得
func (m *DockerAwareMonitor) GetSystemInfo() (*SystemInfo, error) {
	if m.isInDocker {
		return m.getHostSystemInfo()
	}

	// Docker外の場合は通常のモニターを使用
	defaultMonitor := NewDefaultMonitor()
	return defaultMonitor.GetSystemInfo()
}

// getHostSystemInfo はDocker内からホストのシステム情報を取得
func (m *DockerAwareMonitor) getHostSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}

	// CPU情報を取得
	cpuUsage, err := m.getHostCPUUsage()
	if err == nil {
		info.CPUUsagePercent = cpuUsage
	}

	// メモリ情報を取得
	memInfo, err := m.getHostMemoryInfo()
	if err == nil {
		info.MemoryTotalGB = memInfo.totalGB
		info.MemoryUsedGB = memInfo.usedGB
		info.MemoryUsedPercent = (memInfo.usedGB / memInfo.totalGB) * 100
	}

	// ディスク情報を取得（ホストのルートファイルシステム）
	diskInfo, err := m.getHostDiskInfo()
	if err == nil {
		info.DiskTotalGB = diskInfo.totalGB
		info.DiskFreeGB = diskInfo.freeGB
		info.DiskUsedPercent = ((diskInfo.totalGB - diskInfo.freeGB) / diskInfo.totalGB) * 100
	}

	return info, nil
}

// getHostCPUUsage はホストのCPU使用率を取得
func (m *DockerAwareMonitor) getHostCPUUsage() (float64, error) {
	// 最初のサンプリング
	stat1, err := m.readCPUStat()
	if err != nil {
		return 0, err
	}

	// 1秒待機
	time.Sleep(1 * time.Second)

	// 2回目のサンプリング
	stat2, err := m.readCPUStat()
	if err != nil {
		return 0, err
	}

	// CPU使用率を計算
	totalDiff := stat2.total - stat1.total
	idleDiff := stat2.idle - stat1.idle

	if totalDiff == 0 {
		return 0, nil
	}

	usage := 100.0 * (1.0 - float64(idleDiff)/float64(totalDiff))
	return usage, nil
}

type cpuStat struct {
	total uint64
	idle  uint64
}

// readCPUStat は/proc/statからCPU統計を読み取る
func (m *DockerAwareMonitor) readCPUStat() (*cpuStat, error) {
	statPath := filepath.Join(m.hostProcPath, "stat")
	file, err := os.Open(statPath)
	if err != nil {
		// ホストのprocがマウントされていない場合は通常のprocを試す
		file, err = os.Open("/proc/stat")
		if err != nil {
			return nil, err
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 5 {
				return nil, fmt.Errorf("invalid cpu stat format")
			}

			var total uint64
			for i := 1; i < len(fields); i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					continue
				}
				total += val
			}

			idle, _ := strconv.ParseUint(fields[4], 10, 64)

			return &cpuStat{
				total: total,
				idle:  idle,
			}, nil
		}
	}

	return nil, fmt.Errorf("cpu stat not found")
}

type memoryInfo struct {
	totalGB float64
	usedGB  float64
}

// getHostMemoryInfo はホストのメモリ情報を取得
func (m *DockerAwareMonitor) getHostMemoryInfo() (*memoryInfo, error) {
	meminfoPath := filepath.Join(m.hostProcPath, "meminfo")
	file, err := os.Open(meminfoPath)
	if err != nil {
		// ホストのprocがマウントされていない場合は通常のprocを試す
		file, err = os.Open("/proc/meminfo")
		if err != nil {
			return nil, err
		}
	}
	defer file.Close()

	var totalKB, freeKB, buffersKB, cachedKB uint64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			totalKB, _ = strconv.ParseUint(fields[1], 10, 64)
		case "MemFree:":
			freeKB, _ = strconv.ParseUint(fields[1], 10, 64)
		case "Buffers:":
			buffersKB, _ = strconv.ParseUint(fields[1], 10, 64)
		case "Cached:":
			cachedKB, _ = strconv.ParseUint(fields[1], 10, 64)
		}
	}

	totalGB := float64(totalKB) / 1024 / 1024
	availableKB := freeKB + buffersKB + cachedKB
	usedKB := totalKB - availableKB
	usedGB := float64(usedKB) / 1024 / 1024

	return &memoryInfo{
		totalGB: totalGB,
		usedGB:  usedGB,
	}, nil
}

type diskInfo struct {
	totalGB float64
	freeGB  float64
}

// getHostDiskInfo はホストのディスク情報を取得
func (m *DockerAwareMonitor) getHostDiskInfo() (*diskInfo, error) {
	// dfコマンドを使用してホストのディスク情報を取得
	// Docker内では直接ホストのディスク情報にアクセスできないため、
	// マウントされたホストファイルシステムの情報を使用

	// まずは/hostディレクトリの情報を試す
	if _, err := os.Stat("/host"); err == nil {
		// /hostディレクトリが存在する場合
		return m.getDiskUsage("/host")
	}

	// 通常のルートファイルシステムの情報を返す
	return m.getDiskUsage("/")
}

// getDiskUsage は指定されたパスのディスク使用状況を取得
func (m *DockerAwareMonitor) getDiskUsage(path string) (*diskInfo, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	// syscallを使用してファイルシステムの統計情報を取得
	// この実装はLinux固有
	var statfs syscallStatfs
	if err := syscallStatfsFunc(path, &statfs); err != nil {
		return nil, err
	}

	// ブロックサイズとブロック数から容量を計算
	blockSize := uint64(statfs.Bsize)
	totalBlocks := statfs.Blocks
	freeBlocks := statfs.Bavail

	totalBytes := blockSize * totalBlocks
	freeBytes := blockSize * freeBlocks

	return &diskInfo{
		totalGB: float64(totalBytes) / 1024 / 1024 / 1024,
		freeGB:  float64(freeBytes) / 1024 / 1024 / 1024,
	}, nil
}
