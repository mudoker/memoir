package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DatabasePath string `yaml:"database_path"`
	Theme        Theme  `yaml:"theme"`
}

type Theme struct {
	PrimaryColor    string `yaml:"primary_color"`    // Highlight / selected item
	SecondaryColor  string `yaml:"secondary_color"`  // Active borders, accents
	BackgroundColor string `yaml:"background_color"`
	TextColor       string `yaml:"text_color"`
}

func DefaultConfig() Config {
	return Config{
		DatabasePath: "", // Will be filled with default ~/.config/flashtui/data.db
		Theme: Theme{
			PrimaryColor:    "#875faf", // Purple/Violet accent
			SecondaryColor:  "#005f87", // Deep Cyan accent
			BackgroundColor: "#1c1c1c", // Sleek dark gray
			TextColor:       "#bcbcbc", // Off-white
		},
	}
}

func LoadConfig() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return DefaultConfig(), err
	}
	configDir := filepath.Join(home, ".config", "flashtui")
	configPath := filepath.Join(configDir, "config.yaml")

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return DefaultConfig(), err
	}

	cfg := DefaultConfig()
	cfg.DatabasePath = filepath.Join(configDir, "data.db")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Save default config
		data, err := yaml.Marshal(cfg)
		if err == nil {
			_ = os.WriteFile(configPath, data, 0644)
		}
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, err
	}

	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return cfg, err
	}

	if fileCfg.DatabasePath != "" {
		cfg.DatabasePath = fileCfg.DatabasePath
	}
	if fileCfg.Theme.PrimaryColor != "" {
		cfg.Theme.PrimaryColor = fileCfg.Theme.PrimaryColor
	}
	if fileCfg.Theme.SecondaryColor != "" {
		cfg.Theme.SecondaryColor = fileCfg.Theme.SecondaryColor
	}
	if fileCfg.Theme.BackgroundColor != "" {
		cfg.Theme.BackgroundColor = fileCfg.Theme.BackgroundColor
	}
	if fileCfg.Theme.TextColor != "" {
		cfg.Theme.TextColor = fileCfg.Theme.TextColor
	}

	return cfg, nil
}
