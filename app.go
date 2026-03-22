package main

import (
"context"
"os/exec"
"runtime"

wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct. All exported methods are callable from the frontend.
type App struct {
	ctx          context.Context
	ollamaURL    string
	billyURL     string
	popOut       bool
	currentConvID string
}

func NewApp() *App {
return &App{
ollamaURL: "http://127.0.0.1:11434",
billyURL:  "http://127.0.0.1:7437",
}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	initTray(a)
}

func (a *App) shutdown(ctx context.Context) {}

// GetStatus returns current health info for the UI.
func (a *App) GetStatus() StatusInfo {
return getStatus(a)
}

// SendMessage streams an AI response via events: "chat:token", "chat:done", "chat:error".
// chat:done emits the convID so the frontend can track the active conversation.
func (a *App) SendMessage(req ChatRequest) {
	// Create a new conversation on first message of a session
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

// SetActiveConversation resumes an existing conversation (from history sidebar or CLI).
// All subsequent messages will be appended to this conversation.
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

// GetConversations returns recent conversations from ~/.localai/history.db.
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
