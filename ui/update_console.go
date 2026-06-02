package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) UpdateConsole(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		cmdText := m.ConsoleInput.Value()
		m.UIMode = ModeDashboard
		return m.executeConsoleCommand(cmdText)

	case "esc":
		m.UIMode = ModeDashboard
		return m, nil

	default:
		m.ConsoleInput, cmd = m.ConsoleInput.Update(msg)
		return m, cmd
	}
}

func (m Model) executeConsoleCommand(cmdText string) (tea.Model, tea.Cmd) {
	cmdText = strings.TrimSpace(cmdText)
	if cmdText == "" {
		return m, nil
	}

	parts := strings.Split(cmdText, " ")
	op := parts[0]

	switch op {
	case ":w":
		m.SetStatus("Synced all session database states.", false)
		m.RefreshData()

	case ":q":
		_ = m.Database.Close()
		return m, tea.Quit

	case ":import":
		if len(parts) < 2 {
			m.SetStatus("Usage: :import <path>", true)
			return m, nil
		}
		if len(m.Decks) == 0 {
			m.SetStatus("Please create a deck first before importing cards.", true)
			return m, nil
		}
		path := strings.Join(parts[1:], " ")
		deck := m.Decks[m.SelectedDeckIdx]
		count, err := ImportMarkdown(m.Database, deck.ID, path)
		if err != nil {
			m.SetStatus(fmt.Sprintf("Import error: %v", err), true)
		} else {
			m.SetStatus(fmt.Sprintf("Successfully imported %d cards into '%s'.", count, deck.Name), false)
			m.RefreshData()
		}

	case ":export":
		if len(parts) < 3 {
			m.SetStatus("Usage: :export <deck_name> <path>", true)
			return m, nil
		}
		deckName := parts[1]
		path := strings.Join(parts[2:], " ")

		var targetDeck *db.Deck
		for _, d := range m.Decks {
			if strings.EqualFold(d.Name, deckName) {
				td := d
				targetDeck = &td
				break
			}
		}

		if targetDeck == nil {
			m.SetStatus(fmt.Sprintf("Deck '%s' not found.", deckName), true)
			return m, nil
		}

		err := ExportDeckToPath(m.Database, targetDeck.ID, path)
		if err != nil {
			m.SetStatus(fmt.Sprintf("Export error: %v", err), true)
		} else {
			m.SetStatus(fmt.Sprintf("Successfully exported '%s' to '%s'.", deckName, path), false)
		}

	default:
		m.SetStatus("Unknown command: "+op, true)
	}

	return m, nil
}

func (m Model) UpdateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter", "esc":
		m.UIMode = ModeDashboard
		m.RefreshData()
		return m, nil

	default:
		m.SearchInput, cmd = m.SearchInput.Update(msg)
		m.ApplySearchFilter()
		m.SelectedCardIdx = 0
		return m, cmd
	}
}
