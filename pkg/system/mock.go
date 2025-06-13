package system

// MockMonitor はテスト用のモック実装
type MockMonitor struct {
	SystemInfo *SystemInfo
	Err        error
}

// GetSystemInfo はモックのシステム情報を返す
func (m *MockMonitor) GetSystemInfo() (*SystemInfo, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.SystemInfo, nil
}
