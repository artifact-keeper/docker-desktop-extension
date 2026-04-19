FROM golang:1.24-alpine AS builder
ENV CGO_ENABLED=0
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
    com.docker.extension.changelog="Initial release" \
    com.docker.extension.detailed-description="Artifact Keeper is a universal artifact registry that supports 45+ package formats including Maven, npm, PyPI, Docker/OCI, Cargo, Go, NuGet, Helm, and more. This extension runs the full stack locally with one click: backend API, web UI, PostgreSQL, Meilisearch (full-text search), and optional vulnerability scanning via Trivy and OpenSCAP. Perfect for local development, CI/CD testing, or evaluating Artifact Keeper before deploying to production." \
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
