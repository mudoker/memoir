package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderLeftPanel(leftW, panelH int) string {
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
				badge = MutedBadgeStyle(d.DueCount)
			}

			if isSel {
				if m.ActivePanel == PanelDecks {
					nameText = CursorStyle.Render(nameText)
				} else {
					nameText = AccentStyle.Bold(true).Render(nameText)
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
	return leftStyle.Render(decksStr.String())
}

func MutedBadgeStyle(count int) string {
	return GrayLightStyle.Render(fmt.Sprintf("[%d]", count))
}
