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

func getConfigDir() string {
	// Try to find project root by looking for go.mod starting from current working directory
	if dir, err := os.Getwd(); err == nil {
		for {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return filepath.Join(dir, ".flashtui")
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	// Fallback: Try to find project root by looking for go.mod starting from executable directory
	if execPath, err := os.Executable(); err == nil {
		dir := filepath.Dir(execPath)
		for {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return filepath.Join(dir, ".flashtui")
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	// Fallback to user home directory
	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(home, ".config", "flashtui")
	}

	return ".flashtui"
}

func migrateFromOldConfigDir(newDir string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	oldDir := filepath.Join(home, ".config", "flashtui")
	if oldDir == newDir {
		return
	}

	// If old directory doesn't exist, nothing to migrate
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		return
	}

	// Ensure new directory exists
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return
	}

	copyFile := func(src, dst string) error {
		if _, err := os.Stat(dst); err == nil {
			// Destination already exists, do not overwrite
			return nil
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0644)
	}

	_ = copyFile(filepath.Join(oldDir, "config.yaml"), filepath.Join(newDir, "config.yaml"))
	_ = copyFile(filepath.Join(oldDir, "data.db"), filepath.Join(newDir, "data.db"))
	_ = copyFile(filepath.Join(oldDir, "data.db-wal"), filepath.Join(newDir, "data.db-wal"))
	_ = copyFile(filepath.Join(oldDir, "data.db-shm"), filepath.Join(newDir, "data.db-shm"))
}

func LoadConfig() (Config, error) {
	configDir := getConfigDir()

	// Try to migrate files from old config folder if new one doesn't exist or is empty
	migrateFromOldConfigDir(configDir)

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
		home, err := os.UserHomeDir()
		if err == nil {
			oldDefaultDB := filepath.Join(home, ".config", "flashtui", "data.db")
			if fileCfg.DatabasePath == oldDefaultDB {
				cfg.DatabasePath = filepath.Join(configDir, "data.db")
				// Update config file to point to the new workspace database path
				cfg.ThemeName = fileCfg.ThemeName
				cfg.GeminiAPIKey = fileCfg.GeminiAPIKey
				cfg.Theme = fileCfg.Theme
				_ = SaveConfig(cfg)
			} else {
				cfg.DatabasePath = fileCfg.DatabasePath
			}
		} else {
			cfg.DatabasePath = fileCfg.DatabasePath
		}
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
	configDir := getConfigDir()
	configPath := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
