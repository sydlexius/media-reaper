package connection

import (
	"context"
	"fmt"

	"github.com/sydlexius/media-reaper/internal/arrclient"
	"github.com/sydlexius/media-reaper/internal/emby"
	"github.com/sydlexius/media-reaper/internal/repository"
)

// TestResult represents the outcome of a connection test.
type TestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	AppName string `json:"appName,omitempty"`
	Version string `json:"version,omitempty"`
}

// TestConnection tests connectivity for a given type, URL, and plaintext API key.
func TestConnection(ctx context.Context, connType, url, apiKey string) (*TestResult, error) {
	switch repository.ConnectionType(connType) {
	case repository.ConnectionTypeSonarr:
		client := arrclient.NewSonarrClient(url, apiKey)
		info, err := client.TestConnection(ctx)
		if err != nil {
			return &TestResult{Success: false, Message: err.Error()}, nil
		}
		return &TestResult{Success: true, AppName: info.AppName, Version: info.Version}, nil

	case repository.ConnectionTypeRadarr:
		client := arrclient.NewRadarrClient(url, apiKey)
		info, err := client.TestConnection(ctx)
		if err != nil {
			return &TestResult{Success: false, Message: err.Error()}, nil
		}
		return &TestResult{Success: true, AppName: info.AppName, Version: info.Version}, nil

	case repository.ConnectionTypeEmby:
		client := emby.New(url, apiKey)
		info, err := client.TestConnection(ctx)
		if err != nil {
			return &TestResult{Success: false, Message: err.Error()}, nil
		}
		return &TestResult{
			Success: true,
			AppName: "Emby (" + info.ServerName + ")",
			Version: info.Version,
		}, nil

	default:
		return nil, fmt.Errorf("unknown connection type: %s", connType)
	}
}
