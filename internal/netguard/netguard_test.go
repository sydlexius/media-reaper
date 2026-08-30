package netguard

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr error // nil means "any error", non-nil means "exactly this error"
		ok      bool
	}{
		{name: "valid http LAN url", url: "http://192.168.1.50:8096", ok: true},
		{name: "valid https LAN url", url: "https://10.0.0.5:8989", ok: true},
		{name: "valid hostname", url: "http://emby.local:8096", ok: true},
		{name: "missing scheme", url: "192.168.1.50:8096", ok: false},
		{name: "ftp scheme rejected", url: "ftp://192.168.1.50/", ok: false, wantErr: ErrInvalidScheme},
		{name: "file scheme rejected", url: "file:///etc/passwd", ok: false, wantErr: ErrInvalidScheme},
		{name: "gopher scheme rejected", url: "gopher://192.168.1.50/", ok: false, wantErr: ErrInvalidScheme},
		{name: "embedded credentials rejected", url: "http://user:pass@192.168.1.50", ok: false, wantErr: ErrURLCredentials},
		{name: "embedded username only rejected", url: "http://admin@192.168.1.50", ok: false, wantErr: ErrURLCredentials},
		{name: "empty host rejected", url: "http:///path", ok: false, wantErr: ErrNoHost},
		{name: "unparseable url rejected", url: "http://[::1", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateURL(tt.url)
			if tt.ok && err != nil {
				t.Errorf("ValidateURL(%q) = %v, want no error", tt.url, err)
			}
			if !tt.ok && err == nil {
				t.Errorf("ValidateURL(%q) = nil, want error", tt.url)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateURL(%q) = %v, want error wrapping %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		blocked bool
	}{
		// Blocked: no legitimate connection-target use case.
		{name: "IPv4 loopback", ip: "127.0.0.1", blocked: true},
		{name: "IPv4 loopback range", ip: "127.255.255.255", blocked: true},
		{name: "IPv6 loopback", ip: "::1", blocked: true},
		{name: "IPv4 link-local", ip: "169.254.1.1", blocked: true},
		{name: "cloud metadata endpoint", ip: "169.254.169.254", blocked: true},
		{name: "IPv6 link-local", ip: "fe80::1", blocked: true},
		{name: "unspecified IPv4", ip: "0.0.0.0", blocked: true},
		{name: "unspecified IPv6", ip: "::", blocked: true},
		{name: "IPv4 multicast", ip: "224.0.0.1", blocked: true},
		{name: "IPv6 multicast", ip: "ff02::1", blocked: true},
		// Allowed: LAN-first product, private ranges are legitimate targets.
		{name: "RFC1918 10/8", ip: "10.0.0.5", blocked: false},
		{name: "RFC1918 172.16/12", ip: "172.16.0.5", blocked: false},
		{name: "RFC1918 192.168/16", ip: "192.168.1.50", blocked: false},
		{name: "IPv6 unique local", ip: "fd00::1", blocked: false},
		{name: "public IPv4", ip: "8.8.8.8", blocked: false},
		{name: "public IPv6", ip: "2001:4860:4860::8888", blocked: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("net.ParseIP(%q) returned nil", tt.ip)
			}
			got := isBlockedIP(ip)
			if got != tt.blocked {
				t.Errorf("isBlockedIP(%q) = %v, want %v", tt.ip, got, tt.blocked)
			}
		})
	}
}

// TestHTTPClientBlocksLoopbackDial proves the http.Client returned by
// NewHTTPClient refuses to dial a loopback target end to end (not just that
// isBlockedIP classifies it correctly).
func TestHTTPClientBlocksLoopbackDial(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(5*time.Second, false)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	resp, err := client.Do(req) //nolint:bodyclose // request is expected to fail before a response body exists
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("expected dial to loopback target to fail, got success")
	}
}

// TestHTTPClientAllowsPermittedTarget proves a target that isBlockedIP
// allows completes successfully end to end. It uses an injected blocked
// predicate (via newHTTPClient) that treats the httptest.Server's loopback
// address as allowed, so the test exercises the real dial/redirect
// machinery without requiring outbound network access to a non-loopback
// address in CI.
func TestHTTPClientAllowsPermittedTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	allowAll := func(net.IP) bool { return false }
	client := newHTTPClient(5*time.Second, false, allowAll)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected request to a permitted target to succeed, got: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestHTTPClientRevalidatesRedirects proves a redirect to a target that
// ValidateURL would reject (here, a disallowed scheme injected via a
// malformed Location) is refused rather than followed, even though the
// initial request target is permitted.
func TestHTTPClientRevalidatesRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			w.Header().Set("Location", "gopher://192.168.1.50/evil")
			w.WriteHeader(http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	allowAll := func(net.IP) bool { return false }
	client := newHTTPClient(5*time.Second, false, allowAll)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/start", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	resp, err := client.Do(req) //nolint:bodyclose // request is expected to fail before a usable response body exists
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("expected redirect to a disallowed scheme to be refused, got success")
	}
}

// TestHTTPClientFollowsRedirectToPermittedTarget proves a redirect to a
// target that both ValidateURL and the blocked predicate allow is followed
// successfully, so the redirect handling isn't simply refusing everything.
func TestHTTPClientFollowsRedirectToPermittedTarget(t *testing.T) {
	var finalURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		finalURL = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	allowAll := func(net.IP) bool { return false }
	client := newHTTPClient(5*time.Second, false, allowAll)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/start", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected redirect to a permitted target to succeed, got: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if finalURL != "/final" {
		t.Errorf("redirect target: got %q, want %q", finalURL, "/final")
	}
}
