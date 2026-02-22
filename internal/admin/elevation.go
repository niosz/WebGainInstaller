package admin

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	procMessageBox = user32.NewProc("MessageBoxW")
)

const (
	mbOK        = 0x00000000
	mbIconError = 0x00000010
)

func IsElevated() (bool, error) {
	var token windows.Token
	proc := windows.CurrentProcess()
	err := windows.OpenProcessToken(proc, windows.TOKEN_QUERY, &token)
	if err != nil {
		return false, fmt.Errorf("OpenProcessToken: %w", err)
	}
	defer token.Close()

	var elevation struct {
		TokenIsElevated uint32
	}
	var size uint32
	err = windows.GetTokenInformation(
		token,
		windows.TokenElevation,
		(*byte)(unsafe.Pointer(&elevation)),
		uint32(unsafe.Sizeof(elevation)),
		&size,
	)
	if err != nil {
		return false, fmt.Errorf("GetTokenInformation: %w", err)
	}

	return elevation.TokenIsElevated != 0, nil
}

func RequireAdmin() {
	elevated, err := IsElevated()
	if err != nil || !elevated {
		showErrorBox(
			"WebGain Installer",
			"Il programma deve essere lanciato con un utenza avente diritti amministrativi, impossibile continuare.",
		)
		os.Exit(1)
	}
}

func showErrorBox(title, message string) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	msgPtr, _ := syscall.UTF16PtrFromString(message)
	procMessageBox.Call(
		0,
		uintptr(unsafe.Pointer(msgPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(mbOK|mbIconError),
	)
}
