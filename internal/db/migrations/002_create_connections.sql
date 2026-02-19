-- +goose Up
CREATE TABLE connections (
    id                TEXT PRIMARY KEY,
    name              TEXT NOT NULL,
    type              TEXT NOT NULL CHECK(type IN ('sonarr', 'radarr', 'emby')),
    url               TEXT NOT NULL,
    encrypted_api_key TEXT NOT NULL,
    enabled           INTEGER NOT NULL DEFAULT 1,
    status            TEXT NOT NULL DEFAULT 'unknown' CHECK(status IN ('healthy', 'unhealthy', 'unknown')),
    last_checked_at   TIMESTAMP,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_connections_type ON connections(type);
CREATE INDEX idx_connections_enabled ON connections(enabled);

-- +goose Down
DROP INDEX IF EXISTS idx_connections_enabled;
DROP INDEX IF EXISTS idx_connections_type;
DROP TABLE IF EXISTS connections;
