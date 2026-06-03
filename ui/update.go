package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case GeminiResultMsg:
		if msg.Err != nil {
			m.SetStatus(fmt.Sprintf("Gemini Error: %v", msg.Err), true)
			return m, nil
		}

		deckName := fmt.Sprintf("AI-%s", msg.Topic)
		var deckID int64
		var found bool
		for _, d := range m.Decks {
			if d.Name == deckName && d.ParentID == nil {
				deckID = d.ID
				found = true
				break
			}
		}

		var err error
		if !found {
			deckID, err = m.Database.CreateDeck(deckName, nil)
			if err != nil {
				m.SetStatus(fmt.Sprintf("Failed to create deck: %v", err), true)
				return m, nil
			}
		}

		addedCount := 0
		for _, c := range msg.Cards {
			_, err = m.Database.CreateCard(deckID, c.Front, c.Back, c.Hint, c.Tags)
			if err == nil {
				addedCount++
			}
		}

		m.SetStatus(fmt.Sprintf("Gemini generated %d cards in deck '%s'!", addedCount, deckName), false)
		m.RefreshData()
		return m, nil

	case GeminiAdviceMsg:
		if msg.Err != nil {
			m.SetStatus(fmt.Sprintf("Gemini Error: %v", msg.Err), true)
			return m, nil
		}
		m.GeminiAdvice = msg.Advice
		m.UIMode = ModeAdvice
		m.AdviceScrollOffset = 0
		m.SetStatus("Gemini study advice ready!", false)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			_ = m.Database.Close()
			return m, tea.Quit
		}

		if time.Since(m.StatusTime) > 3*time.Second {
			m.StatusMsg = ""
		}

		switch m.UIMode {
		case ModeConsole:
			return m.UpdateConsole(msg)
		case ModeSearch:
			return m.UpdateSearch(msg)
		case ModeFormDeck:
			return m.UpdateFormDeck(msg)
		case ModeFormCard:
			return m.UpdateFormCard(msg)
		case ModeReview:
			return m.UpdateReview(msg)
		case ModeHelp:
			return m.UpdateHelp(msg)
		case ModeAdvice:
			return m.UpdateAdvice(msg)
		case ModeFormKey:
			return m.UpdateFormKey(msg)
		case ModeConfirmDelete:
			return m.UpdateConfirmDelete(msg)
		default:
			return m.UpdateDashboard(msg)
		}
	}

	return m, nil
}
