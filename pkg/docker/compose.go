// Package docker はDockerコンテナおよびDocker Composeの操作を提供します
package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// DefaultComposeService implements ComposeService using Docker API
type DefaultComposeService struct {
	client      *client.Client
	projectName string
}

// NewDefaultComposeService creates a new DefaultComposeService
func NewDefaultComposeService() (*DefaultComposeService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &DefaultComposeService{
		client: cli,
	}, nil
}

// SetProjectName sets the Docker Compose project name
func (s *DefaultComposeService) SetProjectName(name string) {
	s.projectName = name
}

// getProjectName returns the project name to use
func (s *DefaultComposeService) getProjectName(composePath string) string {
	if s.projectName != "" {
		return s.projectName
	}
	// デフォルトはディレクトリ名を使用
	defaultName := filepath.Base(filepath.Dir(composePath))
	return defaultName
}

// listContainersWithFilter は指定されたフィルターでコンテナを一覧表示する内部メソッド
func (s *DefaultComposeService) listContainersWithFilter(filterArgs filters.Args) ([]ContainerInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ListOperationTimeout)
	defer cancel()

	containers, err := s.client.ContainerList(ctx, container.ListOptions{
		All:     true, // 停止中のコンテナも含む
		Filters: filterArgs,
	})
	if err != nil {
		return nil, err
	}

	var result []ContainerInfo
	for i := range containers {
		// コンテナの詳細情報を取得
		inspect, err := s.client.ContainerInspect(ctx, containers[i].ID)
		if err != nil {
			continue
		}

		// サービス名を取得
		serviceName := containers[i].Labels[LabelDockerComposeService]

		// ポート情報を整形
		var ports []string
		for _, p := range containers[i].Ports {
			if p.PublicPort > 0 {
				ports = append(ports, fmt.Sprintf("%d:%d/%s", p.PublicPort, p.PrivatePort, p.Type))
			}
		}

		// 状態を判定
		state := strings.ToLower(inspect.State.Status)

		// ヘルスチェックステータス
		healthStatus := "none"
		if inspect.State.Health != nil {
			healthStatus = inspect.State.Health.Status
		}

		// 稼働時間を計算
		var runningFor string
		if state == containerStateRunning && inspect.State.StartedAt != "" {
			startedAt, err := time.Parse(time.RFC3339Nano, inspect.State.StartedAt)
			if err == nil {
				duration := time.Since(startedAt)
				runningFor = formatDuration(duration)
			}
		}

		info := ContainerInfo{
			ID:           containers[i].ID[:12],
			Name:         strings.TrimPrefix(containers[i].Names[0], "/"),
			Service:      serviceName,
			Image:        containers[i].Image,
			Status:       containers[i].Status,
			State:        state,
			RunningFor:   runningFor,
			Ports:        ports,
			HealthStatus: healthStatus,
			CreatedAt:    time.Unix(containers[i].Created, 0),
		}

		result = append(result, info)
	}

	return result, nil
}

// ListContainers lists all containers managed by docker-compose
func (s *DefaultComposeService) ListContainers(composePath string) ([]ContainerInfo, error) {
	projectName := s.getProjectName(composePath)

	// Docker Composeのラベルでフィルター
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("%s=%s", LabelDockerComposeProject, projectName))

	containers, err := s.listContainersWithFilter(filterArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	return containers, nil
}

// ListGameContainers lists only game containers (with game.type label)
// Returns containers that have the "game.type" label, excluding
// infrastructure containers like watchdog
func (s *DefaultComposeService) ListGameContainers(composePath string) ([]ContainerInfo, error) {
	projectName := s.getProjectName(composePath)

	// Docker Composeのラベルとgame.typeラベルでフィルター
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("%s=%s", LabelDockerComposeProject, projectName))
	filterArgs.Add("label", LabelGameType)

	containers, err := s.listContainersWithFilter(filterArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers with %s label: %w", LabelGameType, err)
	}

	return containers, nil
}

// StartService starts a specific service
func (s *DefaultComposeService) StartService(composePath, serviceName string) error {
	return s.executeServiceOperation(composePath, serviceName, "start",
		func(ctx context.Context, c container.Summary) error {
			return s.client.ContainerStart(ctx, c.ID, container.StartOptions{})
		})
}

