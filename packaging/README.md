# Billy Desktop Packaging Notes

The release goal is to install the Wails desktop app and the Billy CLI together.

## Native installer strategy

- macOS: bundle `Billy.app` and `billy` in one archive; a future `.pkg` can install the app into `/Applications` and `billy` into `/usr/local/bin`
- Windows: bundle `Billy.exe` and `billy.exe` into the same zip/installer payload
- Linux: bundle the GUI binary and `billy` together so `.deb` or tar installs can place the GUI in `/opt` and the CLI on `PATH`

## Power-user installers

- Homebrew cask installs the app and the bundled top-level CLI from the same macOS archive
- Scoop installs the GUI app and the CLI from the same Windows bundle archive
- `scripts/install.sh` installs the macOS or Linux bundle for users who still want a shell installer

## Release-day manifest update

After building release assets, render the package-manager manifests with:

```bash
scripts/release/render_manifests.sh \
  v0.1.0 \
  dist/billy-macos-universal.tar.gz \
  dist/billy-windows-amd64-bundle.zip
```

This updates:

- `packaging/homebrew/Casks/billy.rb`
- `packaging/scoop/billy.json`

If the local tap repos exist, the script can also sync into:

- `../homebrew-billy`
- `../scoop-billy`

To build local release assets from the current `billy-app` + `billy-wails` repos in one step:

```bash
scripts/release/build_release_assets.sh darwin
scripts/release/build_release_assets.sh windows amd64
```

## Why archive bundles instead of app-only assets?

The archive needs to contain both the GUI app and the CLI so:

- Homebrew can install the app and `billy` from the same asset
- Scoop can install the Windows desktop shortcut and `billy.exe`
- the desktop app can locate the bundled CLI and start `billy serve` automatically

## Native package scripts

Once the GUI artifact and CLI artifact exist, native package helpers are available:

- `scripts/release/build_macos_pkg.sh` builds a macOS `.pkg`
- `scripts/release/build_linux_deb.sh` builds Linux packages through `nfpm`

## Linux note

Native Linux GUI builds need a Linux CGO toolchain. From macOS or Windows, use Docker:

```bash
/Users/jonathanforrider/go/bin/wails3 task setup:docker
/Users/jonathanforrider/go/bin/wails3 task linux:build
```
