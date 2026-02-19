# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Go backend with Echo server and health endpoint
- SQLite database with goose migrations and repository interfaces
- Backend authentication with bcrypt and session cookies
- React frontend with Vite, TypeScript, and Tailwind CSS v4
- Login page and auth state management with TanStack Query
- App shell with sidebar navigation and dark/light theme
- Docker multi-stage build (node, go, distroless)
- Development docker-compose with hot-reload
- Pre-commit hooks (golangci-lint, eslint, prettier, gitleaks)
- GitHub Actions CI pipeline (lint, test, Docker build)
- OpenAPI documentation scaffolding with Swagger UI
- Bruno API collection for auth and health endpoints
- README with acknowledgements and skeleton documentation
