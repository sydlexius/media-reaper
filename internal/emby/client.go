package emby

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sydlexius/media-reaper/internal/netguard"
)

const defaultTimeout = 30 * time.Second

// Client provides access to the Emby REST API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New creates an Emby API client. The caller-supplied baseURL is expected to
// have already been validated (scheme, credentials, host) by
// netguard.ValidateURL at the connection-test choke point
// (internal/connection/tester.go); the http client returned here additionally
// resolves and validates the destination IP immediately before every dial
// (including redirect hops), which is the layer that actually defends
// against SSRF and DNS rebinding. Emby does not default to self-signed
// certificates the way the *arr stack does, so TLS verification stays on.
// This client had no CheckRedirect before this fix (net/http's default
// follow-and-revalidate behavior), so following redirects with per-hop
// validation here is a real policy choice, not a preserved regression.
func New(baseURL, apiKey string) *Client {
	return newWithHTTPClient(baseURL, apiKey, netguard.NewHTTPClient(defaultTimeout, false, true))
}

// newWithHTTPClient builds a Client around an explicit http.Client. It
// exists so the package's own unit tests can exercise request-building and
// response-decoding logic against an httptest.Server, which always listens
// on loopback and would otherwise be rejected by the SSRF-guarded client
// New returns. SSRF/redirect/DNS-rebinding policy itself is covered by
// internal/netguard's tests, not here.
func newWithHTTPClient(baseURL, apiKey string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

// TestConnection verifies connectivity by fetching system info.
func (c *Client) TestConnection(ctx context.Context) (*SystemInfo, error) {
	var info SystemInfo
	if err := c.get(ctx, "/System/Info/Public", nil, &info); err != nil {
		return nil, fmt.Errorf("emby connection test failed: %w", err)
	}
	return &info, nil
}

// GetUsers returns all users (requires admin API key).
func (c *Client) GetUsers(ctx context.Context) ([]*User, error) {
	var users []*User
	if err := c.get(ctx, "/Users", nil, &users); err != nil {
		return nil, fmt.Errorf("getting users: %w", err)
	}
	return users, nil
}

// GetLibraries returns all media libraries (folders).
func (c *Client) GetLibraries(ctx context.Context) ([]Library, error) {
	var resp MediaFoldersResponse
	if err := c.get(ctx, "/Library/MediaFolders", nil, &resp); err != nil {
		return nil, fmt.Errorf("getting libraries: %w", err)
	}
	return resp.Items, nil
}

// GetUserItems returns items for a specific user with optional query parameters.
func (c *Client) GetUserItems(ctx context.Context, userID string, params *ItemQuery) (*ItemsResult, error) {
	qp := make(map[string]string)
	if params != nil {
		if params.ParentID != "" {
			qp["ParentId"] = params.ParentID
		}
		if params.IncludeTypes != "" {
			qp["IncludeItemTypes"] = params.IncludeTypes
		}
		if params.Recursive {
			qp["Recursive"] = "true"
		}
		if params.Limit > 0 {
			qp["Limit"] = strconv.Itoa(params.Limit)
		}
		if params.StartIndex > 0 {
			qp["StartIndex"] = strconv.Itoa(params.StartIndex)
		}
		if params.SortBy != "" {
			qp["SortBy"] = params.SortBy
		}
		if params.SortOrder != "" {
			qp["SortOrder"] = params.SortOrder
		}
		if params.IsPlayed != nil {
			qp["IsPlayed"] = strconv.FormatBool(*params.IsPlayed)
		}
		if params.Fields != "" {
			qp["Fields"] = params.Fields
		}
	}

	var result ItemsResult
	if err := c.get(ctx, "/Users/"+userID+"/Items", qp, &result); err != nil {
		return nil, fmt.Errorf("getting user items: %w", err)
	}
	return &result, nil
}

// get performs a GET request with the Emby API key header and decodes the JSON response.
func (c *Client) get(ctx context.Context, path string, queryParams map[string]string, result any) error {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("X-Emby-Token", c.apiKey)
	req.Header.Set("Accept", "application/json")

	if len(queryParams) > 0 {
		q := req.URL.Query()
		for k, v := range queryParams {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.httpClient.Do(req) //nolint:gosec // G704: gosec's taint analysis cannot see into c.httpClient's Transport; the SSRF/DNS-rebinding defense is netguard.NewHTTPClient's DialContext, which resolves and validates every destination IP (including redirect hops) immediately before dialing it -- see internal/netguard/netguard.go. CodeQL alert #3 on this same line is dismissed by the PR author with the same rationale, not fixed by code.
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	return nil
}
