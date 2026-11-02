#!/usr/bin/env bash
# Sanity-check a built extension image: metadata.json parses and has the keys
# Docker Desktop needs, the files it references exist in the image, and the
# required extension LABELs are present (JSON-valued labels must parse).
#
# This stands in for `docker extension validate`, which needs a running Docker
# Desktop (it talks to the extension-manager socket) and so cannot run on a
# plain CI runner.
#
# Usage: scripts/check-extension-image.sh IMAGE
set -euo pipefail

image="${1:?usage: $0 IMAGE}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
labels="$(docker image inspect --format '{{json .Config.Labels}}' "$image")"
files="$(docker run --rm --entrypoint sh "$image" -c 'ls -1 / /ui' 2>/dev/null || true)"

LABELS="$labels" FILES="$files" METADATA="$root/metadata.json" MAINGO="$root/backend/main.go" python3 - <<'PY'
import json, os, re, sys

errors = []
labels = json.loads(os.environ["LABELS"] or "{}") or {}
files = set(os.environ["FILES"].split())

try:
    meta = json.load(open(os.environ["METADATA"]))
except Exception as exc:
    sys.exit(f"::error::metadata.json does not parse: {exc}")

for key in ("icon", "ui", "vm"):
    if key not in meta:
        errors.append(f"metadata.json: missing '{key}'")
for name in (meta.get("icon"), meta.get("vm", {}).get("composefile"), "metadata.json"):
    if name and name not in files:
        errors.append(f"image: /{name} missing")
if "index.html" not in files:
    errors.append("image: /ui/index.html missing")

required = [
    "org.opencontainers.image.title",
    "org.opencontainers.image.description",
    "org.opencontainers.image.vendor",
    "com.docker.desktop.extension.api.version",
    "com.docker.desktop.extension.icon",
    "com.docker.extension.categories",
    "com.docker.extension.publisher-url",
    "com.docker.extension.changelog",
    "com.docker.extension.detailed-description",
    "com.docker.extension.screenshots",
    "com.docker.extension.additional-urls",
]
for key in required:
    if not labels.get(key):
        errors.append(f"label {key} missing or empty")
for key in ("com.docker.extension.screenshots", "com.docker.extension.additional-urls"):
    try:
        json.loads(labels.get(key) or "")
    except ValueError:
        errors.append(f"label {key} is not valid JSON")

changelog = labels.get("com.docker.extension.changelog", "")
m = re.match(r"<h2>v(\d+\.\d+\.\d+)</h2>", changelog)
version = re.search(r'^\s*version\s*=\s*"([^"]+)"', open(os.environ["MAINGO"]).read(), re.M)
if not m:
    errors.append("changelog label does not start with <h2>vX.Y.Z</h2>")
elif version and m.group(1) != version.group(1):
    print(f"::warning::changelog top entry is v{m.group(1)} but backend/main.go version is {version.group(1)}")

for e in errors:
    print(f"::error::{e}")
if errors:
    sys.exit(1)
print(f"Extension image OK ({len(labels)} labels, metadata.json valid)")
PY
