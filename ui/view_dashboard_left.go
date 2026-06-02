package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderLeftPanel(leftW, panelH int) string {
	isActive := m.ActivePanel == PanelDecks && m.UIMode == ModeDashboard

	var borderColor lipgloss.Color
	if isActive {
		borderColor = AccentColor
	} else {
		borderColor = GrayMidColor
	}

	leftStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(leftW - 2).
		Height(panelH)

	var sb strings.Builder

	// ── Section header ─────────────────────────────────────────────────────
	var headerLabel string
	if isActive {
		headerLabel = lipgloss.NewStyle().
			Bold(true).Foreground(GrayDarkColor).Background(AccentColor).
			Padding(0, 1).Render("DECKS")
		headerLabel += " " + lipgloss.NewStyle().
			Foreground(AccentColor).Render(fmt.Sprintf("(%d)", len(m.Decks)))
	} else {
		headerLabel = lipgloss.NewStyle().
			Bold(true).Foreground(GrayLightColor).Background(GrayMidColor).
			Padding(0, 1).Render("DECKS")
		headerLabel += " " + GrayLightStyle.Render(fmt.Sprintf("(%d)", len(m.Decks)))
	}
	sb.WriteString(headerLabel + "\n")
	sb.WriteString(GrayLightStyle.Render(strings.Repeat("─", leftW-4)) + "\n")

	// ── Deck list ──────────────────────────────────────────────────────────
	innerW := leftW - 4
	if innerW < 8 {
		innerW = 8
	}

	if len(m.Decks) == 0 {
		sb.WriteString("\n" + GrayLightStyle.Render("  No decks yet.") + "\n")
		sb.WriteString(GrayLightStyle.Render("  Press 'a' to create one."))
	} else {
		visible := panelH - 3
		start := m.DeckScrollOffset
		end := start + visible
		if end > len(m.Decks) {
			end = len(m.Decks)
		}

		for idx := start; idx < end; idx++ {
			d := m.Decks[idx]
			isSel := m.SelectedDeckIdx == idx

			// Tree prefix
			prefix := ""
			if d.Depth > 0 {
				prefix = strings.Repeat("  ", d.Depth-1)
				if isLastChild(m.Decks, idx) {
					prefix += "└ "
				} else {
					prefix += "├ "
				}
			}

			// Badge text (due / total)
			var badgeRaw string
			if d.CardCount == 0 {
				badgeRaw = "empty"
			} else if d.DueCount > 0 {
				badgeRaw = fmt.Sprintf("%d/%d", d.DueCount, d.CardCount)
			} else {
				badgeRaw = fmt.Sprintf("0/%d", d.CardCount)
			}

			nameMaxW := innerW - len(prefix) - len(badgeRaw) - 3
			if nameMaxW < 4 {
				nameMaxW = 4
			}
			nameText := truncate(d.Name, nameMaxW)

			rowText := fmt.Sprintf(" %s%s", prefix, nameText)
			rowText = padRight(rowText, innerW-len(badgeRaw)-1) + " " + badgeRaw

			var rowRendered string
			if isSel {
				if isActive {
					rowRendered = lipgloss.NewStyle().
						Bold(true).
						Foreground(GrayDarkColor).
						Background(AccentColor).
						Width(innerW).
						Render(rowText)
				} else {
					rowRendered = lipgloss.NewStyle().
						Foreground(WhiteColor).
						Background(GrayMidColor).
						Width(innerW).
						Render(rowText)
				}
			} else {
				// Color-code the badge
				var coloredBadge string
				if d.CardCount == 0 {
					coloredBadge = GrayLightStyle.Render(badgeRaw)
				} else if d.DueCount > 0 {
					coloredBadge = GreenStyle.Bold(true).Render(fmt.Sprintf("%d", d.DueCount)) +
						GrayLightStyle.Render(fmt.Sprintf("/%d", d.CardCount))
				} else {
					coloredBadge = GrayLightStyle.Render(badgeRaw)
				}

				nameRendered := truncate(d.Name, nameMaxW)
				prefixRendered := GrayLightStyle.Render(prefix)
				rowRendered = " " + prefixRendered + nameRendered
				padding := innerW - lipgloss.Width(rowRendered) - lipgloss.Width(coloredBadge) - 1
				if padding > 0 {
					rowRendered += strings.Repeat(" ", padding)
				}
				rowRendered += " " + coloredBadge
			}
			sb.WriteString(rowRendered + "\n")
		}

		// Scroll indicator
		if len(m.Decks) > visible {
			info := fmt.Sprintf(" %d-%d / %d", start+1, end, len(m.Decks))
			sb.WriteString("\n" + GrayLightStyle.Render(info))
		}
	}

	return leftStyle.Render(sb.String())
}

func MutedBadgeStyle(count int) string {
	return GrayLightStyle.Render(fmt.Sprintf("[%d]", count))
}
