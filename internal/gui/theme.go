package gui

import (
	"image/color"

	"fyne.io/fyne/v2/theme"
)

const (
	SettingTheme = "setting.theme"
	ThemeDark    = "dark"
	ThemeLight   = "light"
)

// applyTheme sets the Fyne theme and persists the choice.
// Uses DarkTheme/LightTheme intentionally so the user choice overrides the OS preference.
func applyTheme(name string) {
	if MyApp.App == nil {
		return
	}

	switch name {
	case ThemeLight:
		MyApp.App.Settings().SetTheme(theme.LightTheme())
	default:
		name = ThemeDark
		MyApp.App.Settings().SetTheme(theme.DarkTheme())
	}

	MyApp.App.Preferences().SetString(SettingTheme, name)
}

// loadAndApplyTheme applies the stored theme preference (default: dark).
func loadAndApplyTheme() {
	name := MyApp.App.Preferences().StringWithFallback(SettingTheme, ThemeDark)
	applyTheme(name)
}

// currentThemeName returns the stored theme preference.
func currentThemeName() string {
	if MyApp.App == nil {
		return ThemeDark
	}
	return MyApp.App.Preferences().StringWithFallback(SettingTheme, ThemeDark)
}

// themeForeground returns the current theme's foreground color for canvas.Text usage.
func themeForeground() color.Color {
	return theme.ForegroundColor()
}
