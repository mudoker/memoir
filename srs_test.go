package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"flashtui/db"
	"flashtui/srs"
)

func TestSM2Algorithm(t *testing.T) {
	now := time.Now()

	card := db.Card{
		ID:              1,
		DeckID:          1,
		Front:           "Test Front",
		Back:            "Test Back",
		Hint:            "Hint",
		Tags:            []string{"test"},
		RepetitionCount: 0,
		Interval:        0,
		EaseFactor:      2.5,
		DueAt:           now,
	}

	failed := srs.CalculateSM2(card, 2, now)
	if failed.RepetitionCount != 0 {
		t.Errorf("Expected RepetitionCount to reset to 0, got %d", failed.RepetitionCount)
	}
	if failed.Interval != 1 {
		t.Errorf("Expected Interval to reset to 1 day, got %d", failed.Interval)
	}
	if failed.EaseFactor != 2.3 {
		t.Errorf("Expected EaseFactor to penalize to 2.3, got %f", failed.EaseFactor)
	}

	success1 := srs.CalculateSM2(card, 4, now)
	if success1.RepetitionCount != 1 {
		t.Errorf("Expected RepetitionCount to increment to 1, got %d", success1.RepetitionCount)
	}
	if success1.Interval != 1 {
		t.Errorf("Expected Interval for first milestone to be 1 day, got %d", success1.Interval)
	}

	success2 := srs.CalculateSM2(success1, 5, now)
	if success2.RepetitionCount != 2 {
		t.Errorf("Expected RepetitionCount to increment to 2, got %d", success2.RepetitionCount)
	}
	if success2.Interval != 6 {
		t.Errorf("Expected Interval for second milestone to be 6 days, got %d", success2.Interval)
	}

	success3 := srs.CalculateSM2(success2, 5, now)
	if success3.RepetitionCount != 3 {
		t.Errorf("Expected RepetitionCount to increment to 3, got %d", success3.RepetitionCount)
	}
	if success3.Interval != 17 {
		t.Errorf("Expected Interval to be 17 days, got %d (EF: %f)", success3.Interval, success3.EaseFactor)
	}
}

func TestDatabaseOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "flashtui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	database, err := db.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer database.Close()

	deckID, err := database.CreateDeck("Languages", nil)
	if err != nil {
		t.Fatalf("CreateDeck failed: %v", err)
	}

	subDeckID, err := database.CreateDeck("Japanese", &deckID)
	if err != nil {
		t.Fatalf("Create sub-deck failed: %v", err)
	}

	ids, err := database.GetSubdeckIDs(deckID)
	if err != nil {
		t.Fatalf("GetSubdeckIDs failed: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("Expected 2 subdeck IDs, got %v", ids)
	}

	cardID, err := database.CreateCard(subDeckID, "Inu", "Dog", "Woof", []string{"animal", "vocab"})
	if err != nil {
		t.Fatalf("CreateCard failed: %v", err)
	}

	card, err := database.GetCard(cardID)
	if err != nil {
		t.Fatalf("GetCard failed: %v", err)
	}
	if card.Front != "Inu" || card.Back != "Dog" || card.Hint != "Woof" {
		t.Errorf("Card data mismatch: %+v", card)
	}

	err = database.DeleteDeck(deckID)
	if err != nil {
		t.Fatalf("DeleteDeck failed: %v", err)
	}

	tree, err := database.GetDeckTree()
	if err != nil {
		t.Fatalf("GetDeckTree failed: %v", err)
	}
	if len(tree) != 0 {
		t.Errorf("Expected deck tree to be empty after cascade delete, got %v", tree)
	}

	_, err = database.GetCard(cardID)
	if err == nil {
		t.Error("Expected card to be deleted via cascade, but it still exists")
	}
}
