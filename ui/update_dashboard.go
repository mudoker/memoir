package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) UpdateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if m.GPressed {
		m.GPressed = false
		if key == "g" {
			if m.ActivePanel == PanelDecks {
				m.SelectedDeckIdx = 0
				m.DeckScrollOffset = 0
			} else {
				m.SelectedCardIdx = 0
				m.CardScrollOffset = 0
			}
			m.RefreshData()
			return m, nil
		}
	}

	switch key {
	case "q":
		_ = m.Database.Close()
		return m, tea.Quit

	case ":":
		m.UIMode = ModeConsole
		m.ConsoleInput.SetValue("")
		m.ConsoleInput.Focus()
		return m, textinput.Blink

	case "/":
		m.UIMode = ModeSearch
		m.SearchInput.Focus()
		m.ActivePanel = PanelCards
		return m, textinput.Blink

	case "h":
		m.ActivePanel = PanelDecks
		m.RefreshData()
	case "l":
		m.ActivePanel = PanelCards
		m.RefreshData()

	case "j", "down":
		if m.ActivePanel == PanelDecks {
			if len(m.Decks) > 0 {
				m.SelectedDeckIdx = (m.SelectedDeckIdx + 1) % len(m.Decks)
				m.adjustDeckScroll()
			}
		} else {
			if len(m.FilteredCards) > 0 {
				m.SelectedCardIdx = (m.SelectedCardIdx + 1) % len(m.FilteredCards)
				m.adjustCardScroll()
			}
		}
		m.RefreshData()

	case "k", "up":
		if m.ActivePanel == PanelDecks {
			if len(m.Decks) > 0 {
				m.SelectedDeckIdx = (m.SelectedDeckIdx - 1 + len(m.Decks)) % len(m.Decks)
				m.adjustDeckScroll()
			}
		} else {
			if len(m.FilteredCards) > 0 {
				m.SelectedCardIdx = (m.SelectedCardIdx - 1 + len(m.FilteredCards)) % len(m.FilteredCards)
				m.adjustCardScroll()
			}
		}
		m.RefreshData()

	case "g":
		m.GPressed = true
		return m, nil

	case "G":
		if m.ActivePanel == PanelDecks {
			if len(m.Decks) > 0 {
				m.SelectedDeckIdx = len(m.Decks) - 1
				m.adjustDeckScroll()
			}
		} else {
			if len(m.FilteredCards) > 0 {
				m.SelectedCardIdx = len(m.FilteredCards) - 1
				m.adjustCardScroll()
			}
		}
		m.RefreshData()

	case "ctrl+d":
		step := (m.Height - 10) / 2
		if m.ActivePanel == PanelDecks {
			m.SelectedDeckIdx += step
			if m.SelectedDeckIdx >= len(m.Decks) {
				m.SelectedDeckIdx = len(m.Decks) - 1
			}
			m.adjustDeckScroll()
		} else {
			m.SelectedCardIdx += step
			if m.SelectedCardIdx >= len(m.FilteredCards) {
				m.SelectedCardIdx = len(m.FilteredCards) - 1
			}
			m.adjustCardScroll()
		}
		m.RefreshData()

	case "ctrl+u":
		step := (m.Height - 10) / 2
		if m.ActivePanel == PanelDecks {
			m.SelectedDeckIdx -= step
			if m.SelectedDeckIdx < 0 {
				m.SelectedDeckIdx = 0
			}
			m.adjustDeckScroll()
		} else {
			m.SelectedCardIdx -= step
			if m.SelectedCardIdx < 0 {
				m.SelectedCardIdx = 0
			}
			m.adjustCardScroll()
		}
		m.RefreshData()

	case "a", "e", "d d", "enter":
		return m.handleDashboardActions(key)
	}

	return m, nil
}
