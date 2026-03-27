#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT_DIR="${4:-$ROOT_DIR/dist}"

fail() {
  printf '[bundle] %s\n' "$*" >&2
  exit 1
}

TARGET="${1:-}"
GUI_ARTIFACT="${2:-}"
CLI_ARTIFACT="${3:-}"

[[ -n "$TARGET" && -n "$GUI_ARTIFACT" && -n "$CLI_ARTIFACT" ]] || fail \
  "Usage: scripts/release/package_bundle.sh <darwin|linux|windows> <gui_artifact> <cli_artifact> [out_dir]"

[[ -e "$GUI_ARTIFACT" ]] || fail "Missing GUI artifact: $GUI_ARTIFACT"
[[ -e "$CLI_ARTIFACT" ]] || fail "Missing CLI artifact: $CLI_ARTIFACT"

mkdir -p "$OUT_DIR"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

case "$TARGET" in
  darwin)
    APP_DIR="$TMP_DIR/Billy.app"
    cp -R "$GUI_ARTIFACT" "$APP_DIR"
    mkdir -p "$APP_DIR/Contents/Resources"
    install -m 0755 "$CLI_ARTIFACT" "$APP_DIR/Contents/Resources/billy"
    install -m 0755 "$CLI_ARTIFACT" "$TMP_DIR/billy"
    tar -czf "$OUT_DIR/billy-macos-universal.tar.gz" -C "$TMP_DIR" Billy.app billy
    ;;
  linux)
    install -m 0755 "$GUI_ARTIFACT" "$TMP_DIR/Billy"
    install -m 0755 "$CLI_ARTIFACT" "$TMP_DIR/billy"
    tar -czf "$OUT_DIR/billy-linux-amd64-bundle.tar.gz" -C "$TMP_DIR" Billy billy
    ;;
  windows)
    cp "$GUI_ARTIFACT" "$TMP_DIR/Billy.exe"
    cp "$CLI_ARTIFACT" "$TMP_DIR/billy.exe"
    (cd "$TMP_DIR" && zip -q "$OUT_DIR/billy-windows-amd64-bundle.zip" Billy.exe billy.exe)
    ;;
  *)
    fail "Unknown target: $TARGET"
    ;;
esac

printf '[bundle] Wrote bundle(s) to %s\n' "$OUT_DIR"
