package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

// UI style definitions
var (
	accentColor     = lipgloss.Color("#875faf") // Vibrant Violet/Purple
	accentSecColor  = lipgloss.Color("#0087af") // Deep Cyan/Teal
	grayDarkColor   = lipgloss.Color("#1c1c1c") // Dark background
	grayMidColor    = lipgloss.Color("#303030") // Active/Inactive boundaries
	grayLightColor  = lipgloss.Color("#8a8a8a") // Muted text
	textColor       = lipgloss.Color("#d0d0d0") // Standard text
	whiteColor      = lipgloss.Color("#ffffff") // Active text/borders
	greenColor      = lipgloss.Color("#5fdf87") // Success / Correct
	redColor        = lipgloss.Color("#df5f87") // Failed / Warning
	yellowColor     = lipgloss.Color("#faf089") // Hint / Highlight

	// Style Helpers
	redStyle        = lipgloss.NewStyle().Foreground(redColor)
	greenStyle      = lipgloss.NewStyle().Foreground(greenColor)
	yellowStyle      = lipgloss.NewStyle().Foreground(yellowColor)
	grayLightStyle  = lipgloss.NewStyle().Foreground(grayLightColor)
	accentStyle     = lipgloss.NewStyle().Foreground(accentColor)
	accentSecStyle  = lipgloss.NewStyle().Foreground(accentSecColor)

	// Main Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(whiteColor).
			Background(accentColor).
			Padding(0, 2)

	activeBorderColor   = accentColor
	inactiveBorderColor = grayMidColor

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(inactiveBorderColor).
			Padding(1, 2)

	activePanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(activeBorderColor).
			Padding(1, 2)

	cursorStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			Background(accentColor).
			Bold(true)

	// Bottom Stats panel
	statsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentSecColor).
			Padding(1, 2)
)

type Model struct {
	db     *DB
	config Config

	Width  int
	Height int

	UIMode      UIMode
	ActivePanel ActivePanel

	// Data
	Decks             []Deck
	SelectedDeckIdx   int
	Cards             []Card
	FilteredCards     []Card
	SelectedCardIdx   int
	DeckScrollOffset  int
	CardScrollOffset  int

	// Active review session
	Session *Session

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
	gPressed   bool
}

func NewModel(db *DB, config Config) Model {
	// Setup text inputs
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
		db:            db,
		config:        config,
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

	m.refreshData()
	return m
}

