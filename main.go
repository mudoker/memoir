package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// 1. Load config
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 2. Open DB
	db, err := OpenDB(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 3. First-run bootstrap (if database is empty, seed demo cards matching SRS specifications)
	if err := bootstrapIfEmpty(db); err != nil {
		fmt.Fprintf(os.Stderr, "Error bootstrapping database: %v\n", err)
		os.Exit(1)
	}

	// 4. Run TUI program
	p := tea.NewProgram(NewModel(db, cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running FlashTUI: %v\n", err)
		os.Exit(1)
	}
}

func bootstrapIfEmpty(db *DB) error {
	decks, err := db.GetDeckTree()
	if err != nil {
		return err
	}

	// If decks exist, no need to seed
	if len(decks) > 0 {
		return nil
	}

	// Create Go-Core deck
	parentID, err := db.CreateDeck("Go-Core", nil)
	if err != nil {
		return err
	}

	// Create sub-decks matching wireframe: Concurrency, Memory
	concurrencyID, err := db.CreateDeck("Concurrency", &parentID)
	if err != nil {
		return err
	}

	_, err = db.CreateDeck("Memory", &parentID)
	if err != nil {
		return err
	}

	// Create other root decks: SQL-Optimize, System-Design
	_, err = db.CreateDeck("SQL-Optimize", nil)
	if err != nil {
		return err
	}
	_, err = db.CreateDeck("System-Design", nil)
	if err != nil {
		return err
	}

	// Seed cards inside Go-Core
	_, err = db.CreateCard(parentID, "What is a goroutine?", "A lightweight thread of execution managed by the Go runtime.", "Go thread runtime-level", []string{"core"})
	if err != nil {
		return err
	}

	// Seed cards inside Concurrency
	_, err = db.CreateCard(concurrencyID, "Explain channels", "Channels are the pipes that connect concurrent goroutines. You can send values into channels from one goroutine and receive those values into another goroutine.", "concurrency pipe", []string{"chan"})
	if err != nil {
		return err
	}
	_, err = db.CreateCard(concurrencyID, "What is a Mutex lock?", "A Mutual Exclusion lock. It is used to provide synchronization and prevent race conditions when multiple goroutines access shared memory.", "lock synchronization", []string{"sync"})
	if err != nil {
		return err
	}

	return nil
}
