package connection

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sydlexius/media-reaper/internal/repository"
)

type createConnectionRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	URL    string `json:"url"`
	APIKey string `json:"apiKey"` //nolint:gosec // request DTO, not a hardcoded secret
}

type updateConnectionRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	URL     string `json:"url"`
	APIKey  string `json:"apiKey,omitempty"` //nolint:gosec // request DTO, not a hardcoded secret
	Enabled *bool  `json:"enabled,omitempty"`
}

type connectionResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	URL           string `json:"url"`
	MaskedAPIKey  string `json:"maskedApiKey"`
	Enabled       bool   `json:"enabled"`
	Status        string `json:"status"`
	LastCheckedAt string `json:"lastCheckedAt,omitempty"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

func (s *Service) toResponse(conn *repository.Connection) connectionResponse {
	resp := connectionResponse{
		ID:        conn.ID,
		Name:      conn.Name,
		Type:      string(conn.Type),
		URL:       conn.URL,
		MaskedAPIKey: MaskAPIKey(conn.EncryptedAPIKey, s.encryptor),
		Enabled:   conn.Enabled,
		Status:    string(conn.Status),
		CreatedAt: conn.CreatedAt,
		UpdatedAt: conn.UpdatedAt,
	}
	if conn.LastCheckedAt != nil {
		resp.LastCheckedAt = *conn.LastCheckedAt
	}
	return resp
}

// CreateHandler creates a new connection.
// @Summary Create connection
// @Description Create a new Sonarr, Radarr, or Emby connection with encrypted API key
// @Tags connections
// @Accept json
// @Produce json
// @Param request body createConnectionRequest true "Connection details"
// @Success 201 {object} connectionResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security SessionCookie
// @Router /connections [post]
func (s *Service) CreateHandler(c echo.Context) error {
	var req createConnectionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.Name == "" || req.Type == "" || req.URL == "" || req.APIKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name, type, url, and apiKey are required"})
	}

	if !isValidConnectionType(req.Type) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "type must be sonarr, radarr, or emby"})
	}

	conn, err := s.Create(c.Request().Context(), req.Name, req.Type, req.URL, req.APIKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create connection"})
	}

	return c.JSON(http.StatusCreated, s.toResponse(conn))
}

// ListHandler lists all connections.
// @Summary List connections
// @Description List all connections with masked API keys
// @Tags connections
// @Produce json
// @Success 200 {array} connectionResponse
// @Security SessionCookie
// @Router /connections [get]
func (s *Service) ListHandler(c echo.Context) error {
	connections, err := s.GetAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list connections"})
	}

	responses := make([]connectionResponse, 0, len(connections))
	for _, conn := range connections {
		responses = append(responses, s.toResponse(conn))
	}

	return c.JSON(http.StatusOK, responses)
}

// GetHandler returns a single connection.
// @Summary Get connection
// @Description Get a connection by ID with masked API key
// @Tags connections
// @Produce json
// @Param id path string true "Connection ID"
// @Success 200 {object} connectionResponse
// @Failure 404 {object} map[string]string
// @Security SessionCookie
// @Router /connections/{id} [get]
func (s *Service) GetHandler(c echo.Context) error {
	id := c.Param("id")
	conn, err := s.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get connection"})
	}
	if conn == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "connection not found"})
	}

	return c.JSON(http.StatusOK, s.toResponse(conn))
}

// UpdateHandler updates an existing connection.
// @Summary Update connection
// @Description Update a connection. API key is optional (omit to keep current).
// @Tags connections
// @Accept json
// @Produce json
// @Param id path string true "Connection ID"
// @Param request body updateConnectionRequest true "Connection details"
// @Success 200 {object} connectionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security SessionCookie
// @Router /connections/{id} [put]
func (s *Service) UpdateHandler(c echo.Context) error {
	id := c.Param("id")

	var req updateConnectionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.Name == "" || req.Type == "" || req.URL == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name, type, and url are required"})
	}

	if !isValidConnectionType(req.Type) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "type must be sonarr, radarr, or emby"})
	}

	conn, err := s.Update(c.Request().Context(), id, req.Name, req.Type, req.URL, req.APIKey, req.Enabled)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update connection"})
	}
	if conn == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "connection not found"})
	}

	return c.JSON(http.StatusOK, s.toResponse(conn))
}

// DeleteHandler deletes a connection.
// @Summary Delete connection
// @Description Delete a connection by ID
// @Tags connections
// @Param id path string true "Connection ID"
// @Success 204
// @Failure 500 {object} map[string]string
// @Security SessionCookie
// @Router /connections/{id} [delete]
func (s *Service) DeleteHandler(c echo.Context) error {
	id := c.Param("id")
	if err := s.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete connection"})
	}
	return c.NoContent(http.StatusNoContent)
}

func isValidConnectionType(t string) bool {
	switch repository.ConnectionType(t) {
	case repository.ConnectionTypeSonarr, repository.ConnectionTypeRadarr, repository.ConnectionTypeEmby:
		return true
	}
	return false
}
