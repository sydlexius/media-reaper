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
)

const defaultTimeout = 30 * time.Second

// Client provides access to the Emby REST API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New creates an Emby API client.
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: defaultTimeout},
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

	resp, err := c.httpClient.Do(req) //nolint:gosec // URL is constructed from admin-configured baseURL, not user input
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