func (m *Model) refreshData() {
	decks, err := m.db.GetDeckTree()
	if err != nil {
		m.setStatus("DB Error loading decks: "+err.Error(), true)
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

	// Boundary check selected index
	if m.SelectedDeckIdx >= len(m.Decks) {
		m.SelectedDeckIdx = len(m.Decks) - 1
	}
	if m.SelectedDeckIdx < 0 {
		m.SelectedDeckIdx = 0
	}

	deck := m.Decks[m.SelectedDeckIdx]
	cards, err := m.db.GetCardsInDeck(deck.ID)
	if err != nil {
		m.setStatus("DB Error loading cards: "+err.Error(), true)
		return
	}
	m.Cards = cards

	// Apply search filter if any
	m.applySearchFilter()

	// Bound selected card index
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

func (m *Model) applySearchFilter() {
	query := strings.TrimSpace(strings.ToLower(m.SearchInput.Value()))
	if query == "" {
		m.FilteredCards = m.Cards
		return
	}

	var filtered []Card
	for _, c := range m.Cards {
		if strings.Contains(strings.ToLower(c.Front), query) || strings.Contains(strings.ToLower(c.Back), query) || strings.Contains(strings.ToLower(strings.Join(c.Tags, " ")), query) {
			filtered = append(filtered, c)
		}
	}
	m.FilteredCards = filtered
}

func (m *Model) setStatus(msg string, isError bool) {
	if isError {
		m.StatusMsg = redStyle.Render(msg)
	} else {
		m.StatusMsg = greenStyle.Render(msg)
	}
	m.StatusTime = time.Now()
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Clear status msg after 3 seconds
		if time.Since(m.StatusTime) > 3*time.Second {
			m.StatusMsg = ""
		}

		// Handle UI modes separately
		switch m.UIMode {
		case ModeConsole:
			return m.updateConsole(msg)
		case ModeSearch:
			return m.updateSearch(msg)
		case ModeFormDeck:
			return m.updateFormDeck(msg)
		case ModeFormCard:
			return m.updateFormCard(msg)
		case ModeReview:
			return m.updateReview(msg)
		default:
			return m.updateDashboard(msg)
		}
	}

	return m, nil
}

func (m Model) updateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle standard vim key combination sequences (e.g. gg)
	if m.gPressed {
		m.gPressed = false
		if key == "g" {
			if m.ActivePanel == PanelDecks {
				m.SelectedDeckIdx = 0
				m.DeckScrollOffset = 0
			} else {
				m.SelectedCardIdx = 0
				m.CardScrollOffset = 0
			}
			m.refreshData()
			return m, nil
		}
	}

	switch key {
	case ":":
		m.UIMode = ModeConsole
		m.ConsoleInput.SetValue("")
		m.ConsoleInput.Focus()
		return m, textinput.Blink

	case "/":
		m.UIMode = ModeSearch
		m.SearchInput.Focus()
		m.ActivePanel = PanelCards // Switch to cards panel to see filtered results
		return m, textinput.Blink

	case "h":
		m.ActivePanel = PanelDecks
		m.refreshData()
	case "l":
		m.ActivePanel = PanelCards
		m.refreshData()

	case "j", "down":
		if m.ActivePanel == PanelDecks {
			if len(m.Decks) > 0 {
				m.SelectedDeckIdx = (m.SelectedDeckIdx + 1) % len(m.Decks)
				m.adjustDeckScroll()
			}
		} else {
			if len(m.FilteredCards) > 0 {
				m.SelectedCardIdx = (m.SelectedCardIdx + 1) % len(m.FilteredCards)
				m.adjustCardScroll()
			}
		}
		m.refreshData()

	case "k", "up":
		if m.ActivePanel == PanelDecks {
			if len(m.Decks) > 0 {
				m.SelectedDeckIdx = (m.SelectedDeckIdx - 1 + len(m.Decks)) % len(m.Decks)
				m.adjustDeckScroll()
			}
		} else {
			if len(m.FilteredCards) > 0 {
				m.SelectedCardIdx = (m.SelectedCardIdx - 1 + len(m.FilteredCards)) % len(m.FilteredCards)
				m.adjustCardScroll()
			}
		}
		m.refreshData()

	case "g":
		m.gPressed = true
		return m, nil

	case "G":
		if m.ActivePanel == PanelDecks {
			if len(m.Decks) > 0 {
				m.SelectedDeckIdx = len(m.Decks) - 1
				m.adjustDeckScroll()
			}
		} else {
			if len(m.FilteredCards) > 0 {
				m.SelectedCardIdx = len(m.FilteredCards) - 1
				m.adjustCardScroll()
			}
		}
		m.refreshData()

	case "ctrl+d": // Half-page down
		step := (m.Height - 10) / 2
		if m.ActivePanel == PanelDecks {
			m.SelectedDeckIdx += step
			if m.SelectedDeckIdx >= len(m.Decks) {
				m.SelectedDeckIdx = len(m.Decks) - 1
			}
			m.adjustDeckScroll()
		} else {
			m.SelectedCardIdx += step
			if m.SelectedCardIdx >= len(m.FilteredCards) {
				m.SelectedCardIdx = len(m.FilteredCards) - 1
			}
			m.adjustCardScroll()
		}
		m.refreshData()

	case "ctrl+u": // Half-page up
		step := (m.Height - 10) / 2
		if m.ActivePanel == PanelDecks {
			m.SelectedDeckIdx -= step
			if m.SelectedDeckIdx < 0 {
				m.SelectedDeckIdx = 0
			}
			m.adjustDeckScroll()
		} else {
			m.SelectedCardIdx -= step
			if m.SelectedCardIdx < 0 {
				m.SelectedCardIdx = 0
			}
			m.adjustCardScroll()
		}
		m.refreshData()

	case "a": // Create
		if m.ActivePanel == PanelDecks {
			m.UIMode = ModeFormDeck
			m.FormEditID = 0
			m.FormDeckName.SetValue("")
			m.FormDeckName.Focus()
			return m, textinput.Blink
		} else {
			if len(m.Decks) == 0 {
				m.setStatus("Must create a deck first!", true)
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

	case "e": // Edit
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

	case "d": // Potential delete combo "dd"
		// Handled simply as standard "dd"
		return m, nil

	case "d d": // Purge selection
		if m.ActivePanel == PanelDecks {
			if len(m.Decks) == 0 {
				return m, nil
			}
			d := m.Decks[m.SelectedDeckIdx]
			if err := m.db.DeleteDeck(d.ID); err != nil {
				m.setStatus("Delete Deck error: "+err.Error(), true)
			} else {
				m.setStatus(fmt.Sprintf("Deleted deck '%s' and all contents recursively.", d.Name), false)
				m.SelectedDeckIdx = 0
				m.refreshData()
			}
		} else {
			if len(m.FilteredCards) == 0 {
				return m, nil
			}
			c := m.FilteredCards[m.SelectedCardIdx]
			if err := m.db.DeleteCard(c.ID); err != nil {
				m.setStatus("Delete Card error: "+err.Error(), true)
			} else {
				m.setStatus("Deleted card.", false)
				m.SelectedCardIdx = 0
				m.refreshData()
			}
		}

	case "enter": // Study selected deck
		if m.ActivePanel == PanelDecks && len(m.Decks) > 0 {
			deck := m.Decks[m.SelectedDeckIdx]
			dueCards, err := m.db.GetDueCards(deck.ID)
			if err != nil {
				m.setStatus("Load due cards error: "+err.Error(), true)
				return m, nil
			}
			if len(dueCards) == 0 {
				m.setStatus(fmt.Sprintf("No due cards in '%s' (including subdecks)!", deck.Name), false)
				return m, nil
			}

			m.Session = NewSession(dueCards)
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
	visibleRows := height - 2 // Padding + border
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
	visibleRows := height - 3 // Padding + border + header row
	if m.SelectedCardIdx < m.CardScrollOffset {
		m.CardScrollOffset = m.SelectedCardIdx
	} else if m.SelectedCardIdx >= m.CardScrollOffset+visibleRows {
		m.CardScrollOffset = m.SelectedCardIdx - visibleRows + 1
	}
}

// --- Console Mode Update ---
func (m Model) updateConsole(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		cmdText := m.ConsoleInput.Value()
		m.UIMode = ModeDashboard
		return m.executeConsoleCommand(cmdText)

	case "esc":
		m.UIMode = ModeDashboard
		return m, nil

	default:
		m.ConsoleInput, cmd = m.ConsoleInput.Update(msg)
		return m, cmd
	}
}

func (m Model) executeConsoleCommand(cmdText string) (tea.Model, tea.Cmd) {
	cmdText = strings.TrimSpace(cmdText)
	if cmdText == "" {
		return m, nil
	}

	parts := strings.Split(cmdText, " ")
	op := parts[0]

	switch op {
	case ":w":
		m.setStatus("Synced all session database states.", false)
		m.refreshData()

	case ":q":
		_ = m.db.Close()
		return m, tea.Quit

	case ":import":
		if len(parts) < 2 {
			m.setStatus("Usage: :import <path>", true)
			return m, nil
		}
		if len(m.Decks) == 0 {
			m.setStatus("Please create a deck first before importing cards.", true)
			return m, nil
		}
		path := strings.Join(parts[1:], " ")
		deck := m.Decks[m.SelectedDeckIdx]
		count, err := ImportMarkdown(m.db, deck.ID, path)
		if err != nil {
			m.setStatus(fmt.Sprintf("Import error: %v", err), true)
		} else {
			m.setStatus(fmt.Sprintf("Successfully imported %d cards into '%s'.", count, deck.Name), false)
			m.refreshData()
		}

	case ":export":
		if len(parts) < 3 {
			m.setStatus("Usage: :export <deck_name> <path>", true)
			return m, nil
		}
		deckName := parts[1]
		path := strings.Join(parts[2:], " ")

		// Find deck
		var targetDeck *Deck
		for _, d := range m.Decks {
			if strings.EqualFold(d.Name, deckName) {
				td := d
				targetDeck = &td
				break
			}
		}

		if targetDeck == nil {
			m.setStatus(fmt.Sprintf("Deck '%s' not found.", deckName), true)
			return m, nil
		}

		err := ExportDeckToPath(m.db, targetDeck.ID, path)
		if err != nil {
			m.setStatus(fmt.Sprintf("Export error: %v", err), true)
		} else {
			m.setStatus(fmt.Sprintf("Successfully exported '%s' to '%s'.", deckName, path), false)
		}

	default:
		m.setStatus("Unknown command: "+op, true)
	}

	return m, nil
}

// --- Search Mode Update ---
func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter", "esc":
		m.UIMode = ModeDashboard
		m.refreshData()
		return m, nil

	default:
		m.SearchInput, cmd = m.SearchInput.Update(msg)
		m.applySearchFilter()
		m.SelectedCardIdx = 0
		return m, cmd
	}
}

// --- Form Deck Mode Update ---
func (m Model) updateFormDeck(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(m.FormDeckName.Value())
		if name == "" {
			m.UIMode = ModeDashboard
			return m, nil
		}

		if m.FormEditID == 0 {
			// Creating deck. Check selection. If Left panel is focused,
			// let's check if the user wants to nest it under the selected deck.
			// To avoid confusion, let's ask or support nesting if a parent deck is selected.
			// Let's create it as a child of the selected deck if SelectedDeckIdx exists,
			// or ask: for now, if Left panel is selected and we have decks, we make it nested.
			// Wait, let's do a simple check: if we have decks, create it nested under the selected deck if parent_id is wanted,
			// or create as a top-level deck. Let's make top-level default, but if user wants nested we can let them.
			// Let's support creating top-level decks by default. If we edit it we can keep its parent.
			var parentID *int64 = nil
			// To keep it simple, top-level creation:
			_, err := m.db.CreateDeck(name, parentID)
			if err != nil {
				m.setStatus("Create Deck error: "+err.Error(), true)
			} else {
				m.setStatus(fmt.Sprintf("Created deck '%s'.", name), false)
			}
		} else {
			// Editing deck
			if err := m.db.RenameDeck(m.FormEditID, name); err != nil {
				m.setStatus("Rename Deck error: "+err.Error(), true)
			} else {
				m.setStatus("Renamed deck.", false)
			}
		}
		m.UIMode = ModeDashboard
		m.refreshData()
		return m, nil

	case "esc":
		m.UIMode = ModeDashboard
		return m, nil

	default:
		m.FormDeckName, cmd = m.FormDeckName.Update(msg)
		return m, cmd
	}
}

