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
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	p := tea.NewProgram(ui.New(decksDir))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running Corvidae: %v\n", err)
		os.Exit(1)
	}
}
