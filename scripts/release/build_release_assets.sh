#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BILLY_APP_DIR="${BILLY_APP_DIR:-$ROOT_DIR/../billy-app}"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/dist}"
VERSION="${VERSION:-0.2.0-dev}"
TARGET="${1:-darwin}"
TARGET_ARCH="${2:-}"

PATH="/usr/local/opt/node@20/bin:/usr/local/opt/node@23/bin:/opt/homebrew/opt/node@20/bin:/opt/homebrew/opt/node@23/bin:$PATH"

if command -v wails3 >/dev/null 2>&1; then
  WAILS3_BIN="$(command -v wails3)"
elif [[ -x "$HOME/go/bin/wails3" ]]; then
  WAILS3_BIN="$HOME/go/bin/wails3"
else
  echo "wails3 is required on PATH or at \$HOME/go/bin/wails3" >&2
  exit 1
fi

case "$TARGET" in
  darwin|linux|windows) ;;
  *)
    echo "Usage: scripts/release/build_release_assets.sh <darwin|linux|windows> [amd64|arm64]" >&2
    exit 1
    ;;
esac

case "${TARGET_ARCH:-$(uname -m)}" in
  x86_64) TARGET_ARCH="amd64" ;;
  aarch64|arm64) TARGET_ARCH="arm64" ;;
  amd64|arm64) ;;
  *)
    echo "Unsupported target arch: ${TARGET_ARCH}" >&2
    exit 1
    ;;
esac

[[ -d "$BILLY_APP_DIR" ]] || {
  echo "Billy app repo not found: $BILLY_APP_DIR" >&2
  exit 1
}

mkdir -p "$OUT_DIR"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

case "$TARGET" in
  windows)
    CLI_OUT="$TMP_DIR/billy.exe"
    ;;
  *)
    CLI_OUT="$TMP_DIR/billy"
    ;;
esac

echo "[release] Testing billy-app"
(
  cd "$BILLY_APP_DIR"
  go test ./...
)

echo "[release] Building billy CLI for $TARGET/$TARGET_ARCH -> $CLI_OUT"
(
  cd "$BILLY_APP_DIR"
  GOOS="$TARGET" GOARCH="$TARGET_ARCH" CGO_ENABLED=0 \
    go build -trimpath -ldflags="-s -w" -o "$CLI_OUT" ./cmd/billy
)

echo "[release] Building Billy Desktop for $TARGET"
(
  cd "$ROOT_DIR"
  case "$TARGET" in
    darwin)
      "$WAILS3_BIN" package
      GUI_OUT="$ROOT_DIR/bin/Billy.app"
      ;;
    windows)
      "$WAILS3_BIN" task windows:build
      GUI_OUT="$ROOT_DIR/bin/Billy.exe"
      ;;
    linux)
      "$WAILS3_BIN" task linux:build
      GUI_OUT="$ROOT_DIR/bin/Billy"
      ;;
  esac

  echo "[release] Creating bundle archive"
  scripts/release/package_bundle.sh "$TARGET" "$GUI_OUT" "$CLI_OUT" "$OUT_DIR"

  if [[ "$TARGET" == "darwin" ]]; then
    echo "[release] Creating macOS pkg"
    VERSION="$VERSION" scripts/release/build_macos_pkg.sh "$GUI_OUT" "$CLI_OUT"
  fi
)

echo "[release] Wrote assets to $OUT_DIR"
