package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) UpdateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.UIMode = ModeDashboard
		m.RefreshData()
		return m, nil
	}
	return m, nil
}
