package system

import (
	"errors"
	"testing"
)

func TestMockMonitor_GetSystemInfo(t *testing.T) {
	tests := []struct {
		name       string
		systemInfo *SystemInfo
		err        error
		wantErr    bool
	}{
		{
			name: "正常系 - システム情報を返す",
			systemInfo: &SystemInfo{
				CPUUsagePercent:   50.5,
				MemoryUsedGB:      4.2,
				MemoryTotalGB:     8.0,
				MemoryUsedPercent: 52.5,
				DiskFreeGB:        100.5,
				DiskTotalGB:       500.0,
				DiskUsedPercent:   79.9,
			},
			err:     nil,
			wantErr: false,
		},
		{
			name:       "エラー系 - エラーを返す",
			systemInfo: nil,
			err:        errors.New("mock error"),
			wantErr:    true,
		},
		{
			name: "エラーがない場合はシステム情報を返す",
			systemInfo: &SystemInfo{
				CPUUsagePercent:   10.0,
				MemoryUsedGB:      2.0,
				MemoryTotalGB:     16.0,
				MemoryUsedPercent: 12.5,
				DiskFreeGB:        200.0,
				DiskTotalGB:       1000.0,
				DiskUsedPercent:   80.0,
			},
			err:     nil,
			wantErr: false,
		},
		{
			name:       "nilのシステム情報とエラーなし",
			systemInfo: nil,
			err:        nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &MockMonitor{
				SystemInfo: tt.systemInfo,
				Err:        tt.err,
			}

			got, err := m.GetSystemInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSystemInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				// エラーが期待される場合
				if err != tt.err {
					t.Errorf("GetSystemInfo() error = %v, want %v", err, tt.err)
				}
				if got != nil {
					t.Errorf("GetSystemInfo() = %v, want nil when error", got)
				}
			} else if got != tt.systemInfo {
				// エラーが期待されない場合
				t.Errorf("GetSystemInfo() = %v, want %v", got, tt.systemInfo)
			}
		})
	}
}

func TestMockMonitor_ImplementsInterface(_ *testing.T) {
	// MockMonitorがMonitorインターフェースを実装していることを確認
	var _ Monitor = &MockMonitor{}
	var _ Monitor = (*MockMonitor)(nil)
}

func TestMockMonitor_NilReceiver(t *testing.T) {
	// nilレシーバーでのテスト（パニックすることを確認）
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil receiver, but didn't panic")
		}
	}()

	var m *MockMonitor
	// nilレシーバーでパニックすることを確認
	_, _ = m.GetSystemInfo()
}

func TestMockMonitor_ConcurrentAccess(t *testing.T) {
	// 並行アクセスのテスト
	m := &MockMonitor{
		SystemInfo: &SystemInfo{
			CPUUsagePercent:   75.0,
			MemoryUsedGB:      6.0,
			MemoryTotalGB:     8.0,
			MemoryUsedPercent: 75.0,
			DiskFreeGB:        50.0,
			DiskTotalGB:       200.0,
			DiskUsedPercent:   75.0,
		},
		Err: nil,
	}

	// 複数のゴルーチンから同時にアクセス
	const goroutines = 10
	const iterations = 100

	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				info, err := m.GetSystemInfo()
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if info == nil {
					t.Error("Expected non-nil SystemInfo")
				}
			}
			done <- true
		}()
	}

	// すべてのゴルーチンの完了を待つ
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

// ベンチマークテスト
func BenchmarkMockMonitor_GetSystemInfo(b *testing.B) {
	m := &MockMonitor{
		SystemInfo: &SystemInfo{
			CPUUsagePercent:   50.0,
			MemoryUsedGB:      4.0,
			MemoryTotalGB:     8.0,
			MemoryUsedPercent: 50.0,
			DiskFreeGB:        100.0,
			DiskTotalGB:       500.0,
			DiskUsedPercent:   80.0,
		},
		Err: nil,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := m.GetSystemInfo()
		if err != nil {
			b.Fatalf("GetSystemInfo() failed: %v", err)
		}
	}
}

func BenchmarkMockMonitor_GetSystemInfo_WithError(b *testing.B) {
	m := &MockMonitor{
		SystemInfo: nil,
		Err:        errors.New("mock error"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := m.GetSystemInfo()
		if err == nil {
			b.Fatal("Expected error, got nil")
		}
	}
}

// 並列ベンチマーク
func BenchmarkMockMonitor_GetSystemInfo_Parallel(b *testing.B) {
	m := &MockMonitor{
		SystemInfo: &SystemInfo{
			CPUUsagePercent:   50.0,
			MemoryUsedGB:      4.0,
			MemoryTotalGB:     8.0,
			MemoryUsedPercent: 50.0,
			DiskFreeGB:        100.0,
			DiskTotalGB:       500.0,
			DiskUsedPercent:   80.0,
		},
		Err: nil,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := m.GetSystemInfo()
			if err != nil {
				b.Errorf("GetSystemInfo() failed: %v", err)
			}
		}
	})
}
