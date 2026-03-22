package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Message is a single chat message (user or assistant).
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is sent from the frontend to initiate a streaming response.
type ChatRequest struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`
}

type ollamaStreamPayload struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ollamaChunk struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

// streamMessage is the goroutine that drives AI streaming.
// It prefers billy serve if available, falling back to Ollama directly.
func streamMessage(a *App, req ChatRequest) {
	if billyServingNow(a.billyURL) {
		streamViaBilly(a, req)
		return
	}
	streamViaOllama(a, req)
}

func streamViaBilly(a *App, req ChatRequest) {
	body, _ := json.Marshal(req)
	resp, err := http.Post(a.billyURL+"/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "chat:error", err.Error())
		return
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 6 && line[:6] == "data: " {
			token := line[6:]
			if token == "[DONE]" {
				break
			}
			wailsRuntime.EventsEmit(a.ctx, "chat:token", token)
		}
	}
	wailsRuntime.EventsEmit(a.ctx, "chat:done", "")
}

func streamViaOllama(a *App, req ChatRequest) {
	model := req.Model
	if model == "" {
		model = defaultModel()
	}
	payload := ollamaStreamPayload{Model: model, Messages: req.Messages, Stream: true}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(a.ollamaURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "chat:error", fmt.Sprintf("Ollama unreachable: %s", err.Error()))
		return
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var chunk ollamaChunk
		if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
			continue
		}
		if chunk.Message.Content != "" {
			wailsRuntime.EventsEmit(a.ctx, "chat:token", chunk.Message.Content)
		}
		if chunk.Done {
			break
		}
	}
	wailsRuntime.EventsEmit(a.ctx, "chat:done", "")
}
