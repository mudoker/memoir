package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) ViewDashboard() string {
	var b strings.Builder

	// 1. Header
	title := TitleStyle.Render(fmt.Sprintf("FlashTUI ─ v1.0.0 (Local Time: %s)", time.Now().Format("15:04:05")))
	headerText := fmt.Sprintf(" ╭%s╮\n", strings.Repeat("─", m.Width-2))
	headerMid := fmt.Sprintf(" │  %-*s │\n", m.Width-6, title)
	headerText += headerMid
	headerText += fmt.Sprintf(" ╰%s╯", strings.Repeat("─", m.Width-2))
	b.WriteString(headerText + "\n")

	// 2. Dual Panels
	leftW := int(float64(m.Width) * 0.3)
	rightW := m.Width - leftW - 2
	panelH := m.Height - 11

	leftStyle := PanelStyle.Width(leftW - 4).Height(panelH)
	if m.ActivePanel == PanelDecks && m.UIMode == ModeDashboard {
		leftStyle = ActivePanelStyle.Width(leftW - 4).Height(panelH)
	}

	var decksStr strings.Builder
	decksStr.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).Render("DECKS MANAGER (Normal Mode)") + "\n\n")

	if len(m.Decks) == 0 {
		decksStr.WriteString(" (No decks created)\n Press 'a' to create.")
	} else {
		visibleHeight := panelH - 2
		start := m.DeckScrollOffset
		end := start + visibleHeight
		if end > len(m.Decks) {
			end = len(m.Decks)
		}

		for idx := start; idx < end; idx++ {
			d := m.Decks[idx]
			isSel := m.SelectedDeckIdx == idx
			nameText := d.Name
			badge := fmt.Sprintf("[%d]", d.DueCount)
			if d.DueCount > 0 {
				badge = GreenStyle.Render(badge)
			} else {
				badge = GrayLightStyle.Render(badge)
			}

			if isSel {
				if m.ActivePanel == PanelDecks {
					nameText = CursorStyle.Render(nameText)
				} else {
					nameText = lipgloss.NewStyle().Foreground(AccentColor).Bold(true).Render(nameText)
				}
			}

			prefix := ""
			if d.Depth > 0 {
				prefix = strings.Repeat("│   ", d.Depth-1)
				if isLastChild(m.Decks, idx) {
					prefix += "└── "
				} else {
					prefix += "├── "
				}
			}

			if isSel {
				decksStr.WriteString(fmt.Sprintf(" ▶  %s%s %s\n", prefix, nameText, badge))
			} else {
				decksStr.WriteString(fmt.Sprintf("    %s%s %s\n", prefix, nameText, badge))
			}
		}
	}
	leftView := leftStyle.Render(decksStr.String())

	rightStyle := PanelStyle.Width(rightW - 4).Height(panelH)
	if m.ActivePanel == PanelCards && m.UIMode == ModeDashboard {
		rightStyle = ActivePanelStyle.Width(rightW - 4).Height(panelH)
	}

	var cardsStr strings.Builder
	cardsStr.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).Render("CARDS IN SELECTION") + "\n\n")

	colIdW := 6
	colDueW := 12
	colTagsW := 15
	colFrontW := rightW - 4 - colIdW - colDueW - colTagsW - 8
	if colFrontW < 10 {
		colFrontW = 10
	}

	headerRow := fmt.Sprintf("%-*s %-*s %-*s %-*s\n", colIdW, "ID", colFrontW, "FRONT", colDueW, "DUE", colTagsW, "TAGS")
	cardsStr.WriteString(lipgloss.NewStyle().Bold(true).Foreground(WhiteColor).Render(headerRow))

	if len(m.FilteredCards) == 0 {
		cardsStr.WriteString("\n (No cards matching filter/selection)\n Press 'a' to add a card.")
	} else {
		visibleHeight := panelH - 3
		start := m.CardScrollOffset
		end := start + visibleHeight
		if end > len(m.FilteredCards) {
			end = len(m.FilteredCards)
		}

		for idx := start; idx < end; idx++ {
			c := m.FilteredCards[idx]
			isSel := m.SelectedCardIdx == idx
			idStr := fmt.Sprintf("%03d", c.ID)
			frontText := c.Front
			dueText := formatDue(c.DueAt)

			var tagStrs []string
			for _, t := range c.Tags {
				if t != "" {
					tagStrs = append(tagStrs, "#"+t)
				}
			}
			tagsText := strings.Join(tagStrs, " ")

			if len(frontText) > colFrontW {
				frontText = frontText[:colFrontW-1] + "…"
			}
			if len(tagsText) > colTagsW {
				tagsText = tagsText[:colTagsW-1] + "…"
			}

			row := fmt.Sprintf("%-*s %-*s %-*s %-*s", colIdW, idStr, colFrontW, frontText, colDueW, dueText, colTagsW, tagsText)

			if isSel {
				if m.ActivePanel == PanelCards {
					cardsStr.WriteString(CursorStyle.Render(row) + "\n")
				} else {
					cardsStr.WriteString(AccentStyle.Render(row) + "\n")
				}
			} else {
				cardsStr.WriteString(row + "\n")
			}
		}
	}
	rightView := rightStyle.Render(cardsStr.String())

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView) + "\n")

	streak, _ := m.Database.GetDailyStreak()
	ret, _ := m.Database.GetRetentionAccuracy()
	totalCards, mastered, _ := m.Database.GetMasteryStats()

	barW := 40
	pct := 0.0
	if totalCards > 0 {
		pct = float64(mastered) / float64(totalCards)
	}
	barStr := renderProgressBar(barW, pct)

	statsView := StatsStyle.Width(m.Width - 4).Render(
		fmt.Sprintf("Current Daily Streak: %s %d Days | Retention Accuracy Score: %.1f%%\nMastered Cards:       [%s] %.1f%% (%d/%d)",
			GreenStyle.Render("🔥"), streak, ret, barStr, pct*100.0, mastered, totalCards,
		),
	)
	b.WriteString(statsView + "\n")

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

func renderProgressBar(width int, ratio float64) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filledLen := int(math.Round(float64(width) * ratio))
	emptyLen := width - filledLen
	if filledLen < 0 {
		filledLen = 0
	}
	if emptyLen < 0 {
		emptyLen = 0
	}
	return strings.Repeat("█", filledLen) + strings.Repeat("░", emptyLen)
}

func formatDue(dueAt time.Time) string {
	now := time.Now()
	if dueAt.Before(now) {
		return "Instantly"
	}
	diff := dueAt.Sub(now)
	days := int(math.Ceil(diff.Hours() / 24.0))
	if days <= 1 {
		return "Tomorrow"
	}
	return fmt.Sprintf("%d Days", days)
}

func isLastChild(decks []db.Deck, idx int) bool {
	depth := decks[idx].Depth
	for i := idx + 1; i < len(decks); i++ {
		if decks[i].Depth == depth {
			return false
		}
		if decks[i].Depth < depth {
			return true
		}
	}
	return true
}
