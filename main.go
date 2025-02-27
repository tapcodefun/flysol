package main

import (
	"embed"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/application"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

var version = "0.0.1"

const appName = "Next Solana"

func main() {
	// Create an instance of the app structure
	app := NewApp()
	agent := NewAgent()

	servers := application.NewWithOptions(&options.App{
		Title:  "Solana MevBot",
		Width:  1200,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app, agent,
		},
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title:   fmt.Sprintf("%s %s", appName, version),
				Message: "A Solana MevBot desktop client.\n\nCopyright © 2025",
				Icon:    icon,
			},
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableFramelessWindowDecorations: false,
		},
	})

	// Create application with options
	err := servers.Run()
	if err != nil {
		println("Error:", err.Error())
	}
}
