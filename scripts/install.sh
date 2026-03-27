#!/usr/bin/env bash
set -euo pipefail

REPO="${REPO:-jd4rider/billy-wails}"
INSTALL_DIR="${BILLY_INSTALL_DIR:-$HOME/.local/bin}"
APP_DIR="${BILLY_APP_DIR:-$HOME/Applications}"

info() { printf '[billy-desktop] %s\n' "$*"; }
fail() { printf '[billy-desktop] %s\n' "$*" >&2; exit 1; }

detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux) echo "linux" ;;
    *) fail "Unsupported OS for shell install. Use the native installer from the release page." ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) fail "Unsupported architecture." ;;
  esac
}

latest_tag() {
  curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' \
    | head -n 1
}

main() {
  local os arch tag asset tmpdir
  os="$(detect_os)"
  arch="$(detect_arch)"
  tag="$(latest_tag)"

  [[ -n "$tag" ]] || fail "Could not resolve the latest release tag."

  if [[ "$os" == "darwin" ]]; then
    asset="billy-macos-universal.tar.gz"
  else
    asset="billy-linux-${arch}-bundle.tar.gz"
  fi

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  info "Downloading $asset from $tag"
  curl -fsSL "https://github.com/$REPO/releases/download/$tag/$asset" -o "$tmpdir/archive.tar.gz"
  tar -xzf "$tmpdir/archive.tar.gz" -C "$tmpdir"

  mkdir -p "$INSTALL_DIR"

  if [[ "$os" == "darwin" ]]; then
    mkdir -p "$APP_DIR"
    rm -rf "$APP_DIR/Billy.app"
    cp -R "$tmpdir/Billy.app" "$APP_DIR/Billy.app"
    if [[ -f "$tmpdir/billy" ]]; then
      install -m 0755 "$tmpdir/billy" "$INSTALL_DIR/billy"
    elif [[ -f "$APP_DIR/Billy.app/Contents/Resources/billy" ]]; then
      install -m 0755 "$APP_DIR/Billy.app/Contents/Resources/billy" "$INSTALL_DIR/billy"
    else
      fail "Bundled Billy CLI was not found in the macOS archive."
    fi
    cat >"$INSTALL_DIR/billy-desktop" <<EOF
#!/usr/bin/env bash
exec open "$APP_DIR/Billy.app" --args "\$@"
EOF
    chmod 0755 "$INSTALL_DIR/billy-desktop"
    info "Installed app to $APP_DIR/Billy.app"
    info "Installed CLI to $INSTALL_DIR/billy"
  else
    install -m 0755 "$tmpdir/Billy" "$INSTALL_DIR/Billy"
    install -m 0755 "$tmpdir/billy" "$INSTALL_DIR/billy"
    info "Installed desktop binary and CLI to $INSTALL_DIR"
  fi

  info "If $INSTALL_DIR is not on PATH, add it to your shell profile."
}

main "$@"
