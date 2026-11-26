package config

import (
	"path/filepath"
	"testing"
)

func TestLoadMissingReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.yaml")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error for missing config, got %v", err)
	}
	if cfg.DecksDir != DefaultDecksDir {
		t.Errorf("expected DecksDir %q, got %q", DefaultDecksDir, cfg.DecksDir)
	}
	if cfg.MemorizeOptions != 5 {
		t.Errorf("expected MemorizeOptions 5, got %d", cfg.MemorizeOptions)
	}
	if cfg.MemorizeMode != MemorizeModeOpts {
		t.Errorf("expected MemorizeMode %q, got %q", MemorizeModeOpts, cfg.MemorizeMode)
	}
	if cfg.QuizSM2Sync || cfg.MatchSM2Sync {
		t.Errorf("expected SM-2 sync flags to default to false")
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "config.yaml")
	cfg := Default()
	cfg.FilePath = path
	cfg.QuizSM2Sync = true
	cfg.MatchSM2Sync = true
	cfg.MatchHidePinyin = true
	cfg.MemorizeOptions = 3
	cfg.MemorizeMode = MemorizeModeTyping
	cfg.DecksDir = "/custom/decks"

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !loaded.QuizSM2Sync || !loaded.MatchSM2Sync || !loaded.MatchHidePinyin {
		t.Errorf("expected bool flags to persist as true, got %+v", loaded)
	}
	if loaded.MemorizeOptions != 3 {
		t.Errorf("expected MemorizeOptions 3, got %d", loaded.MemorizeOptions)
	}
	if loaded.MemorizeMode != MemorizeModeTyping {
		t.Errorf("expected MemorizeMode %q, got %q", MemorizeModeTyping, loaded.MemorizeMode)
	}
	if loaded.DecksDir != "/custom/decks" {
		t.Errorf("expected DecksDir /custom/decks, got %q", loaded.DecksDir)
	}
}

func TestNormalizeClampsInvalidValues(t *testing.T) {
	cfg := &Config{
		MemorizeOptions: 99,
		MemorizeMode:    "unknown",
	}
	cfg.Normalize()
	if cfg.MemorizeOptions != 5 {
		t.Errorf("expected clamped MemorizeOptions 5, got %d", cfg.MemorizeOptions)
	}
	if cfg.MemorizeMode != MemorizeModeOpts {
		t.Errorf("expected fallback MemorizeMode %q, got %q", MemorizeModeOpts, cfg.MemorizeMode)
	}
	if cfg.StatsFile != DefaultStatsFile {
		t.Errorf("expected default StatsFile %q, got %q", DefaultStatsFile, cfg.StatsFile)
	}
}
