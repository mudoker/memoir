package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"

	"flashtui/config"
	"flashtui/db"
	"flashtui/srs"
)

type UIMode int

const (
	ModeDashboard UIMode = iota
	ModeReview
	ModeFormDeck
	ModeFormCard
	ModeConsole
	ModeSearch
)

type ActivePanel int

const (
	PanelDecks ActivePanel = iota
	PanelCards
)

type Model struct {
	Database *db.DB
	Config   config.Config

	Width  int
	Height int

	UIMode      UIMode
	ActivePanel ActivePanel

	// Data
	Decks            []db.Deck
	SelectedDeckIdx  int
	Cards            []db.Card
	FilteredCards    []db.Card
	SelectedCardIdx  int
	DeckScrollOffset int
	CardScrollOffset int

	// Active review session
	Session *srs.Session

	// Interactive Inputs
	ConsoleInput  textinput.Model
	SearchInput   textinput.Model
	FormDeckName  textinput.Model
	FormCardFront textinput.Model
	FormCardBack  textinput.Model
	FormCardHint  textinput.Model
	FormCardTags  textinput.Model

	FormActiveField int   // 0: Front, 1: Back, 2: Hint, 3: Tags
	FormEditID      int64 // 0 if creating, otherwise ID being edited

	// State indicators
	StatusMsg  string
	StatusTime time.Time
	GPressed   bool
}

func NewModel(database *db.DB, cfg config.Config) Model {
	consoleInput := textinput.New()
	consoleInput.Prompt = ":"
	consoleInput.CharLimit = 120

	searchInput := textinput.New()
	searchInput.Prompt = "Search: /"
	searchInput.CharLimit = 50

	formDeckName := textinput.New()
	formDeckName.Prompt = "Deck Name: "
	formDeckName.CharLimit = 50

	formCardFront := textinput.New()
	formCardFront.Prompt = "Front: "
	formCardFront.CharLimit = 200

	formCardBack := textinput.New()
	formCardBack.Prompt = "Back: "
	formCardBack.CharLimit = 500

	formCardHint := textinput.New()
	formCardHint.Prompt = "Hint (Optional): "
	formCardHint.CharLimit = 200

	formCardTags := textinput.New()
	formCardTags.Prompt = "Tags (comma-separated): "
	formCardTags.CharLimit = 100

	m := Model{
		Database:      database,
		Config:        cfg,
		UIMode:        ModeDashboard,
		ActivePanel:   PanelDecks,
		ConsoleInput:  consoleInput,
		SearchInput:   searchInput,
		FormDeckName:  formDeckName,
		FormCardFront: formCardFront,
		FormCardBack:  formCardBack,
		FormCardHint:  formCardHint,
		FormCardTags:  formCardTags,
	}

	m.RefreshData()
	return m
}

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
	if query == "" {
		m.FilteredCards = m.Cards
		return
	}

	var filtered []db.Card
	for _, c := range m.Cards {
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
