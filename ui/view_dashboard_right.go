package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderRightPanel(rightW, panelH int) string {
	rightStyle := PanelStyle.Width(rightW - 4).Height(panelH)
	if m.ActivePanel == PanelCards && m.UIMode == ModeDashboard {
		rightStyle = ActivePanelStyle.Width(rightW - 4).Height(panelH)
	}

	var cardsStr strings.Builder
	headerTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(WhiteColor).
		Background(AccentSecColor).
		Padding(0, 1).
		Render(" CARDS IN SELECTION ")
	if m.TagFilter != "" {
		badgeText := fmt.Sprintf(" #%s ", m.TagFilter)
		headerTitle += " " + AccentSecStyle.Bold(true).Background(GrayMidColor).Render(badgeText)
	}
	cardsStr.WriteString(headerTitle + "\n\n")

	colIdW := 5
	colDueW := 11
	colEaseW := 6
	colRepW := 5
	colTagsW := 12
	colFrontW := rightW - 4 - colIdW - colDueW - colEaseW - colRepW - colTagsW - 10
	if colFrontW < 10 {
		colFrontW = 10
	}

	headerRow := fmt.Sprintf("%s %s %s %s %s %s\n",
		padRight("ID", colIdW),
		padRight("FRONT", colFrontW),
		padRight("DUE", colDueW),
		padRight("EASE", colEaseW),
		padRight("REP", colRepW),
		padRight("TAGS", colTagsW),
	)
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
			easeText := fmt.Sprintf("%.2f", c.EaseFactor)
			repText := fmt.Sprintf("%d", c.RepetitionCount)

			var tagStrs []string
			for _, t := range c.Tags {
				if t != "" {
					tagStrs = append(tagStrs, "#"+t)
				}
			}
			tagsText := strings.Join(tagStrs, " ")

			// 1. Truncate raw strings to fit columns
			frontText = truncate(frontText, colFrontW)
			tagsText = truncate(tagsText, colTagsW)

			// 2. Pad raw strings to column widths (safe from ANSI length drift)
			paddedFront := padRight(frontText, colFrontW)
			paddedDue := padRight(dueText, colDueW)
			paddedEase := padRight(easeText, colEaseW)
			paddedRep := padRight(repText, colRepW)
			paddedTags := padRight(tagsText, colTagsW)
			paddedId := padRight(idStr, colIdW)

			// 3. Highlight query and apply column coloring
			query := m.SearchInput.Value()
			coloredFront := HighlightQuery(paddedFront, query)
			coloredTags := HighlightQuery(paddedTags, query)
			if query == "" && tagsText != "" {
				coloredTags = AccentSecStyle.Render(paddedTags)
			}

			var coloredDue string
			if dueText == "Instantly" {
				coloredDue = GreenStyle.Bold(true).Render(paddedDue)
			} else {
				coloredDue = GrayLightStyle.Render(paddedDue)
			}

			row := fmt.Sprintf("%s %s %s %s %s %s", paddedId, coloredFront, coloredDue, paddedEase, paddedRep, coloredTags)

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
	return rightStyle.Render(cardsStr.String())
}
