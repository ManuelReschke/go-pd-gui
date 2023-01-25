package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

const (
	MenuItemMain         = "> Menu"
	MenuItemMainSettings = "Settings"
	MenuItemMainAbout    = "About"
	AboutText            = "@Author: Manuel Reschke\n " +
		"@Github: https://github.com/ManuelReschke/go-pd-gui\n\n " +
		"This tool was made by Manuel Reschke under MIT Licence.\n\n " +
		"Version: " + Version
)

func BuildMainMenu() *fyne.MainMenu {
	// Main Menu
	menuItemSettings := fyne.NewMenuItem(MenuItemMainSettings, menuActionSetting)
	menuItemAbout := fyne.NewMenuItem(MenuItemMainAbout, menuActionAbout)

	appMainMenu := fyne.NewMenu(MenuItemMain, menuItemSettings, menuItemAbout)
	return fyne.NewMainMenu(appMainMenu)
}

func menuActionSetting() {
	BuildSettingsWindow()
}

func menuActionAbout() {
	dialog.ShowInformation(MenuItemMainAbout, AboutText, MyApp.Window)
}