// --- Form Card Mode Update ---
func (m Model) updateFormCard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	key := msg.String()

	switch key {
	case "enter":
		// Save card
		front := strings.TrimSpace(m.FormCardFront.Value())
		back := strings.TrimSpace(m.FormCardBack.Value())
		hint := strings.TrimSpace(m.FormCardHint.Value())
		tagsRaw := m.FormCardTags.Value()

		if front == "" || back == "" {
			m.setStatus("Front and Back cannot be empty!", true)
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
			// Create
			deck := m.Decks[m.SelectedDeckIdx]
			_, err := m.db.CreateCard(deck.ID, front, back, hint, tags)
			if err != nil {
				m.setStatus("Create Card error: "+err.Error(), true)
			} else {
				m.setStatus("Created card.", false)
			}
		} else {
			// Edit
			card, err := m.db.GetCard(m.FormEditID)
			if err == nil {
				card.Front = front
				card.Back = back
				card.Hint = hint
				card.Tags = tags
				err = m.db.UpdateCard(nil, card)
			}
			if err != nil {
				m.setStatus("Edit Card error: "+err.Error(), true)
			} else {
				m.setStatus("Updated card.", false)
			}
		}

		m.UIMode = ModeDashboard
		m.refreshData()
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

// --- Review Mode Update ---
func (m Model) updateReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.Session == nil {
		m.UIMode = ModeDashboard
		return m, nil
	}

	key := msg.String()

	switch key {
	case "esc":
		m.UIMode = ModeDashboard
		m.refreshData()
		return m, nil

	case " ":
		// Flip card
		m.Session.IsFlipped = true
		return m, nil

	case "h":
		// Show hint
		m.Session.ShowHint = true
		return m, nil

	case "s":
		// Shuffle queue
		m.Session.ShuffleQueue()
		m.setStatus("Remaining queue shuffled.", false)
		return m, nil

	case "u":
		// Undo
		if err := m.Session.Undo(m.db); err != nil {
			m.setStatus("Undo error: "+err.Error(), true)
		} else {
			m.setStatus("Rolled back last assessment.", false)
		}
		return m, nil

	case "1", "2", "3", "4", "5":
		grade := int(key[0] - '0')
		if m.Session.ActiveCard != nil {
			// If not flipped, we can force flip or process
			// Typically, user rates *after* flipping. Let's make sure it's flipped.
			if !m.Session.IsFlipped {
				m.Session.IsFlipped = true
				return m, nil
			}

			err := m.Session.GradeActiveCard(grade, m.db)
			if err != nil {
				m.setStatus("Database write error: "+err.Error(), true)
			}
		}
		return m, nil
	}

	return m, nil
}

