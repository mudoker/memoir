package ui

import (
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
	ModeHelp
	ModeAdvice
	ModeFormKey
	ModeConfirmDelete
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
	HelpScrollOffset int
	AdviceScrollOffset int

	// Gemini State
	GeminiAdvice       string
	PendingGeminiCmd   string // Track what command the user was trying to run: "generate" or "advice"
	PendingGeminiTopic string // Track the topic for generate

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
	FormGeminiKey textinput.Model

	FormActiveField int   // 0: Front, 1: Back, 2: Hint, 3: Tags
	FormEditID      int64 // 0 if creating, otherwise ID being edited
	TagFilter       string

	// State indicators
	StatusMsg  string
	StatusTime time.Time
	GPressed   bool
	DPressed   bool
}

func NewModel(database *db.DB, cfg config.Config) Model {
	consoleInput := textinput.New()
	consoleInput.Prompt = ":"
	consoleInput.CharLimit = 120

	searchInput := textinput.New()
	searchInput.Prompt = "Search: /"
	searchInput.CharLimit = 50

	formDeckName := textinput.New()
	formDeckName.Prompt = ""
	formDeckName.CharLimit = 50

	formCardFront := textinput.New()
	formCardFront.Prompt = ""
	formCardFront.CharLimit = 200

	formCardBack := textinput.New()
	formCardBack.Prompt = ""
	formCardBack.CharLimit = 500

	formCardHint := textinput.New()
	formCardHint.Prompt = ""
	formCardHint.CharLimit = 200

	formCardTags := textinput.New()
	formCardTags.Prompt = ""
	formCardTags.CharLimit = 100

	formGeminiKey := textinput.New()
	formGeminiKey.Prompt = "API Key: "
	formGeminiKey.EchoMode = textinput.EchoPassword
	formGeminiKey.CharLimit = 100

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
		FormGeminiKey: formGeminiKey,
	}

	InitStyles(cfg.Theme)
	m.RefreshData()
	return m
}
