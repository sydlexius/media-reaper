package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sydlexius/media-reaper/internal/config"
)

type Server struct {
	echo *echo.Echo
	cfg  *config.Config
}

func New(cfg *config.Config) *Server {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	s := &Server{echo: e, cfg: cfg}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	api := s.echo.Group("/api")
	api.GET("/health", s.healthHandler)
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) Start() error {
	return s.echo.Start(fmt.Sprintf(":%d", s.cfg.Port))
}
