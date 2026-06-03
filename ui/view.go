package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Width < 100 || m.Height < 30 {
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(RedColor).Bold(true).Render(
				fmt.Sprintf("Terminal Size Too Small!\n\nRequires: 100x30 minimum\nCurrent:  %dx%d\n\nPlease enlarge your terminal window.", m.Width, m.Height),
			),
		)
	}

	switch m.UIMode {
	case ModeFormDeck:
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.ViewFormDeck())
	case ModeFormCard:
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.ViewFormCard())
	case ModeReview:
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.ViewReview())
	case ModeHelp:
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.ViewHelpPanel())
	case ModeAdvice:
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.ViewAdvicePanel())
	case ModeFormKey:
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.ViewFormKey())
	case ModeConfirmDelete:
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.ViewConfirmDelete())
	default:
		return m.ViewDashboard()
	}
}
