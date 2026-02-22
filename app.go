package main

import (
	"context"
	"io/fs"
	"syscall"
	"time"
	"unsafe"

	"WebGainInstaller/internal/setup"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	user32Dll       = syscall.NewLazyDLL("user32.dll")
	procMessageBoxW = user32Dll.NewProc("MessageBoxW")
	procFindWindowW = user32Dll.NewProc("FindWindowW")
)

const (
	mbOK          = 0x00000000
	mbYesNo       = 0x00000004
	mbIconError   = 0x00000010
	mbIconWarning = 0x00000030
	idYes         = 6
)

type App struct {
	ctx              context.Context
	configFS         fs.FS
	webgainRoot      string
	hwnd             uintptr
	skipCloseConfirm bool
}

func NewApp(configFS fs.FS) *App {
	return &App{
		configFS: configFS,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) getHWND() uintptr {
	if a.hwnd == 0 {
		title, _ := syscall.UTF16PtrFromString("WebGain Installer")
		a.hwnd, _, _ = procFindWindowW.Call(0, uintptr(unsafe.Pointer(title)))
	}
	return a.hwnd
}

func (a *App) ConfirmCancel() bool {
	title, _ := syscall.UTF16PtrFromString("Conferma Annullamento")
	msg, _ := syscall.UTF16PtrFromString("Si è sicuri di voler annullare l'installazione?")
	ret, _, _ := procMessageBoxW.Call(
		a.getHWND(),
		uintptr(unsafe.Pointer(msg)),
		uintptr(unsafe.Pointer(title)),
		uintptr(mbYesNo|mbIconWarning),
	)
	if int(ret) == idYes {
		wailsRuntime.Quit(a.ctx)
		return false
	}
	return true
}

func (a *App) beforeClose(ctx context.Context) bool {
	if a.skipCloseConfirm {
		return false
	}
	title, _ := syscall.UTF16PtrFromString("Conferma Annullamento")
	msg, _ := syscall.UTF16PtrFromString("Si è sicuri di voler annullare l'installazione?")
	ret, _, _ := procMessageBoxW.Call(
		a.getHWND(),
		uintptr(unsafe.Pointer(msg)),
		uintptr(unsafe.Pointer(title)),
		uintptr(mbYesNo|mbIconWarning),
	)
	return int(ret) != idYes
}

func (a *App) RunSetupSteps() {
	time.Sleep(1 * time.Second)

	wailsRuntime.EventsEmit(a.ctx, "setup:step", "Preparazione installazione...")
	time.Sleep(1 * time.Second)

	root, err := setup.PrepareRoot()
	if err != nil {
		a.fatalCorruptError()
		return
	}
	a.webgainRoot = root

	wailsRuntime.EventsEmit(a.ctx, "setup:step", "Verifica moduli...")
	time.Sleep(1 * time.Second)

	if err := setup.VerifyModules(a.configFS, a.webgainRoot); err != nil {
		a.fatalCorruptError()
		return
	}

	wailsRuntime.EventsEmit(a.ctx, "setup:done", nil)
}

func (a *App) GetEulaText() string {
	data, err := fs.ReadFile(a.configFS, "eula.txt")
	if err != nil {
		return "Errore caricamento EULA."
	}
	return string(data)
}

func (a *App) fatalCorruptError() {
	a.skipCloseConfirm = true
	wailsRuntime.EventsEmit(a.ctx, "setup:fatal", nil)
	time.Sleep(200 * time.Millisecond)
	title, _ := syscall.UTF16PtrFromString("Installazione Corrotta")
	msg, _ := syscall.UTF16PtrFromString("L'installazione risulta corrotta, provare a recuperare il pacchetto o contattare il supporto tecnico.")
	procMessageBoxW.Call(
		a.getHWND(),
		uintptr(unsafe.Pointer(msg)),
		uintptr(unsafe.Pointer(title)),
		uintptr(mbOK|mbIconError),
	)
	wailsRuntime.Quit(a.ctx)
}
