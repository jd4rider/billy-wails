package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// ConversationSummary is sent to the frontend for the history sidebar.
type ConversationSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Model     string `json:"model"`
	UpdatedAt string `json:"updatedAt"`
}

// HistoryMessage is a single message from a past conversation.
type HistoryMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

// MemoryItem is a single remembered fact from the CLI.
type MemoryItem struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

func dbPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".localai", "history.db")
}

func openDB() (*sql.DB, error) {
	path := dbPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil // DB doesn't exist yet — billy CLI hasn't been run
	}
	return sql.Open("sqlite", "file:"+path+"?mode=ro&_journal=WAL")
}

func readConversations() ([]ConversationSummary, error) {
	db, err := openDB()
	if err != nil || db == nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT id, title, model, updated_at FROM conversations ORDER BY updated_at DESC LIMIT 100`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ConversationSummary
	for rows.Next() {
		var c ConversationSummary
		var updatedAt time.Time
		if err := rows.Scan(&c.ID, &c.Title, &c.Model, &updatedAt); err != nil {
			continue
		}
		c.UpdatedAt = updatedAt.Format("Jan 2")
		out = append(out, c)
	}
	return out, rows.Err()
}

func readMessages(convID string) ([]HistoryMessage, error) {
	db, err := openDB()
	if err != nil || db == nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT role, content, created_at FROM messages WHERE conversation_id = ? ORDER BY created_at ASC`,
		convID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []HistoryMessage
	for rows.Next() {
		var m HistoryMessage
		var createdAt time.Time
		if err := rows.Scan(&m.Role, &m.Content, &createdAt); err != nil {
			continue
		}
		m.CreatedAt = createdAt.Format("15:04")
		out = append(out, m)
	}
	return out, rows.Err()
}

func readMemories() ([]MemoryItem, error) {
	db, err := openDB()
	if err != nil || db == nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT id, content, created_at FROM memories ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []MemoryItem
	for rows.Next() {
		var m MemoryItem
		var createdAt time.Time
		if err := rows.Scan(&m.ID, &m.Content, &createdAt); err != nil {
			continue
		}
		m.CreatedAt = createdAt.Format("Jan 2, 2006")
		out = append(out, m)
	}
	return out, rows.Err()
}

// buildSystemPrompt constructs the system message injected before every chat.
// It includes Billy's identity and any memories stored by the CLI.
func buildSystemPrompt() string {
	base := `You are Billy, a local AI coding assistant built by billy.sh. You help developers write, debug, explain, and improve code. You are concise, practical, and prefer showing code examples over long explanations. You run entirely locally — no data ever leaves the user's machine.`

	mems, err := readMemories()
	if err != nil || len(mems) == 0 {
		return base
	}

	base += "\n\nThings you remember about this user:"
	for _, m := range mems {
		base += "\n- " + m.Content
	}
	return base
}
