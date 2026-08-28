package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"safenote/internal/app"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	safeApp := app.NewApp()

	err := wails.Run(&options.App{
		Title:     "SafeNote",
		Width:     1100,
		Height:    750,
		MinWidth:  900,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 17, G: 17, B: 17, A: 1},
		OnStartup:        safeApp.Startup,
		OnShutdown:       safeApp.Shutdown,
		Bind: []interface{}{
			safeApp,
		},
	})
	if err != nil {
		log.Fatalf("safenote: %v", err)
	}
}
