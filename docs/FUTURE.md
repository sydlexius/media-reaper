# Future Features

This document tracks features that are planned but not yet implemented. These are deliberately deferred from the current scope to keep v1 focused and manageable.

## Authentication

- OIDC/OAuth2 support for enterprise SSO
- Emby admin credential passthrough authentication

## Non-Admin Features

- "Leaving Soon" Emby collection (created via POST /Collections)
- Keep-request system (non-admin users request, admin approves/denies)

## Automation

- Scheduled automatic rule execution (cron-style with optional approval queue)
- Emby webhook integration for real-time watch status updates

## Media Management

- Direct Emby deletion for unmanaged items (DELETE /Items/{Id})
- Poster view with watch progress bars
- Streaming availability checks before deletion

## Infrastructure

- PostgreSQL support (behind existing repository interfaces)
- Prometheus metrics endpoint
- Helm chart for Kubernetes deployment
