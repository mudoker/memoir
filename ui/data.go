package ui

import (
	"strings"
	"time"

	"flashtui/db"
)

func (m *Model) RefreshData() {
	decks, err := m.Database.GetDeckTree()
	if err != nil {
		m.SetStatus("DB Error loading decks: "+err.Error(), true)
		return
	}
	m.Decks = decks

	if len(m.Decks) == 0 {
		m.SelectedDeckIdx = 0
		m.Cards = nil
		m.FilteredCards = nil
		m.SelectedCardIdx = 0
		m.DeckScrollOffset = 0
		m.CardScrollOffset = 0
		return
	}

	if m.SelectedDeckIdx >= len(m.Decks) {
		m.SelectedDeckIdx = len(m.Decks) - 1
	}
	if m.SelectedDeckIdx < 0 {
		m.SelectedDeckIdx = 0
	}

	deck := m.Decks[m.SelectedDeckIdx]
	cards, err := m.Database.GetCardsInDeck(deck.ID)
	if err != nil {
		m.SetStatus("DB Error loading cards: "+err.Error(), true)
		return
	}
	m.Cards = cards

	m.ApplySearchFilter()

	if len(m.FilteredCards) == 0 {
		m.SelectedCardIdx = 0
		m.CardScrollOffset = 0
	} else {
		if m.SelectedCardIdx >= len(m.FilteredCards) {
			m.SelectedCardIdx = len(m.FilteredCards) - 1
		}
		if m.SelectedCardIdx < 0 {
			m.SelectedCardIdx = 0
		}
	}
}

func (m *Model) ApplySearchFilter() {
	query := strings.TrimSpace(strings.ToLower(m.SearchInput.Value()))
	var baseCards []db.Card

	if m.TagFilter != "" {
		for _, c := range m.Cards {
			hasTag := false
			for _, t := range c.Tags {
				if strings.EqualFold(t, m.TagFilter) {
					hasTag = true
					break
				}
			}
			if hasTag {
				baseCards = append(baseCards, c)
			}
		}
	} else {
		baseCards = m.Cards
	}

	if query == "" {
		m.FilteredCards = baseCards
		return
	}

	var filtered []db.Card
	for _, c := range baseCards {
		if strings.Contains(strings.ToLower(c.Front), query) ||
			strings.Contains(strings.ToLower(c.Back), query) ||
			strings.Contains(strings.ToLower(strings.Join(c.Tags, " ")), query) {
			filtered = append(filtered, c)
		}
	}
	m.FilteredCards = filtered
}

func (m *Model) SetStatus(msg string, isError bool) {
	if isError {
		m.StatusMsg = RedStyle.Render(msg)
	} else {
		m.StatusMsg = GreenStyle.Render(msg)
	}
	m.StatusTime = time.Now()
}
