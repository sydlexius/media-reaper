package arrclient

import (
	"context"
	"fmt"

	"golift.io/starr"
	"golift.io/starr/sonarr"
)

// SonarrClient wraps the golift/starr sonarr client.
type SonarrClient struct {
	client *sonarr.Sonarr
}

// NewSonarrClient creates a Sonarr client from a URL and decrypted API key.
func NewSonarrClient(url, apiKey string) *SonarrClient {
	config := starr.New(apiKey, url, defaultTimeout)
	return &SonarrClient{client: sonarr.New(config)}
}

// TestConnection verifies connectivity by calling GetSystemStatus.
func (s *SonarrClient) TestConnection(ctx context.Context) (*SystemInfo, error) {
	status, err := s.client.GetSystemStatusContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("sonarr connection test failed: %w", err)
	}
	return &SystemInfo{
		AppName: status.AppName,
		Version: status.Version,
	}, nil
}

// GetAllSeries returns all series from Sonarr.
func (s *SonarrClient) GetAllSeries(ctx context.Context) ([]*sonarr.Series, error) {
	series, err := s.client.GetAllSeriesContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting all series: %w", err)
	}
	return series, nil
}

// GetSeriesEpisodes returns episodes for a series.
func (s *SonarrClient) GetSeriesEpisodes(ctx context.Context, seriesID int64) ([]*sonarr.Episode, error) {
	episodes, err := s.client.GetSeriesEpisodesContext(ctx, &sonarr.GetEpisode{
		SeriesID: seriesID,
	})
	if err != nil {
		return nil, fmt.Errorf("getting episodes for series %d: %w", seriesID, err)
	}
	return episodes, nil
}

// MonitorEpisode sets monitoring status for episodes.
func (s *SonarrClient) MonitorEpisode(ctx context.Context, episodeIDs []int64, monitor bool) error {
	_, err := s.client.MonitorEpisodeContext(ctx, episodeIDs, monitor)
	if err != nil {
		return fmt.Errorf("setting episode monitor status: %w", err)
	}
	return nil
}

// DeleteEpisodeFile removes an episode file by its ID.
func (s *SonarrClient) DeleteEpisodeFile(ctx context.Context, episodeFileID int64) error {
	err := s.client.DeleteEpisodeFileContext(ctx, episodeFileID)
	if err != nil {
		return fmt.Errorf("deleting episode file %d: %w", episodeFileID, err)
	}
	return nil
}

// DeleteSeries removes a series with options for deleting files and adding an import exclusion.
func (s *SonarrClient) DeleteSeries(ctx context.Context, seriesID int, deleteFiles, addImportExclude bool) error {
	err := s.client.DeleteSeriesContext(ctx, seriesID, deleteFiles, addImportExclude)
	if err != nil {
		return fmt.Errorf("deleting series %d: %w", seriesID, err)
	}
	return nil
}

// GetRemotePathMappings returns the remote path mappings configured in Sonarr.
func (s *SonarrClient) GetRemotePathMappings(ctx context.Context) ([]*starr.RemotePathMapping, error) {
	mappings, err := s.client.GetRemotePathMappingsContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting remote path mappings: %w", err)
	}
	return mappings, nil
}
