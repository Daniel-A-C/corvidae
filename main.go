package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"flashcards/internal/ui"
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

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	p := tea.NewProgram(ui.NewWithOptions(decksDir, textsDir, readingDir))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running Corvidae: %v\n", err)
		os.Exit(1)
	}
}
