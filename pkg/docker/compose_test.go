package docker

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestNewDefaultComposeService(t *testing.T) {
	t.Parallel()

	service := NewDefaultComposeService()
	if service == nil {
		t.Error("NewDefaultComposeService() returned nil")
	}
}

func TestDefaultComposeService_ListContainers(t *testing.T) {
	t.Parallel()

	// This test requires docker-compose to be installed
	// We'll skip it in CI environments without docker-compose
	tests := []struct {
		name        string
		composePath string
		wantErr     bool
		skipReason  string
	}{
		{
			name:        "invalid compose path",
			composePath: "/invalid/path/that/does/not/exist/docker-compose.yml",
			wantErr:     true,
		},
		{
			name:        "empty path uses default",
			composePath: "",
			wantErr:     true, // Will fail if docker-compose.yml doesn't exist
			skipReason:  "Requires docker-compose.yml in working directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			s := NewDefaultComposeService()
			got, err := s.ListContainers(tt.composePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListContainers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Error("ListContainers() returned nil containers on success")
			}
		})
	}
}

func TestMockComposeService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mockFunc       func(string) ([]ContainerInfo, error)
		expectedResult []ContainerInfo
		expectedError  error
	}{
		{
			name: "mock returns containers",
			mockFunc: func(path string) ([]ContainerInfo, error) {
				return []ContainerInfo{
					{Name: "test-container"},
				}, nil
			},
			expectedResult: []ContainerInfo{
				{Name: "test-container"},
			},
			expectedError: nil,
		},
		{
			name:           "mock with nil function",
			mockFunc:       nil,
			expectedResult: nil,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockComposeService{
				ListContainersFunc: tt.mockFunc,
			}

			got, err := mock.ListContainers("test-path")

			if err != tt.expectedError {
				t.Errorf("ListContainers() error = %v, want %v", err, tt.expectedError)
			}

			if !reflect.DeepEqual(got, tt.expectedResult) {
				t.Errorf("ListContainers() = %v, want %v", got, tt.expectedResult)
			}
		})
	}
}

// MockCommandExecutor is a mock implementation of CommandExecutor
type MockCommandExecutor struct {
	OutputFunc        func(name string, args ...string) ([]byte, error)
	OutputContextFunc func(ctx context.Context, name string, args ...string) ([]byte, error)
	LookPathFunc      func(file string) (string, error)
}

func (m *MockCommandExecutor) Output(name string, args ...string) ([]byte, error) {
	if m.OutputFunc != nil {
		return m.OutputFunc(name, args...)
	}
	return nil, nil
}

func (m *MockCommandExecutor) OutputContext(ctx context.Context, name string, args ...string) ([]byte, error) {
	if m.OutputContextFunc != nil {
		return m.OutputContextFunc(ctx, name, args...)
	}
	if m.OutputFunc != nil {
		return m.OutputFunc(name, args...)
	}
	return nil, nil
}

func (m *MockCommandExecutor) LookPath(file string) (string, error) {
	if m.LookPathFunc != nil {
		return m.LookPathFunc(file)
	}
	return "", nil
}

