package main

import (
	"context"
	"io/fs"
	"log"
	"os/exec"
	"syscall"
	"time"
	"unsafe"

	"WebGainInstaller/internal/logger"
	"WebGainInstaller/internal/setup"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var devMode = "false"

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
		logger.Error("Creazione WEBGAINROOT fallita: %v", err)
		a.fatalCorruptError()
		return
	}
	a.webgainRoot = root

	if err := logger.Init(root); err != nil {
		log.Printf("Impossibile inizializzare log: %v", err)
	}
	logger.Info("WEBGAINROOT creata: %s", root)
	logger.Info("Modalita: %s", func() string {
		if devMode == "true" {
			return "sviluppo"
		}
		return "produzione"
	}())

	if devMode == "true" {
		logger.Info("Apertura Explorer su WEBGAINROOT (dev mode)")
		exec.Command("explorer", root).Start()
	}

	wailsRuntime.EventsEmit(a.ctx, "setup:step", "Verifica moduli...")
	time.Sleep(1 * time.Second)

	logger.Info("Avvio verifica moduli...")
	webgainOnline, err := setup.VerifyModules(a.configFS, a.webgainRoot)
	if err != nil {
		logger.Error("Verifica moduli fallita: %v", err)
		a.fatalCorruptError()
		return
	}
	logger.Info("Verifica moduli completata (online=%v)", webgainOnline)

	wailsRuntime.EventsEmit(a.ctx, "setup:step", "Inizializzazione moduli...")
	time.Sleep(1 * time.Second)

	logger.Info("Avvio inizializzazione moduli...")
	modules, err := setup.InitModules(a.configFS, a.webgainRoot, webgainOnline)
	if err != nil {
		logger.Error("Inizializzazione moduli fallita: %v", err)
		a.fatalCorruptError()
		return
	}
	logger.Info("Inizializzazione moduli completata: %d moduli pronti", len(modules))
	_ = modules

	wailsRuntime.EventsEmit(a.ctx, "setup:done", nil)
	logger.Info("Setup completato")
}

func (a *App) GetEulaText() string {
	data, err := fs.ReadFile(a.configFS, "eula.txt")
	if err != nil {
		return "Errore caricamento EULA."
	}
	return string(data)
}

func (a *App) fatalCorruptError() {
	logger.Error("Errore fatale: installazione corrotta")
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
	logger.Error("Applicazione terminata per errore fatale")
	logger.Close()
	wailsRuntime.Quit(a.ctx)
}
