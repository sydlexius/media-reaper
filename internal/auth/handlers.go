package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"` //nolint:gosec // login request DTO, not a hardcoded secret
}

type userResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// LoginHandler authenticates a user and creates a session.
// @Summary Login
// @Description Authenticate with username and password to create a session
// @Tags auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login credentials"
// @Success 200 {object} userResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 429 {object} map[string]string
// @Router /auth/login [post]
func (s *Service) LoginHandler(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username and password are required"})
	}

	user, err := s.Authenticate(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	if err := s.CreateSession(c.Response(), c.Request(), user.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
	}

	return c.JSON(http.StatusOK, userResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
}

// LogoutHandler destroys the current session.
// @Summary Logout
// @Description Destroy the current session and clear the session cookie
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (s *Service) LogoutHandler(c echo.Context) error {
	if err := s.DestroySession(c.Response(), c.Request()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to destroy session"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "logged out"})
}

// MeHandler returns the currently authenticated user.
// @Summary Current user
// @Description Get the currently authenticated user from the session
// @Tags auth
// @Produce json
// @Success 200 {object} userResponse
// @Failure 401 {object} map[string]string
// @Security SessionCookie
// @Router /auth/me [get]
func (s *Service) MeHandler(c echo.Context) error {
	user, err := s.GetUserFromSession(c.Request().Context(), c.Request())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
	}

	return c.JSON(http.StatusOK, userResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
}
