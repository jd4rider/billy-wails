#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
APP_BUNDLE="${1:-$ROOT_DIR/build/bin/Billy.app}"
CLI_BINARY="${2:-}"
VERSION="${VERSION:-0.1.0-dev}"
PKG_IDENTIFIER="${PKG_IDENTIFIER:-online.billysh.installer}"
PKG_OUTPUT="${PKG_OUTPUT:-$ROOT_DIR/dist/billy-macos-universal.pkg}"
APP_INSTALL_DIR="${APP_INSTALL_DIR:-/Applications}"

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "This script must run on macOS." >&2
  exit 1
fi

[[ -d "$APP_BUNDLE" ]] || { echo "Missing app bundle: $APP_BUNDLE" >&2; exit 1; }
[[ -n "$CLI_BINARY" && -f "$CLI_BINARY" ]] || { echo "Usage: build_macos_pkg.sh <Billy.app> <billy>" >&2; exit 1; }

mkdir -p "$(dirname "$PKG_OUTPUT")"

PKG_ROOT="$(mktemp -d)"
trap 'rm -rf "$PKG_ROOT"' EXIT

mkdir -p "$PKG_ROOT$APP_INSTALL_DIR" "$PKG_ROOT/usr/local/bin"

cp -R "$APP_BUNDLE" "$PKG_ROOT$APP_INSTALL_DIR/Billy.app"
cp "$CLI_BINARY" "$PKG_ROOT/usr/local/bin/billy"
chmod 755 "$PKG_ROOT/usr/local/bin/billy"

cat >"$PKG_ROOT/usr/local/bin/billy-desktop" <<EOF
#!/usr/bin/env bash
exec open "$APP_INSTALL_DIR/Billy.app" --args "\$@"
EOF
chmod 755 "$PKG_ROOT/usr/local/bin/billy-desktop"

pkgbuild \
  --root "$PKG_ROOT" \
  --identifier "$PKG_IDENTIFIER" \
  --version "$VERSION" \
  --install-location "/" \
  "$PKG_OUTPUT"

echo "Created $PKG_OUTPUT"
