package ui

import (
	"math/rand"
	"path/filepath"
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

// LoadSelectedDecks loads all currently marked decks into memory.
func (m *Model) LoadSelectedDecks() error {
	m.Decks = make(map[string]deck.Deck)
	for file, isSelected := range m.SelectedFiles {
		if isSelected {
			fullPath := filepath.Join(m.BaseDir, m.SelectedDir, file)
			loaded, err := deck.LoadDeck(fullPath)
			if err != nil {
				return err
			}
			m.Decks[fullPath] = loaded
		}
	}
	return nil
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
