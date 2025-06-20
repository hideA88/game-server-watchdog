package docker

import (
	"errors"
	"testing"
)

const (
	testDockerCompose = "docker-compose.yml"
	testServiceName   = "web"
)

func TestMockComposeService_StartService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		composePath string
		serviceName string
		mockFunc    func(string, string) error
		wantErr     bool
	}{
		{
			name:        "successful start",
			composePath: testDockerCompose,
			serviceName: testServiceName,
			mockFunc: func(_, _ string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:        "start error",
			composePath: testDockerCompose,
			serviceName: testServiceName,
			mockFunc: func(_, _ string) error {
				return errors.New("start failed")
			},
			wantErr: true,
		},
		{
			name:        "nil function returns nil",
			composePath: testDockerCompose,
			serviceName: testServiceName,
			mockFunc:    nil,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockComposeService{
				StartServiceFunc: tt.mockFunc,
			}
			err := m.StartService(tt.composePath, tt.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("StartService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMockComposeService_StopService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		composePath string
		serviceName string
		mockFunc    func(string, string) error
		wantErr     bool
	}{
		{
			name:        "successful stop",
			composePath: "docker-compose.yml",
			serviceName: "web",
			mockFunc: func(path, service string) error {
				if path != "docker-compose.yml" {
					t.Errorf("unexpected composePath: %s", path)
				}
				if service != "web" {
					t.Errorf("unexpected serviceName: %s", service)
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:        "stop error",
			composePath: testDockerCompose,
			serviceName: testServiceName,
			mockFunc: func(_, _ string) error {
				return errors.New("stop failed")
			},
			wantErr: true,
		},
		{
			name:        "nil function returns nil",
			composePath: testDockerCompose,
			serviceName: testServiceName,
			mockFunc:    nil,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockComposeService{
				StopServiceFunc: tt.mockFunc,
			}
			err := m.StopService(tt.composePath, tt.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("StopService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