// --- Views Rendering ---

func (m Model) View() string {
	// Canvas size checks
	if m.Width < 100 || m.Height < 30 {
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(redColor).Bold(true).Render(
				fmt.Sprintf("Terminal Size Too Small!\n\nRequires: 100x30 minimum\nCurrent:  %dx%d\n\nPlease enlarge your terminal window.", m.Width, m.Height),
			),
		)
	}

	// Layout views
	switch m.UIMode {
	case ModeFormDeck:
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.viewFormDeck())
	case ModeFormCard:
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.viewFormCard())
	case ModeReview:
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.viewReview())
	default:
		return m.viewDashboard()
	}
}

func (m Model) viewDashboard() string {
	var b strings.Builder

	// 1. Header
	title := titleStyle.Render(fmt.Sprintf("FlashTUI ─ v1.0.0 (Local Time: %s)", time.Now().Format("15:04:05")))
	headerText := fmt.Sprintf(" ╭%s╮\n", strings.Repeat("─", m.Width-2))
	headerMid := fmt.Sprintf(" │  %-*s │\n", m.Width-6, title)
	headerText += headerMid
	headerText += fmt.Sprintf(" ╰%s╯", strings.Repeat("─", m.Width-2))
	b.WriteString(headerText + "\n")

	// 2. Dual Panels
	leftW := int(float64(m.Width) * 0.3)
	rightW := m.Width - leftW - 2

	panelH := m.Height - 11 // Adjust height

	// Render Left Panel: Decks
	leftStyle := panelStyle.Width(leftW - 4).Height(panelH)
	if m.ActivePanel == PanelDecks && m.UIMode == ModeDashboard {
		leftStyle = activePanelStyle.Width(leftW - 4).Height(panelH)
	}

	var decksStr strings.Builder
	decksStr.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).Render("DECKS MANAGER (Normal Mode)") + "\n\n")

	if len(m.Decks) == 0 {
		decksStr.WriteString(" (No decks created)\n Press 'a' to create.")
	} else {
		// Draw decks tree
		visibleHeight := panelH - 2
		start := m.DeckScrollOffset
		end := start + visibleHeight
		if end > len(m.Decks) {
			end = len(m.Decks)
		}

		for idx := start; idx < end; idx++ {
			d := m.Decks[idx]

			// Formatting deck tree
			isSel := m.SelectedDeckIdx == idx
			nameText := d.Name
			badge := fmt.Sprintf("[%d]", d.DueCount)
			if d.DueCount > 0 {
				badge = greenStyle.Render(badge)
			} else {
				badge = grayLightStyle.Render(badge)
			}

			if isSel {
				if m.ActivePanel == PanelDecks {
					nameText = cursorStyle.Render(nameText)
				} else {
					nameText = lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(nameText)
				}
			}

			// Sub-deck indentation prefix
			prefix := ""
			if d.Depth > 0 {
				prefix = strings.Repeat("│   ", d.Depth-1)
				if isLastChild(m.Decks, idx) {
					prefix += "└── "
				} else {
					prefix += "├── "
				}
			}

			if isSel {
				decksStr.WriteString(fmt.Sprintf(" ▶  %s%s %s\n", prefix, nameText, badge))
			} else {
				decksStr.WriteString(fmt.Sprintf("    %s%s %s\n", prefix, nameText, badge))
			}
		}
	}
	leftView := leftStyle.Render(decksStr.String())

	// Render Right Panel: Cards in Selection
	rightStyle := panelStyle.Width(rightW - 4).Height(panelH)
	if m.ActivePanel == PanelCards && m.UIMode == ModeDashboard {
		rightStyle = activePanelStyle.Width(rightW - 4).Height(panelH)
	}

	var cardsStr strings.Builder
	cardsStr.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).Render("CARDS IN SELECTION") + "\n\n")

	// Table Headers
	colIdW := 6
	colDueW := 12
	colTagsW := 15
	colFrontW := rightW - 4 - colIdW - colDueW - colTagsW - 8 // padding adjusts
	if colFrontW < 10 {
		colFrontW = 10
	}

	headerRow := fmt.Sprintf("%-*s %-*s %-*s %-*s\n", colIdW, "ID", colFrontW, "FRONT", colDueW, "DUE", colTagsW, "TAGS")
	cardsStr.WriteString(lipgloss.NewStyle().Bold(true).Foreground(whiteColor).Render(headerRow))

	if len(m.FilteredCards) == 0 {
		cardsStr.WriteString("\n (No cards matching filter/selection)\n Press 'a' to add a card.")
	} else {
		visibleHeight := panelH - 3
		start := m.CardScrollOffset
		end := start + visibleHeight
		if end > len(m.FilteredCards) {
			end = len(m.FilteredCards)
		}

		for idx := start; idx < end; idx++ {
			c := m.FilteredCards[idx]

			isSel := m.SelectedCardIdx == idx
			idStr := fmt.Sprintf("%03d", c.ID)
			frontText := c.Front
			dueText := formatDue(c.DueAt)

			// Tags with '#'
			var tagStrs []string
			for _, t := range c.Tags {
				if t != "" {
					tagStrs = append(tagStrs, "#"+t)
				}
			}
			tagsText := strings.Join(tagStrs, " ")

			// Truncate Front text
			if len(frontText) > colFrontW {
				frontText = frontText[:colFrontW-1] + "…"
			}
			if len(tagsText) > colTagsW {
				tagsText = tagsText[:colTagsW-1] + "…"
			}

			row := fmt.Sprintf("%-*s %-*s %-*s %-*s", colIdW, idStr, colFrontW, frontText, colDueW, dueText, colTagsW, tagsText)

			if isSel {
				if m.ActivePanel == PanelCards {
					cardsStr.WriteString(cursorStyle.Render(row) + "\n")
				} else {
					cardsStr.WriteString(lipgloss.NewStyle().Foreground(accentColor).Render(row) + "\n")
				}
			} else {
				cardsStr.WriteString(row + "\n")
			}
		}
	}
	rightView := rightStyle.Render(cardsStr.String())

	// Combine left and right panels side-by-side
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView) + "\n")

	// 3. Stats Panel
	streak, _ := m.db.GetDailyStreak()
	ret, _ := m.db.GetRetentionAccuracy()
	totalCards, mastered, _ := m.db.GetMasteryStats()

	// Progress bar
	barW := 40
	pct := 0.0
	if totalCards > 0 {
		pct = float64(mastered) / float64(totalCards)
	}
	barStr := renderProgressBar(barW, pct)

	statsView := statsStyle.Width(m.Width - 4).Render(
		fmt.Sprintf("Current Daily Streak: %s %d Days | Retention Accuracy Score: %.1f%%\nMastered Cards:       [%s] %.1f%% (%d/%d)",
			greenStyle.Render("█"), streak, ret, barStr, pct*100.0, mastered, totalCards,
		),
	)
	b.WriteString(statsView + "\n")

	// 4. Console / Help / Status bar
	if m.UIMode == ModeSearch {
		b.WriteString(m.SearchInput.View())
	} else {
		// Normal mode
		helpLine := " j/k: Navigation | h/l: Panes | a: Create Deck/Card | e: Edit | dd: Purge | /: Search | : cmd"
		statusText := m.StatusMsg
		if statusText == "" {
			b.WriteString(lipgloss.NewStyle().Foreground(grayLightColor).Render(helpLine))
		} else {
			b.WriteString(statusText)
		}
	}

	return b.String()
}

