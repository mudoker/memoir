package srs

import (
	"database/sql"
	"time"

	"flashtui/db"
)

type SessionState struct {
	DueQueue                  []db.Card
	LearningQueue             []db.Card
	GraduatedPool             []db.Card
	ActiveCard                *db.Card
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

func (s *Session) SaveState(wasDbWrite bool, cardID int64, preRep int, preInt int, preEF float64, preDue time.Time, logID int64) {
	copyCards := func(src []db.Card) []db.Card {
		dst := make([]db.Card, len(src))
		copy(dst, src)
		return dst
	}

	var activeCardCopy *db.Card
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

func (s *Session) Undo(database *db.DB) error {
	if len(s.UndoStack) == 0 {
		return nil
	}

	lastIdx := len(s.UndoStack) - 1
	state := s.UndoStack[lastIdx]
	s.UndoStack = s.UndoStack[:lastIdx]

	if state.WasDbWrite && database != nil {
		err := database.Transaction(func(tx *sql.Tx) error {
			card, err := database.GetCard(state.DbCardID)
			if err != nil {
				return err
			}
			card.RepetitionCount = state.DbPreRepetitionCount
			card.Interval = state.DbPreInterval
			card.EaseFactor = state.DbPreEaseFactor
			card.DueAt = state.DbPreDueAt
			if err := database.UpdateCard(tx, card); err != nil {
				return err
			}

			_, err = tx.Exec("DELETE FROM review_logs WHERE id = ?", state.DbReviewLogID)
			return err
		})
		if err != nil {
			return err
		}
	}

	s.DueQueue = state.DueQueue
	s.LearningQueue = state.LearningQueue
	s.GraduatedPool = state.GraduatedPool
	s.ActiveCard = state.ActiveCard
	s.CardsReviewedSinceLastLrn = state.CardsReviewedSinceLastLrn
	s.CompletedCount = state.CompletedCount
	s.IsFlipped = state.IsFlipped
	s.ShowHint = state.ShowHint

	// If card is restored, we reduce lapses if we track them
	if s.ActiveCard != nil {
		s.Lapses[s.ActiveCard.ID] = s.Lapses[s.ActiveCard.ID] - 1
		if s.Lapses[s.ActiveCard.ID] < 0 {
			s.Lapses[s.ActiveCard.ID] = 0
		}
	}

	return nil
}
