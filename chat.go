package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	ConvID   string    `json:"convId"`   // empty = start new conversation
	UserText string    `json:"userText"` // raw user text, used as conversation title seed
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
// Conversation creation and convID assignment happen in app.go before this runs.
func streamMessage(a *App, req ChatRequest) {
	model := req.Model
	if model == "" {
		model = defaultModel()
	}

	// Save the user message
	_ = writeMessage(newID(), req.ConvID, "user", req.UserText)

	// Prepend system prompt
	if len(req.Messages) == 0 || req.Messages[0].Role != "system" {
		systemMsg := Message{Role: "system", Content: buildSystemPrompt()}
		req.Messages = append([]Message{systemMsg}, req.Messages...)
	}

	// Stream and accumulate the assistant response
	var assistantContent string
	if billyServingNow(a.billyURL) {
		assistantContent = streamViaBilly(a, req)
	} else {
		assistantContent = streamViaOllama(a, req, model)
	}

	// Save assistant message
	if assistantContent != "" {
		_ = writeMessage(newID(), req.ConvID, "assistant", assistantContent)
	}

	// After the first exchange (2 messages), generate an AI title asynchronously
	if countMessages(req.ConvID) == 2 {
		go generateAndSetTitle(a, req.ConvID, model, req.UserText, assistantContent)
	}

	// Emit done — frontend uses convID to stay in the same conversation
	a.window.EmitEvent("chat:done", req.ConvID)
}

// generateAndSetTitle asks Ollama for a short title and updates the DB + notifies the frontend.
func generateAndSetTitle(a *App, convID, model, userMsg, assistantMsg string) {
	title := generateTitle(a.ollamaURL, model, userMsg, assistantMsg)
	if title == "" {
		return
	}
	_ = updateConversationTitle(convID, title)
	a.window.EmitEvent("conv:titled", map[string]string{
		"id":    convID,
		"title": title,
	})
}

// generateTitle calls Ollama non-streaming to produce a short conversation title.
func generateTitle(ollamaURL, model, userMsg, assistantMsg string) string {
	if len(userMsg) > 300 {
		userMsg = userMsg[:300]
	}
	if len(assistantMsg) > 300 {
		assistantMsg = assistantMsg[:300]
	}
	prompt := "Generate a short title (3–6 words) for a chat that begins with:\n" +
		"User: " + userMsg + "\n" +
		"Assistant: " + assistantMsg + "\n\n" +
		"Reply with ONLY the title. No quotes, no trailing punctuation, no explanation."

	payload := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}
	title := strings.TrimSpace(result.Response)
	// Strip surrounding quotes if the model added them anyway
	title = strings.Trim(title, `"'`)
	if len(title) > 80 {
		title = title[:80]
	}
	return title
}

func streamViaBilly(a *App, req ChatRequest) string {
	body, _ := json.Marshal(req)
	resp, err := http.Post(a.billyURL+"/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		a.window.EmitEvent("chat:error", err.Error())
		return ""
	}
	defer resp.Body.Close()
	var buf strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 6 && line[:6] == "data: " {
			token := line[6:]
			if token == "[DONE]" {
				break
			}
			buf.WriteString(token)
			a.window.EmitEvent("chat:token", token)
		}
	}
	return buf.String()
}

func streamViaOllama(a *App, req ChatRequest, model string) string {
	if model == "" {
		model = defaultModel()
	}
	payload := ollamaStreamPayload{Model: model, Messages: req.Messages, Stream: true}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(a.ollamaURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		a.window.EmitEvent("chat:error", fmt.Sprintf("Ollama unreachable: %s", err.Error()))
		return ""
	}
	defer resp.Body.Close()
	var buf strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var chunk ollamaChunk
		if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
			continue
		}
		if chunk.Message.Content != "" {
			buf.WriteString(chunk.Message.Content)
			a.window.EmitEvent("chat:token", chunk.Message.Content)
		}
		if chunk.Done {
			break
		}
	}
	return buf.String()
}
