#!/usr/bin/env bash
# Smoke test of the pinned upstream stack (postgres, meilisearch, backend,
# web) from docker-compose.yaml, without the Docker Desktop extension-service.
#
# Passes when backend /health returns 200 or 503 and web / returns 200
# within SMOKE_TIMEOUT seconds. Dumps `docker compose logs` on failure.
#
# Env:
#   SMOKE_TIMEOUT   seconds to wait for both endpoints (default 300)
#   SMOKE_PROJECT   compose project name (default ak-smoke)
#   SMOKE_KEEP=1    leave the stack running afterwards
#   WEB_PORT        host port web is published on (default 3000, from compose)
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
timeout="${SMOKE_TIMEOUT:-300}"
project="${SMOKE_PROJECT:-ak-smoke}"
web_port="${WEB_PORT:-3000}"
services=(postgres meilisearch smoke-seed backend web)

dc() {
  docker compose -p "$project" \
    -f "$root/docker-compose.yaml" -f "$root/.github/smoke/compose.smoke.yaml" "$@"
}

# shellcheck disable=SC2329  # invoked via trap
cleanup() {
  local rc=$?
  if [ "$rc" -ne 0 ]; then
    echo "::group::docker compose ps"; dc ps -a || true; echo "::endgroup::"
    echo "::group::docker compose logs"; dc logs --no-color --timestamps || true; echo "::endgroup::"
  fi
  if [ "${SMOKE_KEEP:-0}" != 1 ]; then
    dc down -v --remove-orphans >/dev/null 2>&1 || true
  fi
  exit "$rc"
}
trap cleanup EXIT

echo "Pinned images:"
grep -E '^[[:space:]]*image:[[:space:]]*artifactkeeper/(backend|web):' "$root/docker-compose.yaml"

dc config --quiet
dc pull --quiet "${services[@]}"
dc up -d "${services[@]}"

deadline=$(( $(date +%s) + timeout ))
backend_ok=false
web_ok=false
backend_code=000
web_code=000
while [ "$(date +%s)" -lt "$deadline" ]; do
  if ! $backend_ok; then
    backend_code="$(dc exec -T backend curl -s -o /dev/null -w '%{http_code}' \
      "http://localhost:${PORT:-8080}/health" 2>/dev/null || true)"
    case "$backend_code" in 200|503) backend_ok=true; echo "backend /health -> $backend_code" ;; esac
  fi
  if ! $web_ok; then
    web_code="$(curl -sL -o /dev/null -w '%{http_code}' --max-time 5 \
      "http://localhost:${web_port}/" 2>/dev/null || true)"
    [ "$web_code" = 200 ] && { web_ok=true; echo "web / -> $web_code"; }
  fi
  if $backend_ok && $web_ok; then
    echo "Smoke test passed."
    exit 0
  fi
  sleep 5
done

echo "::error::Smoke test timed out after ${timeout}s (backend /health=${backend_code:-none}, web /=${web_code:-none})"
exit 1
