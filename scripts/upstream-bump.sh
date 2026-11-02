#!/usr/bin/env bash
# Detect new upstream Artifact Keeper releases and bump the extension pins.
#
# Looks up the latest GitHub release of artifact-keeper/artifact-keeper
# (backend) and artifact-keeper/artifact-keeper-web (web), checks that the
# matching Docker Hub tags exist, and compares them with the pins in
# docker-compose.yaml. When something changed it:
#   - rewrites the backend/web `image:` lines in docker-compose.yaml
#   - bumps the patch version in backend/main.go
#   - prepends a changelog entry to the com.docker.extension.changelog LABEL
#
# Usage:
#   scripts/upstream-bump.sh [--check] [--backend X.Y.Z] [--web X.Y.Z] [--root DIR]
#
#   --check        detect only, do not edit any file
#   --backend V    use V for the backend instead of the latest GitHub release
#   --web V        use V for the web UI instead of the latest GitHub release
#   --root DIR     repository root (default: the parent of this script's dir)
#
# Without overrides, a release older than the current pin is ignored (no
# downgrades). Explicit overrides are applied as given.
#
# When GITHUB_OUTPUT is set, writes: changed, backend_old, backend_new,
# backend_changed, web_old, web_new, web_changed, ext_old, ext_new.
#
# Auth: uses `gh api` when gh is available, otherwise curl with GH_TOKEN or
# GITHUB_TOKEN if set.
set -euo pipefail

BACKEND_REPO="artifact-keeper/artifact-keeper"
WEB_REPO="artifact-keeper/artifact-keeper-web"
BACKEND_IMAGE="artifactkeeper/backend"
WEB_IMAGE="artifactkeeper/web"

check_only=false
backend_override=""
web_override=""
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() { sed -n '2,/^set -euo/p' "${BASH_SOURCE[0]}" | sed '$d; s/^# \{0,1\}//'; }

while [ $# -gt 0 ]; do
  case "$1" in
    --check) check_only=true ;;
    --backend) backend_override="${2:?--backend needs a version}"; shift ;;
    --web) web_override="${2:?--web needs a version}"; shift ;;
    --root) root="${2:?--root needs a directory}"; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
  shift
done

compose="$root/docker-compose.yaml"
maingo="$root/backend/main.go"
dockerfile="$root/Dockerfile"
for f in "$compose" "$maingo" "$dockerfile"; do
  [ -f "$f" ] || { echo "error: missing $f" >&2; exit 1; }
done

log() { echo "$*" >&2; }
semver_re='^[0-9]+\.[0-9]+\.[0-9]+$'

# Latest non-draft, non-prerelease release tag of a repo, without leading "v".
latest_release() {
  local repo="$1" tag
  if command -v gh >/dev/null 2>&1; then
    tag="$(gh api "repos/$repo/releases/latest" --jq .tag_name)"
  else
    local auth=()
    local token="${GH_TOKEN:-${GITHUB_TOKEN:-}}"
    [ -n "$token" ] && auth=(-H "Authorization: Bearer $token")
    tag="$(curl -fsSL "${auth[@]}" -H "Accept: application/vnd.github+json" \
      "https://api.github.com/repos/$repo/releases/latest" | jq -r .tag_name)"
  fi
  echo "${tag#v}"
}

# Prints the HTTP status of a Docker Hub tag lookup (200 = exists).
hub_tag_status() {
  curl -s -o /dev/null -w '%{http_code}' --retry 3 --retry-all-errors \
    "https://hub.docker.com/v2/repositories/$1/tags/$2" || true
}

current_pin() {
  local image="$1" v
  v="$(sed -nE "s#^[[:space:]]*image:[[:space:]]*${image}:([^[:space:]]+)[[:space:]]*\$#\\1#p" "$compose")"
  if [ "$(printf '%s\n' "$v" | grep -c .)" -ne 1 ]; then
    log "error: expected exactly one '$image:<tag>' image line in $compose"
    exit 1
  fi
  echo "$v"
}

# version_gt A B: true when A > B (semver, numeric).
version_gt() { [ "$1" != "$2" ] && [ "$(printf '%s\n%s\n' "$1" "$2" | sort -V | tail -n1)" = "$1" ]; }

