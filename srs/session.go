package srs

import (
	"database/sql"
	"math/rand"
	"time"

	"flashtui/db"
)

type Session struct {
	AllCards                  []db.Card
	DueQueue                  []db.Card
	LearningQueue             []db.Card
	GraduatedPool             []db.Card
	UndoStack                 []SessionState
	ActiveCard                *db.Card
	IsFlipped                 bool
	ShowHint                  bool
	CardsReviewedSinceLastLrn int
	TotalSessionCards         int
	CompletedCount            int

	// Enhanced logic features
	Lapses     map[int64]int // Maps card ID to number of session lapses (fails)
	LeechCount int           // Total unique cards marked as leech in this session
	Ratings    []int         // Track graded score history for statistics
}

func NewSession(cards []db.Card) *Session {
	s := &Session{
		AllCards:                  cards,
		DueQueue:                  make([]db.Card, len(cards)),
		LearningQueue:             []db.Card{},
		GraduatedPool:             []db.Card{},
		UndoStack:                 []SessionState{},
		CardsReviewedSinceLastLrn: 0,
		TotalSessionCards:         len(cards),
		CompletedCount:            0,
		Lapses:                    make(map[int64]int),
		Ratings:                   []int{},
	}
	copy(s.DueQueue, cards)

	s.NextCard()
	return s
}

func (s *Session) ShuffleQueue() {
	if len(s.DueQueue) > 0 {
		rand.Shuffle(len(s.DueQueue), func(i, j int) {
			s.DueQueue[i], s.DueQueue[j] = s.DueQueue[j], s.DueQueue[i]
		})
	}
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
		s.ActiveCard = &s.LearningQueue[0]
		s.LearningQueue = s.LearningQueue[1:]
		s.CardsReviewedSinceLastLrn = 0
	} else if len(s.LearningQueue) == 0 {
		s.ActiveCard = &s.DueQueue[0]
		s.DueQueue = s.DueQueue[1:]
		s.CardsReviewedSinceLastLrn++
	} else {
		// Inject from learning queue every 3 reviews
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

func (s *Session) GradeActiveCard(q int, database *db.DB) error {
	if s.ActiveCard == nil {
		return nil
	}

	card := *s.ActiveCard
	now := time.Now()

	s.Ratings = append(s.Ratings, q)

	if q < 3 {
		// Recall lapse: increment lapses counter
		s.Lapses[card.ID] = s.Lapses[card.ID] + 1
		updatedCard := CalculateSM2(card, q, now)

		s.SaveState(false, card.ID, card.RepetitionCount, card.Interval, card.EaseFactor, card.DueAt, 0)
		s.LearningQueue = append(s.LearningQueue, updatedCard)

		s.NextCard()
		return nil
	} else {
		updatedCard := CalculateSM2(card, q, now)
		var logID int64
		var err error

		err = database.Transaction(func(tx *sql.Tx) error {
			if err := database.UpdateCard(tx, &updatedCard); err != nil {
				return err
			}

			log := db.ReviewLog{
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
			logID, err = database.CreateReviewLog(tx, &log)
			return err
		})
		if err != nil {
			return err
		}

		s.SaveState(true, card.ID, card.RepetitionCount, card.Interval, card.EaseFactor, card.DueAt, logID)
		s.GraduatedPool = append(s.GraduatedPool, updatedCard)
		s.CompletedCount++

		s.NextCard()
		return nil
	}
}

func (s *Session) IsLeech(cardID int64) bool {
	return s.Lapses[cardID] >= 3
}
