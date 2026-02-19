package server

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/sydlexius/media-reaper/internal/auth"
	"github.com/sydlexius/media-reaper/internal/config"
	authmw "github.com/sydlexius/media-reaper/internal/server/middleware"
	"github.com/sydlexius/media-reaper/web"

	_ "github.com/sydlexius/media-reaper/docs"
)

type Server struct {
	echo        *echo.Echo
	cfg         *config.Config
	authService *auth.Service
}

func New(cfg *config.Config, authService *auth.Service) *Server {
	e := echo.New()
	e.HideBanner = true

	e.Use(echomw.Logger()) //nolint:staticcheck // TODO(#59): replace with slog RequestLogger
	e.Use(echomw.Recover())

	s := &Server{echo: e, cfg: cfg, authService: authService}
	s.registerRoutes()
	s.registerSPA()
	return s
}

func (s *Server) registerRoutes() {
	api := s.echo.Group("/api")

	// Public routes
	api.GET("/health", s.healthHandler)
	api.GET("/docs/*", echoSwagger.WrapHandler)

	// Auth routes (public, login has rate limiting)
	authGroup := api.Group("/auth")
	authGroup.POST("/login", s.authService.LoginHandler, s.loginRateLimiter())
	authGroup.POST("/logout", s.authService.LogoutHandler)
	authGroup.GET("/me", s.authService.MeHandler)

	// Protected routes (for future use)
	_ = api.Group("", authmw.RequireAuth(s.authService))
}

func (s *Server) registerSPA() {
	distFS, err := fs.Sub(web.Assets, "dist")
	if err != nil {
		log.Println("No embedded frontend assets found (dev mode)")
		return
	}

	entries, err := fs.ReadDir(distFS, ".")
	if err != nil || len(entries) == 0 {
		log.Println("No embedded frontend assets found (dev mode)")
		return
	}

	log.Println("Serving embedded frontend assets")
	fileServer := http.FileServer(http.FS(distFS))

	s.echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			// Skip API routes
			if strings.HasPrefix(path, "/api") {
				return next(c)
			}

			// Try to serve static file
			if path != "/" {
				f, err := distFS.(fs.ReadFileFS).ReadFile(strings.TrimPrefix(path, "/"))
				if err == nil && f != nil {
					fileServer.ServeHTTP(c.Response(), c.Request())
					return nil
				}
			}

			// SPA fallback: serve index.html for all non-file routes
			c.Request().URL.Path = "/"
			fileServer.ServeHTTP(c.Response(), c.Request())
			return nil
		}
	})
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

// healthHandler returns the service health status.
// @Summary Health check
// @Description Returns service health status
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) Start() error {
	return s.echo.Start(fmt.Sprintf(":%d", s.cfg.Port))
}
