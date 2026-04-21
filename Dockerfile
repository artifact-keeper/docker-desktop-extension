FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder
ARG TARGETOS TARGETARCH
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH
WORKDIR /backend
COPY backend/go.* .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download
COPY backend/. .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o bin/service

FROM --platform=$BUILDPLATFORM node:24-alpine AS client-builder
WORKDIR /ui
# cache packages in layer
COPY ui/package.json /ui/package.json
COPY ui/package-lock.json /ui/package-lock.json
RUN --mount=type=cache,target=/usr/src/app/.npm \
    npm set cache /usr/src/app/.npm && \
    npm ci
# install
COPY ui /ui
RUN npm run build

FROM alpine
LABEL org.opencontainers.image.title="Artifact Keeper" \
    org.opencontainers.image.description="One-click local artifact registry with 45+ package formats, search, and vulnerability scanning." \
    org.opencontainers.image.vendor="Artifact Keeper" \
    org.opencontainers.image.url="https://artifactkeeper.com" \
    com.docker.desktop.extension.api.version="0.4.2" \
    com.docker.desktop.extension.icon="https://raw.githubusercontent.com/artifact-keeper/docker-desktop-extension/main/icon.png" \
    com.docker.extension.categories="development-tools" \
    com.docker.extension.publisher-url="https://github.com/artifact-keeper" \
    com.docker.extension.changelog="<h2>v1.0.6</h2><ul><li>Backend <a href='https://github.com/artifact-keeper/artifact-keeper/releases/tag/v1.1.7'>v1.1.7</a>: auto-version from git tags, clippy fixes</li><li>Web <a href='https://github.com/artifact-keeper/artifact-keeper-web/releases/tag/v1.1.3'>v1.1.3</a>: version display fix, Docker cache bust</li><li>Fixed storage permission issues on fresh installs</li><li>Fixed web healthcheck (node-based, no curl/wget needed)</li><li>Semver-only upgrade detection (ignores SHA/branch tags)</li><li>Numeric version comparison (1.1.0 not flagged as upgrade for 1.1.2)</li></ul><h2>v1.0.0</h2><ul><li>Backend <a href='https://github.com/artifact-keeper/artifact-keeper/releases/tag/v1.1.6'>v1.1.6</a>: PyPI normalization, npm version endpoint, Go/Terraform/Swift 404 fixes</li><li>Web <a href='https://github.com/artifact-keeper/artifact-keeper-web/releases/tag/v1.1.0'>v1.1.0</a>: 20 issues closed, password policy, quarantine status, auth badges</li><li>Initial Docker Desktop extension release</li><li>Service health dashboard with real-time status</li><li>Settings panel: port config, LAN toggle, optional service toggles</li><li>Auto-generated admin credentials with copy button</li></ul>" \
    com.docker.extension.detailed-description="<h2>Artifact Keeper</h2><p>Run a complete artifact registry locally with one click. Supports <b>45+ package formats</b> including Maven, npm, PyPI, Docker/OCI, Cargo, Go, NuGet, Helm, Conan, and many more.</p><h3>What's included</h3><ul><li><b>Backend API</b> (Rust) with 45+ format handlers</li><li><b>Web UI</b> (Next.js) for browsing and managing artifacts</li><li><b>PostgreSQL 16</b> for metadata storage</li><li><b>Meilisearch</b> for full-text search across all packages</li></ul><h3>Optional services (toggle in settings)</h3><ul><li><b>Trivy</b> for vulnerability scanning</li><li><b>OpenSCAP</b> for compliance scanning</li><li><b>Dependency-Track</b> for SBOM analysis</li><li><b>Jaeger</b> for distributed tracing</li></ul><h3>Features</h3><ul><li>One-click install, ready in under a minute</li><li>Real-time service health monitoring</li><li>Upgrade detection for all services</li><li>Configurable port and network access</li><li>Auto-generated secure admin credentials</li></ul><p>Perfect for local development, CI/CD testing, air-gapped environments, or evaluating Artifact Keeper before deploying to production.</p>" \
    com.docker.extension.screenshots='[{"alt":"Dashboard showing services and health status","url":"https://raw.githubusercontent.com/artifact-keeper/docker-desktop-extension/main/screenshots/dashboard.png"}]' \
    com.docker.extension.additional-urls='[{"title":"Documentation","url":"https://artifactkeeper.com/docs"},{"title":"GitHub","url":"https://github.com/artifact-keeper/artifact-keeper"},{"title":"Report a Bug","url":"https://github.com/artifact-keeper/artifact-keeper/issues"}]'

RUN apk add --no-cache curl
COPY --from=builder /backend/bin/service /
COPY docker-compose.yaml .
COPY metadata.json .
COPY icon.png .
COPY --from=client-builder /ui/build ui
VOLUME /data/config
CMD ["/service", "-socket", "/run/guest-services/backend.sock"]
