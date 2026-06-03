package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"flashtui/config"
	"flashtui/db"
	"flashtui/ui"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	database, err := db.OpenDB(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := bootstrapIfEmpty(database); err != nil {
		fmt.Fprintf(os.Stderr, "Error bootstrapping database: %v\n", err)
		os.Exit(1)
	}

	uiModel := ui.NewModel(database, cfg)
	p := tea.NewProgram(uiModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running FlashTUI: %v\n", err)
		os.Exit(1)
	}
}

func bootstrapIfEmpty(database *db.DB) error {
	bootstrapped, err := database.IsBootstrapped()
	if err != nil {
		return err
	}
	if bootstrapped {
		return nil
	}

	decks, err := database.GetDeckTree()
	if err != nil {
		return err
	}

	if len(decks) > 0 {
		return database.SetBootstrapped()
	}

	parentID, err := database.CreateDeck("Go-Core", nil)
	if err != nil {
		return err
	}

	concurrencyID, err := database.CreateDeck("Concurrency", &parentID)
	if err != nil {
		return err
	}

	_, err = database.CreateDeck("Memory", &parentID)
	if err != nil {
		return err
	}

	_, err = database.CreateDeck("SQL-Optimize", nil)
	if err != nil {
		return err
	}
	_, err = database.CreateDeck("System-Design", nil)
	if err != nil {
		return err
	}

	_, err = database.CreateCard(parentID, "What is a goroutine?", "A lightweight thread of execution managed by the Go runtime.", "Go thread runtime-level", []string{"core"})
	if err != nil {
		return err
	}

	_, err = database.CreateCard(concurrencyID, "Explain channels", "Channels are the pipes that connect concurrent goroutines. You can send values into channels from one goroutine and receive those values into another goroutine.", "concurrency pipe", []string{"chan"})
	if err != nil {
		return err
	}
	_, err = database.CreateCard(concurrencyID, "What is a Mutex lock?", "A Mutual Exclusion lock. It is used to provide synchronization and prevent race conditions when multiple goroutines access shared memory.", "lock synchronization", []string{"sync"})
	if err != nil {
		return err
	}

	return database.SetBootstrapped()
}
