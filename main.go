package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"WebGainInstaller/internal/admin"
	"WebGainInstaller/internal/font"
	"WebGainInstaller/internal/screen"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed font/*
var fontFS embed.FS

//go:embed media/*
var mediaFS embed.FS

//go:embed config/*
var configFS embed.FS

func main() {
	admin.RequireAdmin()

	fontSubFS, err := fs.Sub(fontFS, "font")
	if err != nil {
		log.Fatal("Errore caricamento font: " + err.Error())
	}
	if err := font.InstallFonts(fontSubFS); err != nil {
		log.Printf("Attenzione: installazione font fallita: %v", err)
	}

	winW, winH := screen.CalculateWindowSize(1150, 900)

	configSubFS, _ := fs.Sub(configFS, "config")
	app := NewApp(configSubFS)

	mediaSubFS, _ := fs.Sub(mediaFS, "media")
	mediaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		data, err := fs.ReadFile(mediaSubFS, path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		switch {
		case strings.HasSuffix(path, ".mp4"):
			w.Header().Set("Content-Type", "video/mp4")
		case strings.HasSuffix(path, ".ico"):
			w.Header().Set("Content-Type", "image/x-icon")
		case strings.HasSuffix(path, ".png"):
			w.Header().Set("Content-Type", "image/png")
		case strings.HasSuffix(path, ".svg"):
			w.Header().Set("Content-Type", "image/svg+xml")
		}
		w.Write(data)
	})

	err = wails.Run(&options.App{
		Title:            "WebGain Installer",
		Width:            winW,
		Height:           winH,
		DisableResize:    true,
		Fullscreen:       false,
		BackgroundColour: &options.RGBA{R: 13, G: 17, B: 23, A: 255},
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: mediaHandler,
		},
		OnStartup:     app.startup,
		OnBeforeClose: app.beforeClose,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		log.Fatal("Errore avvio applicazione: " + err.Error())
	}
}
