package db

import (
	"database/sql"
	"time"
)

func (db *DB) CreateDeck(name string, parentID *int64) (int64, error) {
	var res sql.Result
	var err error
	if parentID != nil {
		res, err = db.Conn.Exec("INSERT INTO decks (name, parent_id) VALUES (?, ?)", name, *parentID)
	} else {
		res, err = db.Conn.Exec("INSERT INTO decks (name, parent_id) VALUES (?, NULL)", name)
	}
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) DeleteDeck(id int64) error {
	_, err := db.Conn.Exec("DELETE FROM decks WHERE id = ?", id)
	return err
}

func (db *DB) RenameDeck(id int64, newName string) error {
	_, err := db.Conn.Exec("UPDATE decks SET name = ? WHERE id = ?", newName, id)
	return err
}

func (db *DB) GetSubdeckIDs(deckID int64) ([]int64, error) {
	rows, err := db.Conn.Query(`
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

func (db *DB) GetDeckTree() ([]Deck, error) {
	rows, err := db.Conn.Query("SELECT id, name, parent_id FROM decks ORDER BY name ASC")
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

	now := time.Now().Unix()
	cardCounts := make(map[int64]int)
	dueCounts := make(map[int64]int)

	cRows, err := db.Conn.Query("SELECT deck_id, COUNT(*), SUM(CASE WHEN due_at <= ? THEN 1 ELSE 0 END) FROM cards GROUP BY deck_id", now)
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

	for i := range all {
		t, dCount := getCounts(all[i].ID)
		all[i].CardCount = t
		all[i].DueCount = dCount
	}

	var buildTree func(parentID *int64, depth int) []Deck
	buildTree = func(parentID *int64, depth int) []Deck {
		var result []Deck
		for _, d := range all {
			match := (parentID == nil && d.ParentID == nil) || (parentID != nil && d.ParentID != nil && *parentID == *d.ParentID)
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
