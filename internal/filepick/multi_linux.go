//go:build linux && !android

package filepick

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rymdport/portal/filechooser"
)

// OpenMultiple opens a native multi-file dialog using exactly one backend.
//
// Important: we must not chain portal → kdialog → zenity after a dialog has
// already been shown. On Kubuntu/KDE that produced two pickers (portal + Dolphin
// via kdialog). Backend choice is fixed up front from the desktop environment.
//
// Returns (nil, nil) if the user cancelled.
// Returns (nil, ErrUnsupported) if no multi-file backend is available.
func OpenMultiple(title string) ([]string, error) {
	if title == "" {
		title = "Select files"
	}

	switch chooseBackend() {
	case backendKDialog:
		return normalize(openKDialog(title))
	case backendPortal:
		return normalize(openPortal(title))
	case backendZenity:
		return normalize(openZenity(title))
	default:
		return nil, ErrUnsupported
	}
}

type backend int

const (
	backendNone backend = iota
	backendKDialog
	backendPortal
	backendZenity
)

// chooseBackend picks a single multi-file implementation.
// KDE/Plasma → kdialog only (Dolphin file dialog). Chaining portal+kdialog
// opens two windows on Kubuntu.
func chooseBackend() backend {
	desktop := strings.ToLower(
		os.Getenv("XDG_CURRENT_DESKTOP") + " " +
			os.Getenv("XDG_SESSION_DESKTOP") + " " +
			os.Getenv("DESKTOP_SESSION"),
	)
	isKDE := strings.Contains(desktop, "kde") ||
		strings.Contains(desktop, "plasma")

	_, kdialogErr := exec.LookPath("kdialog")
	hasKDialog := kdialogErr == nil
	_, zenityErr := exec.LookPath("zenity")
	hasZenity := zenityErr == nil

	if isKDE && hasKDialog {
		return backendKDialog
	}
	// Portal works well on GNOME and most other desktops with xdg-desktop-portal.
	if portalLikelyAvailable() {
		return backendPortal
	}
	if hasKDialog {
		return backendKDialog
	}
	if hasZenity {
		return backendZenity
	}
	// Last resort: try portal anyway (may still work without our probe).
	return backendPortal
}

// portalLikelyAvailable is a cheap probe so we do not open kdialog after a
// half-started portal dialog. Presence of the session bus name is enough.
func portalLikelyAvailable() bool {
	// dbus-send is widely available; failure means "don't prefer portal".
	if _, err := exec.LookPath("dbus-send"); err != nil {
		// No probe tool — still allow portal as a candidate when not on KDE.
		return true
	}
	cmd := exec.Command(
		"dbus-send", "--session", "--print-reply", "--dest=org.freedesktop.DBus",
		"/org/freedesktop/DBus", "org.freedesktop.DBus.NameHasOwner",
		"string:org.freedesktop.portal.Desktop",
	)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	// Reply is like: boolean true
	return bytes.Contains(out, []byte("true"))
}

func normalize(paths []string, err error) ([]string, error) {
	if err == nil {
		return paths, nil
	}
	if isCancel(err) {
		return nil, nil
	}
	return nil, err
}

func openPortal(title string) ([]string, error) {
	uris, err := filechooser.OpenFile("", title, &filechooser.OpenFileOptions{
		Multiple: true,
	})
	if err != nil {
		return nil, err
	}
	if len(uris) == 0 {
		return nil, nil
	}
	return urisToPaths(uris)
}

func openKDialog(title string) ([]string, error) {
	if _, err := exec.LookPath("kdialog"); err != nil {
		return nil, err
	}
	// --separate-output: one path per line (handles spaces in paths).
	cmd := exec.Command("kdialog", "--title", title, "--multiple", "--separate-output", "--getopenfilename", homeOrDot())
	out, err := cmd.Output()
	if err != nil {
		// Exit code 1 = cancel for kdialog.
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return nil, errCancelled{}
		}
		return nil, err
	}
	return splitLines(out), nil
}

func openZenity(title string) ([]string, error) {
	if _, err := exec.LookPath("zenity"); err != nil {
		return nil, err
	}
	cmd := exec.Command("zenity", "--file-selection", "--multiple", "--separator=\n", "--title="+title)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return nil, errCancelled{}
		}
		return nil, err
	}
	return splitLines(out), nil
}

func homeOrDot() string {
	if h, err := os.UserHomeDir(); err == nil && h != "" {
		return h
	}
	if h, err := filepath.Abs("."); err == nil {
		return h
	}
	return "."
}

func splitLines(out []byte) []string {
	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return nil
	}
	var paths []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	return paths
}

func urisToPaths(uris []string) ([]string, error) {
	paths := make([]string, 0, len(uris))
	for _, u := range uris {
		p, err := uriToPath(u)
		if err != nil {
			return nil, fmt.Errorf("parse selected URI %q: %w", u, err)
		}
		if p != "" {
			paths = append(paths, p)
		}
	}
	return paths, nil
}

func uriToPath(raw string) (string, error) {
	// Portal may return file:///path or already-decoded paths.
	if strings.HasPrefix(raw, "file://") {
		u, err := url.Parse(raw)
		if err != nil {
			return "", err
		}
		return u.Path, nil
	}
	if strings.HasPrefix(raw, "file:") {
		// file:/path form
		return strings.TrimPrefix(raw, "file:"), nil
	}
	return raw, nil
}

type errCancelled struct{}

func (errCancelled) Error() string { return "cancelled" }

func isCancel(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(errCancelled); ok {
		return true
	}
	// Portal / dbus often include "cancelled" / "canceled" in the message.
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "cancel")
}
