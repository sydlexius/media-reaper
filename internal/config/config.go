package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                int
	DBPath              string
	SessionSecret       string //nolint:gosec // config field name, not a hardcoded secret
	AdminUser           string
	AdminPass           string
	SecureCookies       bool
	MasterKey           string //nolint:gosec // config field name, not a hardcoded secret
	HealthCheckInterval time.Duration
}

func Load() *Config {
	cfg := &Config{
		Port:                8080,
		DBPath:              "./data/media-reaper.db",
		SessionSecret:       os.Getenv("MEDIA_REAPER_SESSION_SECRET"),
		AdminUser:           os.Getenv("MEDIA_REAPER_ADMIN_USER"),
		AdminPass:           os.Getenv("MEDIA_REAPER_ADMIN_PASS"),
		SecureCookies:       true,
		MasterKey:           os.Getenv("MEDIA_REAPER_MASTER_KEY"),
		HealthCheckInterval: 5 * time.Minute,
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

	if cfg.MasterKey == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Fatalf("Failed to generate master key: %v", err)
		}
		cfg.MasterKey = hex.EncodeToString(b)
		log.Println("WARNING: No master key configured (MEDIA_REAPER_MASTER_KEY). Generated a random one. Encrypted data will not be recoverable across restarts.")
	}

	if h := os.Getenv("MEDIA_REAPER_HEALTH_CHECK_INTERVAL"); h != "" {
		if d, err := time.ParseDuration(h); err == nil {
			cfg.HealthCheckInterval = d
		}
	}

	return cfg
}
