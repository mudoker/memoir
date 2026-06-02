package main

import (
	"database/sql"
	"math"
	"math/rand"
	"time"
)

// CalculateSM2 updates card metrics based on the modified SM-2 algorithm.
// q is the grade (0 to 5)
func CalculateSM2(card Card, q int, now time.Time) Card {
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
		// EF_new = EF_old + (0.1 - (5 - q) * (0.08 + (5 - q) * 0.02))
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

type SessionState struct {
	DueQueue                  []Card
	LearningQueue             []Card
	GraduatedPool             []Card
	ActiveCard                *Card
	CardsReviewedSinceLastLrn int
	CompletedCount            int
	IsFlipped                 bool
	ShowHint                  bool
	WasDbWrite                bool
	DbCardID                  int64
	DbPreRepetitionCount      int
	DbPreInterval             int
	DbPreEaseFactor           float64
	DbPreDueAt                time.Time
	DbReviewLogID             int64
}

type Session struct {
	AllCards                  []Card
	DueQueue                  []Card
	LearningQueue             []Card
	GraduatedPool             []Card
	UndoStack                 []SessionState
	ActiveCard                *Card
	IsFlipped                 bool
	ShowHint                  bool
	CardsReviewedSinceLastLrn int
	TotalSessionCards         int
	CompletedCount            int
}

func NewSession(cards []Card) *Session {
	s := &Session{
		AllCards:                  cards,
		DueQueue:                  make([]Card, len(cards)),
		LearningQueue:             []Card{},
		GraduatedPool:             []Card{},
		UndoStack:                 []SessionState{},
		CardsReviewedSinceLastLrn: 0,
		TotalSessionCards:         len(cards),
		CompletedCount:            0,
	}
	copy(s.DueQueue, cards)

	s.NextCard()
	return s
}

func (s *Session) ShuffleQueue() {
	if len(s.DueQueue) > 0 {
		// Seed is auto-seeded in modern Go, but we can shuffle directly
		rand.Shuffle(len(s.DueQueue), func(i, j int) {
			s.DueQueue[i], s.DueQueue[j] = s.DueQueue[j], s.DueQueue[i]
		})
	}
	// Also reset active card to the newly shuffled first item if active card was popped from due queue
	// For simplicity, we just shuffle the due queue. NextCard calls will draw from it.
}

func (s *Session) NextCard() {
	if len(s.DueQueue) == 0 && len(s.LearningQueue) == 0 {
		s.ActiveCard = nil
		s.IsFlipped = false
		s.ShowHint = false
		return
	}

	s.IsFlipped = false
	s.ShowHint = false

	if len(s.DueQueue) == 0 {
		// Only learning queue has cards
		s.ActiveCard = &s.LearningQueue[0]
		s.LearningQueue = s.LearningQueue[1:]
		s.CardsReviewedSinceLastLrn = 0
	} else if len(s.LearningQueue) == 0 {
		// Only due queue has cards
		s.ActiveCard = &s.DueQueue[0]
		s.DueQueue = s.DueQueue[1:]
		s.CardsReviewedSinceLastLrn++
	} else {
		// Both queues have cards. Injected back every 3 to 5 cards (e.g. when CardsReviewedSinceLastLrn >= 3)
		if s.CardsReviewedSinceLastLrn >= 3 {
			s.ActiveCard = &s.LearningQueue[0]
			s.LearningQueue = s.LearningQueue[1:]
			s.CardsReviewedSinceLastLrn = 0
		} else {
			s.ActiveCard = &s.DueQueue[0]
			s.DueQueue = s.DueQueue[1:]
			s.CardsReviewedSinceLastLrn++
		}
	}
}

func (s *Session) saveState(wasDbWrite bool, cardID int64, preRep int, preInt int, preEF float64, preDue time.Time, logID int64) {
	copyCards := func(src []Card) []Card {
		dst := make([]Card, len(src))
		copy(dst, src)
		return dst
	}

	var activeCardCopy *Card
	if s.ActiveCard != nil {
		c := *s.ActiveCard
		activeCardCopy = &c
	}

	s.UndoStack = append(s.UndoStack, SessionState{
		DueQueue:                  copyCards(s.DueQueue),
		LearningQueue:             copyCards(s.LearningQueue),
		GraduatedPool:             copyCards(s.GraduatedPool),
		ActiveCard:                activeCardCopy,
		CardsReviewedSinceLastLrn: s.CardsReviewedSinceLastLrn,
		CompletedCount:            s.CompletedCount,
		IsFlipped:                 s.IsFlipped,
		ShowHint:                  s.ShowHint,
		WasDbWrite:                wasDbWrite,
		DbCardID:                  cardID,
		DbPreRepetitionCount:      preRep,
		DbPreInterval:             preInt,
		DbPreEaseFactor:           preEF,
		DbPreDueAt:                preDue,
		DbReviewLogID:             logID,
	})
}

func (s *Session) GradeActiveCard(q int, db *DB) error {
	if s.ActiveCard == nil {
		return nil
	}

	card := *s.ActiveCard
	now := time.Now()

	if q < 3 {
		// Fail: updates metrics, goes to learning queue, bypasses DB writes
		updatedCard := CalculateSM2(card, q, now)

		// Cache state before changing queues
		s.saveState(false, card.ID, card.RepetitionCount, card.Interval, card.EaseFactor, card.DueAt, 0)

		// Add to back of learning queue
		s.LearningQueue = append(s.LearningQueue, updatedCard)

		s.NextCard()
		return nil
	} else {
		// Success: graduates, updates saved to DB, logs written
		updatedCard := CalculateSM2(card, q, now)

		var logID int64
		var err error

		// Write to DB atomically
		err = db.Transaction(func(tx *sql.Tx) error {
			if err := db.UpdateCard(tx, &updatedCard); err != nil {
				return err
			}

			log := ReviewLog{
				CardID:              card.ID,
				ReviewedAt:          now,
				PreRepetitionCount:  card.RepetitionCount,
				PreInterval:         card.Interval,
				PreEaseFactor:       card.EaseFactor,
				Grade:               q,
				PostRepetitionCount: updatedCard.RepetitionCount,
				PostInterval:        updatedCard.Interval,
				PostEaseFactor:      updatedCard.EaseFactor,
			}
			logID, err = db.CreateReviewLog(tx, &log)
			return err
		})
		if err != nil {
			return err
		}

		// Cache state for Undo
		s.saveState(true, card.ID, card.RepetitionCount, card.Interval, card.EaseFactor, card.DueAt, logID)

		// Add to graduated pool
		s.GraduatedPool = append(s.GraduatedPool, updatedCard)
		s.CompletedCount++

		s.NextCard()
		return nil
	}
}

func (s *Session) Undo(db *DB) error {
	if len(s.UndoStack) == 0 {
		return nil
	}

	lastIdx := len(s.UndoStack) - 1
	state := s.UndoStack[lastIdx]
	s.UndoStack = s.UndoStack[:lastIdx]

	// Rollback database changes if any
	if state.WasDbWrite && db != nil {
		err := db.Transaction(func(tx *sql.Tx) error {
			// Restore card's old parameters
			card, err := db.GetCard(state.DbCardID)
			if err != nil {
				return err
			}
			card.RepetitionCount = state.DbPreRepetitionCount
			card.Interval = state.DbPreInterval
			card.EaseFactor = state.DbPreEaseFactor
			card.DueAt = state.DbPreDueAt
			if err := db.UpdateCard(tx, card); err != nil {
				return err
			}

			// Delete the review log
			_, err = tx.Exec("DELETE FROM review_logs WHERE id = ?", state.DbReviewLogID)
			return err
		})
		if err != nil {
			return err
		}
	}

	// Restore in-memory queue states
	s.DueQueue = state.DueQueue
	s.LearningQueue = state.LearningQueue
	s.GraduatedPool = state.GraduatedPool
	s.ActiveCard = state.ActiveCard
	s.CardsReviewedSinceLastLrn = state.CardsReviewedSinceLastLrn
	s.CompletedCount = state.CompletedCount
	s.IsFlipped = state.IsFlipped
	s.ShowHint = state.ShowHint

	return nil
}
