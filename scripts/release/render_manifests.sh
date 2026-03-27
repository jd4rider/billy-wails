#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
RELEASE_REPO="${RELEASE_REPO:-jd4rider/billy-wails}"
HOMEBREW_REPO="${HOMEBREW_REPO:-$ROOT_DIR/../homebrew-billy}"
SCOOP_REPO="${SCOOP_REPO:-$ROOT_DIR/../scoop-billy}"

fail() {
  printf '[release] %s\n' "$*" >&2
  exit 1
}

sha256_file() {
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
  elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    fail "No SHA256 tool found. Install shasum or sha256sum."
  fi
}

VERSION_INPUT="${1:-}"
MACOS_ARCHIVE="${2:-}"
WINDOWS_BUNDLE="${3:-}"

[[ -n "$VERSION_INPUT" && -n "$MACOS_ARCHIVE" && -n "$WINDOWS_BUNDLE" ]] || fail \
  "Usage: scripts/release/render_manifests.sh <version> <macos-tar.gz> <windows-bundle.zip>"

[[ -f "$MACOS_ARCHIVE" ]] || fail "Missing macOS archive: $MACOS_ARCHIVE"
[[ -f "$WINDOWS_BUNDLE" ]] || fail "Missing Windows bundle: $WINDOWS_BUNDLE"

VERSION="${VERSION_INPUT#v}"
MACOS_SHA="$(sha256_file "$MACOS_ARCHIVE")"
WINDOWS_SHA="$(sha256_file "$WINDOWS_BUNDLE")"

mkdir -p "$ROOT_DIR/packaging/homebrew/Casks" "$ROOT_DIR/packaging/scoop"

cat >"$ROOT_DIR/packaging/homebrew/Casks/billy.rb" <<EOF
# typed: false
# frozen_string_literal: true

cask "billy" do
  version "$VERSION"
  sha256 "$MACOS_SHA"

  url "https://github.com/$RELEASE_REPO/releases/download/v#{version}/billy-macos-universal.tar.gz"
  name "Billy"
  desc "Local AI coding assistant desktop app with bundled Billy CLI"
  homepage "https://billysh.online"

  app "Billy.app"
  binary "billy"
end
EOF

cat >"$ROOT_DIR/packaging/scoop/billy.json" <<EOF
{
  "version": "$VERSION",
  "description": "Local AI coding assistant desktop app with bundled Billy CLI.",
  "homepage": "https://billysh.online",
  "license": "MIT",
  "architecture": {
    "64bit": {
      "url": "https://github.com/$RELEASE_REPO/releases/download/v$VERSION/billy-windows-amd64-bundle.zip",
      "hash": "$WINDOWS_SHA",
      "bin": [
        "billy.exe"
      ],
      "shortcuts": [
        [
          "Billy.exe",
          "Billy"
        ]
      ]
    }
  }
}
EOF

if [[ -d "$HOMEBREW_REPO" ]]; then
  mkdir -p "$HOMEBREW_REPO/Casks"
  cp "$ROOT_DIR/packaging/homebrew/Casks/billy.rb" "$HOMEBREW_REPO/Casks/billy.rb"
fi

if [[ -d "$SCOOP_REPO" ]]; then
  mkdir -p "$SCOOP_REPO/bucket"
  cp "$ROOT_DIR/packaging/scoop/billy.json" "$SCOOP_REPO/bucket/billy.json"
fi

printf '[release] Updated Billy desktop manifests for %s against %s\n' "$VERSION" "$RELEASE_REPO"
