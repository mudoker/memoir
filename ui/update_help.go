package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) UpdateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if m.GPressed {
		m.GPressed = false
		if key == "g" {
			m.HelpScrollOffset = 0
			return m, nil
		}
	}

	switch key {
	case "esc", "q":
		m.UIMode = ModeDashboard
		m.RefreshData()
		return m, nil

	case "j", "down":
		m.HelpScrollOffset++
		return m, nil

	case "k", "up":
		m.HelpScrollOffset--
		if m.HelpScrollOffset < 0 {
			m.HelpScrollOffset = 0
		}
		return m, nil

	case "ctrl+d", "pgdown":
		m.HelpScrollOffset += 10
		return m, nil

	case "ctrl+u", "pgup":
		m.HelpScrollOffset -= 10
		if m.HelpScrollOffset < 0 {
			m.HelpScrollOffset = 0
		}
		return m, nil

	case "g":
		m.GPressed = true
		return m, nil

	case "G":
		// Set to a large number; the view renderer will clamp it to the max available offset
		m.HelpScrollOffset = 999
		return m, nil
	}

	return m, nil
}
