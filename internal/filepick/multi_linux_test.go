//go:build linux && !android

package filepick

import (
	"testing"
)

func TestURIToPath(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"file:///home/dev/a.txt", "/home/dev/a.txt"},
		{"file:///tmp/foo%20bar.txt", "/tmp/foo bar.txt"},
		{"/already/path", "/already/path"},
	}
	for _, tt := range tests {
		got, err := uriToPath(tt.in)
		if err != nil {
			t.Fatalf("uriToPath(%q): %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("uriToPath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSplitLines(t *testing.T) {
	got := splitLines([]byte("a\nb\n\nc\n"))
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("got %#v", got)
	}
	if splitLines(nil) != nil && len(splitLines(nil)) != 0 {
		t.Fatalf("empty input should yield empty")
	}
}
