package docker

import "time"

// ContainerInfo represents information about a Docker container
type ContainerInfo struct {
	ID           string
	Name         string
	Service      string
	Image        string
	Status       string
	State        string
	RunningFor   string
	Ports        []string
	HealthStatus string
	CreatedAt    time.Time
}

// ComposeService represents Docker Compose operations
type ComposeService interface {
	// ListContainers returns a list of containers managed by docker-compose
	ListContainers(composePath string) ([]ContainerInfo, error)
}
