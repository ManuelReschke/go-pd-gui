//go:build linux && !android

package filepick

import (
	"os"
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
	if len(splitLines(nil)) != 0 {
		t.Fatalf("empty input should yield empty")
	}
}

func TestChooseBackendKDEPrefersKDialog(t *testing.T) {
	t.Setenv("XDG_CURRENT_DESKTOP", "KDE")
	t.Setenv("XDG_SESSION_DESKTOP", "KDE")
	t.Setenv("DESKTOP_SESSION", "plasma")

	// Only meaningful when kdialog exists (true on this Kubuntu host).
	if _, err := os.Stat("/usr/bin/kdialog"); err != nil {
		t.Skip("kdialog not installed")
	}
	if got := chooseBackend(); got != backendKDialog {
		t.Fatalf("on KDE want kdialog, got %v", got)
	}
}

func TestChooseBackendNonKDENotForcedKDialog(t *testing.T) {
	t.Setenv("XDG_CURRENT_DESKTOP", "GNOME")
	t.Setenv("XDG_SESSION_DESKTOP", "gnome")
	t.Setenv("DESKTOP_SESSION", "gnome")

	got := chooseBackend()
	// Must not be "none"; portal or zenity/kdialog as fallback is fine.
	if got == backendNone {
		t.Fatalf("expected some backend on GNOME-like env")
	}
	// Critical: we only force kdialog on KDE.
	// On GNOME we prefer portal when available.
	if portalLikelyAvailable() && got != backendPortal {
		t.Fatalf("GNOME with portal available: want portal, got %v", got)
	}
}
