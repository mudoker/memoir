package srs

import (
	"math"
	"time"

	"flashtui/db"
)

// CalculateSM2 updates card metrics based on the modified SM-2 algorithm.
// q is the grade (0 to 5)
func CalculateSM2(card db.Card, q int, now time.Time) db.Card {
	updated := card

	if q < 3 {
		// Recall lapse
		updated.RepetitionCount = 0
		updated.Interval = 1
		// Penalize ease factor by 0.2
		efPenalized := card.EaseFactor - 0.2
		if efPenalized < 1.3 {
			efPenalized = 1.3
		}
		updated.EaseFactor = efPenalized
		// Due in 1 day
		updated.DueAt = now.AddDate(0, 0, 1)
	} else {
		// Recall success
		// Ease Factor calculation
		qDiff := float64(5 - q)
		efNew := card.EaseFactor + (0.1 - qDiff*(0.08+qDiff*0.02))
		if efNew < 1.3 {
			efNew = 1.3
		}
		updated.EaseFactor = efNew

		if card.RepetitionCount == 0 {
			updated.Interval = 1
		} else if card.RepetitionCount == 1 {
			updated.Interval = 6
		} else {
			// math.Ceil(I_old * EF)
			updated.Interval = int(math.Ceil(float64(card.Interval) * efNew))
		}
		updated.RepetitionCount = card.RepetitionCount + 1
		// Due in Interval days
		updated.DueAt = now.AddDate(0, 0, updated.Interval)
	}

	return updated
}
