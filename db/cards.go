package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (db *DB) CreateCard(deckID int64, front, back, hint string, tags []string) (int64, error) {
	tagsStr := strings.Join(tags, ",")
	now := time.Now().Unix()
	res, err := db.Conn.Exec(`
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
		_, err = db.Conn.Exec(query, card.DeckID, card.Front, card.Back, card.Hint, tagsStr,
			card.RepetitionCount, card.Interval, card.EaseFactor, card.DueAt.Unix(), card.ID)
	}
	return err
}

func (db *DB) DeleteCard(id int64) error {
	_, err := db.Conn.Exec("DELETE FROM cards WHERE id = ?", id)
	return err
}

func (db *DB) GetCard(id int64) (*Card, error) {
	var c Card
	var tagsStr string
	var dueAtUnix int64
	err := db.Conn.QueryRow(`
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
	rows, err := db.Conn.Query(`
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
	deckIDs, err := db.GetSubdeckIDs(deckID)
	if err != nil {
		return nil, err
	}

	if len(deckIDs) == 0 {
		return nil, nil
	}

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

	rows, err := db.Conn.Query(query, args...)
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
