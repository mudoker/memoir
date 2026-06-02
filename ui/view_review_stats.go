package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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

	b.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).Render("Rating Distribution:") + "\n")
	ratingCounts := make(map[int]int)
	for _, r := range m.Session.Ratings {
		ratingCounts[r]++
	}
	ratingLabels := []string{"Forgot", "Hard  ", "Good  ", "Easy  ", "Perf  "}
	for r := 1; r <= 5; r++ {
		count := ratingCounts[r]
		bar := ""
		if count > 0 {
			bar = strings.Repeat("█", count)
		}
		var barColorStyle lipgloss.Style
		switch r {
		case 1:
			barColorStyle = RedStyle
		case 2:
			barColorStyle = YellowStyle
		case 3:
			barColorStyle = AccentSecStyle
		default:
			barColorStyle = GreenStyle
		}
		b.WriteString(fmt.Sprintf("  %d (%s): %s (%d)\n", r, ratingLabels[r-1], barColorStyle.Render(bar), count))
	}
	b.WriteString("\n")

	b.WriteString(GrayLightStyle.Render("[Press Esc to return to the Deck Manager]"))

	statsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(GreenColor).
		Padding(2, 6).
		Width(54).
		Align(lipgloss.Left).
		Render(b.String())

	return AddShadow(statsBox)
}
