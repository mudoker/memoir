package ui

import "github.com/charmbracelet/lipgloss"

// Palette colors
var (
	AccentColor     = lipgloss.Color("#875faf") // Vibrant Violet/Purple
	AccentSecColor  = lipgloss.Color("#0087af") // Deep Cyan/Teal
	GrayDarkColor   = lipgloss.Color("#1c1c1c") // Dark background
	GrayMidColor    = lipgloss.Color("#303030") // Active/Inactive boundaries
	GrayLightColor  = lipgloss.Color("#8a8a8a") // Muted text
	TextColor       = lipgloss.Color("#d0d0d0") // Standard text
	WhiteColor      = lipgloss.Color("#ffffff") // Active text/borders
	GreenColor      = lipgloss.Color("#5fdf87") // Success / Correct
	RedColor        = lipgloss.Color("#df5f87") // Failed / Leech warning
	YellowColor     = lipgloss.Color("#faf089") // Hint / Highlight
)

// Style Helpers
var (
	RedStyle       = lipgloss.NewStyle().Foreground(RedColor)
	GreenStyle     = lipgloss.NewStyle().Foreground(GreenColor)
	YellowStyle    = lipgloss.NewStyle().Foreground(YellowColor)
	GrayLightStyle = lipgloss.NewStyle().Foreground(GrayLightColor)
	AccentStyle    = lipgloss.NewStyle().Foreground(AccentColor)
	AccentSecStyle = lipgloss.NewStyle().Foreground(AccentSecColor)

	// Leech label styling
	LeechStyle = lipgloss.NewStyle().
			Foreground(WhiteColor).
			Background(RedColor).
			Padding(0, 1).
			Bold(true)
)

// Structural Styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(WhiteColor).
			Background(AccentColor).
			Padding(0, 2)

	ActiveBorderColor   = AccentColor
	InactiveBorderColor = GrayMidColor

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(InactiveBorderColor).
			Padding(1, 2)

	ActivePanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ActiveBorderColor).
			Padding(1, 2)

	CursorStyle = lipgloss.NewStyle().
			Foreground(WhiteColor).
			Background(AccentColor).
			Bold(true)

	// Bottom Stats panel
	StatsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentSecColor).
			Padding(1, 2)
)
