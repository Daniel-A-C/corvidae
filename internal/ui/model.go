package ui

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"flashcards/internal/deck"
	"flashcards/internal/quiz"
)

const (
	StateModeSelect = iota
	StateDirSelect
	StateDeckSelect
	StateReview
	StateQuiz
)

const (
	ModeReview = iota
	ModeQuiz
)

// Model represents the Bubble Tea state model for Corvidae.
type Model struct {
	BaseDir         string
	State           int
	Mode            int
	Dirs            []string
	DirCursor       int
	SelectedDir     string
	DeckFiles       []string
	Cursor          int
	SelectedFiles   map[string]bool

	Decks           map[string]deck.Deck
	ActiveCards     []deck.CardRef
	CurrentIndex    int
	ShowAnswer      bool
	ShowExplanation bool

	// Quiz state
	QuizOptions  []string
	CorrectIndex int
	ShowFeedback bool
	IsCorrect    bool

	Err    error
	Width  int
	Height int
}

// New initializes and returns a new Model.
func New(baseDir string) Model {
	if baseDir == "" {
		baseDir = "decks"
	}
	dirs, err := deck.GetDeckDirectories(baseDir)
	return Model{
		BaseDir:       baseDir,
		State:         StateModeSelect,
		Mode:          ModeReview,
		Dirs:          dirs,
		SelectedFiles: make(map[string]bool),
		Err:           err,
	}
}

// Init sets up the terminal alternate screen on launch.
func (m Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

// LoadSelectedDecks loads all currently marked decks into memory across any directories.
func (m *Model) LoadSelectedDecks() error {
	m.Decks = make(map[string]deck.Deck)
	loadedPaths := make(map[string]bool)
	for key, isSelected := range m.SelectedFiles {
		if !isSelected {
			continue
		}
		var fullPath string
		if filepath.IsAbs(key) {
			fullPath = key
		} else {
			candidate1 := filepath.Join(m.BaseDir, key)
			if _, err := os.Stat(candidate1); err == nil {
				fullPath = candidate1
			} else if m.SelectedDir != "" {
				candidate2 := filepath.Join(m.BaseDir, m.SelectedDir, key)
				if _, err := os.Stat(candidate2); err == nil {
					fullPath = candidate2
				}
			}
		}
		if fullPath != "" && !loadedPaths[fullPath] {
			loaded, err := deck.LoadDeck(fullPath)
			if err != nil {
				return err
			}
			m.Decks[fullPath] = loaded
			loadedPaths[fullPath] = true
		}
	}
	return nil
}

// isDeckSelected reports whether a deck file in the current SelectedDir is marked as selected.
func (m Model) isDeckSelected(file string) bool {
	relPath := file
	if m.SelectedDir != "" && !strings.HasPrefix(file, m.SelectedDir) {
		relPath = filepath.Join(m.SelectedDir, file)
	}
	return m.SelectedFiles[relPath] || m.SelectedFiles[file]
}

// hasAnySelectedDeck reports whether at least one deck is selected across all directories.
func (m Model) hasAnySelectedDeck() bool {
	for _, selected := range m.SelectedFiles {
		if selected {
			return true
		}
	}
	return false
}

// countSelectedDecks returns the number of uniquely selected deck files across all directories.
func (m Model) countSelectedDecks() int {
	count := 0
	for _, selected := range m.SelectedFiles {
		if selected {
			count++
		}
	}
	return count
}

// toggleDeck toggles the selection status of a deck file within the current directory.
func (m *Model) toggleDeck(file string) {
	relPath := file
	if m.SelectedDir != "" && !strings.HasPrefix(file, m.SelectedDir) {
		relPath = filepath.Join(m.SelectedDir, file)
	}
	if m.isDeckSelected(file) {
		delete(m.SelectedFiles, relPath)
		delete(m.SelectedFiles, file)
	} else {
		m.SelectedFiles[relPath] = true
	}
}

// SetupReview prepares cards scheduled for today or unreviewed cards.
func (m *Model) SetupReview() {
	today := time.Now().Format("2006-01-02")
	var due []deck.CardRef

	for file, d := range m.Decks {
		for i, card := range d.Cards {
			if card.NextReview == "" || card.NextReview <= today {
				due = append(due, deck.CardRef{Filename: file, OrigIdx: i})
			}
		}
	}

	rand.Shuffle(len(due), func(i, j int) { due[i], due[j] = due[j], due[i] })

	m.ActiveCards = due
	m.CurrentIndex = 0
	m.ShowAnswer = false
	m.ShowExplanation = false
}

// SetupQuiz prepares all cards in the selected decks for a quiz session.
func (m *Model) SetupQuiz() {
	var all []deck.CardRef
	for file, d := range m.Decks {
		for i := range d.Cards {
			all = append(all, deck.CardRef{Filename: file, OrigIdx: i})
		}
	}

	rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
	m.ActiveCards = all
	m.CurrentIndex = 0
	m.ShowFeedback = false
	m.IsCorrect = false
	m.GenerateQuizOptions()
}

// GenerateQuizOptions generates random multiple-choice distractors for the current question.
func (m *Model) GenerateQuizOptions() {
	if m.CurrentIndex >= len(m.ActiveCards) {
		return
	}

	ref := m.ActiveCards[m.CurrentIndex]
	targetCard := m.Decks[ref.Filename].Cards[ref.OrigIdx]

	var allCards []deck.Flashcard
	for _, d := range m.Decks {
		allCards = append(allCards, d.Cards...)
	}

	m.QuizOptions, m.CorrectIndex = quiz.GenerateOptions(targetCard, allCards, 5)
}
