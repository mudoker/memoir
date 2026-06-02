package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) UpdateFormDeck(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(m.FormDeckName.Value())
		if name == "" {
			m.UIMode = ModeDashboard
			return m, nil
		}

		if m.FormEditID == 0 {
			_, err := m.Database.CreateDeck(name, nil)
			if err != nil {
				m.SetStatus("Create Deck error: "+err.Error(), true)
			} else {
				m.SetStatus(fmt.Sprintf("Created deck '%s'.", name), false)
			}
		} else {
			if err := m.Database.RenameDeck(m.FormEditID, name); err != nil {
				m.SetStatus("Rename Deck error: "+err.Error(), true)
			} else {
				m.SetStatus("Renamed deck.", false)
			}
		}
		m.UIMode = ModeDashboard
		m.RefreshData()
		return m, nil

	case "esc":
		m.UIMode = ModeDashboard
		return m, nil

	default:
		m.FormDeckName, cmd = m.FormDeckName.Update(msg)
		return m, cmd
	}
}

func (m Model) UpdateFormCard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	key := msg.String()

	switch key {
	case "enter":
		front := strings.TrimSpace(m.FormCardFront.Value())
		back := strings.TrimSpace(m.FormCardBack.Value())
		hint := strings.TrimSpace(m.FormCardHint.Value())
		tagsRaw := m.FormCardTags.Value()

		if front == "" || back == "" {
			m.SetStatus("Front and Back cannot be empty!", true)
			return m, nil
		}

		var tags []string
		for _, t := range strings.Split(tagsRaw, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}

		if m.FormEditID == 0 {
			deck := m.Decks[m.SelectedDeckIdx]
			_, err := m.Database.CreateCard(deck.ID, front, back, hint, tags)
			if err != nil {
				m.SetStatus("Create Card error: "+err.Error(), true)
			} else {
				m.SetStatus("Created card.", false)
			}
		} else {
			card, err := m.Database.GetCard(m.FormEditID)
			if err == nil {
				card.Front = front
				card.Back = back
				card.Hint = hint
				card.Tags = tags
				err = m.Database.UpdateCard(nil, card)
			}
			if err != nil {
				m.SetStatus("Edit Card error: "+err.Error(), true)
			} else {
				m.SetStatus("Updated card.", false)
			}
		}

		m.UIMode = ModeDashboard
		m.RefreshData()
		return m, nil

	case "esc":
		m.UIMode = ModeDashboard
		return m, nil

	case "tab", "down":
		m.FormActiveField = (m.FormActiveField + 1) % 4
		m.focusCardFormField()
		return m, nil

	case "shift+tab", "up":
		m.FormActiveField = (m.FormActiveField - 1 + 4) % 4
		m.focusCardFormField()
		return m, nil

	default:
		switch m.FormActiveField {
		case 0:
			m.FormCardFront, cmd = m.FormCardFront.Update(msg)
		case 1:
			m.FormCardBack, cmd = m.FormCardBack.Update(msg)
		case 2:
			m.FormCardHint, cmd = m.FormCardHint.Update(msg)
		case 3:
			m.FormCardTags, cmd = m.FormCardTags.Update(msg)
		}
		return m, cmd
	}
}

func (m *Model) focusCardFormField() {
	m.FormCardFront.Blur()
	m.FormCardBack.Blur()
	m.FormCardHint.Blur()
	m.FormCardTags.Blur()

	switch m.FormActiveField {
	case 0:
		m.FormCardFront.Focus()
	case 1:
		m.FormCardBack.Focus()
	case 2:
		m.FormCardHint.Focus()
	case 3:
		m.FormCardTags.Focus()
	}
}
