package netguard

import (
	"context"
	"errors"
	"fmt"
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

	client := NewHTTPClient(5*time.Second, false, true)
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

// TestSafeDialContextFallsBackToNextPermittedIP proves that when a host
// resolves to multiple permitted addresses and the first one fails to
// dial, safeDialContext tries the next one rather than failing outright.
// Each dial targets a specific IP literal, so net/http's own multi-address
// fallback (which only engages when dialing a hostname) never runs here --
// this behavior has to be provided by safeDialContext itself.
func TestSafeDialContextFallsBackToNextPermittedIP(t *testing.T) {
	unreachable := net.ParseIP("192.0.2.1") // TEST-NET-1 (RFC 5737), not blocked by policy
	reachable := net.ParseIP("192.0.2.2")

	lookup := func(context.Context, string) ([]net.IP, error) {
		return []net.IP{unreachable, reachable}, nil
	}

	var dialedIPs []string
	dial := func(_ context.Context, _, addr string) (net.Conn, error) {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			t.Fatalf("splitting dialed addr %q: %v", addr, err)
		}
		dialedIPs = append(dialedIPs, host)
		if host == unreachable.String() {
			return nil, fmt.Errorf("simulated dial failure for %s", host)
		}
		return &net.TCPConn{}, nil
	}

	allowAll := func(net.IP) bool { return false }
	dialCtx := safeDialContext(dial, lookup, allowAll)

	conn, err := dialCtx(context.Background(), "tcp", "example.invalid:80")
	if err != nil {
		t.Fatalf("expected fallback to the second permitted IP to succeed, got: %v", err)
	}
	if conn == nil {
		t.Fatal("expected a non-nil connection")
	}

	want := []string{unreachable.String(), reachable.String()}
	if len(dialedIPs) != len(want) || dialedIPs[0] != want[0] || dialedIPs[1] != want[1] {
		t.Errorf("dial order: got %v, want %v", dialedIPs, want)
	}
}

// TestSafeDialContextDistinguishesBlockedFromUnreachable proves the two
// failure modes stay distinct: ErrBlockedHost when every resolved address
// was blocked by policy (nothing was ever dialed), versus a plain dial
// error when a permitted address was tried and failed to connect.
func TestSafeDialContextDistinguishesBlockedFromUnreachable(t *testing.T) {
	t.Run("all resolved addresses blocked", func(t *testing.T) {
		lookup := func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}
		dialCalled := false
		dial := func(context.Context, string, string) (net.Conn, error) {
			dialCalled = true
			return nil, errors.New("should not be called")
		}

		dialCtx := safeDialContext(dial, lookup, isBlockedIP)
		_, err := dialCtx(context.Background(), "tcp", "example.invalid:80")
		if !errors.Is(err, ErrBlockedHost) {
			t.Errorf("expected ErrBlockedHost, got %v", err)
		}
		if dialCalled {
			t.Error("dial should never be called when every resolved address is blocked")
		}
	})

	t.Run("permitted address unreachable", func(t *testing.T) {
		lookup := func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("192.0.2.1")}, nil
		}
		wantErr := errors.New("connection refused")
		dial := func(context.Context, string, string) (net.Conn, error) {
			return nil, wantErr
		}

		allowAll := func(net.IP) bool { return false }
		dialCtx := safeDialContext(dial, lookup, allowAll)
		_, err := dialCtx(context.Background(), "tcp", "example.invalid:80")
		if errors.Is(err, ErrBlockedHost) {
			t.Error("expected a dial error, not ErrBlockedHost, for a permitted-but-unreachable address")
		}
		if !errors.Is(err, wantErr) {
			t.Errorf("expected the underlying dial error to be wrapped, got %v", err)
		}
	})
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
	client := newHTTPClient(5*time.Second, false, true, allowAll)

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
	client := newHTTPClient(5*time.Second, false, true, allowAll)

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
	client := newHTTPClient(5*time.Second, false, true, allowAll)

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

// TestHTTPClientRefusesRedirectsWhenNotFollowing proves a client built with
// followRedirects=false (the policy used for the Sonarr/Radarr clients, to
// preserve starr.Client's own never-follow behavior) does not follow a
// redirect to an otherwise-permitted target, and instead returns the 3xx
// response itself.
func TestHTTPClientRefusesRedirectsWhenNotFollowing(t *testing.T) {
	redirectTargetHit := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		redirectTargetHit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	allowAll := func(net.IP) bool { return false }
	client := newHTTPClient(5*time.Second, false, false, allowAll)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/start", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected the client to return the redirect response rather than error, got: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("status: got %d, want %d (redirect should not have been followed)", resp.StatusCode, http.StatusFound)
	}
	if redirectTargetHit {
		t.Error("redirect target was hit; a followRedirects=false client must not follow redirects")
	}
}
