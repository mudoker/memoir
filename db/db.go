package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	Conn *sql.DB
}

type Deck struct {
	ID        int64
	Name      string
	ParentID  *int64
	Depth     int
	Children  []Deck
	CardCount int
	DueCount  int
}

type Card struct {
	ID              int64
	DeckID          int64
	Front           string
	Back            string
	Hint            string
	Tags            []string
	RepetitionCount int
	Interval        int
	EaseFactor      float64
	DueAt           time.Time
}

type ReviewLog struct {
	ID                  int64
	CardID              int64
	ReviewedAt          time.Time
	PreRepetitionCount  int
	PreInterval         int
	PreEaseFactor       float64
	Grade               int
	PostRepetitionCount int
	PostInterval        int
	PostEaseFactor      float64
}

func OpenDB(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{Conn: conn}

	_, err = conn.Exec(`
		PRAGMA journal_mode = WAL;
		PRAGMA foreign_keys = ON;
	`)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set pragmas: %w", err)
	}

	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.Conn.Close()
}

func (db *DB) Transaction(fn func(tx *sql.Tx) error) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS decks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			parent_id INTEGER,
			FOREIGN KEY(parent_id) REFERENCES decks(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			deck_id INTEGER NOT NULL,
			front TEXT NOT NULL,
			back TEXT NOT NULL,
			hint TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '',
			repetition_count INTEGER NOT NULL DEFAULT 0,
			interval INTEGER NOT NULL DEFAULT 0,
			ease_factor REAL NOT NULL DEFAULT 2.5,
			due_at INTEGER NOT NULL,
			FOREIGN KEY(deck_id) REFERENCES decks(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS review_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			card_id INTEGER NOT NULL,
			reviewed_at INTEGER NOT NULL,
			pre_repetition_count INTEGER NOT NULL,
			pre_interval INTEGER NOT NULL,
			pre_ease_factor REAL NOT NULL,
			grade INTEGER NOT NULL,
			post_repetition_count INTEGER NOT NULL,
			post_interval INTEGER NOT NULL,
			post_ease_factor REAL NOT NULL,
			FOREIGN KEY(card_id) REFERENCES cards(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_decks_parent ON decks(parent_id);`,
		`CREATE INDEX IF NOT EXISTS idx_cards_deck ON cards(deck_id);`,
		`CREATE INDEX IF NOT EXISTS idx_cards_due ON cards(due_at);`,
		`CREATE INDEX IF NOT EXISTS idx_review_logs_card ON review_logs(card_id);`,
	}

	for _, q := range queries {
		if _, err := db.Conn.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
