package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// CommandExecutor is an interface for executing commands
type CommandExecutor interface {
	Output(name string, args ...string) ([]byte, error)
	OutputContext(ctx context.Context, name string, args ...string) ([]byte, error)
	LookPath(file string) (string, error)
}

// RealCommandExecutor implements CommandExecutor with real exec commands
type RealCommandExecutor struct{}

// Output executes a command and returns its output
func (e *RealCommandExecutor) Output(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

// OutputContext executes a command with context and returns its output
func (e *RealCommandExecutor) OutputContext(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

// LookPath searches for an executable in PATH
func (e *RealCommandExecutor) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// DefaultComposeService implements ComposeService using docker compose CLI
type DefaultComposeService struct {
	executor CommandExecutor
}

// NewDefaultComposeService creates a new DefaultComposeService
func NewDefaultComposeService() *DefaultComposeService {
	return &DefaultComposeService{
		executor: &RealCommandExecutor{},
	}
}

// dockerComposeJSON represents the JSON output from docker compose ps
type dockerComposeJSON struct {
	ID         string     `json:"ID"`
	Name       string     `json:"Name"`
	Service    string     `json:"Service"`
	Status     string     `json:"Status"`
	State      string     `json:"State"`
	Health     string     `json:"Health"`
	ExitCode   int        `json:"ExitCode"`
	Publishers []portInfo `json:"Publishers"`
}

type portInfo struct {
	URL           string `json:"URL"`
	TargetPort    int    `json:"TargetPort"`
	PublishedPort int    `json:"PublishedPort"`
	Protocol      string `json:"Protocol"`
}

// ListContainers returns a list of containers managed by docker compose
func (s *DefaultComposeService) ListContainers(composePath string) ([]ContainerInfo, error) {
	// Validate compose file path
	if composePath == "" {
		composePath = "docker-compose.yml"
	}

	absPath, err := filepath.Abs(composePath)
	if err != nil {
		return nil, fmt.Errorf("invalid compose path: %w", err)
	}

	// Run docker compose ps command with JSON format (Docker Compose V2)
	output, err := s.executor.Output("docker", "compose", "-f", absPath, "ps", "--format", "json")
	if err != nil {
		// Check if docker is installed
		if _, err := s.executor.LookPath("docker"); err != nil {
			return nil, fmt.Errorf("docker not found: %w", err)
		}
		return nil, fmt.Errorf("failed to execute docker compose: %w", err)
	}

	// Parse JSON output line by line
	var containers []ContainerInfo
	trimmedOutput := strings.TrimSpace(string(output))
	if trimmedOutput == "" {
		return containers, nil
	}
	lines := strings.Split(trimmedOutput, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		var dcContainer dockerComposeJSON
		if err := json.Unmarshal([]byte(line), &dcContainer); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}

		// Convert to ContainerInfo
		info := ContainerInfo{
			ID:           dcContainer.ID,
			Name:         dcContainer.Name,
			Service:      dcContainer.Service,
			Status:       dcContainer.Status,
			State:        dcContainer.State,
			HealthStatus: dcContainer.Health,
		}

		// Parse ports
		for _, port := range dcContainer.Publishers {
			portStr := fmt.Sprintf("%d", port.TargetPort)
			if port.PublishedPort != 0 {
				portStr = fmt.Sprintf("%d:%d", port.PublishedPort, port.TargetPort)
			}
			if port.Protocol != "" && port.Protocol != "tcp" {
				portStr += "/" + port.Protocol
			}
			info.Ports = append(info.Ports, portStr)
		}

		// Extract running time from status
		if strings.Contains(info.Status, "Up") {
			parts := strings.Split(info.Status, " ")
			if len(parts) >= 2 {
				info.RunningFor = strings.Join(parts[1:], " ")
			}
		}

		containers = append(containers, info)
	}

	return containers, nil
}

// StartService starts a specific service using docker compose
func (s *DefaultComposeService) StartService(composePath string, serviceName string) error {
	// Validate service name for security
	if !IsValidServiceName(serviceName) {
		return fmt.Errorf("%w: %s", ErrInvalidServiceName, serviceName)
	}
	
	// Validate compose file path
	if composePath == "" {
		composePath = "docker-compose.yml"
	}

	absPath, err := filepath.Abs(composePath)
	if err != nil {
		return fmt.Errorf("invalid compose path: %w", err)
	}

	// Run docker compose start command for specific service with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ServiceOperationTimeout)
	defer cancel()
	
	output, err := s.executor.OutputContext(ctx, "docker", "compose", "-f", absPath, "start", serviceName)
	if err != nil {
		// Check if it's a timeout
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("start service %s timeout after %v", serviceName, ServiceOperationTimeout)
		}
		// Check if docker is installed
		if _, err := s.executor.LookPath("docker"); err != nil {
			return fmt.Errorf("docker not found: %w", err)
		}
		return fmt.Errorf("failed to start service %s: %w (output: %s)", serviceName, err, string(output))
	}

	return nil
}

// StopService stops a specific service using docker compose
func (s *DefaultComposeService) StopService(composePath string, serviceName string) error {
	// Validate service name for security
	if !IsValidServiceName(serviceName) {
		return fmt.Errorf("%w: %s", ErrInvalidServiceName, serviceName)
	}
	
	// Validate compose file path
	if composePath == "" {
		composePath = "docker-compose.yml"
	}

	absPath, err := filepath.Abs(composePath)
	if err != nil {
		return fmt.Errorf("invalid compose path: %w", err)
	}

	// Run docker compose stop command for specific service with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ServiceOperationTimeout)
	defer cancel()
	
	output, err := s.executor.OutputContext(ctx, "docker", "compose", "-f", absPath, "stop", serviceName)
	if err != nil {
		// Check if it's a timeout
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("stop service %s timeout after %v", serviceName, ServiceOperationTimeout)
		}
		// Check if docker is installed
		if _, err := s.executor.LookPath("docker"); err != nil {
			return fmt.Errorf("docker not found: %w", err)
		}
		return fmt.Errorf("failed to stop service %s: %w (output: %s)", serviceName, err, string(output))
	}

	return nil
}
