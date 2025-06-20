package docker

// MockComposeService is a mock implementation of ComposeService for testing
type MockComposeService struct {
	ListContainersFunc func(composePath string) ([]ContainerInfo, error)
	StartServiceFunc   func(composePath string, serviceName string) error
	StopServiceFunc    func(composePath string, serviceName string) error
}

// ListContainers calls the mock function
func (m *MockComposeService) ListContainers(composePath string) ([]ContainerInfo, error) {
	if m.ListContainersFunc != nil {
		return m.ListContainersFunc(composePath)
	}
	return nil, nil
}

// StartService calls the mock function
func (m *MockComposeService) StartService(composePath string, serviceName string) error {
	if m.StartServiceFunc != nil {
		return m.StartServiceFunc(composePath, serviceName)
	}
	return nil
}

// StopService calls the mock function
func (m *MockComposeService) StopService(composePath string, serviceName string) error {
	if m.StopServiceFunc != nil {
		return m.StopServiceFunc(composePath, serviceName)
	}
	return nil
}
