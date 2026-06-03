package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"flashtui/db"
)

func (m Model) UpdateConsole(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	if !strings.HasPrefix(op, ":") {
		op = ":" + op
	}

	switch op {
	case ":w", ":w!":
		m.SetStatus("Synced all session database states.", false)
		m.RefreshData()

	case ":q", ":q!", ":quit", ":quit!", ":exit", ":exit!":
		_ = m.Database.Close()
		return m, tea.Quit

	case ":wq", ":wq!", ":x", ":x!":
		m.SetStatus("Synced all session database states.", false)
		m.RefreshData()
		_ = m.Database.Close()
		return m, tea.Quit

	case ":import":
		if len(parts) < 2 {
			m.SetStatus("Usage: :import <path>", true)
			return m, nil
		}
		if len(m.Decks) == 0 {
			m.SetStatus("Please create a deck first before importing cards.", true)
			return m, nil
		}
		path := strings.Join(parts[1:], " ")
		deck := m.Decks[m.SelectedDeckIdx]
		count, err := ImportMarkdown(m.Database, deck.ID, path)
		if err != nil {
			m.SetStatus(fmt.Sprintf("Import error: %v", err), true)
		} else {
			m.SetStatus(fmt.Sprintf("Successfully imported %d cards into '%s'.", count, deck.Name), false)
			m.RefreshData()
		}

	case ":tag":
		if len(parts) < 2 {
			m.TagFilter = ""
			m.SetStatus("Cleared tag filter.", false)
		} else {
			m.TagFilter = parts[1]
			m.SetStatus(fmt.Sprintf("Filtering cards by tag: #%s", m.TagFilter), false)
		}
		m.RefreshData()

	case ":tags":
		tagMap := make(map[string]bool)
		for _, c := range m.Cards {
			for _, t := range c.Tags {
				if t != "" {
					tagMap[t] = true
				}
			}
		}
		if len(tagMap) == 0 {
			m.SetStatus("No tags found in the current deck.", false)
		} else {
			var tags []string
			for t := range tagMap {
				tags = append(tags, "#"+t)
			}
			m.SetStatus(fmt.Sprintf("Available tags in deck: %s", strings.Join(tags, ", ")), false)
		}

	case ":export":
		if len(parts) < 3 {
			m.SetStatus("Usage: :export <deck_name> <path>", true)
			return m, nil
		}
		deckName := parts[1]
		path := strings.Join(parts[2:], " ")

		var targetDeck *db.Deck
		for _, d := range m.Decks {
			if strings.EqualFold(d.Name, deckName) {
				td := d
				targetDeck = &td
				break
			}
		}

		if targetDeck == nil {
			m.SetStatus(fmt.Sprintf("Deck '%s' not found.", deckName), true)
			return m, nil
		}

		err := ExportDeckToPath(m.Database, targetDeck.ID, path)
		if err != nil {
			m.SetStatus(fmt.Sprintf("Export error: %v", err), true)
		} else {
			m.SetStatus(fmt.Sprintf("Successfully exported '%s' to '%s'.", deckName, path), false)
		}

	case ":theme":
		if len(parts) < 2 {
			var themeNames []string
			for t := range Themes {
				themeNames = append(themeNames, t)
			}
			m.SetStatus(fmt.Sprintf("Usage: :theme <name> (available: %s) [current: %s]", strings.Join(themeNames, ", "), m.Config.ThemeName), false)
		} else {
			name := strings.ToLower(parts[1])
			if m.ApplyTheme(name) {
				m.SetStatus(fmt.Sprintf("Theme changed to '%s'.", name), false)
				m.RefreshData()
			} else {
				m.SetStatus(fmt.Sprintf("Theme '%s' not found.", name), true)
			}
		}

	case ":help", ":h":
		m.UIMode = ModeHelp
		m.HelpScrollOffset = 0
		m.SetStatus("Opened help panel.", false)

	case ":generate", ":gen":
		if len(parts) < 2 {
			m.SetStatus("Usage: :generate <topic_name>", true)
			return m, nil
		}
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			m.SetStatus("Error: GEMINI_API_KEY environment variable is not set.", true)
			return m, nil
		}
		topic := strings.Join(parts[1:], " ")
		m.SetStatus(fmt.Sprintf("Contacting Gemini to generate cards on '%s'...", topic), false)
		m.UIMode = ModeDashboard
		return m, GenerateCardsCmd(apiKey, topic)

	case ":advice", ":coach":
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			m.SetStatus("Error: GEMINI_API_KEY environment variable is not set.", true)
			return m, nil
		}
		m.SetStatus("Asking Gemini for study coach advice...", false)
		m.UIMode = ModeDashboard

		totalCards, mastered, _ := m.Database.GetMasteryStats()
		streak, _ := m.Database.GetDailyStreak()
		ret, _ := m.Database.GetRetentionAccuracy()
		activity, _ := m.Database.GetLast7DaysActivity()

		var sb strings.Builder
		sb.WriteString("Decks list:\n")
		for _, d := range m.Decks {
			sb.WriteString(fmt.Sprintf("- Deck: %s (Cards: %d, Due: %d)\n", d.Name, d.CardCount, d.DueCount))
		}

		prompt := fmt.Sprintf(`Analyze my study progress for these flashcards:
- Total Decks: %d
- Total Cards: %d
- Mastered Cards: %d (Mastery Rate: %.1f%%)
- Daily Streak: %d days
- Memory Retention/Accuracy: %.1f%%
- Last 7 days review activity: %v

Decks breakdown:
%s

Provide concise, encouraging, and highly actionable advice (max 200 words) on:
1. What I am doing well.
2. What I should focus on next (which decks need attention).
3. Tips for optimizing retention.
Keep the tone encouraging, study-focused, and friendly like a memory coach. Use bullet points. Keep it clear.
`, len(m.Decks), totalCards, mastered, float64(mastered)/float64(totalCards)*100, streak, ret, activity, sb.String())

		return m, GetAdviceCmd(apiKey, prompt)

	default:
		m.SetStatus("Unknown command: "+op, true)
	}

	return m, nil
}

func (m Model) UpdateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter", "esc":
		m.UIMode = ModeDashboard
		m.RefreshData()
		return m, nil

	default:
		m.SearchInput, cmd = m.SearchInput.Update(msg)
		m.ApplySearchFilter()
		m.SelectedCardIdx = 0
		return m, cmd
	}
}
