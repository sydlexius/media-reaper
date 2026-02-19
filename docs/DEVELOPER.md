# Media Reaper Developer Guide

## Architecture

*Coming soon*

## Setup

### Prerequisites

- Go 1.23+
- Node.js 22+
- Docker (optional, for containerized development)

### Local Development

```bash
# Install Go dependencies
go mod download

# Install frontend dependencies
cd web && npm install && cd ..

# Start backend with hot-reload (requires air)
air -c .air.toml

# Or start without hot-reload
MEDIA_REAPER_ADMIN_USER=admin MEDIA_REAPER_ADMIN_PASS=changeme \
  MEDIA_REAPER_SECURE_COOKIES=false \
  go run -tags dev ./cmd/server/

# Start frontend dev server
cd web && npm run dev
```

### Docker Development

```bash
cd docker
docker compose up
```

### Pre-commit Hooks

```bash
pip install pre-commit
pre-commit install
```

## API Reference

API documentation is available via Swagger UI at `/api/docs/` when running in development mode.

Regenerate Swagger docs:

```bash
swag init -g cmd/server/main.go -o docs
```

## Database Schema

Managed via goose migrations in `internal/db/migrations/`.

## Contributing

*Coming soon*
