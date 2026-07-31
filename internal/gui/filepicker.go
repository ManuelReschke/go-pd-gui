package gui

const (
	SettingFilePicker = "setting.filepicker"

	// FilePickerFyne uses Fyne's built-in single-file dialog.
	FilePickerFyne = "fyne"
	// FilePickerNative uses the OS multi-file picker (portal/kdialog/Win32).
	FilePickerNative = "native"

	// Labels shown in Settings.
	FilePickerOptionFyne   = "Fyne default"
	FilePickerOptionNative = "OS Native Multipicker"
)

// setFilePickerMode persists the file picker preference.
// Unknown values fall back to native (multi-select on desktop).
func setFilePickerMode(name string) {
	if MyApp.App == nil {
		return
	}
	switch name {
	case FilePickerFyne:
		// ok
	default:
		name = FilePickerNative
	}
	MyApp.App.Preferences().SetString(SettingFilePicker, name)
}

// currentFilePickerMode returns the stored file picker preference (default: native).
func currentFilePickerMode() string {
	if MyApp.App == nil {
		return FilePickerNative
	}
	name := MyApp.App.Preferences().StringWithFallback(SettingFilePicker, FilePickerNative)
	if name == FilePickerFyne {
		return FilePickerFyne
	}
	return FilePickerNative
}

// filePickerModeFromLabel maps a Settings radio label to a stored mode value.
func filePickerModeFromLabel(label string) string {
	if label == FilePickerOptionFyne {
		return FilePickerFyne
	}
	return FilePickerNative
}

// filePickerLabelFromMode maps a stored mode to a Settings radio label.
func filePickerLabelFromMode(mode string) string {
	if mode == FilePickerFyne {
		return FilePickerOptionFyne
	}
	return FilePickerOptionNative
}
