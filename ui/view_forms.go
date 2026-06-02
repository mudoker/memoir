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

	var b strings.Builder
	b.WriteString(fmt.Sprintf("╭%s╮\n", strings.Repeat("─", 50)))
	b.WriteString(fmt.Sprintf("│ %-*s │\n", 48, lipgloss.NewStyle().Bold(true).Foreground(WhiteColor).Render(title)))
	b.WriteString(fmt.Sprintf("├%s┤\n", strings.Repeat("─", 50)))
	b.WriteString("│                                                  │\n")
	b.WriteString(fmt.Sprintf("│  %s │\n", m.FormDeckName.View()))
	b.WriteString("│                                                  │\n")
	b.WriteString(fmt.Sprintf("├%s┤\n", strings.Repeat("─", 50)))
	b.WriteString("│  [Enter] Confirm  |  [Esc] Cancel                │\n")
	b.WriteString(fmt.Sprintf("╰%s╯", strings.Repeat("─", 50)))

	return b.String()
}

func (m Model) ViewFormCard() string {
	title := "CREATE NEW CARD"
	if m.FormEditID != 0 {
		title = "EDIT CARD"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("╭%s╮\n", strings.Repeat("─", 70)))
	b.WriteString(fmt.Sprintf("│ %-*s │\n", 68, lipgloss.NewStyle().Bold(true).Foreground(WhiteColor).Render(title)))
	b.WriteString(fmt.Sprintf("├%s┤\n", strings.Repeat("─", 70)))
	b.WriteString("│                                                                    │\n")

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

		b.WriteString(fmt.Sprintf("│ %s %-6s : %-56s │\n", activeIndicator, name, viewStr))
		b.WriteString("│                                                                    │\n")
	}

	b.WriteString(fmt.Sprintf("├%s┤\n", strings.Repeat("─", 70)))
	b.WriteString("│  [Tab] Cycle Fields  |  [Enter] Save  |  [Esc] Cancel              │\n")
	b.WriteString(fmt.Sprintf("╰%s╯", strings.Repeat("─", 70)))

	return b.String()
}

func getActiveIndicator(isActive bool) string {
	if isActive {
		return "▶"
	}
	return " "
}
