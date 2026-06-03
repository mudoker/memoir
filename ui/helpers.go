package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"flashtui/db"
)

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

func renderStyledProgressBar(width int, ratio float64) string {
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
	filledStr := GreenStyle.Render(strings.Repeat("█", filledLen))
	emptyStr := GrayLightStyle.Render(strings.Repeat("░", emptyLen))
	return filledStr + emptyStr
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

func padRight(s string, width int) string {
	w := runewidth.StringWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func truncate(s string, width int) string {
	w := runewidth.StringWidth(s)
	if w <= width {
		return s
	}
	return runewidth.Truncate(s, width-1, "…")
}

func HighlightQuery(s string, query string) string {
	if query == "" {
		return s
	}
	lowerS := strings.ToLower(s)
	lowerQuery := strings.ToLower(query)

	idx := strings.Index(lowerS, lowerQuery)
	if idx == -1 {
		return s
	}

	var result strings.Builder
	lastIdx := 0
	for idx != -1 {
		result.WriteString(s[lastIdx:idx])
		match := s[idx : idx+len(query)]
		result.WriteString(YellowStyle.Render(match))
		lastIdx = idx + len(query)

		nextIdx := strings.Index(lowerS[lastIdx:], lowerQuery)
		if nextIdx == -1 {
			break
		}
		idx = lastIdx + nextIdx
	}
	result.WriteString(s[lastIdx:])
	return result.String()
}

func AddShadow(box string) string {
	box = strings.TrimSuffix(box, "\n")
	lines := strings.Split(box, "\n")
	if len(lines) == 0 {
		return box
	}
	var result []string
	for i, line := range lines {
		if i == 0 {
			result = append(result, line)
		} else {
			result = append(result, line+"░")
		}
	}
	width := lipgloss.Width(lines[0])
	if width > 1 {
		result = append(result, " "+strings.Repeat("░", width-1))
	}
	return strings.Join(result, "\n")
}
