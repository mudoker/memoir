package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) ViewHelpPanel() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(GreenColor).Underline(true).Render("💡 FLASHTUI COMMANDS & HOTKEYS") + "\n\n")

	// Keyboard Shortcuts Section
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(WhiteColor).Render("Keyboard Navigation:") + "\n")
	b.WriteString("  h / l      : Switch Panels (Decks / Cards)\n")
	b.WriteString("  j / k      : Navigate List selections\n")
	b.WriteString("  gg / G     : Jump to top / bottom of list\n")
	b.WriteString("  ctrl+u/d   : Scroll list half-page up / down\n")
	b.WriteString("  a          : Add new Deck or Card\n")
	b.WriteString("  e          : Edit selected Deck or Card\n")
	b.WriteString("  dd         : Delete selected Deck or Card\n")
	b.WriteString("  /          : Enter Search mode\n")
	b.WriteString("  :          : Enter Console command mode\n")
	b.WriteString("  Enter      : Start review session on selected Deck\n\n")

	// Console Commands Section
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(WhiteColor).Render("Console Commands (:):") + "\n")
	b.WriteString("  :w         : Sync session database changes\n")
	b.WriteString("  :q / :q!   : Safe exit / force exit\n")
	b.WriteString("  :wq / :x   : Save changes and exit\n")
	b.WriteString("  :theme <n> : Change theme (catppuccin, tokyonight, gruvbox, etc.)\n")
	b.WriteString("  :tag <tag> : Filter cards list by specific tag\n")
	b.WriteString("  :tags      : List all unique tags in current deck\n")
	b.WriteString("  :import <p>: Parse cards from external Markdown file\n")
	b.WriteString("  :export <d> <p>: Export deck recursively to JSON file\n")
	b.WriteString("  :help / :h : Open this help overview panel\n\n")

	b.WriteString(GrayLightStyle.Render("[Press Esc or q to return to the Dashboard]"))

	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor).
		Padding(2, 6).
		Width(60).
		Align(lipgloss.Left).
		Render(b.String())

	return AddShadow(helpBox)
}
