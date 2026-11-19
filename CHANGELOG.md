# Changelog

All notable changes to **Go-PD-GUI (DRAINY)** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project aims to follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [1.4.2] - 2026-11-19

### Changed

- Upgrade UI toolkit from Fyne v2.8.0 to v2.8.2.

### Fixed

- Linux `make install` and `make user-install` failed on the v1.4.1 tarball because that package was built with a Fyne CLI older than v1.7.3. The old CLI wrote the icon filename without `.png`. Release builds now pull `fyneio/fyne-cross-images:linux`, whose Fyne CLI v1.7.3 writes `com.gopdgui.app.png`. The desktop entry stays `Icon=com.gopdgui.app`. (#9)

### Security

- Update `golang.org/x/crypto` to v0.55.0 (CVE-2026-56854) and `golang.org/x/image` to v0.45.0 (CVE-2026-46599, CVE-2026-33809, and related TIFF, WebP, and BMP issues).

## [1.4.1] - 2026-07-31

### Added

- Automated GitHub Releases on version tags (`v*`) via fyne-cross + GoReleaser.
- **Dark / Light mode switch** in Settings (persisted via app preferences; default remains dark). (#4)
- **Multi-file upload**: stage multiple files (Add files), review/remove them in a queue, then upload sequentially with overall progress. Copy pastes all result URLs (one per line). (#3)
- **Native multi-select** when adding files on Linux (XDG portal / kdialog / zenity) and Windows (common dialog). Android keeps the single-file Fyne picker.
- Linux multi-select uses **one** backend only (KDE → kdialog/Dolphin; elsewhere portal when available) so portal+kdialog no longer open two pickers on Kubuntu.
- Settings: **Filepicker** choice — *Fyne default* (single file) or *OS Native Multipicker* (multi-select on desktop).

### Changed

- Download links in the README now point to GitHub Release assets instead of committed `fyne-cross/` binaries.
- UI labels that used hardcoded white canvas text now use theme-aware widgets so Light mode stays readable.
- Main screen upload actions are split into **Add files** and **Upload** / **Upload all** (Fyne’s file dialog only returns one file per pick).

### Fixed

- **Android uploads** no longer fail with `open /document/primary%3A…: no such file or directory` when picking photos via the Storage Access Framework. Content URIs are streamed into a temp file before `go-pd` upload. (#5)

### Notes

- Tag `v1.4.0` was created the same day but not published. The release workflow failed on that tag; `v1.4.1` is the build that shipped.

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

[Unreleased]: https://github.com/ManuelReschke/go-pd-gui/compare/v1.4.2...HEAD
[1.4.2]: https://github.com/ManuelReschke/go-pd-gui/releases/tag/v1.4.2
[1.4.1]: https://github.com/ManuelReschke/go-pd-gui/releases/tag/v1.4.1
[1.3.0]: https://github.com/ManuelReschke/go-pd-gui/releases/tag/v1.3.0
[1.2.0]: https://github.com/ManuelReschke/go-pd-gui/releases/tag/v1.2.0
[1.1.0]: https://github.com/ManuelReschke/go-pd-gui/releases/tag/v1.1.0
[1.0.0]: https://github.com/ManuelReschke/go-pd-gui/releases/tag/v1.0.0
