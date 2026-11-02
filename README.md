# Artifact Keeper Docker Desktop Extension

<p align="center">
  <img alt="Artifact Keeper" src="https://raw.githubusercontent.com/artifact-keeper/.github/main/brand/banners/github-readme_1600x600.png" width="100%">
</p>

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

### Automation

Three GitHub Actions workflows keep the extension current:

- **CI** (`ci.yml`): on every PR and push to `main`, builds the extension image, checks `metadata.json` and the extension LABELs, and runs a compose smoke test of the pinned backend/web stack (`.github/smoke/smoke.sh`, also runnable locally).
- **Upstream bump** (`upstream-bump.yml`): every 6 hours (or on demand) runs `scripts/upstream-bump.sh` to find new `artifact-keeper` / `artifact-keeper-web` releases. If the matching Docker Hub images exist, it updates `docker-compose.yaml`, bumps the patch version in `backend/main.go`, adds a changelog entry, builds and smoke-tests, then opens or updates a PR from the `bot/upstream-bump` branch.
- **Release** (`release.yml`): on merge to `main`, if `artifactkeeper/docker-desktop:<version>` is not on Docker Hub yet, pushes `<version>` and `latest` (amd64 + arm64) and creates the `v<version>` GitHub release. Needs the `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` repo secrets.

Trigger a bump by hand (omit either flag to use the latest release), or check locally without editing:

```bash
gh workflow run upstream-bump.yml -f backend_version=1.11.10 -f web_version=1.11.2
scripts/upstream-bump.sh --check
```

## Links

- [Artifact Keeper](https://artifactkeeper.com)
- [Documentation](https://artifactkeeper.com/docs)
- [GitHub](https://github.com/artifact-keeper/artifact-keeper)
- [Report a Bug](https://github.com/artifact-keeper/artifact-keeper/issues)
