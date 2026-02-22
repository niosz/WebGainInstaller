package font

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const (
	wmFontChange    = 0x001D
	hwndBroadcast   = uintptr(0xFFFF)
	smtoAbortIfHung = 0x0002
)

var (
	gdi32                = syscall.NewLazyDLL("gdi32.dll")
	user32               = syscall.NewLazyDLL("user32.dll")
	procAddFontResourceW = gdi32.NewProc("AddFontResourceW")
	procSendMessageTimeout = user32.NewProc("SendMessageTimeoutW")
)

func IsFontInstalled(fontFileName string) bool {
	fontsDir := os.Getenv("WINDIR") + `\Fonts`
	destPath := filepath.Join(fontsDir, fontFileName)
	_, err := os.Stat(destPath)
	return err == nil
}

func InstallFonts(fontFS fs.FS) error {
	fontsDir := os.Getenv("WINDIR") + `\Fonts`

	regKey, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion\Fonts`,
		registry.SET_VALUE|registry.QUERY_VALUE,
	)
	if err != nil {
		return fmt.Errorf("impossibile aprire chiave registro font: %w", err)
	}
	defer regKey.Close()

	entries, err := fs.ReadDir(fontFS, ".")
	if err != nil {
		return fmt.Errorf("impossibile leggere cartella Font: %w", err)
	}

	installed := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(name), ".ttf") {
			continue
		}

		destPath := filepath.Join(fontsDir, name)
		if _, err := os.Stat(destPath); err == nil {
			continue
		}

		data, err := fs.ReadFile(fontFS, name)
		if err != nil {
			return fmt.Errorf("impossibile leggere font %s: %w", name, err)
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return fmt.Errorf("impossibile copiare font %s in %s: %w", name, fontsDir, err)
		}

		fontDisplayName := strings.TrimSuffix(name, filepath.Ext(name))
		if err := regKey.SetStringValue(fontDisplayName+" (TrueType)", name); err != nil {
			return fmt.Errorf("impossibile registrare font %s: %w", name, err)
		}

		destPathUTF16, _ := syscall.UTF16PtrFromString(destPath)
		procAddFontResourceW.Call(uintptr(unsafe.Pointer(destPathUTF16)))
		installed++
	}

	if installed > 0 {
		var result uintptr
		procSendMessageTimeout.Call(
			hwndBroadcast,
			wmFontChange,
			0, 0,
			smtoAbortIfHung,
			1000,
			uintptr(unsafe.Pointer(&result)),
		)
	}

	return nil
}
