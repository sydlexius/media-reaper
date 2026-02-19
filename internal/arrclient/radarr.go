package arrclient

import (
	"context"
	"fmt"

	"golift.io/starr"
	"golift.io/starr/radarr"
)

// RadarrClient wraps the golift/starr radarr client.
type RadarrClient struct {
	client *radarr.Radarr
}

// NewRadarrClient creates a Radarr client from a URL and decrypted API key.
func NewRadarrClient(url, apiKey string) *RadarrClient {
	config := starr.New(apiKey, url, defaultTimeout)
	return &RadarrClient{client: radarr.New(config)}
}

// TestConnection verifies connectivity by calling GetSystemStatus.
func (r *RadarrClient) TestConnection(ctx context.Context) (*SystemInfo, error) {
	status, err := r.client.GetSystemStatusContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("radarr connection test failed: %w", err)
	}
	return &SystemInfo{
		AppName: status.AppName,
		Version: status.Version,
	}, nil
}

// GetMovie returns movies from Radarr. Pass nil to get all movies.
func (r *RadarrClient) GetMovie(ctx context.Context, params *radarr.GetMovie) ([]*radarr.Movie, error) {
	movies, err := r.client.GetMovieContext(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("getting movies: %w", err)
	}
	return movies, nil
}

// EditMovies performs bulk edits on movies (e.g., toggling monitored status).
func (r *RadarrClient) EditMovies(ctx context.Context, edit *radarr.BulkEdit) ([]*radarr.Movie, error) {
	movies, err := r.client.EditMoviesContext(ctx, edit)
	if err != nil {
		return nil, fmt.Errorf("editing movies: %w", err)
	}
	return movies, nil
}

// DeleteMovie removes a movie with options for deleting files and adding an import exclusion.
func (r *RadarrClient) DeleteMovie(ctx context.Context, movieID int64, deleteFiles, addImportExclusion bool) error {
	err := r.client.DeleteMovieContext(ctx, movieID, deleteFiles, addImportExclusion)
	if err != nil {
		return fmt.Errorf("deleting movie %d: %w", movieID, err)
	}
	return nil
}

// GetRemotePathMappings returns the remote path mappings configured in Radarr.
func (r *RadarrClient) GetRemotePathMappings(ctx context.Context) ([]*starr.RemotePathMapping, error) {
	mappings, err := r.client.GetRemotePathMappingsContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting remote path mappings: %w", err)
	}
	return mappings, nil
}
