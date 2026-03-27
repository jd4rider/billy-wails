# Billy Desktop

Billy Desktop is the Wails-based desktop companion for Billy.

The release goal is the same one used by Logos:

- ship the desktop app and the `billy` CLI together
- let the desktop app launch `billy serve` itself when the bundled CLI is present
- publish bundle archives that package managers can install cleanly

## Development

```bash
cd billy-wails/frontend
npm install
cd ..
wails dev
```

## Production build

```bash
wails build
```

That produces the desktop app/executable. For release bundles, pair the built GUI artifact with the `billy` CLI from `billy-app`.

## Packaging

See:

- [packaging/README.md](./packaging/README.md)
- [scripts/install.sh](./scripts/install.sh)
- [scripts/release/package_bundle.sh](./scripts/release/package_bundle.sh)
- [scripts/release/render_manifests.sh](./scripts/release/render_manifests.sh)
- [scripts/release/build_macos_pkg.sh](./scripts/release/build_macos_pkg.sh)
- [scripts/release/build_linux_deb.sh](./scripts/release/build_linux_deb.sh)

The intended shipped artifacts are:

- macOS: `billy-macos-universal.tar.gz`
- Linux: `billy-linux-amd64-bundle.tar.gz`
- Windows: `billy-windows-amd64-bundle.zip`

Each bundle includes:

- the desktop app
- the `billy` CLI
- a layout that allows the desktop app to locate and start `billy serve`
