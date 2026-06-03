package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"flashtui/config"
)

func (m Model) UpdateFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		newKey := strings.TrimSpace(m.FormGeminiKey.Value())
		m.UIMode = ModeDashboard

		if newKey == "" {
			m.SetStatus("Gemini API key cannot be empty.", true)
			return m, nil
		}

		m.Config.GeminiAPIKey = newKey
		err := config.SaveConfig(m.Config)
		if err != nil {
			m.SetStatus(fmt.Sprintf("Failed to save config: %v", err), true)
			return m, nil
		}

		m.SetStatus("Gemini API key saved to config.yaml.", false)

		// Resume the pending command!
		if m.PendingGeminiCmd == "generate" {
			topic := m.PendingGeminiTopic
			m.SetStatus(fmt.Sprintf("Contacting Gemini to generate cards on '%s'...", topic), false)
			return m, GenerateCardsCmd(newKey, topic)
		} else if m.PendingGeminiCmd == "advice" {
			m.SetStatus("Asking Gemini for study coach advice...", false)

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

			return m, GetAdviceCmd(newKey, prompt)
		}

		return m, nil

	case "esc":
		m.UIMode = ModeDashboard
		m.SetStatus("Canceled Gemini API key setup.", false)
		return m, nil

	default:
		m.FormGeminiKey, cmd = m.FormGeminiKey.Update(msg)
		return m, cmd
	}
}
