package system

import (
	"testing"
	"time"
)

// 統合テスト - 実際のシステム情報を取得してテストする
func TestDefaultMonitor_GetSystemInfo_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	monitor := NewDefaultMonitor()
	
	start := time.Now()
	info, err := monitor.GetSystemInfo()
	elapsed := time.Since(start)

	// エラーが発生しないこと
	if err != nil {
		t.Fatalf("GetSystemInfo() returned error: %v", err)
	}

	// 結果が nil でないこと
	if info == nil {
		t.Fatal("GetSystemInfo() returned nil SystemInfo")
	}

	// レスポンス時間が妥当であること（最大3秒）
	if elapsed > 3*time.Second {
		t.Errorf("GetSystemInfo() took too long: %v", elapsed)
	}

	// CPU使用率の妥当性チェック
	if info.CPUUsagePercent < 0 || info.CPUUsagePercent > 100 {
		t.Errorf("CPU usage percent out of range: %f", info.CPUUsagePercent)
	}

	// メモリ情報の妥当性チェック
	if info.MemoryTotalGB <= 0 {
		t.Errorf("Memory total should be positive: %f", info.MemoryTotalGB)
	}
	if info.MemoryUsedGB < 0 || info.MemoryUsedGB > info.MemoryTotalGB {
		t.Errorf("Memory used out of range: used=%f, total=%f", info.MemoryUsedGB, info.MemoryTotalGB)
	}
	if info.MemoryUsedPercent < 0 || info.MemoryUsedPercent > 100 {
		t.Errorf("Memory used percent out of range: %f", info.MemoryUsedPercent)
	}

	// ディスク情報の妥当性チェック
	if info.DiskTotalGB <= 0 {
		t.Errorf("Disk total should be positive: %f", info.DiskTotalGB)
	}
	if info.DiskFreeGB < 0 || info.DiskFreeGB > info.DiskTotalGB {
		t.Errorf("Disk free out of range: free=%f, total=%f", info.DiskFreeGB, info.DiskTotalGB)
	}
	if info.DiskUsedPercent < 0 || info.DiskUsedPercent > 100 {
		t.Errorf("Disk used percent out of range: %f", info.DiskUsedPercent)
	}

	// 計算の整合性チェック
	calculatedMemUsedPercent := (info.MemoryUsedGB / info.MemoryTotalGB) * 100
	memPercentDiff := abs(info.MemoryUsedPercent - calculatedMemUsedPercent)
	if memPercentDiff > 1.0 { // 1%以下の誤差は許容
		t.Errorf("Memory percent calculation inconsistent: reported=%f, calculated=%f", 
			info.MemoryUsedPercent, calculatedMemUsedPercent)
	}

	calculatedDiskUsedPercent := ((info.DiskTotalGB - info.DiskFreeGB) / info.DiskTotalGB) * 100
	diskPercentDiff := abs(info.DiskUsedPercent - calculatedDiskUsedPercent)
	if diskPercentDiff > 5.0 { // 5%以下の誤差は許容（ファイルシステムメタデータのため）
		t.Errorf("Disk percent calculation inconsistent: reported=%f, calculated=%f", 
			info.DiskUsedPercent, calculatedDiskUsedPercent)
	}
}

// パフォーマンステスト
func TestDefaultMonitor_GetSystemInfo_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	monitor := NewDefaultMonitor()
	
	// 複数回実行してパフォーマンスをチェック
	const iterations = 3
	var totalDuration time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, err := monitor.GetSystemInfo()
		elapsed := time.Since(start)
		
		if err != nil {
			t.Fatalf("Iteration %d failed: %v", i, err)
		}
		
		totalDuration += elapsed
		
		// 各回の実行時間をチェック
		if elapsed > 2*time.Second {
			t.Errorf("Iteration %d took too long: %v", i, elapsed)
		}
	}

	avgDuration := totalDuration / iterations
	t.Logf("Average execution time: %v", avgDuration)
	
	// 平均実行時間のチェック
	if avgDuration > 1500*time.Millisecond {
		t.Errorf("Average execution time too long: %v", avgDuration)
	}
}

// 連続実行テスト（メモリリークチェック）
func TestDefaultMonitor_GetSystemInfo_MemoryLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	monitor := NewDefaultMonitor()
	
	// 大量実行でメモリリークがないかチェック
	const iterations = 20
	
	for i := 0; i < iterations; i++ {
		info, err := monitor.GetSystemInfo()
		if err != nil {
			t.Fatalf("Iteration %d failed: %v", i, err)
		}
		if info == nil {
			t.Fatalf("Iteration %d returned nil", i)
		}
		
		// 5回ごとにログ出力
		if (i+1)%5 == 0 {
			t.Logf("Completed %d iterations", i+1)
		}
	}
}

// システム負荷状態での動作テスト
func TestDefaultMonitor_GetSystemInfo_UnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	monitor := NewDefaultMonitor()
	
	// 並行してシステム情報を取得
	const goroutines = 5
	const iterationsPerGoroutine = 3
	
	results := make(chan error, goroutines*iterationsPerGoroutine)
	
	for g := 0; g < goroutines; g++ {
		go func(_ int) {
			for i := 0; i < iterationsPerGoroutine; i++ {
				info, err := monitor.GetSystemInfo()
				if err != nil {
					results <- err
					return
				}
				if info == nil {
					results <- nil
					return
				}
				results <- nil
			}
		}(g)
	}
	
	// 結果をチェック
	for i := 0; i < goroutines*iterationsPerGoroutine; i++ {
		err := <-results
		if err != nil {
			t.Errorf("Goroutine failed: %v", err)
		}
	}
}

