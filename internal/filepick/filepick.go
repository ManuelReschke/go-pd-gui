// Package filepick opens a native multi-file picker when the platform supports it.
//
// Fyne's dialog.NewFileOpen only returns a single file. Desktop platforms can
// still multi-select via XDG portal / kdialog / zenity (Linux) or the Win32
// common dialog (Windows). Callers should fall back to Fyne when OpenMultiple
// returns ErrUnsupported (e.g. Android).
package filepick

import "errors"

// ErrUnsupported means this platform has no multi-file picker; use Fyne instead.
var ErrUnsupported = errors.New("multi-file picker not supported on this platform")

// ErrCancelled means the user dismissed the dialog without selecting files.
// OpenMultiple currently returns (nil, nil) on cancel for simplicity; this
// sentinel is reserved if a backend needs an explicit cancel error.
var ErrCancelled = errors.New("file selection cancelled")
