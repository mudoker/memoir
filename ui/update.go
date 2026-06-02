package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			_ = m.Database.Close()
			return m, tea.Quit
		}

		if time.Since(m.StatusTime) > 3*time.Second {
			m.StatusMsg = ""
		}

		switch m.UIMode {
		case ModeConsole:
			return m.UpdateConsole(msg)
		case ModeSearch:
			return m.UpdateSearch(msg)
		case ModeFormDeck:
			return m.UpdateFormDeck(msg)
		case ModeFormCard:
			return m.UpdateFormCard(msg)
		case ModeReview:
			return m.UpdateReview(msg)
		default:
			return m.UpdateDashboard(msg)
		}
	}

	return m, nil
}