func TestDefaultComposeService_ListContainers_WithMock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		composePath  string
		mockExecutor *MockCommandExecutor
		want         []ContainerInfo
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name:        "successful container listing",
			composePath: "docker-compose.yml",
			mockExecutor: &MockCommandExecutor{
				OutputFunc: func(name string, args ...string) ([]byte, error) {
					// Return valid JSON output
					return []byte(`{"ID":"abc123","Name":"test-container","Service":"web","Status":"Up 2 hours","State":"running","Health":"healthy","ExitCode":0,"Publishers":[{"URL":"","TargetPort":80,"PublishedPort":8080,"Protocol":"tcp"}]}
{"ID":"def456","Name":"db-container","Service":"db","Status":"Up 1 hour","State":"running","Health":"","ExitCode":0,"Publishers":[{"URL":"","TargetPort":5432,"PublishedPort":0,"Protocol":"tcp"}]}`), nil
				},
			},
			want: []ContainerInfo{
				{
					ID:           "abc123",
					Name:         "test-container",
					Service:      "web",
					Status:       "Up 2 hours",
					State:        "running",
					HealthStatus: "healthy",
					RunningFor:   "2 hours",
					Ports:        []string{"8080:80"},
				},
				{
					ID:           "def456",
					Name:         "db-container",
					Service:      "db",
					Status:       "Up 1 hour",
					State:        "running",
					HealthStatus: "",
					RunningFor:   "1 hour",
					Ports:        []string{"5432"},
				},
			},
			wantErr: false,
		},
		{
			name:        "docker not found",
			composePath: "docker-compose.yml",
			mockExecutor: &MockCommandExecutor{
				OutputFunc: func(name string, args ...string) ([]byte, error) {
					return nil, errors.New("docker command failed")
				},
				LookPathFunc: func(file string) (string, error) {
					return "", errors.New("docker not found")
				},
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "docker not found",
		},
		{
			name:        "docker compose execution error",
			composePath: "docker-compose.yml",
			mockExecutor: &MockCommandExecutor{
				OutputFunc: func(name string, args ...string) ([]byte, error) {
					return nil, errors.New("docker compose failed")
				},
				LookPathFunc: func(file string) (string, error) {
					return "/usr/bin/docker", nil
				},
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "failed to execute docker compose",
		},
		{
			name:        "invalid JSON output",
			composePath: "docker-compose.yml",
			mockExecutor: &MockCommandExecutor{
				OutputFunc: func(name string, args ...string) ([]byte, error) {
					return []byte("invalid json"), nil
				},
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "failed to parse JSON",
		},
		{
			name:        "empty output",
			composePath: "docker-compose.yml",
			mockExecutor: &MockCommandExecutor{
				OutputFunc: func(name string, args ...string) ([]byte, error) {
					return []byte(""), nil
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:        "container with UDP port",
			composePath: "docker-compose.yml",
			mockExecutor: &MockCommandExecutor{
				OutputFunc: func(name string, args ...string) ([]byte, error) {
					return []byte(`{"ID":"xyz789","Name":"dns-container","Service":"dns","Status":"Up 30 minutes","State":"running","Health":"","ExitCode":0,"Publishers":[{"URL":"","TargetPort":53,"PublishedPort":53,"Protocol":"udp"}]}`), nil
				},
			},
			want: []ContainerInfo{
				{
					ID:           "xyz789",
					Name:         "dns-container",
					Service:      "dns",
					Status:       "Up 30 minutes",
					State:        "running",
					HealthStatus: "",
					RunningFor:   "30 minutes",
					Ports:        []string{"53:53/udp"},
				},
			},
			wantErr: false,
		},
		{
			name:        "container not running",
			composePath: "docker-compose.yml",
			mockExecutor: &MockCommandExecutor{
				OutputFunc: func(name string, args ...string) ([]byte, error) {
					return []byte(`{"ID":"stopped123","Name":"stopped-container","Service":"worker","Status":"Exited (0) 5 minutes ago","State":"exited","Health":"","ExitCode":0,"Publishers":[]}`), nil
				},
			},
			want: []ContainerInfo{
				{
					ID:           "stopped123",
					Name:         "stopped-container",
					Service:      "worker",
					Status:       "Exited (0) 5 minutes ago",
					State:        "exited",
					HealthStatus: "",
					RunningFor:   "",
					Ports:        nil,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DefaultComposeService{
				executor: tt.mockExecutor,
			}

			got, err := s.ListContainers(tt.composePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListContainers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.wantErrMsg != "" {
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("ListContainers() error = %v, want error containing %v", err, tt.wantErrMsg)
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListContainers() = %v, want %v", got, tt.want)
				for i := range got {
					if i < len(tt.want) {
						if !reflect.DeepEqual(got[i], tt.want[i]) {
							t.Errorf("Container[%d] differs:\ngot:  %+v\nwant: %+v", i, got[i], tt.want[i])
						}
					}
				}
			}
		})
	}
}

func TestRealCommandExecutor(t *testing.T) {
	t.Parallel()

	e := &RealCommandExecutor{}

	// Test LookPath with a common command
	path, err := e.LookPath("echo")
	if err != nil {
		t.Errorf("LookPath(\"echo\") error = %v", err)
	}
	if path == "" {
		t.Error("LookPath(\"echo\") returned empty path")
	}

	// Test Output with a simple command
	output, err := e.Output("echo", "test")
	if err != nil {
		t.Errorf("Output(\"echo\", \"test\") error = %v", err)
	}
	if !strings.Contains(string(output), "test") {
		t.Errorf("Output(\"echo\", \"test\") = %v, want output containing \"test\"", string(output))
	}
}
