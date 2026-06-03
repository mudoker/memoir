package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DatabasePath string `yaml:"database_path"`
	ThemeName    string `yaml:"theme_name"`
	Theme        Theme  `yaml:"theme"`
	GeminiAPIKey string `yaml:"gemini_api_key"`
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
		ThemeName:    "catppuccin",
		Theme: Theme{
			PrimaryColor:    "#cba6f7", // Mauve
			SecondaryColor:  "#89b4fa", // Blue
			BackgroundColor: "#1e1e2e", // Mocha base
			TextColor:       "#cdd6f4", // Text
		},
		GeminiAPIKey: "",
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
	if fileCfg.ThemeName != "" {
		cfg.ThemeName = fileCfg.ThemeName
	}
	if fileCfg.GeminiAPIKey != "" {
		cfg.GeminiAPIKey = fileCfg.GeminiAPIKey
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

func SaveConfig(cfg Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".config", "flashtui")
	configPath := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
