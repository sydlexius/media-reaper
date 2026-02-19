package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          int
	DBPath        string
	SessionSecret string
	AdminUser     string
	AdminPass     string
	SecureCookies bool
}

func Load() *Config {
	cfg := &Config{
		Port:          8080,
		DBPath:        "./data/media-reaper.db",
		SessionSecret: os.Getenv("MEDIA_REAPER_SESSION_SECRET"),
		AdminUser:     os.Getenv("MEDIA_REAPER_ADMIN_USER"),
		AdminPass:     os.Getenv("MEDIA_REAPER_ADMIN_PASS"),
		SecureCookies: true,
	}

	if p := os.Getenv("MEDIA_REAPER_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			cfg.Port = v
		}
	}

	if d := os.Getenv("MEDIA_REAPER_DB_PATH"); d != "" {
		cfg.DBPath = d
	}

	if os.Getenv("MEDIA_REAPER_SECURE_COOKIES") == "false" {
		cfg.SecureCookies = false
	}

	return cfg
}
