package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) ViewDashboard() string {
	var b strings.Builder

	// 1. Header with stats badges
	totalDue := 0
	for _, d := range m.Decks {
		if d.ParentID == nil {
			totalDue += d.DueCount
		}
	}
	totalCards, mastered, _ := m.Database.GetMasteryStats()

	badgeDecks := AccentStyle.Render(fmt.Sprintf(" 📂 Decks: %d ", len(m.Decks)))
	badgeCards := AccentSecStyle.Render(fmt.Sprintf(" 🗃️ Cards: %d ", totalCards))
	badgeDue := GreenStyle.Render(fmt.Sprintf(" ⏳ Due: %d ", totalDue))

	title := TitleStyle.Render(" FlashTUI ─ v1.0.0 ") + "  " + badgeDecks + " " + badgeCards + " " + badgeDue
	headerText := fmt.Sprintf(" ╭%s╮\n", strings.Repeat("─", m.Width-2))

	w := lipgloss.Width(title)
	padding := m.Width - 6 - w
	if padding < 0 {
		padding = 0
	}
	headerMid := fmt.Sprintf(" │  %s%s │\n", title, strings.Repeat(" ", padding))
	headerText += headerMid
	headerText += fmt.Sprintf(" ╰%s╯", strings.Repeat("─", m.Width-2))
	b.WriteString(headerText + "\n")

	// 2. Dual Panels (delegated to left/right renderers)
	leftW := int(float64(m.Width) * 0.3)
	rightW := m.Width - leftW - 2
	panelH := m.Height - 11

	leftView := m.renderLeftPanel(leftW, panelH)
	rightView := m.renderRightPanel(rightW, panelH)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView) + "\n")

	// 3. Stats Panel with Weekly Activity Grid
	streak, _ := m.Database.GetDailyStreak()
	ret, _ := m.Database.GetRetentionAccuracy()

	barW := 30
	pct := 0.0
	if totalCards > 0 {
		pct = float64(mastered) / float64(totalCards)
	}
	barStr := renderProgressBar(barW, pct)

	// Fetch 7 day history
	activity, _ := m.Database.GetLast7DaysActivity()
	var actBlocks []string
	weekdays := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	for i, count := range activity {
		block := "░"
		if count > 10 {
			block = "█"
		} else if count > 4 {
			block = "▓"
		} else if count > 0 {
			block = "▒"
		}

		coloredBlock := block
		if count > 0 {
			coloredBlock = GreenStyle.Render(block)
		} else {
			coloredBlock = GrayLightStyle.Render(block)
		}
		actBlocks = append(actBlocks, fmt.Sprintf("%s:%s", weekdays[i], coloredBlock))
	}
	activityStr := strings.Join(actBlocks, " ")

	statsView := StatsStyle.Width(m.Width - 4).Render(
		fmt.Sprintf("Streak Tracker: %s %d Days | Accuracy: %.1f%% | Recent: %s\nMastered Cards: [%s] %.1f%% (%d/%d)",
			GreenStyle.Render("🔥"), streak, ret, activityStr, barStr, pct*100.0, mastered, totalCards,
		),
	)
	b.WriteString(statsView + "\n")

	// 4. Console / Help / Status bar
	if m.UIMode == ModeSearch {
		b.WriteString(m.SearchInput.View())
	} else {
		helpLine := " j/k: Navigation | h/l: Panes | a: Create Deck/Card | e: Edit | dd: Purge | /: Search | : cmd"
		if m.StatusMsg == "" {
			b.WriteString(GrayLightStyle.Render(helpLine))
		} else {
			b.WriteString(m.StatusMsg)
		}
	}

	return b.String()
}
