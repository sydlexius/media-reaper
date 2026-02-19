package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/sydlexius/media-reaper/internal/auth"
	"github.com/sydlexius/media-reaper/internal/config"
	authmw "github.com/sydlexius/media-reaper/internal/server/middleware"
)

type Server struct {
	echo        *echo.Echo
	cfg         *config.Config
	authService *auth.Service
}

func New(cfg *config.Config, authService *auth.Service) *Server {
	e := echo.New()
	e.HideBanner = true

	e.Use(echomw.Logger())
	e.Use(echomw.Recover())

	s := &Server{echo: e, cfg: cfg, authService: authService}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	api := s.echo.Group("/api")

	// Public routes
	api.GET("/health", s.healthHandler)

	// Auth routes (public, login has rate limiting)
	authGroup := api.Group("/auth")
	authGroup.POST("/login", s.authService.LoginHandler, s.loginRateLimiter())
	authGroup.POST("/logout", s.authService.LogoutHandler)
	authGroup.GET("/me", s.authService.MeHandler)

	// Protected routes (for future use)
	_ = api.Group("", authmw.RequireAuth(s.authService))
}

func (s *Server) loginRateLimiter() echo.MiddlewareFunc {
	rateLimiterConfig := echomw.RateLimiterConfig{
		Skipper: echomw.DefaultSkipper,
		Store: echomw.NewRateLimiterMemoryStoreWithConfig(
			echomw.RateLimiterMemoryStoreConfig{
				Rate:      5.0 / 60.0,
				Burst:     5,
				ExpiresIn: 1 * time.Minute,
			},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			return ctx.RealIP(), nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "rate limit error"})
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "too many login attempts, please try again later",
			})
		},
	}
	return echomw.RateLimiterWithConfig(rateLimiterConfig)
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) Start() error {
	return s.echo.Start(fmt.Sprintf(":%d", s.cfg.Port))
}
