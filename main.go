package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"flashcards/internal/config"
	"flashcards/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", config.DefaultConfigPath, "path to persistent config.yaml")
	flag.StringVar(&configPath, "c", config.DefaultConfigPath, "path to persistent config.yaml (shorthand)")

	var decksDir string
	flag.StringVar(&decksDir, "decks", config.DefaultDecksDir, "path to the decks directory")
	flag.StringVar(&decksDir, "d", config.DefaultDecksDir, "path to the decks directory (shorthand)")

	var textsDir string
	flag.StringVar(&textsDir, "texts", config.DefaultTextsDir, "path to the memorization texts directory")
	flag.StringVar(&textsDir, "t", config.DefaultTextsDir, "path to the memorization texts directory (shorthand)")

	var readingDir string
	flag.StringVar(&readingDir, "reading", config.DefaultReadingDir, "path to the reading translation texts directory")
	flag.StringVar(&readingDir, "r", config.DefaultReadingDir, "path to the reading translation texts directory (shorthand)")

	var memorizeOptions int
	flag.IntVar(&memorizeOptions, "options", 5, "number of choices in Memorize by Options mode (2-5)")
	flag.IntVar(&memorizeOptions, "o", 5, "number of choices in Memorize by Options mode (2-5) (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\nOptions:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load %s: %v (using defaults)\n", configPath, err)
		cfg = config.Default()
		cfg.FilePath = configPath
	}

	explicitFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		explicitFlags[f.Name] = true
	})

	if explicitFlags["decks"] || explicitFlags["d"] {
		cfg.DecksDir = decksDir
	}
	if explicitFlags["texts"] || explicitFlags["t"] {
		cfg.TextsDir = textsDir
	}
	if explicitFlags["reading"] || explicitFlags["r"] {
		cfg.ReadingDir = readingDir
	}
	if (explicitFlags["options"] || explicitFlags["o"]) && memorizeOptions >= 2 && memorizeOptions <= 5 {
		cfg.MemorizeOptions = memorizeOptions
	}

	initialModel := ui.NewWithConfig(cfg)
	if (explicitFlags["options"] || explicitFlags["o"]) && memorizeOptions >= 2 && memorizeOptions <= 5 {
		initialModel.SetMemorizeOptionCount(memorizeOptions)
	}

	p := tea.NewProgram(initialModel, tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running Corvidae: %v\n", err)
		os.Exit(1)
	}
}
