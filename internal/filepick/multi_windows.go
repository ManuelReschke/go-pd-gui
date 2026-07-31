//go:build windows

package filepick

import (
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	comdlg32            = windows.NewLazySystemDLL("comdlg32.dll")
	procGetOpenFileName = comdlg32.NewProc("GetOpenFileNameW")
)

// OPENFILENAMEW — see https://learn.microsoft.com/windows/win32/api/commdlg/ns-commdlg-openfilenamew
type openFileNameW struct {
	StructSize      uint32
	Owner           windows.Handle
	Instance        windows.Handle
	Filter          *uint16
	CustomFilter    *uint16
	MaxCustomFilter uint32
	FilterIndex     uint32
	File            *uint16
	MaxFile         uint32
	FileTitle       *uint16
	MaxFileTitle    uint32
	InitialDir      *uint16
	Title           *uint16
	Flags           uint32
	FileOffset      uint16
	FileExtension   uint16
	DefExt          *uint16
	CustData        uintptr
	FnHook          uintptr
	TemplateName    *uint16
	PvReserved      uintptr
	DwReserved      uint32
	FlagsEx         uint32
}

const (
	ofnExplorer         = 0x00080000
	ofnFileMustExist    = 0x00001000
	ofnPathMustExist    = 0x00000800
	ofnAllowMultiSelect = 0x00000200
	ofnHideReadOnly     = 0x00000004
)

// OpenMultiple opens the Win32 multi-file open dialog.
// Returns (nil, nil) if the user cancelled.
func OpenMultiple(title string) ([]string, error) {
	if title == "" {
		title = "Select files"
	}
	if err := procGetOpenFileName.Find(); err != nil {
		return nil, ErrUnsupported
	}

	// Large buffer for directory + many null-separated file names.
	const bufChars = 256 * 1024
	buf := make([]uint16, bufChars)

	titlePtr, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return nil, err
	}
	filterPtr, err := windows.UTF16PtrFromString("All Files (*.*)\x00*.*\x00")
	if err != nil {
		return nil, err
	}

	ofn := openFileNameW{
		StructSize: uint32(unsafe.Sizeof(openFileNameW{})),
		Filter:     filterPtr,
		File:       &buf[0],
		MaxFile:    uint32(len(buf)),
		Title:      titlePtr,
		Flags:      ofnExplorer | ofnFileMustExist | ofnPathMustExist | ofnAllowMultiSelect | ofnHideReadOnly,
	}

	r, _, _ := procGetOpenFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		// Cancel or error — treat as cancel (nil, nil).
		return nil, nil
	}

	return parseMultiSelectBuffer(buf), nil
}

// parseMultiSelectBuffer decodes the double-null-terminated buffer from
// GetOpenFileName with OFN_ALLOWMULTISELECT | OFN_EXPLORER:
//   - single file: "C:\path\to\file.txt\0\0"
//   - multi:       "C:\path\to\dir\0file1.txt\0file2.txt\0\0"
func parseMultiSelectBuffer(buf []uint16) []string {
	parts := splitUTF16Z(buf)
	if len(parts) == 0 {
		return nil
	}
	if len(parts) == 1 {
		return []string{parts[0]}
	}
	dir := parts[0]
	out := make([]string, 0, len(parts)-1)
	for _, name := range parts[1:] {
		out = append(out, filepath.Join(dir, name))
	}
	return out
}

func splitUTF16Z(buf []uint16) []string {
	var parts []string
	start := 0
	for i := 0; i < len(buf); i++ {
		if buf[i] != 0 {
			continue
		}
		if i == start {
			// double null → end
			break
		}
		parts = append(parts, syscall.UTF16ToString(buf[start:i]))
		start = i + 1
	}
	return parts
}
