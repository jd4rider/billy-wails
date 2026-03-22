package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
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

// MemoryItem is a single remembered fact.
type MemoryItem struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

func newID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func dbPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".localai", "history.db")
}

func openDB() (*sql.DB, error) {
	path := dbPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", "file:"+path+"?_journal=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	if err := ensureSchema(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

const schema = `
CREATE TABLE IF NOT EXISTS conversations (
	id                TEXT PRIMARY KEY,
	title             TEXT NOT NULL,
	model             TEXT NOT NULL,
	compacted_summary TEXT NOT NULL DEFAULT '',
	created_at        TIMESTAMP NOT NULL,
	updated_at        TIMESTAMP NOT NULL
);
CREATE TABLE IF NOT EXISTS messages (
	id              TEXT PRIMARY KEY,
	conversation_id TEXT NOT NULL,
	role            TEXT NOT NULL,
	content         TEXT NOT NULL,
	created_at      TIMESTAMP NOT NULL,
	FOREIGN KEY(conversation_id) REFERENCES conversations(id)
);
CREATE TABLE IF NOT EXISTS memories (
	id         TEXT PRIMARY KEY,
	content    TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL
);
`

func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(schema)
	if err != nil {
		return err
	}
	// Migration guard: add column if upgrading from older CLI DB
	_, _ = db.Exec(`ALTER TABLE conversations ADD COLUMN compacted_summary TEXT NOT NULL DEFAULT ''`)
	return nil
}

// ── Read ──────────────────────────────────────────────────────────────────────

func readConversations() ([]ConversationSummary, error) {
	db, err := openDB()
	if err != nil {
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
	if err != nil {
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
	if err != nil {
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

// ── Write ─────────────────────────────────────────────────────────────────────

func writeConversation(id, title, model string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	now := time.Now()
	_, err = db.Exec(
		`INSERT INTO conversations (id, title, model, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		id, title, model, now, now,
	)
	return err
}

func writeMessage(msgID, convID, role, content string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	now := time.Now()
	_, err = db.Exec(
		`INSERT INTO messages (id, conversation_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`,
		msgID, convID, role, content, now,
	)
	if err != nil {
		return err
	}
	_, err = db.Exec(
		`UPDATE conversations SET updated_at = ? WHERE id = ?`, now, convID,
	)
	return err
}

// writeMemory saves a new memory and returns its ID.
func writeMemory(content string) (string, error) {
	db, err := openDB()
	if err != nil {
		return "", err
	}
	defer db.Close()
	id := newID()
	_, err = db.Exec(
		`INSERT INTO memories (id, content, created_at) VALUES (?, ?, ?)`,
		id, content, time.Now(),
	)
	return id, err
}

// deleteMemory removes a memory by ID.
func deleteMemory(id string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`DELETE FROM memories WHERE id = ?`, id)
	return err
}

// ── System prompt ─────────────────────────────────────────────────────────────

// buildSystemPrompt constructs the system message injected before every chat.
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

// updateConversationTitle updates the title of a conversation after AI generation.
func updateConversationTitle(convID, title string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`UPDATE conversations SET title = ? WHERE id = ?`, title, convID)
	return err
}

// countMessages returns how many messages exist for a conversation.
func countMessages(convID string) int {
	db, err := openDB()
	if err != nil {
		return 0
	}
	defer db.Close()
	var n int
	db.QueryRow(`SELECT COUNT(*) FROM messages WHERE conversation_id = ?`, convID).Scan(&n)
	return n
}


