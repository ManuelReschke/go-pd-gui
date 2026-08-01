# Go-PD-GUI -DRAINY
A free easy to use cross plattform upload tool for pixeldrain.com.
* Windows x32/x64
* Linux x64
* Android APK ARM / ARM64

![Go-PD-GUI-DRAINY](assets/go-pd-gui-icon.png)

**Current version:** v1.4.1 · [Changelog](CHANGELOG.md)

## Preview
![Go-PD](assets/go-pd-gui-preview.gif)

Progress bar
![Go-PD-Progress](assets/go-pd-gui-progress-bar.gif)

## Download binaries

Pre-built packages are published on the [GitHub Releases](https://github.com/ManuelReschke/go-pd-gui/releases) page (latest: [v1.4.1](https://github.com/ManuelReschke/go-pd-gui/releases/tag/v1.4.1)).

* Windows:
  * [Windows 64 Bit](https://github.com/ManuelReschke/go-pd-gui/releases/latest/download/DRAINY_Windows_x64.zip)
  * [Windows 32 Bit](https://github.com/ManuelReschke/go-pd-gui/releases/latest/download/DRAINY_Windows_x86.zip)
* Linux:
  * [Linux 64 Bit](https://github.com/ManuelReschke/go-pd-gui/releases/latest/download/DRAINY_Linux_x64.tar.xz)
* Android:
  * [Android ARM64](https://github.com/ManuelReschke/go-pd-gui/releases/latest/download/DRAINY_Android_arm64.apk)
  * [Android ARM](https://github.com/ManuelReschke/go-pd-gui/releases/latest/download/DRAINY_Android_arm.apk)

## Features

### Highlights (new)

* **Multi-file upload** — select several files, stage them in a queue (add / remove), then upload them all in one go
* **OS native multi-select** — on Linux and Windows, pick many files at once in the system file dialog (Ctrl/Shift); sequential upload with overall progress
* **Filepicker setting** — switch in Settings between **Fyne default** (single file) and **OS Native Multipicker**
* **Dark / Light mode** — theme switch in Settings (persisted)
* **Copy all URLs** — after a multi-upload, Copy puts every PixelDrain link on its own line
* **Android SAF uploads** — content URIs from the Storage Access Framework are prepared correctly before upload

### Core

* Upload to [pixeldrain.com](https://pixeldrain.com)
* Optional API-Key authentication (saved in user settings)
* Real progress bar (0–100%) during transfer
* Result hyperlink + **Copy** button
* Last uploaded file on the main screen
* Remote upload history (with API key)
* Cross-platform builds: Windows x32/x64, Linux x64, Android ARM / ARM64

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for release history and upgrade notes.

## License

This software is released under the MIT License, see LICENSE.