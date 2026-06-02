package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

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
