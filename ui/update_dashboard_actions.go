package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"flashtui/srs"
)

func (m Model) handleDashboardActions(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "a":
		if m.ActivePanel == PanelDecks {
			m.UIMode = ModeFormDeck
			m.FormEditID = 0
			m.FormDeckName.SetValue("")
			m.FormDeckName.Focus()
			return m, textinput.Blink
		} else {
			if len(m.Decks) == 0 {
				m.SetStatus("Must create a deck first!", true)
				return m, nil
			}
			m.UIMode = ModeFormCard
			m.FormEditID = 0
			m.FormCardFront.SetValue("")
			m.FormCardBack.SetValue("")
			m.FormCardHint.SetValue("")
			m.FormCardTags.SetValue("")
			m.FormActiveField = 0
			m.FormCardFront.Focus()
			return m, textinput.Blink
		}

	case "e":
		if m.ActivePanel == PanelDecks {
			if len(m.Decks) == 0 {
				return m, nil
			}
			d := m.Decks[m.SelectedDeckIdx]
			m.UIMode = ModeFormDeck
			m.FormEditID = d.ID
			m.FormDeckName.SetValue(d.Name)
			m.FormDeckName.Focus()
			return m, textinput.Blink
		} else {
			if len(m.FilteredCards) == 0 {
				return m, nil
			}
			c := m.FilteredCards[m.SelectedCardIdx]
			m.UIMode = ModeFormCard
			m.FormEditID = c.ID
			m.FormCardFront.SetValue(c.Front)
			m.FormCardBack.SetValue(c.Back)
			m.FormCardHint.SetValue(c.Hint)
			m.FormCardTags.SetValue(strings.Join(c.Tags, ", "))
			m.FormActiveField = 0
			m.FormCardFront.Focus()
			return m, textinput.Blink
		}

	case "d d":
		if m.ActivePanel == PanelDecks {
			if len(m.Decks) == 0 {
				return m, nil
			}
			d := m.Decks[m.SelectedDeckIdx]
			if err := m.Database.DeleteDeck(d.ID); err != nil {
				m.SetStatus("Delete Deck error: "+err.Error(), true)
			} else {
				m.SetStatus(fmt.Sprintf("Deleted deck '%s' and all contents recursively.", d.Name), false)
				m.SelectedDeckIdx = 0
				m.RefreshData()
			}
		} else {
			if len(m.FilteredCards) == 0 {
				return m, nil
			}
			c := m.FilteredCards[m.SelectedCardIdx]
			if err := m.Database.DeleteCard(c.ID); err != nil {
				m.SetStatus("Delete Card error: "+err.Error(), true)
			} else {
				m.SetStatus("Deleted card.", false)
				m.SelectedCardIdx = 0
				m.RefreshData()
			}
		}

	case "enter":
		if m.ActivePanel == PanelDecks && len(m.Decks) > 0 {
			deck := m.Decks[m.SelectedDeckIdx]
			dueCards, err := m.Database.GetDueCards(deck.ID)
			if err != nil {
				m.SetStatus("Load due cards error: "+err.Error(), true)
				return m, nil
			}
			if len(dueCards) == 0 {
				m.SetStatus(fmt.Sprintf("No due cards in '%s' (including subdecks)!", deck.Name), false)
				return m, nil
			}

			m.Session = srs.NewSession(dueCards)
			m.UIMode = ModeReview
		}
	}

	return m, nil
}

func (m *Model) adjustDeckScroll() {
	height := m.Height - 10
	if height <= 2 {
		return
	}
	visibleRows := height - 2
	if m.SelectedDeckIdx < m.DeckScrollOffset {
		m.DeckScrollOffset = m.SelectedDeckIdx
	} else if m.SelectedDeckIdx >= m.DeckScrollOffset+visibleRows {
		m.DeckScrollOffset = m.SelectedDeckIdx - visibleRows + 1
	}
}

func (m *Model) adjustCardScroll() {
	height := m.Height - 10
	if height <= 2 {
		return
	}
	visibleRows := height - 3
	if m.SelectedCardIdx < m.CardScrollOffset {
		m.CardScrollOffset = m.SelectedCardIdx
	} else if m.SelectedCardIdx >= m.CardScrollOffset+visibleRows {
		m.CardScrollOffset = m.SelectedCardIdx - visibleRows + 1
	}
}
