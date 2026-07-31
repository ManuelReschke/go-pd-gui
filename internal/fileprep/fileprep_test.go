package fileprep

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalPathCandidatesDecodesAndroidDocumentPath(t *testing.T) {
	raw := "/document/primary%3APictures%2F1719364964095.png"
	cands := LocalPathCandidates(raw)
	if len(cands) < 2 {
		t.Fatalf("expected raw + decoded candidates, got %#v", cands)
	}
	if cands[0] != filepath.Clean(raw) {
		t.Fatalf("first candidate want cleaned raw path, got %q", cands[0])
	}
	wantDecoded := filepath.Clean("/document/primary:Pictures/1719364964095.png")
	found := false
	for _, c := range cands {
		if c == wantDecoded {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing decoded candidate %q in %#v", wantDecoded, cands)
	}
}

func TestReadableLocalFileIgnoresAndroidDocumentPath(t *testing.T) {
	// Issue #5: SAF document paths are not real filesystem paths.
	if path, ok := ReadableLocalFile("/document/primary%3APictures%2F1719364964095.png"); ok {
		t.Fatalf("android document path must not be treated as readable, got %q", path)
	}
}

func TestReadableLocalFileAcceptsRealFile(t *testing.T) {
	dir := t.TempDir()
	realPath := filepath.Join(dir, "photo.png")
	if err := os.WriteFile(realPath, []byte("png-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, ok := ReadableLocalFile(realPath)
	if !ok {
		t.Fatal("expected real file to be readable")
	}
	if got != realPath {
		t.Fatalf("want %q, got %q", realPath, got)
	}
}

func TestReadableLocalFileAcceptsURLEncodedRealPath(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "My Pictures")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	realPath := filepath.Join(sub, "shot.png")
	if err := os.WriteFile(realPath, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	encoded := strings.ReplaceAll(realPath, " ", "%20")
	got, ok := ReadableLocalFile(encoded)
	if !ok {
		t.Fatalf("expected encoded real path to resolve, encoded=%q", encoded)
	}
	if got != realPath {
		t.Fatalf("want %q, got %q", realPath, got)
	}
}

func TestCopyReaderToTempFile(t *testing.T) {
	payload := []byte("hello-upload-payload")
	path, err := CopyReaderToTempFile(bytes.NewReader(payload), "pic.png")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(path) })

	if filepath.Ext(path) != ".png" {
		t.Fatalf("temp file should keep extension, got %q", path)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("temp content mismatch: got %q want %q", got, payload)
	}
}

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", "upload.bin"},
		{".", "upload.bin"},
		{"  photo.png  ", "photo.png"},
		{"primary%3APictures%2F1719364964095.png", "1719364964095.png"},
		{"/abs/path/file.jpg", "file.jpg"},
	}
	for _, tc := range tests {
		if got := SanitizeFileName(tc.in); got != tc.want {
			t.Errorf("SanitizeFileName(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}
