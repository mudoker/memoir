package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

type Deck struct {
	ID       int64
	Name     string
	ParentID *int64
	Depth    int      // Helper for hierarchical tree drawing
	Children []Deck   // Helper for tree structures
	CardCount int     // Total cards in this deck + subdecks
	DueCount  int     // Due cards in this deck + subdecks
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
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}

	// Enable Write-Ahead Logging (WAL mode) and Foreign Keys
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
	return db.conn.Close()
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
		// Add indices for performance
		`CREATE INDEX IF NOT EXISTS idx_decks_parent ON decks(parent_id);`,
		`CREATE INDEX IF NOT EXISTS idx_cards_deck ON cards(deck_id);`,
		`CREATE INDEX IF NOT EXISTS idx_cards_due ON cards(due_at);`,
		`CREATE INDEX IF NOT EXISTS idx_review_logs_card ON review_logs(card_id);`,
	}

	for _, q := range queries {
		if _, err := db.conn.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// Transaction executes a function in an SQLite transaction
func (db *DB) Transaction(fn func(tx *sql.Tx) error) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

// --- Decks ---

func (db *DB) CreateDeck(name string, parentID *int64) (int64, error) {
	var res sql.Result
	var err error
	if parentID != nil {
		res, err = db.conn.Exec("INSERT INTO decks (name, parent_id) VALUES (?, ?)", name, *parentID)
	} else {
		res, err = db.conn.Exec("INSERT INTO decks (name, parent_id) VALUES (?, NULL)", name)
	}
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) DeleteDeck(id int64) error {
	// Foreign key with ON DELETE CASCADE handles child decks, cards, logs automatically.
	_, err := db.conn.Exec("DELETE FROM decks WHERE id = ?", id)
	return err
}

func (db *DB) RenameDeck(id int64, newName string) error {
	_, err := db.conn.Exec("UPDATE decks SET name = ? WHERE id = ?", newName, id)
	return err
}

// GetSubdeckIDs returns a slice containing the deck ID and all nested subdeck IDs recursively
func (db *DB) GetSubdeckIDs(deckID int64) ([]int64, error) {
	rows, err := db.conn.Query(`
		WITH RECURSIVE subdecks AS (
			SELECT id FROM decks WHERE id = ?
			UNION ALL
			SELECT d.id FROM decks d JOIN subdecks s ON d.parent_id = s.id
		)
		SELECT id FROM subdecks;
	`, deckID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// GetDeckTree retrieves all decks, calculates card counts (including subdecks),
// and returns a flat representation showing hierarchy depth.
func (db *DB) GetDeckTree() ([]Deck, error) {
	// 1. Fetch all decks
	rows, err := db.conn.Query("SELECT id, name, parent_id FROM decks ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Deck
	for rows.Next() {
		var d Deck
		var parentID sql.NullInt64
		if err := rows.Scan(&d.ID, &d.Name, &parentID); err != nil {
			return nil, err
		}
		if parentID.Valid {
			pid := parentID.Int64
			d.ParentID = &pid
		}
		all = append(all, d)
	}

	if len(all) == 0 {
		return nil, nil
	}

	// 2. Fetch counts per deck (cards, and due cards)
	now := time.Now().Unix()
	cardCounts := make(map[int64]int)
	dueCounts := make(map[int64]int)

	cRows, err := db.conn.Query("SELECT deck_id, COUNT(*), SUM(CASE WHEN due_at <= ? THEN 1 ELSE 0 END) FROM cards GROUP BY deck_id", now)
	if err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var deckID int64
			var total, due int
			if err := cRows.Scan(&deckID, &total, &due); err == nil {
				cardCounts[deckID] = total
				dueCounts[deckID] = due
			}
		}
	}

	// Helper to get subdeck counts recursively
	var getCounts func(id int64) (int, int)
	getCounts = func(id int64) (int, int) {
		total := cardCounts[id]
		due := dueCounts[id]
		for _, d := range all {
			if d.ParentID != nil && *d.ParentID == id {
				t, dCount := getCounts(d.ID)
				total += t
				due += dCount
			}
		}
		return total, due
	}

	// Populate counts
	for i := range all {
		t, dCount := getCounts(all[i].ID)
		all[i].CardCount = t
		all[i].DueCount = dCount
	}

	// 3. Build tree and flatten using DFS to preserve hierarchy order
	var buildTree func(parentID *int64, depth int) []Deck
	buildTree = func(parentID *int64, depth int) []Deck {
		var result []Deck
		for _, d := range all {
			match := false
			if parentID == nil && d.ParentID == nil {
				match = true
			} else if parentID != nil && d.ParentID != nil && *parentID == *d.ParentID {
				match = true
			}

			if match {
				d.Depth = depth
				d.Children = buildTree(&d.ID, depth+1)
				result = append(result, d)
			}
		}
		return result
	}

	rootDecks := buildTree(nil, 0)

	var flatten func(decks []Deck) []Deck
	flatten = func(decks []Deck) []Deck {
		var result []Deck
		for _, d := range decks {
			result = append(result, d)
			result = append(result, flatten(d.Children)...)
		}
		return result
	}

	return flatten(rootDecks), nil
}

// --- Cards ---

func (db *DB) CreateCard(deckID int64, front, back, hint string, tags []string) (int64, error) {
	tagsStr := strings.Join(tags, ",")
	now := time.Now().Unix()
	res, err := db.conn.Exec(`
		INSERT INTO cards (deck_id, front, back, hint, tags, repetition_count, interval, ease_factor, due_at)
		VALUES (?, ?, ?, ?, ?, 0, 0, 2.5, ?)
	`, deckID, front, back, hint, tagsStr, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) UpdateCard(tx *sql.Tx, card *Card) error {
	tagsStr := strings.Join(card.Tags, ",")
	query := `
		UPDATE cards SET
			deck_id = ?, front = ?, back = ?, hint = ?, tags = ?,
			repetition_count = ?, interval = ?, ease_factor = ?, due_at = ?
		WHERE id = ?
	`
	var err error
	if tx != nil {
		_, err = tx.Exec(query, card.DeckID, card.Front, card.Back, card.Hint, tagsStr,
			card.RepetitionCount, card.Interval, card.EaseFactor, card.DueAt.Unix(), card.ID)
	} else {
		_, err = db.conn.Exec(query, card.DeckID, card.Front, card.Back, card.Hint, tagsStr,
			card.RepetitionCount, card.Interval, card.EaseFactor, card.DueAt.Unix(), card.ID)
	}
	return err
}

func (db *DB) DeleteCard(id int64) error {
	_, err := db.conn.Exec("DELETE FROM cards WHERE id = ?", id)
	return err
}

func (db *DB) GetCard(id int64) (*Card, error) {
	var c Card
	var tagsStr string
	var dueAtUnix int64
	err := db.conn.QueryRow(`
		SELECT id, deck_id, front, back, hint, tags, repetition_count, interval, ease_factor, due_at
		FROM cards WHERE id = ?
	`, id).Scan(&c.ID, &c.DeckID, &c.Front, &c.Back, &c.Hint, &tagsStr, &c.RepetitionCount, &c.Interval, &c.EaseFactor, &dueAtUnix)
	if err != nil {
		return nil, err
	}

	c.DueAt = time.Unix(dueAtUnix, 0)
	if tagsStr != "" {
		c.Tags = strings.Split(tagsStr, ",")
	} else {
		c.Tags = []string{}
	}
	return &c, nil
}

func (db *DB) GetCardsInDeck(deckID int64) ([]Card, error) {
	rows, err := db.conn.Query(`
		SELECT id, deck_id, front, back, hint, tags, repetition_count, interval, ease_factor, due_at
		FROM cards WHERE deck_id = ? ORDER BY id ASC
	`, deckID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []Card
	for rows.Next() {
		var c Card
		var tagsStr string
		var dueAtUnix int64
		if err := rows.Scan(&c.ID, &c.DeckID, &c.Front, &c.Back, &c.Hint, &tagsStr, &c.RepetitionCount, &c.Interval, &c.EaseFactor, &dueAtUnix); err != nil {
			return nil, err
		}
		c.DueAt = time.Unix(dueAtUnix, 0)
		if tagsStr != "" {
			c.Tags = strings.Split(tagsStr, ",")
		} else {
			c.Tags = []string{}
		}
		cards = append(cards, c)
	}
	return cards, nil
}

func (db *DB) GetDueCards(deckID int64) ([]Card, error) {
	// Find all subdeck IDs
	deckIDs, err := db.GetSubdeckIDs(deckID)
	if err != nil {
		return nil, err
	}

	if len(deckIDs) == 0 {
		return nil, nil
	}

	// Form query with placeholders
	placeholders := make([]string, len(deckIDs))
	args := make([]interface{}, len(deckIDs)+1)
	for i, id := range deckIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	now := time.Now().Unix()
	args[len(deckIDs)] = now

	query := fmt.Sprintf(`
		SELECT id, deck_id, front, back, hint, tags, repetition_count, interval, ease_factor, due_at
		FROM cards
		WHERE deck_id IN (%s) AND due_at <= ?
		ORDER BY due_at ASC
	`, strings.Join(placeholders, ","))

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []Card
	for rows.Next() {
		var c Card
		var tagsStr string
		var dueAtUnix int64
		if err := rows.Scan(&c.ID, &c.DeckID, &c.Front, &c.Back, &c.Hint, &tagsStr, &c.RepetitionCount, &c.Interval, &c.EaseFactor, &dueAtUnix); err != nil {
			return nil, err
		}
		c.DueAt = time.Unix(dueAtUnix, 0)
		if tagsStr != "" {
			c.Tags = strings.Split(tagsStr, ",")
		} else {
			c.Tags = []string{}
		}
		cards = append(cards, c)
	}
	return cards, nil
}

// --- Review Logs ---

func (db *DB) CreateReviewLog(tx *sql.Tx, log *ReviewLog) (int64, error) {
	query := `
		INSERT INTO review_logs (card_id, reviewed_at, pre_repetition_count, pre_interval, pre_ease_factor, grade, post_repetition_count, post_interval, post_ease_factor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.Exec(query, log.CardID, log.ReviewedAt.Unix(), log.PreRepetitionCount, log.PreInterval, log.PreEaseFactor, log.Grade, log.PostRepetitionCount, log.PostInterval, log.PostEaseFactor)
	} else {
		res, err = db.conn.Exec(query, log.CardID, log.ReviewedAt.Unix(), log.PreRepetitionCount, log.PreInterval, log.PreEaseFactor, log.Grade, log.PostRepetitionCount, log.PostInterval, log.PostEaseFactor)
	}
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) DeleteLastReviewLog(tx *sql.Tx, cardID int64) (*ReviewLog, error) {
	// Find the last log for this card
	var log ReviewLog
	var reviewedAtUnix int64
	var querySelect = `
		SELECT id, card_id, reviewed_at, pre_repetition_count, pre_interval, pre_ease_factor, grade, post_repetition_count, post_interval, post_ease_factor
		FROM review_logs
		WHERE card_id = ?
		ORDER BY id DESC LIMIT 1
	`
	var err error
	if tx != nil {
		err = tx.QueryRow(querySelect, cardID).Scan(&log.ID, &log.CardID, &reviewedAtUnix, &log.PreRepetitionCount, &log.PreInterval, &log.PreEaseFactor, &log.Grade, &log.PostRepetitionCount, &log.PostInterval, &log.PostEaseFactor)
	} else {
		err = db.conn.QueryRow(querySelect, cardID).Scan(&log.ID, &log.CardID, &reviewedAtUnix, &log.PreRepetitionCount, &log.PreInterval, &log.PreEaseFactor, &log.Grade, &log.PostRepetitionCount, &log.PostInterval, &log.PostEaseFactor)
	}
	if err != nil {
		return nil, err
	}
	log.ReviewedAt = time.Unix(reviewedAtUnix, 0)

	// Delete it
	var queryDelete = "DELETE FROM review_logs WHERE id = ?"
	if tx != nil {
		_, err = tx.Exec(queryDelete, log.ID)
	} else {
		_, err = db.conn.Exec(queryDelete, log.ID)
	}
	if err != nil {
		return nil, err
	}

	return &log, nil
}

// GetDailyStreak computes the number of consecutive calendar days (in local timezone)
// up to today/yesterday that the user has performed at least one review.
func (db *DB) GetDailyStreak() (int, error) {
	rows, err := db.conn.Query("SELECT reviewed_at FROM review_logs ORDER BY reviewed_at DESC")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	reviewedDates := make(map[string]bool)
	for rows.Next() {
		var u int64
		if err := rows.Scan(&u); err != nil {
			return 0, err
		}
		t := time.Unix(u, 0).Local()
		reviewedDates[t.Format("2006-01-02")] = true
	}

	if len(reviewedDates) == 0 {
		return 0, nil
	}

	streak := 0
	checkDate := time.Now().Local()

	// If no reviews today, check if there was one yesterday.
	// If not even yesterday, then streak is 0.
	todayStr := checkDate.Format("2006-01-02")
	yesterdayStr := checkDate.AddDate(0, 0, -1).Format("2006-01-02")

	if !reviewedDates[todayStr] && !reviewedDates[yesterdayStr] {
		return 0, nil
	}

	if !reviewedDates[todayStr] {
		// Streak continues from yesterday
		checkDate = checkDate.AddDate(0, 0, -1)
	}

	for {
		dateStr := checkDate.Format("2006-01-02")
		if reviewedDates[dateStr] {
			streak++
			checkDate = checkDate.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return streak, nil
}

// GetRetentionAccuracy returns the percentage of reviews with grade >= 3 out of total reviews.
func (db *DB) GetRetentionAccuracy() (float64, error) {
	var total, success int
	err := db.conn.QueryRow("SELECT COUNT(*), SUM(CASE WHEN grade >= 3 THEN 1 ELSE 0 END) FROM review_logs").Scan(&total, &success)
	if err != nil {
		// If table is empty, COUNT(*) returns 0, but Scan might fail if it's NULL (e.g. SUM of 0 rows is NULL, handled below)
		return 0.0, nil
	}
	if total == 0 {
		return 0.0, nil
	}
	return (float64(success) / float64(total)) * 100.0, nil
}

// GetMasteryStats returns the total number of cards and the number of "mastered" cards (EaseFactor >= 2.5 and RepetitionCount >= 3)
func (db *DB) GetMasteryStats() (int, int, error) {
	var total, mastered int
	err := db.conn.QueryRow("SELECT COUNT(*), SUM(CASE WHEN repetition_count >= 3 AND ease_factor >= 2.5 THEN 1 ELSE 0 END) FROM cards").Scan(&total, &mastered)
	if err != nil {
		return 0, 0, nil
	}
	return total, mastered, nil
}