func (m Model) viewFormDeck() string {
	title := "CREATE NEW DECK"
	if m.FormEditID != 0 {
		title = "RENAME DECK"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("╭%s╮\n", strings.Repeat("─", 50)))
	b.WriteString(fmt.Sprintf("│ %-*s │\n", 48, lipgloss.NewStyle().Bold(true).Foreground(whiteColor).Render(title)))
	b.WriteString(fmt.Sprintf("├%s┤\n", strings.Repeat("─", 50)))
	b.WriteString("│                                                  │\n")
	b.WriteString(fmt.Sprintf("│  %s │\n", m.FormDeckName.View()))
	b.WriteString("│                                                  │\n")
	b.WriteString(fmt.Sprintf("├%s┤\n", strings.Repeat("─", 50)))
	b.WriteString("│  [Enter] Confirm  |  [Esc] Cancel                │\n")
	b.WriteString(fmt.Sprintf("╰%s╯", strings.Repeat("─", 50)))

	return b.String()
}

func (m Model) viewFormCard() string {
	title := "CREATE NEW CARD"
	if m.FormEditID != 0 {
		title = "EDIT CARD"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("╭%s╮\n", strings.Repeat("─", 70)))
	b.WriteString(fmt.Sprintf("│ %-*s │\n", 68, lipgloss.NewStyle().Bold(true).Foreground(whiteColor).Render(title)))
	b.WriteString(fmt.Sprintf("├%s┤\n", strings.Repeat("─", 70)))
	b.WriteString("│                                                                    │\n")

	fields := []string{"Front", "Back", "Hint", "Tags"}
	for i, name := range fields {
		var activeIndicator string
		var viewStr string
		switch i {
		case 0:
			viewStr = m.FormCardFront.View()
			if m.FormActiveField == 0 {
				activeIndicator = "▶"
			} else {
				activeIndicator = " "
			}
		case 1:
			viewStr = m.FormCardBack.View()
			if m.FormActiveField == 1 {
				activeIndicator = "▶"
			} else {
				activeIndicator = " "
			}
		case 2:
			viewStr = m.FormCardHint.View()
			if m.FormActiveField == 2 {
				activeIndicator = "▶"
			} else {
				activeIndicator = " "
			}
		case 3:
			viewStr = m.FormCardTags.View()
			if m.FormActiveField == 3 {
				activeIndicator = "▶"
			} else {
				activeIndicator = " "
			}
		}

		b.WriteString(fmt.Sprintf("│ %s %-6s : %-56s │\n", activeIndicator, name, viewStr))
		b.WriteString("│                                                                    │\n")
	}

	b.WriteString(fmt.Sprintf("├%s┤\n", strings.Repeat("─", 70)))
	b.WriteString("│  [Tab] Cycle Fields  |  [Enter] Save  |  [Esc] Cancel              │\n")
	b.WriteString(fmt.Sprintf("╰%s╯", strings.Repeat("─", 70)))

	return b.String()
}

