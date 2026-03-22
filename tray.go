package main

import (
	_ "embed"

	"fyne.io/systray"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed build/trayicon.png
var trayIconData []byte

// runTray initializes the system tray icon.
// Must be called from a goroutine; blocks until systray.Quit() is called.
func runTray(a *App) {
	systray.Run(func() {
		systray.SetIcon(trayIconData)
		systray.SetTitle("")
		systray.SetTooltip("Billy — local AI assistant")

		mShow := systray.AddMenuItem("Show Billy", "Open the Billy window")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit Billy", "Exit the application")

		go func() {
			for {
				select {
				case <-mShow.ClickedCh:
					if a.ctx != nil {
						wailsRuntime.WindowShow(a.ctx)
						wailsRuntime.WindowSetAlwaysOnTop(a.ctx, true)
						wailsRuntime.WindowSetAlwaysOnTop(a.ctx, false)
					}
				case <-mQuit.ClickedCh:
					systray.Quit()
					if a.ctx != nil {
						wailsRuntime.Quit(a.ctx)
					}
				}
			}
		}()
	}, func() {
		// on exit
	})
}
