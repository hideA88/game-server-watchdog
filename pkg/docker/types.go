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

// ContainerStats represents container resource usage statistics
type ContainerStats struct {
	ContainerID   string
	Name          string
	CPUPercent    float64
	MemoryPercent float64
	MemoryUsage   string // e.g., "1.5GiB / 2GiB"
	NetworkIO     string // e.g., "1.2MB / 3.4MB"
	BlockIO       string // e.g., "5.6MB / 7.8MB"
}

// ComposeService represents Docker Compose operations
type ComposeService interface {
	// ListContainers returns a list of containers managed by docker-compose
	ListContainers(composePath string) ([]ContainerInfo, error)
	// ListGameContainers returns a list of game containers (with game.type label)
	ListGameContainers(composePath string) ([]ContainerInfo, error)
	// StartService starts a specific service
	StartService(composePath string, serviceName string) error
	// StopService stops a specific service
	StopService(composePath string, serviceName string) error
	// GetContainerStats gets resource usage stats for a specific container
	GetContainerStats(containerName string) (*ContainerStats, error)
	// GetAllContainersStats gets resource usage stats for all containers
	GetAllContainersStats(composePath string) ([]ContainerStats, error)
	// RestartContainer restarts a specific container
	RestartContainer(composePath string, serviceName string) error
	// GetContainerLogs gets logs from a specific container
	GetContainerLogs(composePath string, serviceName string, lines int) (string, error)
	// Close closes the Docker client connection
	Close() error
}
