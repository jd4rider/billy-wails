#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
GUI_BINARY="${1:-}"
CLI_BINARY="${2:-}"
PACKAGER="${PACKAGER:-deb}"
VERSION="${VERSION:-0.2.0-dev}"
ARCH="${ARCH:-amd64}"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/dist}"

[[ -n "$GUI_BINARY" && -f "$GUI_BINARY" ]] || { echo "Usage: build_linux_deb.sh <Billy> <billy>" >&2; exit 1; }
[[ -n "$CLI_BINARY" && -f "$CLI_BINARY" ]] || { echo "Usage: build_linux_deb.sh <Billy> <billy>" >&2; exit 1; }

if ! command -v nfpm >/dev/null 2>&1; then
  echo "nfpm is required to build Linux packages." >&2
  exit 1
fi

mkdir -p "$OUT_DIR"

GUI_BINARY="$GUI_BINARY" \
CLI_BINARY="$CLI_BINARY" \
VERSION="$VERSION" \
GOARCH="$ARCH" \
nfpm pkg \
  --packager "$PACKAGER" \
  --config "$ROOT_DIR/packaging/linux/nfpm.yaml" \
  --target "$OUT_DIR/billy-desktop-linux-$ARCH.$PACKAGER"

echo "Created $OUT_DIR/billy-desktop-linux-$ARCH.$PACKAGER"
