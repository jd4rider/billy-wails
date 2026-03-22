//go:build darwin

package main

/*
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>

void billyInitTray(const unsigned char *iconBytes, int iconLen);
*/
import "C"

import (
	_ "embed"
	"unsafe"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed build/trayicon.png
var trayIconData []byte

var _trayApp *App

//export trayShowCallback
func trayShowCallback() {
	if _trayApp != nil && _trayApp.ctx != nil {
		wailsRuntime.WindowShow(_trayApp.ctx)
		wailsRuntime.WindowSetAlwaysOnTop(_trayApp.ctx, true)
		wailsRuntime.WindowSetAlwaysOnTop(_trayApp.ctx, false)
	}
}

//export trayQuitCallback
func trayQuitCallback() {
	if _trayApp != nil && _trayApp.ctx != nil {
		wailsRuntime.Quit(_trayApp.ctx)
	}
}

func initTray(a *App) {
	_trayApp = a
	if len(trayIconData) == 0 {
		return
	}
	// billyInitTray copies bytes into NSData synchronously before
	// dispatch_async returns, so it is safe to free immediately after.
	cData := C.CBytes(trayIconData)
	C.billyInitTray((*C.uchar)(cData), C.int(len(trayIconData)))
	C.free(unsafe.Pointer(cData))
}
