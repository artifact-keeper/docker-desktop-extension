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
    com.docker.extension.categories="development-tools" \
    com.docker.extension.publisher-url="https://github.com/artifact-keeper" \
    com.docker.extension.changelog="Initial release"

RUN apk add --no-cache curl
COPY --from=builder /backend/bin/service /
COPY docker-compose.yaml .
COPY metadata.json .
COPY icon.png .
COPY --from=client-builder /ui/build ui
VOLUME /data/config
CMD ["/service", "-socket", "/run/guest-services/backend.sock"]
