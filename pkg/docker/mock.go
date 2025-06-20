package docker

// MockComposeService is a mock implementation of ComposeService for testing
type MockComposeService struct {
	ListContainersFunc        func(composePath string) ([]ContainerInfo, error)
	ListGameContainersFunc    func(composePath string) ([]ContainerInfo, error)
	StartServiceFunc          func(composePath, serviceName string) error
	StopServiceFunc           func(composePath, serviceName string) error
	GetContainerStatsFunc     func(containerName string) (*ContainerStats, error)
	GetAllContainersStatsFunc func(composePath string) ([]ContainerStats, error)
	RestartContainerFunc      func(composePath, serviceName string) error
	GetContainerLogsFunc      func(composePath, serviceName string, lines int) (string, error)
}

// ListContainers calls the mock function
func (m *MockComposeService) ListContainers(composePath string) ([]ContainerInfo, error) {
	if m.ListContainersFunc != nil {
		return m.ListContainersFunc(composePath)
	}
	return nil, nil
}

// ListGameContainers calls the mock function
func (m *MockComposeService) ListGameContainers(composePath string) ([]ContainerInfo, error) {
	if m.ListGameContainersFunc != nil {
		return m.ListGameContainersFunc(composePath)
	}
	// デフォルトではListContainersと同じ動作
	return m.ListContainers(composePath)
}

// StartService calls the mock function
func (m *MockComposeService) StartService(composePath, serviceName string) error {
	if m.StartServiceFunc != nil {
		return m.StartServiceFunc(composePath, serviceName)
	}
	return nil
}

// StopService calls the mock function
func (m *MockComposeService) StopService(composePath, serviceName string) error {
	if m.StopServiceFunc != nil {
		return m.StopServiceFunc(composePath, serviceName)
	}
	return nil
}

// GetContainerStats calls the mock function
func (m *MockComposeService) GetContainerStats(containerName string) (*ContainerStats, error) {
	if m.GetContainerStatsFunc != nil {
		return m.GetContainerStatsFunc(containerName)
	}
	return nil, nil
}

// GetAllContainersStats calls the mock function
func (m *MockComposeService) GetAllContainersStats(composePath string) ([]ContainerStats, error) {
	if m.GetAllContainersStatsFunc != nil {
		return m.GetAllContainersStatsFunc(composePath)
	}
	return nil, nil
}

// RestartContainer calls the mock function
func (m *MockComposeService) RestartContainer(composePath, serviceName string) error {
	if m.RestartContainerFunc != nil {
		return m.RestartContainerFunc(composePath, serviceName)
	}
	return nil
}

// GetContainerLogs calls the mock function
func (m *MockComposeService) GetContainerLogs(composePath, serviceName string, lines int) (string, error) {
	if m.GetContainerLogsFunc != nil {
		return m.GetContainerLogsFunc(composePath, serviceName, lines)
	}
	return "", nil
}

// Close is a no-op for the mock
func (m *MockComposeService) Close() error {
	return nil
}
