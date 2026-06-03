package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) UpdateAdvice(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if m.GPressed {
		m.GPressed = false
		if key == "g" {
			m.AdviceScrollOffset = 0
			return m, nil
		}
	}

	switch key {
	case "esc", "q":
		m.UIMode = ModeDashboard
		m.RefreshData()
		return m, nil

	case "j", "down":
		m.AdviceScrollOffset++
		return m, nil

	case "k", "up":
		m.AdviceScrollOffset--
		if m.AdviceScrollOffset < 0 {
			m.AdviceScrollOffset = 0
		}
		return m, nil

	case "ctrl+d", "pgdown":
		m.AdviceScrollOffset += 10
		return m, nil

	case "ctrl+u", "pgup":
		m.AdviceScrollOffset -= 10
		if m.AdviceScrollOffset < 0 {
			m.AdviceScrollOffset = 0
		}
		return m, nil

	case "g":
		m.GPressed = true
		return m, nil

	case "G":
		m.AdviceScrollOffset = 999
		return m, nil
	}

	return m, nil
}