func (m Model) viewReview() string {
	if m.Session == nil {
		return ""
	}

	// If no active card, session completed
	if m.Session.ActiveCard == nil {
		var b strings.Builder
		b.WriteString("╭──────────────────────────────────────────────────────────╮\n")
		b.WriteString("│                                                          │\n")
		b.WriteString("│                  🎉 SESSION COMPLETED!                   │\n")
		b.WriteString("│                                                          │\n")
		b.WriteString("│        You have successfully reviewed all due cards.     │\n")
		b.WriteString("│                                                          │\n")
		b.WriteString("│        [Press Esc to return to the Deck Manager]        │\n")
		b.WriteString("│                                                          │\n")
		b.WriteString("╰──────────────────────────────────────────────────────────╯")
		return b.String()
	}

	card := m.Session.ActiveCard
	deck := m.Decks[m.SelectedDeckIdx]

	var b strings.Builder

	// Header progress bar
	pct := 0.0
	if m.Session.TotalSessionCards > 0 {
		pct = float64(m.Session.CompletedCount) / float64(m.Session.TotalSessionCards)
	}
	barStr := renderProgressBar(40, pct)
	progressText := fmt.Sprintf("Queue Progress: [%s] %d%% (%d/%d)", barStr, int(pct*100), m.Session.CompletedCount, m.Session.TotalSessionCards)

	title := fmt.Sprintf(" Reviewing: %s ", deck.Name)
	boxW := m.Width - 10
	if boxW > 85 {
		boxW = 85
	}

	b.WriteString(fmt.Sprintf("╭─%s%s─╮\n", title, strings.Repeat("─", boxW-len(title)-4)))
	b.WriteString(fmt.Sprintf("│  %-*s  │\n", boxW-6, progressText))
	b.WriteString(fmt.Sprintf("├%s┤\n", strings.Repeat("─", boxW-2)))
	b.WriteString("│                                                                              │\n")

	// Wrap text inside container
	contentW := boxW - 8
	frontStyle := lipgloss.NewStyle().Width(contentW).Align(lipgloss.Left)

	// Render Question
	b.WriteString(fmt.Sprintf("│  %s  │\n", lipgloss.NewStyle().Bold(true).Render("Question:")))
	wrappedFront := frontStyle.Render(card.Front)
	for _, line := range strings.Split(wrappedFront, "\n") {
		b.WriteString(fmt.Sprintf("│    %-*s  │\n", boxW-8, line))
	}
	b.WriteString("│                                                                              │\n")

	// Render hint if available and requested
	if card.Hint != "" {
		if m.Session.ShowHint {
			b.WriteString(fmt.Sprintf("│  %s  │\n", lipgloss.NewStyle().Italic(true).Foreground(yellowColor).Render("Hint: "+card.Hint)))
		} else {
			b.WriteString(fmt.Sprintf("│  %s  │\n", lipgloss.NewStyle().Italic(true).Foreground(grayLightColor).Render("[Hint Available: Press 'h' to peek]")))
		}
		b.WriteString("│                                                                              │\n")
	}

	// Divider + Back answer if flipped
	if m.Session.IsFlipped {
		b.WriteString(fmt.Sprintf("├%s┤\n", strings.Repeat("─", boxW-2)))
		b.WriteString("│                                                                              │\n")
		b.WriteString(fmt.Sprintf("│  %s  │\n", lipgloss.NewStyle().Bold(true).Render("Answer:")))
		wrappedBack := frontStyle.Render(card.Back)
		for _, line := range strings.Split(wrappedBack, "\n") {
			b.WriteString(fmt.Sprintf("│    %-*s  │\n", boxW-8, line))
		}
		b.WriteString("│                                                                              │\n")
	}

	b.WriteString(fmt.Sprintf("╰%s╯\n", strings.Repeat("─", boxW-2)))

	// Review Menu / Footer
	var footer string
	if m.Session.IsFlipped {
		footer = " [1-5]: Rate performance (1: Forgot, 2: Hard, 3: Good, 4: Easy, 5: Perfect) | u: Undo | Esc: Exit"
	} else {
		footer = " [Space]: Flip Card Back | h: Reveal Hint | s: Shuffle Queue | Esc: Exit"
	}
	b.WriteString(lipgloss.NewStyle().Foreground(grayLightColor).Render(footer))

	return b.String()
}

