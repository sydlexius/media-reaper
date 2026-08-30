package connection

import (
	"context"
	"fmt"

	"github.com/sydlexius/media-reaper/internal/arrclient"
	"github.com/sydlexius/media-reaper/internal/emby"
	"github.com/sydlexius/media-reaper/internal/netguard"
	"github.com/sydlexius/media-reaper/internal/repository"
)

// TestResult represents the outcome of a connection test.
type TestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	AppName string `json:"appName,omitempty"`
	Version string `json:"version,omitempty"`
}

// TestConnection tests connectivity for a given type, URL, and plaintext API
// key. This is the single choke point all three connection types (emby,
// sonarr, radarr) and all three call sites (the unsaved test handler, the
// saved-connection test handler, and the periodic health checker) pass
// through, so the URL is validated here rather than duplicated per client.
// Only the coarse, always-true properties of the URL are checked here
// (scheme, embedded credentials, presence of a host); IP-level SSRF and
// DNS-rebinding defense happens per-dial in the http clients constructed
// by the emby and arrclient packages via netguard.NewHTTPClient, because a
// hostname can legitimately resolve to a different address between this
// check and the connection actually being opened.
func TestConnection(ctx context.Context, connType, url, apiKey string) (*TestResult, error) {
	if _, err := netguard.ValidateURL(url); err != nil {
		return &TestResult{Success: false, Message: err.Error()}, nil
	}

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
