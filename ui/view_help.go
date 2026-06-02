package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) ViewHelpPanel() string {
	w := m.Width - 16
	if w < 50 {
		w = 50
	}
	if w > 90 {
		w = 90
	}

	row := func(key, desc string) string {
		keyStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentSecColor).
			Width(20)
		return keyStyle.Render(key) + GrayLightStyle.Render(desc)
	}

	sectionHeader := func(title string) string {
		return "\n" + lipgloss.NewStyle().
			Bold(true).
			Foreground(WhiteColor).
			Underline(true).
			Render(title) + "\n"
	}

	sep := GrayLightStyle.Render(strings.Repeat("─", w-4))

	var b strings.Builder

	// Title
	b.WriteString(lipgloss.NewStyle().
		Bold(true).Foreground(AccentColor).
		Render(" FlashTUI — Keyboard Reference") + "\n")
	b.WriteString(sep + "\n")

	// Navigation
	b.WriteString(sectionHeader("Navigation"))
	b.WriteString(row("h / l", "Switch panels (Decks ↔ Cards)") + "\n")
	b.WriteString(row("Tab", "Toggle active panel") + "\n")
	b.WriteString(row("j / k  or  ↑ / ↓", "Move selection up / down") + "\n")
	b.WriteString(row("gg / G", "Jump to top / bottom of list") + "\n")
	b.WriteString(row("Ctrl+u / Ctrl+d", "Scroll half-page up / down") + "\n")

	// Actions
	b.WriteString(sectionHeader("Actions"))
	b.WriteString(row("a", "Add new Deck or Card (context-aware)") + "\n")
	b.WriteString(row("e", "Edit selected Deck or Card") + "\n")
	b.WriteString(row("dd", "Delete selected Deck (cascades) or Card") + "\n")
	b.WriteString(row("Enter", "Start review session on selected Deck") + "\n")

	// Modes
	b.WriteString(sectionHeader("Modes"))
	b.WriteString(row("/", "Enter Search mode (filters cards live)") + "\n")
	b.WriteString(row(":", "Enter Command console") + "\n")
	b.WriteString(row("? or :help", "Open this help panel") + "\n")
	b.WriteString(row("q", "Quit (also Ctrl+C)") + "\n")

	b.WriteString("\n" + sep + "\n")

	// Console commands
	b.WriteString(sectionHeader("Console Commands  (:)"))
	b.WriteString(row(":w", "Sync database state") + "\n")
	b.WriteString(row(":q / :q!", "Quit / force quit") + "\n")
	b.WriteString(row(":wq / :x", "Save + quit (Vim-style)") + "\n")
	b.WriteString(row(":tag <name>", "Filter cards by tag  (blank = clear)") + "\n")
	b.WriteString(row(":tags", "List all tags in current deck") + "\n")
	b.WriteString(row(":theme <name>", "Switch colour theme") + "\n")
	b.WriteString(row("", "  catppuccin · tokyonight · gruvbox · nord · monokai") + "\n")
	b.WriteString(row(":import <path>", "Import cards from Markdown file") + "\n")
	b.WriteString(row(":export <d> <p>", "Export deck tree to JSON file") + "\n")

	b.WriteString("\n" + sep + "\n")

	// How DueDate & Tags work
	b.WriteString(sectionHeader("How Due Date & Tags Work"))
	b.WriteString(lipgloss.NewStyle().Foreground(TextColor).Width(w - 4).Render(
		"Due Date  — set automatically. New cards are due immediately (\"Instantly\").\n"+
			"After each review you rate 1–5; the SM-2 algorithm calculates the next\n"+
			"due date (e.g. +1 day, +4 days) and stores it. No manual command needed.\n",
	) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(TextColor).Width(w - 4).Render(
		"Tags  — set manually when creating or editing a card (press 'a'/'e').\n"+
			"Enter comma-separated values in the Tags field, e.g.  math, calculus\n"+
			"Use :import to bulk-load tagged cards from a Markdown file:\n"+
			"  <!-- tags: math, calculus -->  inside the card block.\n"+
			"Use :tag <name> to filter the card list by a tag; :tag alone clears it.",
	) + "\n")

	b.WriteString("\n" + sep + "\n")
	b.WriteString(GrayLightStyle.Render("Press Esc or q to close this panel"))

	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor).
		Padding(1, 3).
		Width(w).
		Render(b.String())

	return AddShadow(helpBox)
}
