package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) ViewAdvicePanel() string {
	w := m.Width - 16
	if w < 50 {
		w = 50
	}
	if w > 90 {
		w = 90
	}

	sep := GrayLightStyle.Render(strings.Repeat("─", w-4))

	var b strings.Builder

	// Title
	b.WriteString(lipgloss.NewStyle().
		Bold(true).Foreground(AccentSecColor).
		Render(" 🧠 Gemini Study Coach — Actionable Advice") + "\n")
	b.WriteString(sep + "\n\n")

	// Render the raw advice text.
	cleanedAdvice := m.GeminiAdvice
	lines := strings.Split(cleanedAdvice, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "- ") {
			bulletContent := strings.TrimPrefix(strings.TrimPrefix(trimmed, "* "), "- ")
			b.WriteString("  • " + HighlightBoldText(bulletContent) + "\n")
		} else if strings.HasPrefix(trimmed, "### ") {
			headerContent := strings.TrimPrefix(trimmed, "### ")
			b.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(WhiteColor).Underline(true).Render(headerContent) + "\n")
		} else if strings.HasPrefix(trimmed, "## ") {
			headerContent := strings.TrimPrefix(trimmed, "## ")
			b.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(WhiteColor).Underline(true).Render(headerContent) + "\n")
		} else {
			b.WriteString(HighlightBoldText(trimmed) + "\n")
		}
	}

	b.WriteString("\n" + sep + "\n")

	// Now split the lines and slice according to the height
	finalLines := strings.Split(b.String(), "\n")
	totalLines := len(finalLines)

	// Available height inside the advice box
	innerMaxHeight := m.Height - 8
	if innerMaxHeight < 10 {
		innerMaxHeight = 10
	}

	maxOffset := totalLines - innerMaxHeight
	if maxOffset < 0 {
		maxOffset = 0
	}

	if m.AdviceScrollOffset > maxOffset {
		m.AdviceScrollOffset = maxOffset
	}
	if m.AdviceScrollOffset < 0 {
		m.AdviceScrollOffset = 0
	}

	endIdx := m.AdviceScrollOffset + innerMaxHeight
	if endIdx > totalLines {
		endIdx = totalLines
	}

	// Slice visible lines
	visibleLines := finalLines[m.AdviceScrollOffset:endIdx]

	// Create scroll status footer
	var scrollFooter string
	if totalLines > innerMaxHeight {
		scrollPct := int(float64(endIdx) / float64(totalLines) * 100)
		scrollFooter = fmt.Sprintf("Scroll: %d-%d / %d (%d%%) · j/k scroll · esc/q close", m.AdviceScrollOffset+1, endIdx, totalLines, scrollPct)
	} else {
		scrollFooter = "esc/q to close advice"
	}

	footerLine := lipgloss.NewStyle().Foreground(AccentSecColor).Render(scrollFooter)

	// Reassemble display content
	displayContent := strings.Join(visibleLines, "\n") + "\n" + sep + "\n" + footerLine

	adviceBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(AccentSecColor).
		Padding(1, 3).
		Width(w).
		Render(displayContent)

	return AddShadow(adviceBox)
}

// HighlightBoldText highlights text inside **...** in AccentColor
func HighlightBoldText(s string) string {
	parts := strings.Split(s, "**")
	if len(parts) < 3 {
		return s
	}
	var res strings.Builder
	for i, p := range parts {
		if i%2 == 1 {
			res.WriteString(lipgloss.NewStyle().Bold(true).Foreground(AccentColor).Render(p))
		} else {
			res.WriteString(p)
		}
	}
	return res.String()
}
