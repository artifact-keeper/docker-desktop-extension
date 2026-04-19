# Artifact Keeper Docker Desktop Extension

Run a full [Artifact Keeper](https://artifactkeeper.com) artifact registry locally with one click. Supports 45+ package formats including Maven, npm, PyPI, Docker/OCI, Cargo, Go, NuGet, Helm, and more.

## Install

```bash
docker extension install artifactkeeper/docker-desktop:1.0.0
```

Or search for "Artifact Keeper" in the Docker Desktop Extensions Marketplace.

## What's Included

**Core services (always running):**

| Service | Purpose |
|---------|---------|
| Backend | Artifact Keeper API server (Rust) |
| PostgreSQL 16 | Metadata storage |
| Web UI | Next.js frontend |
| Meilisearch | Full-text search |

**Optional services (toggle in settings):**

| Service | Purpose | RAM |
|---------|---------|-----|
| Trivy | Vulnerability scanning | ~512 MB |
| OpenSCAP | Compliance scanning | ~256 MB |
| Dependency-Track | SBOM analysis | ~4 GB |
| Jaeger | Distributed tracing | ~256 MB |

## Features

- **One-click install**: full registry stack running in under a minute
- **Service health dashboard**: real-time status of all services
- **Upgrade detection**: checks Docker Hub for newer versions and upgrades in-place
- **Settings panel**: configure port, network access, toggle optional services
- **Admin credentials**: auto-generated secure password shown on first run

## Default Credentials

On first install, the extension generates a random admin password. Find it in the extension dashboard or settings panel.

- **Username:** `admin`
- **Password:** shown in the extension UI (copy button provided)

## Development

### Build and install locally

```bash
make build-extension
make install-extension
```

### Frontend hot reload

```bash
cd ui && npm install && npm run dev
make dev-ui-source
```

### Debug with Chrome DevTools

```bash
make dev-debug
```

### Update after code changes

```bash
make update-extension
```

### Uninstall

```bash
make uninstall
```

## Architecture

The extension packages a Go backend (settings API, health checking, upgrade detection) and a React frontend (MUI, Docker Desktop theme) inside an Alpine image. Service containers are defined in `docker-compose.yaml` with Docker Compose profiles for optional services.

```
Extension Image (Alpine)
  +-- Go backend (Unix socket API)
  +-- React UI (embedded in Docker Desktop tab)
  +-- docker-compose.yaml (service definitions)
  +-- metadata.json (Docker Desktop config)

Compose Services:
  postgres -> backend -> web
              |
              +-> meilisearch
              +-> trivy (optional)
              +-> openscap (optional)
              +-> dependency-track (optional)
              +-> jaeger (optional)
```

## Publishing

```bash
# Build and push multi-arch (amd64 + arm64)
make push-extension
```

## Links

- [Artifact Keeper](https://artifactkeeper.com)
- [Documentation](https://artifactkeeper.com/docs)
- [GitHub](https://github.com/artifact-keeper/artifact-keeper)
- [Report a Bug](https://github.com/artifact-keeper/artifact-keeper/issues)
