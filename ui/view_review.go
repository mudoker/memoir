package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) ViewReview() string {
	if m.Session == nil {
		return ""
	}

	if m.Session.ActiveCard == nil {
		var b strings.Builder
		b.WriteString("╭──────────────────────────────────────────────────────────╮\n")
		b.WriteString("│                                                          │\n")
		b.WriteString("│                  🎉 SESSION COMPLETED!                   │\n")
		b.WriteString("│                                                          │\n")
		b.WriteString("│        You have successfully reviewed all due cards.     │\n")
		b.WriteString("│                                                          │\n")
		b.WriteString("│        [Press Esc to return to the Deck Manager]        │\n")
		b.WriteString("│                                                          │\n")
		b.WriteString("╰──────────────────────────────────────────────────────────╯")
		return b.String()
	}

	card := m.Session.ActiveCard
	deck := m.Decks[m.SelectedDeckIdx]

	var b strings.Builder

	pct := 0.0
	if m.Session.TotalSessionCards > 0 {
		pct = float64(m.Session.CompletedCount) / float64(m.Session.TotalSessionCards)
	}
	barStr := renderProgressBar(40, pct)
	progressText := fmt.Sprintf("Queue Progress: [%s] %d%% (%d/%d)", barStr, int(pct*100), m.Session.CompletedCount, m.Session.TotalSessionCards)

	title := fmt.Sprintf(" Reviewing: %s ", deck.Name)
	boxW := m.Width - 10
	if boxW > 85 {
		boxW = 85
	}

	b.WriteString(fmt.Sprintf("╭─%s%s─╮\n", title, strings.Repeat("─", boxW-len(title)-4)))
	b.WriteString(fmt.Sprintf("│  %-*s  │\n", boxW-6, progressText))
	b.WriteString(fmt.Sprintf("├%s┤\n", strings.Repeat("─", boxW-2)))
	b.WriteString("│                                                                              │\n")

	contentW := boxW - 8
	frontStyle := lipgloss.NewStyle().Width(contentW).Align(lipgloss.Left)

	// Leech Warning indicator
	if m.Session.IsLeech(card.ID) {
		leechLabel := LeechStyle.Render("🔥 STICKY LEECH CARD")
		b.WriteString(fmt.Sprintf("│  %-*s  │\n", boxW-6, leechLabel))
		b.WriteString("│                                                                              │\n")
	}

	b.WriteString(fmt.Sprintf("│  %s  │\n", lipgloss.NewStyle().Bold(true).Render("Question:")))
	wrappedFront := frontStyle.Render(card.Front)
	for _, line := range strings.Split(wrappedFront, "\n") {
		b.WriteString(fmt.Sprintf("│    %-*s  │\n", boxW-8, line))
	}
	b.WriteString("│                                                                              │\n")

	if card.Hint != "" {
		if m.Session.ShowHint {
			b.WriteString(fmt.Sprintf("│  %s  │\n", YellowStyle.Render("Hint: "+card.Hint)))
		} else {
			b.WriteString(fmt.Sprintf("│  %s  │\n", GrayLightStyle.Render("[Hint Available: Press 'h' to peek]")))
		}
		b.WriteString("│                                                                              │\n")
	}

	if m.Session.IsFlipped {
		b.WriteString(fmt.Sprintf("├%s┤\n", strings.Repeat("─", boxW-2)))
		b.WriteString("│                                                                              │\n")
		b.WriteString(fmt.Sprintf("│  %s  │\n", lipgloss.NewStyle().Bold(true).Render("Answer:")))
		wrappedBack := frontStyle.Render(card.Back)
		for _, line := range strings.Split(wrappedBack, "\n") {
			b.WriteString(fmt.Sprintf("│    %-*s  │\n", boxW-8, line))
		}
		b.WriteString("│                                                                              │\n")
	}

	b.WriteString(fmt.Sprintf("╰%s╯\n", strings.Repeat("─", boxW-2)))

	var footer string
	if m.Session.IsFlipped {
		footer = " [1-5]: Rate performance (1: Forgot, 2: Hard, 3: Good, 4: Easy, 5: Perfect) | u: Undo | Esc: Exit"
	} else {
		footer = " [Space]: Flip Card Back | h: Reveal Hint | s: Shuffle Queue | Esc: Exit"
	}
	b.WriteString(GrayLightStyle.Render(footer))

	return b.String()
}
