//go:build !windows && !(linux && !android)

package filepick

// OpenMultiple is not available on this platform (e.g. Android, iOS, wasm).
// Callers should fall back to Fyne's single-file dialog.
func OpenMultiple(title string) ([]string, error) {
	return nil, ErrUnsupported
}
