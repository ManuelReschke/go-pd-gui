//go:build linux && !android

package filepick

import (
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rymdport/portal/filechooser"
)

// OpenMultiple opens a native multi-file dialog.
// Returns (nil, nil) if the user cancelled.
// Returns (nil, ErrUnsupported) if no multi-file backend is available.
func OpenMultiple(title string) ([]string, error) {
	if title == "" {
		title = "Select files"
	}

	if paths, err := openPortal(title); err == nil {
		return paths, nil
	} else if isCancel(err) {
		return nil, nil
	}

	if paths, err := openKDialog(title); err == nil {
		return paths, nil
	} else if isCancel(err) {
		return nil, nil
	}

	if paths, err := openZenity(title); err == nil {
		return paths, nil
	} else if isCancel(err) {
		return nil, nil
	}

	return nil, ErrUnsupported
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
