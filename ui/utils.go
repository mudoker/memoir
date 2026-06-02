package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"flashtui/db"
)

// Markdown Import Logic
func ImportMarkdown(database *db.DB, deckID int64, filePath string) (int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(content), "\n")
	var cardsCreated int
	var currentFront string
	var currentBack []string
	var currentHint string
	var currentTags []string

	saveCurrentCard := func() error {
		if currentFront != "" {
			backStr := strings.TrimSpace(strings.Join(currentBack, "\n"))
			_, err := database.CreateCard(deckID, currentFront, backStr, currentHint, currentTags)
			if err != nil {
				return err
			}
			cardsCreated++
		}
		currentFront = ""
		currentBack = nil
		currentHint = ""
		currentTags = nil
		return nil
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 && strings.HasPrefix(parts[0], "#") {
				if err := saveCurrentCard(); err != nil {
					return cardsCreated, err
				}
				currentFront = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(trimmed, "<!--") && strings.HasSuffix(trimmed, "-->") {
			inner := strings.TrimSpace(trimmed[4 : len(trimmed)-3])
			if strings.HasPrefix(strings.ToLower(inner), "hint:") {
				currentHint = strings.TrimSpace(inner[5:])
			} else if strings.HasPrefix(strings.ToLower(inner), "tags:") {
				tagsPart := strings.TrimSpace(inner[5:])
				rawTags := strings.Split(tagsPart, ",")
				for _, t := range rawTags {
					t = strings.TrimSpace(t)
					if t != "" {
						currentTags = append(currentTags, t)
					}
				}
			}
		} else {
			if currentFront != "" {
				currentBack = append(currentBack, line)
			}
		}
	}

	if err := saveCurrentCard(); err != nil {
		return cardsCreated, err
	}

	return cardsCreated, nil
}

// JSON Serialization Export Logic
type ExportDeck struct {
	Name     string       `json:"name"`
	Cards    []ExportCard `json:"cards,omitempty"`
	SubDecks []ExportDeck `json:"sub_decks,omitempty"`
}

type ExportCard struct {
	Front string   `json:"front"`
	Back  string   `json:"back"`
	Hint  string   `json:"hint"`
	Tags  []string `json:"tags"`
}

func compileExportDeck(database *db.DB, deckID int64) (ExportDeck, error) {
	var name string
	err := database.Conn.QueryRow("SELECT name FROM decks WHERE id = ?", deckID).Scan(&name)
	if err != nil {
		return ExportDeck{}, err
	}

	dbCards, err := database.GetCardsInDeck(deckID)
	if err != nil {
		return ExportDeck{}, err
	}

	cards := make([]ExportCard, len(dbCards))
	for i, c := range dbCards {
		cards[i] = ExportCard{
			Front: c.Front,
			Back:  c.Back,
			Hint:  c.Hint,
			Tags:  c.Tags,
		}
	}

	rows, err := database.Conn.Query("SELECT id FROM decks WHERE parent_id = ?", deckID)
	if err != nil {
		return ExportDeck{}, err
	}
	defer rows.Close()

	var childIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return ExportDeck{}, err
		}
		childIDs = append(childIDs, id)
	}

	var subDecks []ExportDeck
	for _, cid := range childIDs {
		sd, err := compileExportDeck(database, cid)
		if err != nil {
			return ExportDeck{}, err
		}
		subDecks = append(subDecks, sd)
	}

	return ExportDeck{
		Name:     name,
		Cards:    cards,
		SubDecks: subDecks,
	}, nil
}

func ExportDeckToPath(database *db.DB, deckID int64, filePath string) error {
	exportData, err := compileExportDeck(database, deckID)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, jsonData, 0644)
}