// Render console command line view at the bottom
func (m Model) viewConsole() string {
	return m.ConsoleInput.View()
}

// Helper to draw progress bars
func renderProgressBar(width int, ratio float64) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filledLen := int(math.Round(float64(width) * ratio))
	emptyLen := width - filledLen
	if filledLen < 0 {
		filledLen = 0
	}
	if emptyLen < 0 {
		emptyLen = 0
	}
	return strings.Repeat("█", filledLen) + strings.Repeat("░", emptyLen)
}

func formatDue(dueAt time.Time) string {
	now := time.Now()
	if dueAt.Before(now) {
		return "Instantly"
	}
	diff := dueAt.Sub(now)
	days := int(math.Ceil(diff.Hours() / 24.0))
	if days <= 1 {
		return "Tomorrow"
	}
	return fmt.Sprintf("%d Days", days)
}

func isLastChild(decks []Deck, idx int) bool {
	depth := decks[idx].Depth
	for i := idx + 1; i < len(decks); i++ {
		if decks[i].Depth == depth {
			return false
		}
		if decks[i].Depth < depth {
			return true
		}
	}
	return true
}

// Markdown Import Logic
func ImportMarkdown(db *DB, deckID int64, filePath string) (int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(content), "\n")
	var cardsCreated int
	var currentFront string
	var currentBack []string
	var currentHint string
	var currentTags []string

	saveCurrentCard := func() error {
		if currentFront != "" {
			backStr := strings.TrimSpace(strings.Join(currentBack, "\n"))
			_, err := db.CreateCard(deckID, currentFront, backStr, currentHint, currentTags)
			if err != nil {
				return err
			}
			cardsCreated++
		}
		currentFront = ""
		currentBack = nil
		currentHint = ""
		currentTags = nil
		return nil
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Count hash symbols
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 && strings.HasPrefix(parts[0], "#") {
				// Save previous card first
				if err := saveCurrentCard(); err != nil {
					return cardsCreated, err
				}
				currentFront = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(trimmed, "<!--") && strings.HasSuffix(trimmed, "-->") {
			inner := strings.TrimSpace(trimmed[4 : len(trimmed)-3])
			if strings.HasPrefix(strings.ToLower(inner), "hint:") {
				currentHint = strings.TrimSpace(inner[5:])
			} else if strings.HasPrefix(strings.ToLower(inner), "tags:") {
				tagsPart := strings.TrimSpace(inner[5:])
				rawTags := strings.Split(tagsPart, ",")
				for _, t := range rawTags {
					t = strings.TrimSpace(t)
					if t != "" {
						currentTags = append(currentTags, t)
					}
				}
			}
		} else {
			if currentFront != "" {
				currentBack = append(currentBack, line)
			}
		}
	}

	// Save last card
	if err := saveCurrentCard(); err != nil {
		return cardsCreated, err
	}

	return cardsCreated, nil
}

