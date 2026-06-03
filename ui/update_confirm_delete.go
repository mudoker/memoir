package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) UpdateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.UIMode = ModeDashboard
		if m.ActivePanel == PanelDecks {
			if len(m.Decks) == 0 {
				return m, nil
			}
			d := m.Decks[m.SelectedDeckIdx]
			if err := m.Database.DeleteDeck(d.ID); err != nil {
				m.SetStatus("Delete Deck error: "+err.Error(), true)
			} else {
				m.SetStatus(fmt.Sprintf("Deleted deck '%s' and all contents recursively.", d.Name), false)
				m.SelectedDeckIdx = 0
				m.RefreshData()
			}
		} else {
			if len(m.FilteredCards) == 0 {
				return m, nil
			}
			c := m.FilteredCards[m.SelectedCardIdx]
			if err := m.Database.DeleteCard(c.ID); err != nil {
				m.SetStatus("Delete Card error: "+err.Error(), true)
			} else {
				m.SetStatus("Deleted card.", false)
				m.SelectedCardIdx = 0
				m.RefreshData()
			}
		}
		return m, nil

	case "esc", "q":
		m.UIMode = ModeDashboard
		m.SetStatus("Deletion canceled.", false)
		return m, nil
	}

	return m, nil
}
