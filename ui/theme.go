package ui

import (
	"flashtui/config"
)

var Themes = map[string]config.Theme{
	"catppuccin": {
		PrimaryColor:    "#cba6f7", // Mauve
		SecondaryColor:  "#89b4fa", // Blue
		BackgroundColor: "#1e1e2e", // Mocha base
		TextColor:       "#cdd6f4", // Text
	},
	"tokyonight": {
		PrimaryColor:    "#bb9af7", // Purple
		SecondaryColor:  "#7aa2f7", // Blue
		BackgroundColor: "#1a1b26", // Dark base
		TextColor:       "#a9b1d6", // Off-white
	},
	"gruvbox": {
		PrimaryColor:    "#d65d0e", // Orange
		SecondaryColor:  "#fabd2f", // Yellow
		BackgroundColor: "#282828", // Retro base
		TextColor:       "#ebdbb2", // Beige
	},
	"nord": {
		PrimaryColor:    "#88c0d0", // Frost ice blue
		SecondaryColor:  "#81a1c1", // Polar blue
		BackgroundColor: "#2e3440", // Polar night
		TextColor:       "#d8dee9", // Snow text
	},
	"monokai": {
		PrimaryColor:    "#f92672", // Pink
		SecondaryColor:  "#a6e22e", // Lime green
		BackgroundColor: "#272822", // Vintage dark
		TextColor:       "#f8f8f2", // Warm text
	},
	"cyberpunk": {
		PrimaryColor:    "#ff007f", // Neon pink/magenta
		SecondaryColor:  "#00f0ff", // Neon cyan
		BackgroundColor: "#080810", // Cyber dark base
		TextColor:       "#e0e0ff", // Electric light blue-white
	},
	"dracula": {
		PrimaryColor:    "#ff79c6", // Pink
		SecondaryColor:  "#50fa7b", // Green
		BackgroundColor: "#282a36", // Dracula dark bg
		TextColor:       "#f8f8f2", // Light grey-white
	},
	"vintage": {
		PrimaryColor:    "#e78a4e", // Vintage copper orange
		SecondaryColor:  "#a9b665", // Sage green
		BackgroundColor: "#1d2021", // Warm charcoal
		TextColor:       "#d4be98", // Light warm parchment
	},
}

func (m *Model) ApplyTheme(name string) bool {
	t, ok := Themes[name]
	if !ok {
		return false
	}
	m.Config.ThemeName = name
	m.Config.Theme = t
	InitStyles(t)
	_ = config.SaveConfig(m.Config)
	return true
}
