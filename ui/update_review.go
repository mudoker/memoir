package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) UpdateReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.Session == nil {
		m.UIMode = ModeDashboard
		return m, nil
	}

	key := msg.String()

	switch key {
	case "esc", "q":
		m.UIMode = ModeDashboard
		m.RefreshData()
		return m, nil

	case " ":
		m.Session.IsFlipped = true
		return m, nil

	case "h":
		m.Session.ShowHint = true
		return m, nil

	case "s":
		m.Session.ShuffleQueue()
		m.SetStatus("Remaining queue shuffled.", false)
		return m, nil

	case "u":
		if err := m.Session.Undo(m.Database); err != nil {
			m.SetStatus("Undo error: "+err.Error(), true)
		} else {
			m.SetStatus("Rolled back last assessment.", false)
		}
		return m, nil

	case "1", "2", "3", "4", "5":
		grade := int(key[0] - '0')
		if m.Session.ActiveCard != nil {
			if !m.Session.IsFlipped {
				m.Session.IsFlipped = true
				return m, nil
			}

			err := m.Session.GradeActiveCard(grade, m.Database)
			if err != nil {
				m.SetStatus("Database write error: "+err.Error(), true)
			}
		}
		return m, nil
	}

	return m, nil
}
