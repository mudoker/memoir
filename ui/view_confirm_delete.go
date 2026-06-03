package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) ViewConfirmDelete() string {
	var content strings.Builder

	innerW := 50
	boxW := innerW + 10 // Padding (4x2) + Borders (1x2) = 10

	// Use plain text without emojis to prevent terminal width discrepancies
	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(RedColor).Underline(true).Render("CONFIRM DELETION") + "\n\n")

	var targetInfo string
	var warning string

	if m.ActivePanel == PanelDecks {
		if len(m.Decks) > 0 {
			d := m.Decks[m.SelectedDeckIdx]
			deckName := truncate(d.Name, innerW-10)
			targetInfo = fmt.Sprintf("Deck: %s", AccentStyle.Bold(true).Render(deckName))
			warning = "Warning: Deleting this deck will recursively delete all subdecks and nested cards permanently."
		}
	} else {
		if len(m.FilteredCards) > 0 {
			c := m.FilteredCards[m.SelectedCardIdx]
			cardFront := truncate(c.Front, innerW-16)
			targetInfo = fmt.Sprintf("Card Front: %s", AccentSecStyle.Bold(true).Render(cardFront))
			warning = "Warning: This card will be permanently removed from the deck."
		}
	}

	content.WriteString("Are you sure you want to delete the selected item?\n\n")
	content.WriteString("  " + targetInfo + "\n\n")

	// Manually wrap warning text to fit innerW exactly
	wrappedWarning := wrapText(warning, innerW)
	for _, line := range strings.Split(wrappedWarning, "\n") {
		content.WriteString(RedStyle.Render(line) + "\n")
	}
	content.WriteString("\n")

	// Use ASCII hyphens for separator to prevent ambiguous width layout bugs in CJK or customized terminals
	sep := GrayLightStyle.Render(strings.Repeat("-", innerW))
	content.WriteString(sep + "\n")
	content.WriteString(lipgloss.NewStyle().Foreground(GrayLightColor).Render("[Enter] Confirm Delete  •  [Esc] Cancel"))

	confirmBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(RedColor).
		Padding(1, 4).
		Width(boxW).
		Render(content.String())

	return AddShadow(confirmBox)
}

func wrapText(text string, limit int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	var lines []string
	currLine := words[0]
	for _, word := range words[1:] {
		if len(currLine)+1+len(word) > limit {
			lines = append(lines, currLine)
			currLine = word
		} else {
			currLine += " " + word
		}
	}
	lines = append(lines, currLine)
	return strings.Join(lines, "\n")
}