// StopService stops a specific service
func (s *DefaultComposeService) StopService(composePath, serviceName string) error {
	return s.executeServiceOperation(composePath, serviceName, "stop",
		func(ctx context.Context, c container.Summary) error {
			return s.client.ContainerStop(ctx, c.ID, container.StopOptions{})
		})
}

// GetContainerStats gets resource usage stats for a specific container
func (s *DefaultComposeService) GetContainerStats(containerName string) (*ContainerStats, error) {
	// コンテナを名前で検索
	containerSummary, err := s.findContainerByName(containerName)
	if err != nil {
		return nil, err
	}

	// 統計情報を取得
	stats, err := s.getContainerStatsData(containerSummary.ID)
	if err != nil {
		return nil, err
	}

	// 統計情報を計算
	return s.calculateContainerStats(containerSummary, stats), nil
}

// findContainerByName finds a container by its name
func (s *DefaultComposeService) findContainerByName(containerName string) (*container.Summary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), QueryOperationTimeout)
	defer cancel()

	// コンテナ名でフィルター
	filterArgs := filters.NewArgs()
	filterArgs.Add("name", containerName)

	containers, err := s.client.ContainerList(ctx, container.ListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find container: %w", err)
	}

	if len(containers) == 0 {
		return nil, fmt.Errorf("container %s not found", containerName)
	}

	return &containers[0], nil
}

// getContainerStatsData retrieves and parses container stats from Docker API
func (s *DefaultComposeService) getContainerStatsData(containerID string) (*container.StatsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), QueryOperationTimeout)
	defer cancel()

	// Stats APIを呼び出し
	statsResponse, err := s.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer func() {
		if err := statsResponse.Body.Close(); err != nil {
			// ログ出力は避けて、エラーを無視
			_ = err
		}
	}()

	// 統計情報を読み取り
	data, err := io.ReadAll(statsResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read stats: %w", err)
	}

	var stats container.StatsResponse
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse stats: %w", err)
	}

	return &stats, nil
}

// calculateContainerStats calculates all container statistics and returns ContainerStats
func (s *DefaultComposeService) calculateContainerStats(
	containerSummary *container.Summary, stats *container.StatsResponse) *ContainerStats {
	// CPU使用率を計算
	cpuPercent := calculateCPUPercent(stats)

	// メモリ使用率を計算
	memPercent := s.calculateMemoryPercent(stats)
	memUsage := s.formatMemoryUsage(stats)

	// ネットワークI/Oを計算
	networkIO := s.calculateNetworkIO(stats)

	// ブロックI/Oを計算
	blockIO := s.calculateBlockIO(stats)

	return &ContainerStats{
		ContainerID:   containerSummary.ID[:12],
		Name:          strings.TrimPrefix(containerSummary.Names[0], "/"),
		CPUPercent:    cpuPercent,
		MemoryPercent: memPercent,
		MemoryUsage:   memUsage,
		NetworkIO:     networkIO,
		BlockIO:       blockIO,
	}
}

// calculateMemoryPercent calculates memory usage percentage
func (s *DefaultComposeService) calculateMemoryPercent(stats *container.StatsResponse) float64 {
	return float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0
}

// formatMemoryUsage formats memory usage as "used / total"
func (s *DefaultComposeService) formatMemoryUsage(stats *container.StatsResponse) string {
	return fmt.Sprintf("%.2fGiB / %.2fGiB",
		float64(stats.MemoryStats.Usage)/1024/1024/1024,
		float64(stats.MemoryStats.Limit)/1024/1024/1024)
}

// calculateNetworkIO calculates network I/O statistics
func (s *DefaultComposeService) calculateNetworkIO(stats *container.StatsResponse) string {
	var rxBytes, txBytes uint64
	for _, v := range stats.Networks {
		rxBytes += v.RxBytes
		txBytes += v.TxBytes
	}
	return fmt.Sprintf("%s / %s", formatBytes(rxBytes), formatBytes(txBytes))
}

// calculateBlockIO calculates block I/O statistics
func (s *DefaultComposeService) calculateBlockIO(stats *container.StatsResponse) string {
	var readBytes, writeBytes uint64
	for _, v := range stats.BlkioStats.IoServiceBytesRecursive {
		switch v.Op {
		case "read":
			readBytes += v.Value
		case "write":
			writeBytes += v.Value
		}
	}
	return fmt.Sprintf("%s / %s", formatBytes(readBytes), formatBytes(writeBytes))
}

