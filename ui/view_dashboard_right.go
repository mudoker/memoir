package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderRightPanel(rightW, panelH int) string {
	isActive := m.ActivePanel == PanelCards && m.UIMode == ModeDashboard

	var borderType lipgloss.Border
	var borderColor lipgloss.Color
	if isActive {
		borderType = lipgloss.DoubleBorder()
		borderColor = AccentSecColor
	} else {
		borderType = lipgloss.RoundedBorder()
		borderColor = GrayMid2Color
	}

	rightStyle := lipgloss.NewStyle().
		Border(borderType).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(rightW - 2).
		Height(panelH)

	var sb strings.Builder

	// ── Section header ──────────────────────────────────────────────────────
	var headerLabel string
	cardCount := len(m.FilteredCards)
	if isActive {
		headerLabel = lipgloss.NewStyle().
			Bold(true).Foreground(GrayDarkColor).Background(AccentSecColor).
			Padding(0, 1).Render("CARDS")
		headerLabel += " " + lipgloss.NewStyle().
			Foreground(AccentSecColor).Render(fmt.Sprintf("(%d)", cardCount))
	} else {
		headerLabel = lipgloss.NewStyle().
			Bold(true).Foreground(GrayLightColor).Background(GrayMidColor).
			Padding(0, 1).Render("CARDS")
		headerLabel += " " + GrayLightStyle.Render(fmt.Sprintf("(%d)", cardCount))
	}

	if m.TagFilter != "" {
		tagBadge := lipgloss.NewStyle().
			Bold(true).Foreground(AccentSecColor).Background(GrayMidColor).
			Padding(0, 1).Render("#" + m.TagFilter)
		headerLabel += "  " + tagBadge
	}
	sb.WriteString(headerLabel + "\n")

	// ── Breadcrumb path ─────────────────────────────────────────────────────
	var path []string
	if len(m.Decks) > 0 && m.SelectedDeckIdx < len(m.Decks) {
		curr := m.Decks[m.SelectedDeckIdx]
		path = append(path, curr.Name)
		pid := curr.ParentID
		for pid != nil {
			found := false
			for i := range m.Decks {
				if m.Decks[i].ID == *pid {
					path = append([]string{m.Decks[i].Name}, path...)
					pid = m.Decks[i].ParentID
					found = true
					break
				}
			}
			if !found {
				break
			}
		}
	}
	if len(path) > 0 {
		pathText := strings.Join(path, " / ")
		sb.WriteString(GrayLightStyle.Render("  "+pathText) + "\n")
	}

	innerW := rightW - 4
	if innerW < 12 {
		innerW = 12
	}
	sb.WriteString(GrayLightStyle.Render(strings.Repeat("─", innerW)) + "\n")

	// ── Column layout ───────────────────────────────────────────────────────
	colIdW := 4
	colDueW := 12
	colEaseW := 5
	colRepW := 4
	colTagsW := 12
	colFrontW := innerW - colIdW - colDueW - colEaseW - colRepW - colTagsW - 10
	if colFrontW < 8 {
		colFrontW = 8
	}

	// Column headers
	colHeaders := fmt.Sprintf("%s  %s  %s  %s  %s  %s",
		lipgloss.NewStyle().Bold(true).Foreground(AccentSecColor).Render(padRight("ID", colIdW)),
		lipgloss.NewStyle().Bold(true).Foreground(AccentSecColor).Render(padRight("FRONT TEXT", colFrontW)),
		lipgloss.NewStyle().Bold(true).Foreground(AccentSecColor).Render(padRight("SCHEDULED", colDueW)),
		lipgloss.NewStyle().Bold(true).Foreground(AccentSecColor).Render(padRight("EASE", colEaseW)),
		lipgloss.NewStyle().Bold(true).Foreground(AccentSecColor).Render(padRight("REPS", colRepW)),
		lipgloss.NewStyle().Bold(true).Foreground(AccentSecColor).Render(padRight("TAGS", colTagsW)),
	)
	sb.WriteString(colHeaders + "\n")
	sb.WriteString(GrayLightStyle.Render(strings.Repeat("─", innerW)) + "\n")

	// ── Card rows ──────────────────────────────────────────────────────────
	if len(m.FilteredCards) == 0 {
		sb.WriteString("\n")
		sb.WriteString("  " + GrayLightStyle.Render("No cards in this deck.") + "\n")
		sb.WriteString("  " + GrayLightStyle.Render("Press 'a' to add one, or switch deck with j/k."))
	} else {
		visibleH := panelH - 7
		if visibleH < 1 {
			visibleH = 1
		}
		start := m.CardScrollOffset
		end := start + visibleH
		if end > len(m.FilteredCards) {
			end = len(m.FilteredCards)
		}

		for idx := start; idx < end; idx++ {
			c := m.FilteredCards[idx]
			isSel := m.SelectedCardIdx == idx

			idStr := fmt.Sprintf("%d", c.ID)
			rawDue := formatDue(c.DueAt)
			var dueText string
			if rawDue == "Instantly" {
				dueText = "⚡ Now"
			} else if rawDue == "Tomorrow" {
				dueText = "📅 Tomorrow"
			} else {
				dueText = "📅 " + strings.Replace(rawDue, " Days", "d", 1)
			}

			easeText := fmt.Sprintf("%.1f", c.EaseFactor)
			repText := fmt.Sprintf("%d", c.RepetitionCount)
			frontText := truncate(c.Front, colFrontW)

			var tagStrs []string
			for _, t := range c.Tags {
				if t != "" {
					tagStrs = append(tagStrs, "#"+t)
				}
			}
			tagsText := truncate(strings.Join(tagStrs, " "), colTagsW)

			if isSel {
				// Full-width highlight row
				rawRow := fmt.Sprintf("%s  %s  %s  %s  %s  %s",
					padRight(idStr, colIdW),
					padRight(frontText, colFrontW),
					padRight(dueText, colDueW),
					padRight(easeText, colEaseW),
					padRight(repText, colRepW),
					padRight(tagsText, colTagsW),
				)
				rawRow = truncate(rawRow, innerW)
				var selStyle lipgloss.Style
				if isActive {
					selStyle = lipgloss.NewStyle().
						Bold(true).Foreground(GrayDarkColor).Background(AccentSecColor).
						Width(innerW)
				} else {
					selStyle = lipgloss.NewStyle().
						Foreground(WhiteColor).Background(GrayMidColor).
						Width(innerW)
				}
				sb.WriteString(selStyle.Render(rawRow) + "\n")
			} else {
				// Color-coded unselected row
				query := m.SearchInput.Value()
				coloredId := GrayLightStyle.Render(padRight(idStr, colIdW))
				coloredFront := lipgloss.NewStyle().Render(HighlightQuery(padRight(frontText, colFrontW), query))

				var coloredDue string
				if rawDue == "Instantly" {
					coloredDue = GreenStyle.Bold(true).Render(padRight(dueText, colDueW))
				} else {
					coloredDue = GrayLightStyle.Render(padRight(dueText, colDueW))
				}

				var coloredEase string
				if c.EaseFactor < 1.8 {
					coloredEase = RedStyle.Render(padRight(easeText, colEaseW))
				} else if c.EaseFactor >= 2.5 {
					coloredEase = GreenStyle.Render(padRight(easeText, colEaseW))
				} else {
					coloredEase = YellowStyle.Render(padRight(easeText, colEaseW))
				}

				coloredRep := GrayLightStyle.Render(padRight(repText, colRepW))
				coloredTags := AccentSecStyle.Render(HighlightQuery(padRight(tagsText, colTagsW), query))

				row := fmt.Sprintf("%s  %s  %s  %s  %s  %s",
					coloredId, coloredFront, coloredDue, coloredEase, coloredRep, coloredTags)
				sb.WriteString(row + "\n")
			}
		}

		// Scroll position indicator
		if len(m.FilteredCards) > visibleH {
			shown := end - start
			indicator := fmt.Sprintf("  %d–%d of %d", start+1, start+shown, len(m.FilteredCards))
			sb.WriteString("\n" + GrayLightStyle.Render(indicator))
		}
	}

	return rightStyle.Render(sb.String())
}
