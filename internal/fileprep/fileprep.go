// Package fileprep turns Fyne / Android document URIs into local paths that
// go-pd can open with os.Open / SetFile.
package fileprep

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// ReadableLocalFile reports whether uriPath (or its URL-decoded form) is an
// existing regular file on the local filesystem.
func ReadableLocalFile(uriPath string) (string, bool) {
	for _, candidate := range LocalPathCandidates(uriPath) {
		st, err := os.Stat(candidate)
		if err != nil || st.IsDir() {
			continue
		}
		return candidate, true
	}
	return "", false
}

// LocalPathCandidates returns possible filesystem paths derived from a URI path.
// Android document paths often keep percent-encoding (%3A, %2F) in Path().
func LocalPathCandidates(uriPath string) []string {
	if uriPath == "" {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	add := func(p string) {
		p = filepath.Clean(p)
		if p == "" || p == "." {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}

	add(uriPath)
	if decoded, err := url.PathUnescape(uriPath); err == nil {
		add(decoded)
	}
	return out
}

// CopyReaderToTempFile writes r to a unique temp file and returns its path.
// The caller must remove the file when finished (e.g. after upload).
func CopyReaderToTempFile(r io.Reader, fileName string) (string, error) {
	if r == nil {
		return "", fmt.Errorf("reader is nil")
	}
	ext := filepath.Ext(fileName)
	// Keep the extension when present so PixelDrain still sees a plausible name
	// via FileName (the temp path itself is only for SetFile).
	tmp, err := os.CreateTemp("", "gopdgui-upload-*"+ext)
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, r); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	return tmpPath, nil
}

// SanitizeFileName picks a safe non-empty basename for the upload.
func SanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	name = filepath.Base(name)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "upload.bin"
	}
	// Document URIs sometimes surface encoded segments as the "name".
	if decoded, err := url.PathUnescape(name); err == nil && decoded != "" {
		name = filepath.Base(decoded)
	}
	if name == "" || name == "." {
		return "upload.bin"
	}
	return name
}