// GetAllContainersStats gets resource usage stats for all containers
func (s *DefaultComposeService) GetAllContainersStats(composePath string) ([]ContainerStats, error) {
	containers, err := s.ListContainers(composePath)
	if err != nil {
		return nil, err
	}

	var stats []ContainerStats
	for i := range containers {
		if strings.EqualFold(containers[i].State, containerStateRunning) {
			stat, err := s.GetContainerStats(containers[i].Name)
			if err != nil {
				continue
			}
			stats = append(stats, *stat)
		}
	}

	return stats, nil
}

// executeServiceOperation executes a common service operation pattern
func (s *DefaultComposeService) executeServiceOperation(composePath, serviceName, operation string,
	containerOp func(context.Context, container.Summary) error) error {
	if !IsValidServiceName(serviceName) {
		return fmt.Errorf("%w: %s", ErrInvalidServiceName, serviceName)
	}

	projectName := s.getProjectName(composePath)

	// サービスに属するコンテナを検索
	containers, err := s.findServiceContainers(projectName, serviceName)
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return fmt.Errorf("service %s not found", serviceName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ServiceOperationTimeout)
	defer cancel()

	// すべてのコンテナに操作を実行
	for i := range containers {
		if err := containerOp(ctx, containers[i]); err != nil {
			return fmt.Errorf("failed to %s container %s: %w", operation, containers[i].Names[0], err)
		}
	}

	return nil
}

// RestartContainer restarts a specific container
func (s *DefaultComposeService) RestartContainer(composePath, serviceName string) error {
	return s.executeServiceOperation(composePath, serviceName, "restart",
		func(ctx context.Context, c container.Summary) error {
			return s.client.ContainerRestart(ctx, c.ID, container.StopOptions{})
		})
}

// GetContainerLogs gets logs from a specific container
func (s *DefaultComposeService) GetContainerLogs(composePath, serviceName string, lines int) (string, error) {
	if !IsValidServiceName(serviceName) {
		return "", fmt.Errorf("%w: %s", ErrInvalidServiceName, serviceName)
	}

	if lines <= 0 {
		lines = 100
	} else if lines > 1000 {
		lines = 1000
	}

	projectName := s.getProjectName(composePath)

	// サービスに属するコンテナを検索
	containers, err := s.findServiceContainers(projectName, serviceName)
	if err != nil {
		return "", err
	}

	if len(containers) == 0 {
		return "", fmt.Errorf("service %s not found", serviceName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ListOperationTimeout)
	defer cancel()

	// ログを取得
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       strconv.Itoa(lines),
		Timestamps: false,
	}

	logsReader, err := s.client.ContainerLogs(ctx, containers[0].ID, options)
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}
	defer func() {
		if err := logsReader.Close(); err != nil {
			_ = err
		}
	}()

	// ログを読み取り
	logs, err := io.ReadAll(logsReader)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}

	// Docker APIのログは特殊なフォーマットなので、クリーンアップ
	return cleanDockerLogs(string(logs)), nil
}

// findServiceContainers finds containers belonging to a specific service
func (s *DefaultComposeService) findServiceContainers(projectName, serviceName string) ([]container.Summary, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("%s=%s", LabelDockerComposeProject, projectName))
	filterArgs.Add("label", fmt.Sprintf("%s=%s", LabelDockerComposeService, serviceName))

	ctx, cancel := context.WithTimeout(context.Background(), QueryOperationTimeout)
	defer cancel()

	return s.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
}

// calculateCPUPercent calculates CPU usage percentage
func calculateCPUPercent(stats *container.StatsResponse) float64 {
	// CPU使用率の計算（Docker公式の方法）
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent := (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
		return cpuPercent
	}

	return 0.0
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	default:
		days := int(d.Hours()) / 24
		hours := int(d.Hours()) % 24
		return fmt.Sprintf("%dd%dh", days, hours)
	}
}

// formatBytes formats bytes into a human-readable string
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// cleanDockerLogs removes Docker's log format headers
func cleanDockerLogs(logs string) string {
	lines := strings.Split(logs, "\n")
	var cleaned []string

	for _, line := range lines {
		// Docker APIのログは各行の先頭に8バイトのヘッダーがある
		if len(line) > 8 {
			cleaned = append(cleaned, line[8:])
		} else if line != "" {
			cleaned = append(cleaned, line)
		}
	}

	return strings.Join(cleaned, "\n")
}

// Close closes the Docker client connection
func (s *DefaultComposeService) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
