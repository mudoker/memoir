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
		return m.viewSessionStatistics()
	}

	card := m.Session.ActiveCard
	deck := m.Decks[m.SelectedDeckIdx]
	var content strings.Builder

	// Title Header
	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(WhiteColor).Underline(true).Render(fmt.Sprintf("Reviewing Deck: %s", deck.Name)) + "\n\n")

	// Header progress bar
	pct := 0.0
	if m.Session.TotalSessionCards > 0 {
		pct = float64(m.Session.CompletedCount) / float64(m.Session.TotalSessionCards)
	}
	barStr := renderProgressBar(35, pct)
	progressText := fmt.Sprintf("Queue Progress: [%s] %d%% (%d/%d)", barStr, int(pct*100), m.Session.CompletedCount, m.Session.TotalSessionCards)
	content.WriteString(AccentSecStyle.Render(progressText) + "\n\n")

	// Sticky card indicator
	if m.Session.IsLeech(card.ID) {
		content.WriteString(LeechStyle.Render("🔥 STICKY LEECH CARD (Review Lapses: 3+)") + "\n\n")
	}

	// Question Section
	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(WhiteColor).Render("Question:") + "\n")
	wrappedFront := lipgloss.NewStyle().Width(m.Width - 16).Render(card.Front)
	content.WriteString(wrappedFront + "\n\n")

	// Hint Section
	if card.Hint != "" {
		if m.Session.ShowHint {
			content.WriteString(YellowStyle.Render("Hint: "+card.Hint) + "\n\n")
		} else {
			content.WriteString(GrayLightStyle.Render("[Hint Available: Press 'h' to peek]") + "\n\n")
		}
	}

	// Context Divider & Answer Section
	if m.Session.IsFlipped {
		content.WriteString(lipgloss.NewStyle().Foreground(GrayMidColor).Render(strings.Repeat("─", m.Width-16)) + "\n\n")
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(WhiteColor).Render("Answer:") + "\n")
		wrappedBack := lipgloss.NewStyle().Width(m.Width - 16).Render(card.Back)
		content.WriteString(wrappedBack + "\n\n")
	}

	// Build Footer
	var footer string
	if m.Session.IsFlipped {
		footer = "[1-5]: Rate performance (1: Forgot, 2: Hard, 3: Good, 4: Easy, 5: Perfect) | u: Undo | Esc: Exit"
	} else {
		footer = "[Space]: Flip Card Back | h: Reveal Hint | s: Shuffle Queue | Esc: Exit"
	}
	content.WriteString(GrayLightStyle.Render(footer))

	reviewBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor).
		Padding(1, 4).
		Width(m.Width - 8).
		Render(content.String())

	return reviewBox
}

func (m Model) viewSessionStatistics() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(GreenColor).Underline(true).Render("🎉 STUDY SESSION COMPLETED!") + "\n\n")

	// Compute session stats
	totalReviews := len(m.Session.Ratings)
	successReviews := 0
	sumRatings := 0
	for _, r := range m.Session.Ratings {
		sumRatings += r
		if r >= 3 {
			successReviews++
		}
	}

	accuracy := 0.0
	avgScore := 0.0
	if totalReviews > 0 {
		accuracy = (float64(successReviews) / float64(totalReviews)) * 100
		avgScore = float64(sumRatings) / float64(totalReviews)
	}

	uniqueLapses := 0
	for _, count := range m.Session.Lapses {
		if count > 0 {
			uniqueLapses++
		}
	}

	streak, _ := m.Database.GetDailyStreak()

	b.WriteString(fmt.Sprintf("Graduated Cards  : %s\n", GreenStyle.Render(fmt.Sprintf("%d/%d", m.Session.CompletedCount, m.Session.TotalSessionCards))))
	b.WriteString(fmt.Sprintf("Total Reviews    : %d attempts\n", totalReviews))
	b.WriteString(fmt.Sprintf("Session Accuracy : %.1f%%\n", accuracy))
	b.WriteString(fmt.Sprintf("Average Rating   : %.2f / 5.0\n", avgScore))
	b.WriteString(fmt.Sprintf("Lapsed Cards     : %d\n", uniqueLapses))
	b.WriteString(fmt.Sprintf("Current Streak   : 🔥 %d Days\n\n", streak))

	b.WriteString(GrayLightStyle.Render("[Press Esc to return to the Deck Manager]"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(GreenColor).
		Padding(2, 6).
		Width(54).
		Align(lipgloss.Left).
		Render(b.String())
}
