package ui

import (
	"github.com/charmbracelet/lipgloss"

	"flashtui/config"
)

// ── Base palette (defaults, overridden by theme) ───────────────────────────
var (
	AccentColor    = lipgloss.Color("#7c6af7") // Indigo-violet primary
	AccentSecColor = lipgloss.Color("#22b8cf") // Cyan-teal secondary
	GrayDarkColor  = lipgloss.Color("#18181b") // Deepest bg (zinc-900)
	GrayMidColor   = lipgloss.Color("#27272a") // Surface 1 (zinc-800)
	GrayMid2Color  = lipgloss.Color("#3f3f46") // Surface 2 (zinc-700)
	GrayLightColor = lipgloss.Color("#71717a") // Muted text (zinc-500)
	TextColor      = lipgloss.Color("#e4e4e7") // Body text (zinc-200)
	WhiteColor     = lipgloss.Color("#fafafa") // High-contrast
	GreenColor     = lipgloss.Color("#4ade80") // Success
	RedColor       = lipgloss.Color("#f87171") // Error / leech
	YellowColor    = lipgloss.Color("#facc15") // Warn / hint
	OrangeColor    = lipgloss.Color("#fb923c") // Warm accent
)

// ── Semantic style shortcuts ───────────────────────────────────────────────
var (
	RedStyle       = lipgloss.NewStyle().Foreground(RedColor)
	GreenStyle     = lipgloss.NewStyle().Foreground(GreenColor)
	YellowStyle    = lipgloss.NewStyle().Foreground(YellowColor)
	OrangeStyle    = lipgloss.NewStyle().Foreground(OrangeColor)
	GrayLightStyle = lipgloss.NewStyle().Foreground(GrayLightColor)
	AccentStyle    = lipgloss.NewStyle().Foreground(AccentColor)
	AccentSecStyle = lipgloss.NewStyle().Foreground(AccentSecColor)

	// Leech label
	LeechStyle = lipgloss.NewStyle().
			Foreground(WhiteColor).
			Background(RedColor).
			Padding(0, 1).
			Bold(true)
)

// ── Structural styles ──────────────────────────────────────────────────────
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(WhiteColor).
			Background(AccentColor).
			Padding(0, 2)

	ActiveBorderColor   = AccentColor
	InactiveBorderColor = GrayMid2Color

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(InactiveBorderColor).
			Padding(0, 1)

	ActivePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ActiveBorderColor).
				Padding(0, 1)

	CursorStyle = lipgloss.NewStyle().
			Foreground(GrayDarkColor).
			Background(AccentColor).
			Bold(true)

	// Stats panel
	StatsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(GrayMid2Color).
			Background(GrayDarkColor).
			Padding(0, 2)

	// Header badges
	BadgeDecksStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(WhiteColor).
			Background(AccentColor).
			Padding(0, 1)

	BadgeCardsStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(WhiteColor).
			Background(AccentSecColor).
			Padding(0, 1)

	BadgeDueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(GrayDarkColor).
			Background(GreenColor).
			Padding(0, 1)
)

// InitStyles reinitialises all colour-derived style variables from the loaded theme.
func InitStyles(theme config.Theme) {
	AccentColor = lipgloss.Color(theme.PrimaryColor)
	AccentSecColor = lipgloss.Color(theme.SecondaryColor)
	GrayDarkColor = lipgloss.Color(theme.BackgroundColor)
	TextColor = lipgloss.Color(theme.TextColor)

	// Derived greys — keep fixed neutrals so panels stay legible
	GrayMidColor = lipgloss.Color("#27272a")
	GrayMid2Color = lipgloss.Color("#3f3f46")
	GrayLightColor = lipgloss.Color("#71717a")
	WhiteColor = lipgloss.Color("#fafafa")

	// Semantic shortcuts
	RedStyle = lipgloss.NewStyle().Foreground(RedColor)
	GreenStyle = lipgloss.NewStyle().Foreground(GreenColor)
	YellowStyle = lipgloss.NewStyle().Foreground(YellowColor)
	OrangeStyle = lipgloss.NewStyle().Foreground(OrangeColor)
	GrayLightStyle = lipgloss.NewStyle().Foreground(GrayLightColor)
	AccentStyle = lipgloss.NewStyle().Foreground(AccentColor)
	AccentSecStyle = lipgloss.NewStyle().Foreground(AccentSecColor)

	LeechStyle = lipgloss.NewStyle().
		Foreground(WhiteColor).
		Background(RedColor).
		Padding(0, 1).
		Bold(true)

	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(WhiteColor).
		Background(AccentColor).
		Padding(0, 2)

	ActiveBorderColor = AccentColor
	InactiveBorderColor = GrayMid2Color

	PanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(InactiveBorderColor).
		Padding(0, 1)

	ActivePanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ActiveBorderColor).
		Padding(0, 1)

	CursorStyle = lipgloss.NewStyle().
		Foreground(GrayDarkColor).
		Background(AccentColor).
		Bold(true)

	StatsStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(GrayMid2Color).
		Background(GrayDarkColor).
		Padding(0, 2)

	BadgeDecksStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(WhiteColor).
		Background(AccentColor).
		Padding(0, 1)

	BadgeCardsStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(WhiteColor).
		Background(AccentSecColor).
		Padding(0, 1)

	BadgeDueStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(GrayDarkColor).
		Background(GreenColor).
		Padding(0, 1)
}
