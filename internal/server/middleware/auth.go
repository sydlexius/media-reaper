package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sydlexius/media-reaper/internal/auth"
)

func RequireAuth(authService *auth.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, err := authService.GetUserFromSession(c.Request().Context(), c.Request())
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid session"})
			}
			if user == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
			}

			c.Set("user", user)
			return next(c)
		}
	}
}
