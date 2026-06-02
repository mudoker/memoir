package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) ViewFormDeck() string {
	title := "CREATE NEW DECK"
	if m.FormEditID != 0 {
		title = "RENAME DECK"
	}

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(WhiteColor).Underline(true).Render(title) + "\n\n")
	content.WriteString(m.FormDeckName.View() + "\n\n")
	content.WriteString(lipgloss.NewStyle().Foreground(GrayLightColor).Render("[Enter] Confirm  |  [Esc] Cancel"))

	formBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor).
		Padding(1, 4).
		Width(50).
		Render(content.String())

	return formBox
}

func (m Model) ViewFormCard() string {
	title := "CREATE NEW CARD"
	if m.FormEditID != 0 {
		title = "EDIT CARD"
	}

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(WhiteColor).Underline(true).Render(title) + "\n\n")

	fields := []string{"Front", "Back", "Hint", "Tags"}
	for i, name := range fields {
		var activeIndicator string
		var viewStr string
		switch i {
		case 0:
			viewStr = m.FormCardFront.View()
			activeIndicator = getActiveIndicator(m.FormActiveField == 0)
		case 1:
			viewStr = m.FormCardBack.View()
			activeIndicator = getActiveIndicator(m.FormActiveField == 1)
		case 2:
			viewStr = m.FormCardHint.View()
			activeIndicator = getActiveIndicator(m.FormActiveField == 2)
		case 3:
			viewStr = m.FormCardTags.View()
			activeIndicator = getActiveIndicator(m.FormActiveField == 3)
		}

		label := fmt.Sprintf("%s %-6s: ", activeIndicator, name)
		content.WriteString(label + viewStr + "\n\n")
	}

	content.WriteString(lipgloss.NewStyle().Foreground(GrayLightColor).Render("[Tab] Cycle Fields  |  [Enter] Save  |  [Esc] Cancel"))

	formBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor).
		Padding(1, 4).
		Width(72).
		Render(content.String())

	return formBox
}

func getActiveIndicator(isActive bool) string {
	if isActive {
		return "❯"
	}
	return " "
}
