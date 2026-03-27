package main

import (
	"context"
	"os/exec"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// App is the main application struct. All exported methods are callable from the frontend.
type App struct {
	app           *application.App
	window        *application.WebviewWindow
	ollamaURL     string
	billyURL      string
	popOut        bool
	currentConvID string
	billyBinary   string
	startedBilly  *exec.Cmd
}

func NewApp() *App {
	return &App{
		ollamaURL: "http://127.0.0.1:11434",
		billyURL:  "http://127.0.0.1:7437",
	}
}

// ServiceStartup is called by Wails v3 when the service starts.
func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	path, cmd, err := ensureBillyServeRunning(a.billyURL)
	if err == nil {
		a.billyBinary = path
		a.startedBilly = cmd
	} else {
		a.billyBinary = path
		println("Billy desktop startup warning:", err.Error())
	}
	return nil
}

// ServiceShutdown is called by Wails v3 when the service stops.
func (a *App) ServiceShutdown() error {
	if a.startedBilly != nil && a.startedBilly.Process != nil {
		_ = a.startedBilly.Process.Kill()
		_, _ = a.startedBilly.Process.Wait()
	}
	return nil
}

// GetStatus returns current health info for the UI.
func (a *App) GetStatus() StatusInfo {
	return getStatus(a)
}

// SendMessage streams an AI response via events: "chat:token", "chat:done", "chat:error".
func (a *App) SendMessage(req ChatRequest) {
	if a.currentConvID == "" {
		a.currentConvID = newID()
		model := req.Model
		if model == "" {
			model = defaultModel()
		}
		_ = writeConversation(a.currentConvID, "New conversation", model)
	}
	req.ConvID = a.currentConvID
	go streamMessage(a, req)
}

// SetActiveConversation resumes an existing conversation.
func (a *App) SetActiveConversation(convID string) {
	a.currentConvID = convID
}

// NewConversation clears the active conversation so the next message starts a fresh one.
func (a *App) NewConversation() {
	a.currentConvID = ""
}

// ListModels returns locally available Ollama model names.
func (a *App) ListModels() []string {
	return listModels(a.ollamaURL)
}

// PopOut toggles the window between tray-panel size and floating window size.
func (a *App) PopOut() {
	if !a.popOut {
		a.window.SetSize(720, 800)
		a.window.SetAlwaysOnTop(true)
		a.window.Center()
		a.popOut = true
	} else {
		a.window.SetSize(500, 640)
		a.window.SetAlwaysOnTop(false)
		a.popOut = false
	}
}

// OpenInstallPage opens the Billy install page in the default browser.
func (a *App) OpenInstallPage() {
	url := "https://billysh.online/#install"
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

// GetConversations returns recent conversations.
func (a *App) GetConversations() []ConversationSummary {
	convs, _ := readConversations()
	if convs == nil {
		return []ConversationSummary{}
	}
	return convs
}

// GetMessages returns messages for a given conversation ID.
func (a *App) GetMessages(convID string) []HistoryMessage {
	msgs, _ := readMessages(convID)
	if msgs == nil {
		return []HistoryMessage{}
	}
	return msgs
}

// GetMemories returns all memories stored by the CLI.
func (a *App) GetMemories() []MemoryItem {
	mems, _ := readMemories()
	if mems == nil {
		return []MemoryItem{}
	}
	return mems
}

// AddMemory saves a new memory and returns its ID.
func (a *App) AddMemory(content string) string {
	id, _ := writeMemory(content)
	return id
}

// DeleteMemory removes a memory by ID.
func (a *App) DeleteMemory(id string) {
	_ = deleteMemory(id)
}
