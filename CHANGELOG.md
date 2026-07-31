# Changelog

All notable changes to **Go-PD-GUI (DRAINY)** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project aims to follow [Semantic Versioning](https://semver.org/).

## [1.3.0] - 2026-07-31

### Changed

- Upgrade UI toolkit to **Fyne v2.8** (from v2.4).
- Raise Go toolchain to **1.25**.
- Update dependencies: `go-pd` v1.6.0, `req/v3` v3.60, and transitive modules.
- Enable Fyne `fyneDo` migration (`FyneApp.toml`) for the new threading model.
- UI updates from background work now go through `fyne.Do` (copy feedback, history highlight, progress bar).

### Fixed

- File open dialog no longer panics on resize under Fyne 2.8 (`Show` before `Resize`).
- File picker no longer freezes during upload; transfer runs on a background goroutine.
- Upload path uses a real local filesystem path for `go-pd` / `SetFile` (required for progress size and reliable uploads).
- API key is loaded from preferences at startup so upload auth works without opening Settings first.
- Clearer upload error messages.

### Notes

- `theme.DarkTheme()` remains in use but is deprecated upstream (removal planned for Fyne v3). Prefer system/default theme later if needed.

## [1.2.0] - 2024-05-22

### Added

- Real progress bar (0–100%) during PixelDrain upload.
- Progress bar preview in the README.

### Changed

- Go toolchain to 1.21.
- Dependency updates.
- Rebuilt Windows, Linux, and Android binaries.

## [1.1.0] - 2023-01-25

### Added

- Show last uploaded file on the main screen.
- Package / dependency updates around the go-pd client.

## [1.0.0] - 2022-06-30

### Added

- First public release of DRAINY.
- Cross-platform PixelDrain upload UI (Windows, Linux, Android).
- Optional API key authentication via settings.
- Copy link button and result hyperlink.
- Fyne-based desktop/mobile packaging via fyne-cross.

[1.3.0]: https://github.com/ManuelReschke/go-pd-gui/releases/tag/v1.3.0
[1.2.0]: https://github.com/ManuelReschke/go-pd-gui/releases/tag/v1.2.0
[1.1.0]: https://github.com/ManuelReschke/go-pd-gui/releases/tag/v1.1.0
[1.0.0]: https://github.com/ManuelReschke/go-pd-gui/releases/tag/v1.0.0
