package docker

// MockComposeService is a mock implementation of ComposeService for testing
type MockComposeService struct {
	ListContainersFunc func(composePath string) ([]ContainerInfo, error)
}

// ListContainers calls the mock function
func (m *MockComposeService) ListContainers(composePath string) ([]ContainerInfo, error) {
	if m.ListContainersFunc != nil {
		return m.ListContainersFunc(composePath)
	}
	return nil, nil
}