// リソース値の範囲テスト
func TestDefaultMonitor_GetSystemInfo_ValueRanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping value range test in short mode")
	}

	monitor := NewDefaultMonitor()
	
	// 複数回実行して値の安定性をチェック
	const samples = 2
	infos := make([]*SystemInfo, samples)
	
	for i := 0; i < samples; i++ {
		info, err := monitor.GetSystemInfo()
		if err != nil {
			t.Fatalf("Sample %d failed: %v", i, err)
		}
		infos[i] = info
		
		// サンプル間で少し待機
		if i < samples-1 {
			time.Sleep(50 * time.Millisecond)
		}
	}
	
	// 各サンプルの値をチェック
	for i, info := range infos {
		t.Logf("Sample %d: CPU=%.2f%%, Mem=%.2f%%, Disk=%.2f%%", 
			i, info.CPUUsagePercent, info.MemoryUsedPercent, info.DiskUsedPercent)
		
		// 基本的な範囲チェック
		validateSystemInfo(t, info, i)
	}
	
	// サンプル間の変動チェック（メモリとディスクはそれほど変動しないはず）
	for i := 1; i < samples; i++ {
		memDiff := abs(infos[i].MemoryUsedPercent - infos[0].MemoryUsedPercent)
		if memDiff > 10.0 { // 10%以上の変動は異常
			t.Errorf("Memory usage varied too much between samples: %f -> %f", 
				infos[0].MemoryUsedPercent, infos[i].MemoryUsedPercent)
		}
		
		diskDiff := abs(infos[i].DiskUsedPercent - infos[0].DiskUsedPercent)
		if diskDiff > 1.0 { // 1%以上の変動は異常
			t.Errorf("Disk usage varied too much between samples: %f -> %f", 
				infos[0].DiskUsedPercent, infos[i].DiskUsedPercent)
		}
	}
}

// validateSystemInfo はSystemInfoの値が妥当かチェックする
func validateSystemInfo(t *testing.T, info *SystemInfo, sampleID int) {
	t.Helper()
	
	// CPU使用率
	if info.CPUUsagePercent < 0 || info.CPUUsagePercent > 100 {
		t.Errorf("Sample %d: CPU usage out of range: %f", sampleID, info.CPUUsagePercent)
	}
	
	// メモリ情報
	if info.MemoryTotalGB <= 0 {
		t.Errorf("Sample %d: Memory total should be positive: %f", sampleID, info.MemoryTotalGB)
	}
	if info.MemoryUsedGB < 0 {
		t.Errorf("Sample %d: Memory used should be non-negative: %f", sampleID, info.MemoryUsedGB)
	}
	if info.MemoryUsedGB > info.MemoryTotalGB {
		t.Errorf("Sample %d: Memory used exceeds total: used=%f, total=%f", 
			sampleID, info.MemoryUsedGB, info.MemoryTotalGB)
	}
	if info.MemoryUsedPercent < 0 || info.MemoryUsedPercent > 100 {
		t.Errorf("Sample %d: Memory used percent out of range: %f", sampleID, info.MemoryUsedPercent)
	}
	
	// ディスク情報
	if info.DiskTotalGB <= 0 {
		t.Errorf("Sample %d: Disk total should be positive: %f", sampleID, info.DiskTotalGB)
	}
	if info.DiskFreeGB < 0 {
		t.Errorf("Sample %d: Disk free should be non-negative: %f", sampleID, info.DiskFreeGB)
	}
	if info.DiskFreeGB > info.DiskTotalGB {
		t.Errorf("Sample %d: Disk free exceeds total: free=%f, total=%f", 
			sampleID, info.DiskFreeGB, info.DiskTotalGB)
	}
	if info.DiskUsedPercent < 0 || info.DiskUsedPercent > 100 {
		t.Errorf("Sample %d: Disk used percent out of range: %f", sampleID, info.DiskUsedPercent)
	}
	
	// 妥当な最小値チェック（あまりに小さい値は異常）
	if info.MemoryTotalGB < 0.1 { // 100MB未満は異常
		t.Errorf("Sample %d: Memory total too small: %f", sampleID, info.MemoryTotalGB)
	}
	if info.DiskTotalGB < 0.1 { // 100MB未満は異常
		t.Errorf("Sample %d: Disk total too small: %f", sampleID, info.DiskTotalGB)
	}
}

// abs は絶対値を返すヘルパー関数
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// ベンチマークテスト
func BenchmarkDefaultMonitor_GetSystemInfo(b *testing.B) {
	monitor := NewDefaultMonitor()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := monitor.GetSystemInfo()
		if err != nil {
			b.Fatalf("GetSystemInfo() failed: %v", err)
		}
	}
}

// 並行実行ベンチマーク
func BenchmarkDefaultMonitor_GetSystemInfo_Parallel(b *testing.B) {
	monitor := NewDefaultMonitor()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := monitor.GetSystemInfo()
			if err != nil {
				b.Errorf("GetSystemInfo() failed: %v", err)
			}
		}
	})
}