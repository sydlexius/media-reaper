package connection

import (
	"context"
	"log"
	"time"

	"github.com/sydlexius/media-reaper/internal/repository"
)

// HealthChecker periodically tests all enabled connections.
type HealthChecker struct {
	repo      repository.ConnectionRepository
	encryptor *Encryptor
	interval  time.Duration
}

// NewHealthChecker creates a health checker.
func NewHealthChecker(
	repo repository.ConnectionRepository,
	encryptor *Encryptor,
	interval time.Duration,
) *HealthChecker {
	return &HealthChecker{repo: repo, encryptor: encryptor, interval: interval}
}

// Start runs the health check loop until the context is cancelled.
func (h *HealthChecker) Start(ctx context.Context) {
	h.checkAll(ctx)

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Health checker stopped")
			return
		case <-ticker.C:
			h.checkAll(ctx)
		}
	}
}

func (h *HealthChecker) checkAll(ctx context.Context) {
	connections, err := h.repo.GetAllEnabled(ctx)
	if err != nil {
		log.Printf("Health check: failed to fetch connections: %v", err)
		return
	}

	for _, conn := range connections {
		if ctx.Err() != nil {
			return
		}

		apiKey, err := h.encryptor.Decrypt(conn.EncryptedAPIKey)
		if err != nil {
			log.Printf("Health check: failed to decrypt key for %s: %v", conn.Name, err)
			h.updateStatus(ctx, conn.ID, repository.ConnectionStatusUnhealthy)
			continue
		}

		checkCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		result, testErr := TestConnection(checkCtx, string(conn.Type), conn.URL, apiKey)
		cancel()

		if testErr != nil || !result.Success {
			log.Printf("Health check: %s (%s) unhealthy", conn.Name, conn.Type)
			h.updateStatus(ctx, conn.ID, repository.ConnectionStatusUnhealthy)
		} else {
			h.updateStatus(ctx, conn.ID, repository.ConnectionStatusHealthy)
		}
	}
}

func (h *HealthChecker) updateStatus(ctx context.Context, id string, status repository.ConnectionStatus) {
	now := time.Now().UTC().Format(time.RFC3339)
	if err := h.repo.UpdateStatus(ctx, id, status, now); err != nil {
		log.Printf("Health check: failed to update status for %s: %v", id, err)
	}
}
