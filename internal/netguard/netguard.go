// Package netguard provides SSRF protections for outbound HTTP requests to
// caller-supplied connection URLs (Emby, Sonarr, and Radarr targets).
//
// Policy: media-reaper is a LAN-first product. Users routinely point it at
// RFC1918 addresses (192.168.x.x, 10.x.x.x, 172.16-31.x.x) and IPv6 unique
// local addresses for their Emby/Sonarr/Radarr servers, so a blanket
// private-range block would break normal use of the product. Instead this
// package blocks only address classes that have no legitimate connection
// target use case: loopback, link-local (which covers the
// 169.254.169.254 cloud instance metadata endpoint), the unspecified
// address, and multicast. Private/LAN ranges are allowed by default.
package netguard

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	// dialTimeout bounds a single TCP connection attempt.
	dialTimeout = 10 * time.Second
	// maxRedirects matches net/http's own default redirect cap.
	maxRedirects = 10
)

var (
	// ErrInvalidScheme is returned when a URL is not http or https.
	ErrInvalidScheme = errors.New("url scheme must be http or https")
	// ErrURLCredentials is returned when a URL embeds userinfo (user:pass@host).
	ErrURLCredentials = errors.New("url must not contain embedded credentials")
	// ErrNoHost is returned when a URL has no host component.
	ErrNoHost = errors.New("url must include a host")
	// ErrBlockedHost is returned when a host resolves only to blocked addresses.
	ErrBlockedHost = errors.New("url host resolves to a blocked address (loopback, link-local, unspecified, or multicast)")
)

// ValidateURL parses rawURL and rejects anything that is not a plausible
// outbound HTTP(S) target: a non-http(s) scheme, embedded credentials, or a
// missing host. It does not resolve the host -- IP-level blocking happens
// per-dial in the transport returned by NewHTTPClient, which is what
// actually defends against DNS rebinding (a check made here against a
// resolved IP could go stale before the connection is opened).
func ValidateURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parsing url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, ErrInvalidScheme
	}
	if u.User != nil {
		return nil, ErrURLCredentials
	}
	if u.Hostname() == "" {
		return nil, ErrNoHost
	}
	return u, nil
}

// isBlockedIP reports whether ip belongs to an address class that must never
// be an outbound connection target, regardless of the LAN-first policy:
// loopback, link-local unicast/multicast (which covers the
// 169.254.169.254 cloud metadata endpoint), the unspecified address, and
// multicast. RFC1918 private ranges and IPv6 unique-local addresses are
// deliberately NOT blocked -- see the package doc comment.
func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified()
}

// safeDialContext returns a DialContext that resolves the target host and
// dials a specific resolved IP, chosen and validated in the same call. This
// closes the DNS-rebinding gap: there is no separate earlier lookup whose
// result could differ from the address actually connected to, because the
// address that gets validated here is the address that gets dialed. This
// runs for every connection net/http opens, including one opened to follow
// a redirect, so redirect targets get the same IP-level enforcement as the
// original request.
func safeDialContext(dialer *net.Dialer, blocked func(net.IP) bool) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("splitting host/port: %w", err)
		}

		ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
		if err != nil {
			return nil, fmt.Errorf("resolving host: %w", err)
		}

		for _, ip := range ips {
			if blocked(ip) {
				continue
			}
			return dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		}

		return nil, fmt.Errorf("%w: %s", ErrBlockedHost, host)
	}
}

// checkRedirect re-validates the scheme, credentials, and host of every
// redirect hop so a 3xx response cannot be used to steer the request at a
// URL that ValidateURL would have rejected up front. Per-hop IP-level
// blocking is still enforced by the transport's DialContext on the actual
// connection opened to follow the redirect.
func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirects {
		return fmt.Errorf("stopped after %d redirects", maxRedirects)
	}
	if _, err := ValidateURL(req.URL.String()); err != nil {
		return fmt.Errorf("redirect target rejected: %w", err)
	}
	return nil
}

// NewHTTPClient returns an http.Client hardened against SSRF: every dial
// (including one triggered by a followed redirect) resolves and validates
// the destination IP immediately before connecting to it, and every
// redirect hop is re-validated at the URL level before it is followed.
//
// insecureSkipVerify controls TLS certificate verification. It exists
// because the Sonarr/Radarr client library (golift.io/starr) defaults to
// skipping verification for these LAN-deployed services, which commonly run
// self-signed certificates; callers must pass the same value they used
// before adopting this client to avoid a functional regression.
func NewHTTPClient(timeout time.Duration, insecureSkipVerify bool) *http.Client {
	return newHTTPClient(timeout, insecureSkipVerify, isBlockedIP)
}

// newHTTPClient is the internal constructor behind NewHTTPClient. It takes
// an explicit blocked predicate so tests can exercise the redirect/dial
// plumbing against a permissive policy without depending on real network
// access to a non-loopback address.
func newHTTPClient(timeout time.Duration, insecureSkipVerify bool, blocked func(net.IP) bool) *http.Client {
	dialer := &net.Dialer{Timeout: dialTimeout}

	transport := &http.Transport{
		DialContext: safeDialContext(dialer, blocked),
	}
	if insecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // matches golift.io/starr's existing default for LAN-deployed *arr instances with self-signed certificates
	}

	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: checkRedirect,
		Transport:     transport,
	}
}