# resolve NAME REPO IMAGE OLD OVERRIDE -> prints the version to pin.
resolve() {
  local name="$1" repo="$2" image="$3" old="$4" override="$5" new status
  if [ -n "$override" ]; then
    new="${override#v}"
    [[ "$new" =~ $semver_re ]] || { log "error: $name override '$override' is not X.Y.Z"; exit 1; }
    status="$(hub_tag_status "$image" "$new")"
    if [ "$status" != 200 ]; then
      log "error: $name override $new: Docker Hub $image:$new returned HTTP $status"
      exit 1
    fi
    log "$name: $old -> $new (manual override)"
    echo "$new"; return
  fi
  new="$(latest_release "$repo")"
  if ! [[ "$new" =~ $semver_re ]]; then
    log "$name: latest release of $repo is '$new' (not X.Y.Z); keeping $old"
    echo "$old"; return
  fi
  if [ "$new" = "$old" ]; then
    log "$name: $old is the latest release"
    echo "$old"; return
  fi
  if ! version_gt "$new" "$old"; then
    log "$name: latest release $new is older than pin $old; keeping $old"
    echo "$old"; return
  fi
  status="$(hub_tag_status "$image" "$new")"
  if [ "$status" != 200 ]; then
    log "$name: release v$new found but Docker Hub $image:$new returned HTTP $status (image may still be publishing); keeping $old"
    echo "$old"; return
  fi
  log "$name: $old -> $new"
  echo "$new"
}

backend_old="$(current_pin "$BACKEND_IMAGE")"
web_old="$(current_pin "$WEB_IMAGE")"
backend_new="$(resolve backend "$BACKEND_REPO" "$BACKEND_IMAGE" "$backend_old" "$backend_override")"
web_new="$(resolve web "$WEB_REPO" "$WEB_IMAGE" "$web_old" "$web_override")"

ext_old="$(sed -nE 's/^[[:space:]]*version[[:space:]]*=[[:space:]]*"([0-9]+\.[0-9]+\.[0-9]+)".*/\1/p' "$maingo")"
if [ "$(printf '%s\n' "$ext_old" | grep -c .)" -ne 1 ]; then
  log "error: expected exactly one 'version = \"X.Y.Z\"' line in $maingo"
  exit 1
fi
ext_new="${ext_old%.*}.$(( ${ext_old##*.} + 1 ))"

backend_changed=false; [ "$backend_new" != "$backend_old" ] && backend_changed=true
web_changed=false; [ "$web_new" != "$web_old" ] && web_changed=true
changed=false
if $backend_changed || $web_changed; then changed=true; else ext_new="$ext_old"; fi

if [ -n "${GITHUB_OUTPUT:-}" ]; then
  {
    echo "changed=$changed"
    echo "backend_old=$backend_old"; echo "backend_new=$backend_new"; echo "backend_changed=$backend_changed"
    echo "web_old=$web_old"; echo "web_new=$web_new"; echo "web_changed=$web_changed"
    echo "ext_old=$ext_old"; echo "ext_new=$ext_new"
  } >> "$GITHUB_OUTPUT"
fi

if ! $changed; then
  log "Up to date: backend $backend_old, web $web_old. Nothing to do."
  exit 0
fi

log "Extension version: $ext_old -> $ext_new"
if $check_only; then
  log "--check: not editing files."
  exit 0
fi

BACKEND_OLD="$backend_old" BACKEND_NEW="$backend_new" \
WEB_OLD="$web_old" WEB_NEW="$web_new" EXT_NEW="$ext_new" \
COMPOSE="$compose" MAINGO="$maingo" DOCKERFILE="$dockerfile" \
BACKEND_REPO="$BACKEND_REPO" WEB_REPO="$WEB_REPO" \
BACKEND_IMAGE="$BACKEND_IMAGE" WEB_IMAGE="$WEB_IMAGE" \
python3 - <<'PY'
import os, re, sys

e = os.environ

def edit(path, pattern, repl, flags=re.M):
    with open(path, encoding="utf-8") as f:
        text = f.read()
    new, n = re.subn(pattern, repl, text, flags=flags)
    if n != 1:
        sys.exit(f"error: {path}: expected exactly 1 match for {pattern!r}, got {n}")
    with open(path, "w", encoding="utf-8") as f:
        f.write(new)

for image, ver in ((e["BACKEND_IMAGE"], e["BACKEND_NEW"]), (e["WEB_IMAGE"], e["WEB_NEW"])):
    edit(e["COMPOSE"], r"^(\s*image:\s*" + re.escape(image) + r":)\S+[ \t]*$",
         lambda m, v=ver: m.group(1) + v)

edit(e["MAINGO"], r'^(\s*version\s*=\s*")\d+\.\d+\.\d+(")',
     lambda m: m.group(1) + e["EXT_NEW"] + m.group(2))

items = []
for label, repo, old, new in (("Backend", e["BACKEND_REPO"], e["BACKEND_OLD"], e["BACKEND_NEW"]),
                              ("Web", e["WEB_REPO"], e["WEB_OLD"], e["WEB_NEW"])):
    if old != new:
        items.append(f"<li>{label} <a href='https://github.com/{repo}/releases/tag/v{new}'>v{new}</a>"
                     f": upstream update from v{old}</li>")
entry = f"<h2>v{e['EXT_NEW']}</h2><ul>{''.join(items)}</ul>"
assert '"' not in entry and "\n" not in entry
edit(e["DOCKERFILE"], r'(com\.docker\.extension\.changelog=")', lambda m: m.group(1) + entry, flags=0)
PY

log "Updated docker-compose.yaml, backend/main.go and Dockerfile."
