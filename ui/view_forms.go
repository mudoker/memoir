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
	
	label := AccentStyle.Bold(true).Render("❯ ") + lipgloss.NewStyle().Bold(true).Foreground(WhiteColor).Render("Name: ")
	content.WriteString(label + m.FormDeckName.View() + "\n\n")
	content.WriteString(lipgloss.NewStyle().Foreground(GrayLightColor).Render("[Enter] Confirm  |  [Esc] Cancel"))

	formBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(AccentColor).
		Padding(1, 4).
		Width(50).
		Render(content.String())

	return AddShadow(formBox)
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
		isActive := m.FormActiveField == i
		var activeIndicator string
		if isActive {
			activeIndicator = AccentStyle.Bold(true).Render("❯ ")
		} else {
			activeIndicator = "  "
		}

		var viewStr string
		switch i {
		case 0:
			viewStr = m.FormCardFront.View()
		case 1:
			viewStr = m.FormCardBack.View()
		case 2:
			viewStr = m.FormCardHint.View()
		case 3:
			viewStr = m.FormCardTags.View()
		}

		var label string
		if isActive {
			label = fmt.Sprintf("%s%s: ", activeIndicator, lipgloss.NewStyle().Bold(true).Foreground(WhiteColor).Render(fmt.Sprintf("%-6s", name)))
		} else {
			label = fmt.Sprintf("%s%s: ", activeIndicator, GrayLightStyle.Render(fmt.Sprintf("%-6s", name)))
		}
		content.WriteString(label + viewStr + "\n\n")
	}

	content.WriteString(lipgloss.NewStyle().Foreground(GrayLightColor).Render("[Tab] Cycle Fields  •  [Enter] Save  •  [Esc] Cancel"))

	formBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(AccentColor).
		Padding(1, 4).
		Width(72).
		Render(content.String())

	return AddShadow(formBox)
}
