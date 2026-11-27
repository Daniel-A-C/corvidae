// Package config manages persistent user configuration stored in config.yaml.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath  = "config.yaml"
	DefaultDecksDir    = "decks"
	DefaultTextsDir    = "memorizationTexts"
	DefaultReadingDir  = "readingTranslationTexts"
	DefaultForgeDir    = "forge"
	DefaultStatsFile   = "stats.yaml"
	MemorizeModeOpts   = "options"
	MemorizeModeTyping = "typing"
)

// Config holds persisted user preferences for Corvidae.
type Config struct {
	DecksDir        string `yaml:"decks_dir,omitempty"`
	TextsDir        string `yaml:"texts_dir,omitempty"`
	ReadingDir      string `yaml:"reading_dir,omitempty"`
	ForgeDir        string `yaml:"forge_dir,omitempty"`
	StatsFile       string `yaml:"stats_file,omitempty"`
	QuizSM2Sync     bool   `yaml:"quiz_sm2_sync"`
	MatchSM2Sync    bool   `yaml:"match_sm2_sync"`
	ForgeSM2Sync    bool   `yaml:"forge_sm2_sync"`
	MatchHidePinyin bool   `yaml:"match_hide_pinyin"`
	MemorizeOptions int    `yaml:"memorize_options,omitempty"` // 2 to 5
	MemorizeMode    string `yaml:"memorize_mode,omitempty"`    // "options" or "typing"
	FilePath        string `yaml:"-"`
}

// Default returns a Config initialized with sensible defaults.
func Default() *Config {
	return &Config{
		DecksDir:        DefaultDecksDir,
		TextsDir:        DefaultTextsDir,
		ReadingDir:      DefaultReadingDir,
		ForgeDir:        DefaultForgeDir,
		StatsFile:       DefaultStatsFile,
		QuizSM2Sync:     false,
		MatchSM2Sync:    false,
		ForgeSM2Sync:    false,
		MatchHidePinyin: false,
		MemorizeOptions: 5,
		MemorizeMode:    MemorizeModeOpts,
		FilePath:        DefaultConfigPath,
	}
}

// Normalize ensures all fields have valid values within bounds.
func (c *Config) Normalize() {
	if c.DecksDir == "" {
		c.DecksDir = DefaultDecksDir
	}
	if c.TextsDir == "" {
		c.TextsDir = DefaultTextsDir
	}
	if c.ReadingDir == "" {
		c.ReadingDir = DefaultReadingDir
	}
	if c.ForgeDir == "" {
		c.ForgeDir = DefaultForgeDir
	}
	if c.StatsFile == "" {
		c.StatsFile = DefaultStatsFile
	}
	if c.MemorizeOptions < 2 || c.MemorizeOptions > 5 {
		c.MemorizeOptions = 5
	}
	if c.MemorizeMode != MemorizeModeOpts && c.MemorizeMode != MemorizeModeTyping {
		c.MemorizeMode = MemorizeModeOpts
	}
}

// Load reads a Config from path. If the file does not exist, defaults are returned without error.
func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath
	}
	cfg := Default()
	cfg.FilePath = path

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("failed to read config %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config %s: %w", path, err)
	}
	cfg.FilePath = path
	cfg.Normalize()
	return cfg, nil
}

// Save writes the Config to its FilePath on disk.
func (c *Config) Save() error {
	if c == nil || c.FilePath == "" {
		return nil
	}
	c.Normalize()
	dir := filepath.Dir(c.FilePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(c.FilePath, data, 0644)
}
