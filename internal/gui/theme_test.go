package gui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestApplyThemePersistsPreference(t *testing.T) {
	MyApp.App = test.NewApp()
	t.Cleanup(func() {
		MyApp.App.Quit()
		MyApp.App = nil
	})

	applyTheme(ThemeLight)
	if got := currentThemeName(); got != ThemeLight {
		t.Fatalf("want %q, got %q", ThemeLight, got)
	}

	applyTheme(ThemeDark)
	if got := currentThemeName(); got != ThemeDark {
		t.Fatalf("want %q, got %q", ThemeDark, got)
	}

	// Unknown values fall back to dark and are stored as dark.
	applyTheme("unknown")
	if got := currentThemeName(); got != ThemeDark {
		t.Fatalf("want fallback %q, got %q", ThemeDark, got)
	}
}

func TestLoadAndApplyThemeDefaultDark(t *testing.T) {
	MyApp.App = test.NewApp()
	t.Cleanup(func() {
		MyApp.App.Quit()
		MyApp.App = nil
	})

	// Fresh app preferences: no theme key yet.
	loadAndApplyTheme()
	if got := currentThemeName(); got != ThemeDark {
		t.Fatalf("default theme want %q, got %q", ThemeDark, got)
	}
}
