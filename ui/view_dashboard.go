package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) ViewDashboard() string {
	var b strings.Builder

	// ── 1. HEADER ─────────────────────────────────────────────────────────────
	totalDue := 0
	for _, d := range m.Decks {
		if d.ParentID == nil {
			totalDue += d.DueCount
		}
	}
	totalCards, mastered, _ := m.Database.GetMasteryStats()

	// Logo-style brand + version
	brand := lipgloss.NewStyle().
		Bold(true).
		Foreground(WhiteColor).
		Background(AccentColor).
		Padding(0, 2).
		Render("  FlashTUI")
	version := lipgloss.NewStyle().
		Foreground(AccentColor).
		Background(GrayDarkColor).
		Padding(0, 1).
		Render("v1.0.0")

	// Stat badges
	deckBadge := lipgloss.NewStyle().
		Bold(true).Foreground(AccentColor).
		Render(fmt.Sprintf("  %d Decks", len(m.Decks)))
	cardBadge := lipgloss.NewStyle().
		Bold(true).Foreground(AccentSecColor).
		Render(fmt.Sprintf("  %d Cards", totalCards))

	var dueBadge string
	if totalDue > 0 {
		dueBadge = lipgloss.NewStyle().
			Bold(true).Foreground(GreenColor).
			Render(fmt.Sprintf("  %d Due", totalDue))
	} else {
		dueBadge = lipgloss.NewStyle().
			Foreground(GrayLightColor).
			Render("  0 Due")
	}

	headerLeft := lipgloss.JoinHorizontal(lipgloss.Center, brand, " ", version)
	headerRight := deckBadge + "    " + cardBadge + "    " + dueBadge

	spacerW := m.Width - 4 - lipgloss.Width(headerLeft) - lipgloss.Width(headerRight)
	if spacerW < 1 {
		spacerW = 1
	}
	spacer := strings.Repeat(" ", spacerW)

	headerContent := headerLeft + spacer + headerRight
	headerBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(AccentColor).
		Padding(0, 2).
		Render(headerContent)
	b.WriteString(headerBox + "\n")

	// ── 2. PANELS ─────────────────────────────────────────────────────────────
	leftW := int(float64(m.Width) * 0.28)
	rightW := m.Width - leftW - 2
	panelH := m.Height - 11

	leftView := m.renderLeftPanel(leftW, panelH)
	rightView := m.renderRightPanel(rightW, panelH)
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView) + "\n")

	// ── 3. STATS PANEL ────────────────────────────────────────────────────────
	streak, _ := m.Database.GetDailyStreak()
	ret, _ := m.Database.GetRetentionAccuracy()

	// Mastery bar
	pct := 0.0
	if totalCards > 0 {
		pct = float64(mastered) / float64(totalCards)
	}
	barStr := renderCompactBar(26, pct)

	// Weekly activity
	activity, _ := m.Database.GetLast7DaysActivity()
	days := []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	var actParts []string
	for i, count := range activity {
		var block string
		if count > 10 {
			block = GreenStyle.Bold(true).Render("█")
		} else if count > 4 {
			block = GreenStyle.Render("▓")
		} else if count > 0 {
			block = AccentStyle.Render("▒")
		} else {
			block = GrayLightStyle.Render("░")
		}
		actParts = append(actParts, GrayLightStyle.Render(days[i])+block)
	}
	weekGrid := strings.Join(actParts, " ")

	// Build stat cells
	streakCell := fmt.Sprintf("%s  %s",
		GrayLightStyle.Render("Streak"),
		AccentStyle.Bold(true).Render(fmt.Sprintf("%d days", streak)),
	)
	accuracyCell := fmt.Sprintf("%s  %s",
		GrayLightStyle.Render("Accuracy"),
		GreenStyle.Bold(true).Render(fmt.Sprintf("%.1f%%", ret)),
	)
	masteryCell := fmt.Sprintf("%s  %s  %s",
		GrayLightStyle.Render("Mastery"),
		barStr,
		AccentSecStyle.Bold(true).Render(fmt.Sprintf("%.0f%%", pct*100)),
	)
	weekCell := fmt.Sprintf("%s  %s",
		GrayLightStyle.Render("Week"),
		weekGrid,
	)

	sep := GrayLightStyle.Render("  │  ")
	statsLine := streakCell + sep + accuracyCell + sep + masteryCell + sep + weekCell

	statsBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(AccentSecColor).
		Padding(0, 2).
		Render(statsLine)
	b.WriteString(statsBox + "\n")

	// ── 4. STATUS / COMMAND BAR ───────────────────────────────────────────────
	b.WriteString(m.renderStatusBar())

	return b.String()
}

func (m Model) renderStatusBar() string {
	var bar strings.Builder

	switch m.UIMode {
	case ModeSearch:
		modeBadge := lipgloss.NewStyle().
			Bold(true).Foreground(GrayDarkColor).Background(AccentSecColor).
			Padding(0, 1).Render("SEARCH")
		matchInfo := GrayLightStyle.Render(fmt.Sprintf(" %d matches", len(m.FilteredCards)))
		bar.WriteString(modeBadge + "  " + m.SearchInput.View() + matchInfo)

	case ModeConsole:
		modeBadge := lipgloss.NewStyle().
			Bold(true).Foreground(GrayDarkColor).Background(YellowColor).
			Padding(0, 1).Render("COMMAND")
		bar.WriteString(modeBadge + "  " + m.ConsoleInput.View())

	default:
		var panelLabel string
		if m.ActivePanel == PanelDecks {
			panelLabel = lipgloss.NewStyle().
				Bold(true).Foreground(GrayDarkColor).Background(AccentColor).
				Padding(0, 1).Render("DECKS")
		} else {
			panelLabel = lipgloss.NewStyle().
				Bold(true).Foreground(GrayDarkColor).Background(AccentSecColor).
				Padding(0, 1).Render("CARDS")
		}
		bar.WriteString(panelLabel + "  ")

		if m.StatusMsg != "" {
			bar.WriteString(m.StatusMsg)
		} else {
			hints := []string{"j/k nav", "h/l·Tab panes", "a add", "e edit", "dd del", "/ search", ": cmd", "? help"}
			var hintParts []string
			for _, h := range hints {
				hintParts = append(hintParts, GrayLightStyle.Render(h))
			}
			bar.WriteString(strings.Join(hintParts, GrayLightStyle.Render("  ·  ")))
		}
	}

	return bar.String()
}

// renderCompactBar renders a slim, block-character progress bar.
func renderCompactBar(width int, pct float64) string {
	filled := int(float64(width) * pct)
	if filled > width {
		filled = width
	}
	bar := GreenStyle.Render(strings.Repeat("█", filled)) +
		GrayLightStyle.Render(strings.Repeat("░", width-filled))
	return "[" + bar + "]"
}
