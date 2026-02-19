# media-reaper

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

A Docker-based web application that integrates Sonarr/Radarr with Emby to identify watched media eligible for deletion. media-reaper helps home media server administrators reclaim storage by automatically tracking watch status across users and applying configurable rules to flag content for cleanup.

## Features

- **Multi-instance support** - Connect multiple Sonarr, Radarr, and Emby instances
- **Watch tracking** - Track per-user watch status across all Emby users
- **Rules engine** - Configurable rule sets with watch thresholds, temporal criteria, metadata filters, and genre exclusions
- **Grace periods** - Flagged items enter a configurable grace period before becoming actionable
- **Safe deletion** - Always deletes through Sonarr/Radarr (never directly through Emby) with pre-action freshness checks
- **Bulk operations** - Select and act on multiple items with per-item result reporting
- **Dashboard** - Storage breakdown, watch analytics, and "quick wins" panel
- **Dark/light theme** - System preference detection with manual toggle
- **Single binary** - React frontend embedded in Go binary via Docker distroless image

## Quick Start

```bash
docker run -d \
  --name media-reaper \
  -p 8080:8080 \
  -v media-reaper-data:/data \
  -e MEDIA_REAPER_ADMIN_USER=admin \
  -e MEDIA_REAPER_ADMIN_PASS=changeme \
  ghcr.io/sydlexius/media-reaper:latest
```

Or with Docker Compose:

```yaml
services:
  media-reaper:
    image: ghcr.io/sydlexius/media-reaper:latest
    ports:
      - "8080:8080"
    volumes:
      - media-reaper-data:/data
    environment:
      - MEDIA_REAPER_ADMIN_USER=admin
      - MEDIA_REAPER_ADMIN_PASS=changeme

volumes:
  media-reaper-data:
```

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `MEDIA_REAPER_PORT` | `8080` | Server port |
| `MEDIA_REAPER_DB_PATH` | `./data/media-reaper.db` | SQLite database file path |
| `MEDIA_REAPER_ADMIN_USER` | (none) | Initial admin username (first run only) |
| `MEDIA_REAPER_ADMIN_PASS` | (none) | Initial admin password (first run only) |
| `MEDIA_REAPER_SESSION_SECRET` | (random) | Session cookie encryption key |
| `MEDIA_REAPER_SECURE_COOKIES` | `true` | Set to `false` for non-HTTPS environments |

## Screenshots

*Coming soon*

## Development

See [docs/DEVELOPER.md](docs/DEVELOPER.md) for setup instructions.

```bash
# Clone and install dependencies
git clone https://github.com/sydlexius/media-reaper.git
cd media-reaper
go mod download
cd web && npm install && cd ..

# Start backend (port 8080)
MEDIA_REAPER_ADMIN_USER=admin MEDIA_REAPER_ADMIN_PASS=changeme \
  MEDIA_REAPER_SECURE_COOKIES=false \
  go run -tags dev ./cmd/server/

# Start frontend (port 5173, proxies API to 8080)
cd web && npm run dev
```

## API Documentation

When running in development mode, Swagger UI is available at `http://localhost:8080/api/docs/`.

## Acknowledgements

The following projects informed the design of media-reaper:

| Project | License | Contribution |
|---------|---------|-------------|
| [Maintainerr](https://github.com/Maintainerr/Maintainerr) | MIT | Rule composition UI patterns, grace-period concept |
| [MUMC](https://github.com/terrelsa13/MUMC) | GPL-3.0 | Multi-user Emby watch-status logic |
| [EmbyArrSync](https://github.com/Amateur-God/EmbyArrSync) | GPL-3.0 | Emby-to-*arr unmonitor workflow |
| [Jellysweep](https://github.com/jon4hz/jellysweep) | GPL-3.0 | Keep-request UX, Go architecture |
| [Janitorr](https://github.com/Schaka/janitorr) | GPL-3.0 | Per-season TV deletion patterns |
| [rfsbraz/deleterr](https://github.com/rfsbraz/deleterr) | MIT | Smart exclusion patterns |
| [golift/starr](https://github.com/golift/starr) | MIT | Go Sonarr/Radarr API client library |
| [jellyfin-plugin-media-cleaner](https://github.com/shemanaev/jellyfin-plugin-media-cleaner) | MIT | Favorites-as-protection pattern |

## License

This project is licensed under the GNU General Public License v3.0. See [LICENSE](LICENSE) for details.
