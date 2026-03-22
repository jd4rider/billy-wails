package main

import (
"context"
"os/exec"
"runtime"

wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct. All exported methods are callable from the frontend.
type App struct {
ctx      context.Context
ollamaURL string
billyURL  string
popOut    bool
}

func NewApp() *App {
return &App{
ollamaURL: "http://127.0.0.1:11434",
billyURL:  "http://127.0.0.1:7437",
}
}

func (a *App) startup(ctx context.Context) {
a.ctx = ctx
go runTray(a)
}

func (a *App) shutdown(ctx context.Context) {}

// GetStatus returns current health info for the UI.
func (a *App) GetStatus() StatusInfo {
return getStatus(a)
}

// SendMessage streams an AI response via events: "chat:token", "chat:done", "chat:error".
func (a *App) SendMessage(req ChatRequest) {
go streamMessage(a, req)
}

// ListModels returns locally available Ollama model names.
func (a *App) ListModels() []string {
return listModels(a.ollamaURL)
}

// PopOut toggles the window between tray-panel size and floating window size.
func (a *App) PopOut() {
if !a.popOut {
wailsRuntime.WindowSetSize(a.ctx, 720, 800)
wailsRuntime.WindowSetAlwaysOnTop(a.ctx, true)
wailsRuntime.WindowCenter(a.ctx)
a.popOut = true
} else {
wailsRuntime.WindowSetSize(a.ctx, 420, 620)
wailsRuntime.WindowSetAlwaysOnTop(a.ctx, false)
a.popOut = false
}
}

// OpenInstallPage opens the billy.sh install page in the default browser.
func (a *App) OpenInstallPage() {
url := "https://jd4rider.github.io/billy-web/#install"
switch runtime.GOOS {
case "darwin":
exec.Command("open", url).Start()
case "windows":
exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
default:
exec.Command("xdg-open", url).Start()
}
}

// GetPlatform returns the OS ("darwin", "windows", "linux").
func (a *App) GetPlatform() string {
return runtime.GOOS
}
