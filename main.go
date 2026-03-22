package main

import (
	"embed"
	"os"
	"slices"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

//go:embed build/trayicon.png
var trayIconData []byte

func main() {
	// Detect if launched via "login item" (--login flag set when registering the login item)
	launchedAtLogin := slices.Contains(os.Args[1:], "--login")

	appService := NewApp()

	app := application.New(application.Options{
		Name:        "Billy",
		Description: "Local AI coding assistant",
		Icon:        appIcon,
		Services: []application.Service{
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "Billy",
		Width:     500,
		Height:    640,
		MinWidth:  420,
		MinHeight: 480,
		Hidden:    true, // always start hidden; tray click or startup shows it
		Mac: application.MacWindow{
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInsetUnified,
			InvisibleTitleBarHeight: 30,
		},
		BackgroundColour:           application.NewRGBA(13, 13, 18, 255),
		URL:                        "/",
		DefaultContextMenuDisabled: os.Getenv("BILLY_DEV") != "1",
		HideOnFocusLost:            true,
	})

	appService.app = app
	appService.window = window

	// System tray
	tray := app.SystemTray.New()
	if len(trayIconData) > 0 {
		tray.SetTemplateIcon(trayIconData)
	}
	tray.AttachWindow(window).WindowOffset(5)

	// Right-click menu
	menu := app.NewMenu()
	menu.Add("Show Billy").OnClick(func(ctx *application.Context) {
		window.Show()
		window.Focus()
	})
	menu.AddSeparator()
	menu.Add("Quit Billy").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
	tray.SetMenu(menu)

	// On first launch (not login-item), drop the window down from the tray icon
	if !launchedAtLogin {
		app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(e *application.ApplicationEvent) {
			tray.ShowWindow()
		})
	}

	if err := app.Run(); err != nil {
		println("Error:", err.Error())
	}
}
