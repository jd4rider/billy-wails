package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:             "Billy",
		Width:             420,
		Height:            620,
		MinWidth:          360,
		MinHeight:         480,
		Frameless:         false,
		AlwaysOnTop:       false,
		HideWindowOnClose: true, // stay in tray instead of quitting
		BackgroundColour:  &options.RGBA{R: 13, G: 13, B: 18, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			Appearance:           mac.NSAppearanceNameDarkAqua,
			About: &mac.AboutInfo{
				Title:   "Billy",
				Message: "Local AI coding assistant",
				Icon:    appIcon,
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			BackdropType:                      windows.Mica,
			DisableWindowIcon:                 false,
			IsZoomControlEnabled:              false,
			DisablePinchZoom:                  true,
			DisableFramelessWindowDecorations: false,
		},
		// Disable right-click context menu in production
		EnableDefaultContextMenu: os.Getenv("BILLY_DEV") == "1",
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