// JSON Serialization Export Logic
type ExportDeck struct {
	Name     string       `json:"name"`
	Cards    []ExportCard `json:"cards,omitempty"`
	SubDecks []ExportDeck `json:"sub_decks,omitempty"`
}

type ExportCard struct {
	Front string   `json:"front"`
	Back  string   `json:"back"`
	Hint  string   `json:"hint"`
	Tags  []string `json:"tags"`
}

func compileExportDeck(db *DB, deckID int64) (ExportDeck, error) {
	var name string
	err := db.conn.QueryRow("SELECT name FROM decks WHERE id = ?", deckID).Scan(&name)
	if err != nil {
		return ExportDeck{}, err
	}

	dbCards, err := db.GetCardsInDeck(deckID)
	if err != nil {
		return ExportDeck{}, err
	}

	cards := make([]ExportCard, len(dbCards))
	for i, c := range dbCards {
		cards[i] = ExportCard{
			Front: c.Front,
			Back:  c.Back,
			Hint:  c.Hint,
			Tags:  c.Tags,
		}
	}

	// Query child decks
	rows, err := db.conn.Query("SELECT id FROM decks WHERE parent_id = ?", deckID)
	if err != nil {
		return ExportDeck{}, err
	}
	defer rows.Close()

	var childIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return ExportDeck{}, err
		}
		childIDs = append(childIDs, id)
	}

	var subDecks []ExportDeck
	for _, cid := range childIDs {
		sd, err := compileExportDeck(db, cid)
		if err != nil {
			return ExportDeck{}, err
		}
		subDecks = append(subDecks, sd)
	}

	return ExportDeck{
		Name:     name,
		Cards:    cards,
		SubDecks: subDecks,
	}, nil
}

func ExportDeckToPath(db *DB, deckID int64, filePath string) error {
	exportData, err := compileExportDeck(db, deckID)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, jsonData, 0644)
}
