package connection

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestTestConnectionRejectsBlockedURLs proves the shared TestConnection
// choke point rejects SSRF-relevant URLs for all three connection types
// before any outbound request is attempted, and that it does so via a
// failed TestResult (matching how a real connectivity failure is reported)
// rather than a Go error, since a validation failure is exactly the kind
// of thing TestUnsavedHandler and TestSavedHandler already turn into a
// user-facing "Message" field.
func TestTestConnectionRejectsBlockedURLs(t *testing.T) {
	tests := []struct {
		name          string
		connType      string
		url           string
		wantMsgSubstr string // asserted with strings.Contains against result.Message
	}{
		{name: "emby loopback", connType: "emby", url: "http://127.0.0.1:8096", wantMsgSubstr: "blocked address"},
		{name: "sonarr loopback", connType: "sonarr", url: "http://localhost:8989", wantMsgSubstr: "blocked address"},
		{name: "radarr loopback", connType: "radarr", url: "http://127.0.0.1:7878", wantMsgSubstr: "blocked address"},
		{name: "emby cloud metadata", connType: "emby", url: "http://169.254.169.254/latest/meta-data/", wantMsgSubstr: "blocked address"},
		{name: "sonarr cloud metadata", connType: "sonarr", url: "http://169.254.169.254/", wantMsgSubstr: "blocked address"},
		{name: "radarr cloud metadata", connType: "radarr", url: "http://169.254.169.254/", wantMsgSubstr: "blocked address"},
		{name: "emby non-http scheme", connType: "emby", url: "gopher://192.168.1.50/", wantMsgSubstr: "url scheme must be"},
		{name: "emby embedded credentials", connType: "emby", url: "http://user:pass@192.168.1.50", wantMsgSubstr: "embedded credentials"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TestConnection(context.Background(), tt.connType, tt.url, "test-key")
			if err != nil {
				t.Fatalf("TestConnection returned an error (want a failed TestResult instead): %v", err)
			}
			if result.Success {
				t.Errorf("TestConnection(%q, %q) succeeded, want rejection", tt.connType, tt.url)
			}
			// Assert the specific rejection reason, not just that something
			// failed: a plain connection-refused error also produces a
			// non-empty message, so a bare emptiness check would keep
			// passing even if the actual blocking logic regressed entirely.
			if !strings.Contains(result.Message, tt.wantMsgSubstr) {
				t.Errorf("TestConnection(%q, %q) message = %q, want it to contain %q", tt.connType, tt.url, result.Message, tt.wantMsgSubstr)
			}
		})
	}
}

// TestTestConnectionAllowsLANURLs proves normal LAN-deployed Emby/Sonarr/
// Radarr URLs are not rejected by the URL-shape validation at the choke
// point. It does not require the target to actually be reachable: an
// unreachable private-range host still passes validation and fails later,
// for a different (connectivity) reason, which is the behavior the LAN-
// first policy requires.
func TestTestConnectionAllowsLANURLs(t *testing.T) {
	tests := []struct {
		name     string
		connType string
		url      string
	}{
		{name: "emby 192.168", connType: "emby", url: "http://192.168.1.50:8096"},
		{name: "sonarr 10.x", connType: "sonarr", url: "http://10.0.0.5:8989"},
		{name: "radarr 172.16", connType: "radarr", url: "http://172.16.0.5:7878"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Short deadline: these hosts are not reachable in the test
			// environment, and the point of the test is the validation
			// outcome, not how long an unreachable dial takes to time out.
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			result, err := TestConnection(ctx, tt.connType, tt.url, "test-key")
			if err != nil {
				t.Fatalf("TestConnection returned an unexpected error: %v", err)
			}
			// The host is not actually reachable in this test environment, so
			// the connection attempt itself is expected to fail -- what this
			// test asserts is that it fails with a connectivity message, not
			// a "blocked address" or "invalid scheme" validation rejection.
			if result.Success {
				t.Fatalf("unexpectedly succeeded connecting to %q; test assumes it is unreachable", tt.url)
			}
			for _, blockedSubstr := range []string{"blocked address", "url scheme must be", "embedded credentials", "must include a host"} {
				if strings.Contains(result.Message, blockedSubstr) {
					t.Errorf("LAN url %q was rejected by URL validation (message: %q), want a connectivity failure instead", tt.url, result.Message)
				}
			}
		})
	}
}
