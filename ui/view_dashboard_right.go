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
	cardsStr.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).Render("CARDS IN SELECTION") + "\n\n")

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

			// Visual width truncation
			frontText = truncate(frontText, colFrontW)
			tagsText = truncate(tagsText, colTagsW)

			// Search Query highlights
			query := m.SearchInput.Value()
			frontText = HighlightQuery(frontText, query)
			tagsText = HighlightQuery(tagsText, query)

			// Visual width right-padding
			frontText = padRight(frontText, colFrontW)
			dueText = padRight(dueText, colDueW)
			easeText = padRight(easeText, colEaseW)
			repText = padRight(repText, colRepW)
			tagsText = padRight(tagsText, colTagsW)
			paddedId := padRight(idStr, colIdW)

			row := fmt.Sprintf("%s %s %s %s %s %s", paddedId, frontText, dueText, easeText, repText, tagsText)

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
