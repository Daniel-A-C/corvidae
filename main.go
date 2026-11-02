package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"flashcards/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var decksDir string
	flag.StringVar(&decksDir, "decks", "decks", "path to the decks directory")
	flag.StringVar(&decksDir, "d", "decks", "path to the decks directory (shorthand)")

	var textsDir string
	flag.StringVar(&textsDir, "texts", "memorizationTexts", "path to the memorization texts directory")
	flag.StringVar(&textsDir, "t", "memorizationTexts", "path to the memorization texts directory (shorthand)")

	var readingDir string
	flag.StringVar(&readingDir, "reading", "readingTranslationTexts", "path to the reading translation texts directory")
	flag.StringVar(&readingDir, "r", "readingTranslationTexts", "path to the reading translation texts directory (shorthand)")

	var memorizeOptions int
	flag.IntVar(&memorizeOptions, "options", 5, "number of choices in Memorize by Options mode (2-5)")
	flag.IntVar(&memorizeOptions, "o", 5, "number of choices in Memorize by Options mode (2-5) (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\nOptions:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	initialModel := ui.NewWithOptions(decksDir, textsDir, readingDir)
	if memorizeOptions >= 2 && memorizeOptions <= 5 {
		initialModel.SetMemorizeOptionCount(memorizeOptions)
	}

	p := tea.NewProgram(initialModel, tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running Corvidae: %v\n", err)
		os.Exit(1)
	}
}
