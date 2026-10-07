// Package ui implements the terminal user interface for Corvidae using Bubble Tea and Lip Gloss.
package ui

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"flashcards/internal/arabic"
	"flashcards/internal/deck"
	"flashcards/internal/memorize"
	"flashcards/internal/quiz"
	"flashcards/internal/reading"
)

const (
	StateModeSelect = iota
	StateDirSelect
	StateDeckSelect
	StateReview
	StateQuiz
	StateArabicMenu
	StateArabicExplorer
	StateArabicDrill
	StateArabicStages
	StateArabicStageView
	StateMemorizeSelectText
	StateMemorizePortionSelect
	StateMemorize
	StateMemorizeComplete
	StateReadingSelectText
	StateReading
	StateReadingWordDetail
	StateReadingDiscussion
	StateReadingComplete
)

const (
	ModeReview = iota
	ModeQuiz
	ModeArabic
	ModeMemorize
	ModeReading
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

	// Arabic Academy state
	ArabicMenuCursor    int
	ArabicLetterCursor  int // 0 to 27 in arabic.Letters
	ArabicExplorerTab   int // 0: Letters, 1: Auxiliary & Harakat
	ArabicAuxCursor     int // index into arabic.AuxiliaryItems
	ArabicStageCursor   int // 0 to 7 in arabic.Stages
	ArabicDrillType     arabic.DrillType
	ArabicQuestions     []arabic.DrillQuestion
	ArabicQuestionIdx   int
	ArabicScore         int
	ArabicShowFeedback  bool
	ArabicIsCorrect     bool

	// Marked cards state
	MarkedCards       []deck.Flashcard
	PlayingMarkedDeck bool
	StatusMessage     string

	// Memorize by Options state
	TextsDir               string
	MemorizeTexts          []memorize.TextHeader
	MemorizeCursor         int
	MemorizeActiveText     *memorize.Text
	MemorizeSections       []memorize.Section
	PortionStartSection    int // 0-based index into MemorizeSections
	PortionEndSection      int // 0-based index into MemorizeSections (inclusive)
	PortionFocus           int // 0: Start, 1: End
	PortionStartWordIdx    int
	PortionEndWordIdx      int
	PortionCompletedMsg    string
	MemorizeCurrentIndex   int
	MemorizeCurrentOptions []string
	MemorizeCorrectIndex   int
	MemorizeSelectedOption int
	MemorizeMistakes       int
	MemorizeStreak         int
	MemorizeBestStreak     int
	MemorizeShowFeedback   bool
	MemorizeIsCorrect      bool

	// Learn by Reading state
	ReadingDir          string
	ReadingProgress     *reading.ReadingProgress
	ReadingTexts        []reading.TextHeader
	ReadingCursor       int
	ReadingSession      *reading.Session
	ReadingWordFeedback string

	Err    error
	Width  int
	Height int
}

// New initializes and returns a new Model using default directories.
func New(baseDir string) Model {
	return NewWithOptions(baseDir, "memorizationTexts", "readingTranslationTexts")
}

// NewWithOptions initializes and returns a new Model with specified deck, text, and reading directories.
func NewWithOptions(baseDir, textsDir string, readingDir ...string) Model {
	if baseDir == "" {
		baseDir = "decks"
	}
	if textsDir == "" {
		textsDir = "memorizationTexts"
	}
	rd := "readingTranslationTexts"
	if len(readingDir) > 0 && readingDir[0] != "" {
		rd = readingDir[0]
	}

	dirs, err := deck.GetDeckDirectories(baseDir)
	prog, _ := reading.LoadProgress(filepath.Join(rd, "reading_progress.yaml"))

	m := Model{
		BaseDir:         baseDir,
		TextsDir:        textsDir,
		ReadingDir:      rd,
		ReadingProgress: prog,
		State:           StateModeSelect,
		Mode:            ModeReview,
		Dirs:            dirs,
		SelectedFiles:   make(map[string]bool),
		Err:             err,
	}
	m.LoadMarkedDeck()
	m.LoadPersistedSelectedDecks()
	return m
}

// MarkWordAsFlashcard adds a vocabulary word from reading mode to the Marked Cards deck.
func (m *Model) MarkWordAsFlashcard(w reading.WordTranslation) bool {
	card := deck.Flashcard{
		Character:     w.Target,
		Pinyin:        w.Pinyin,
		Pronunciation: w.Pronunciation,
		Meaning:       w.Meaning,
		Explanation:   w.Explanation,
		Language:      w.Language,
		Interval:      1,
		Ease:          2.5,
		Reps:          0,
		NextReview:    time.Now().Format("2006-01-02"),
	}
	for _, c := range m.MarkedCards {
		if SameCard(c, card) {
			return false
		}
	}
	m.MarkedCards = append(m.MarkedCards, card)
	m.saveMarkedDeck()
	return true
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

// toggleDeck toggles the selection status of a deck file within the current directory and persists it.
func (m *Model) toggleDeck(file string) {
	relPath := file
	if m.SelectedDir != "" && !strings.HasPrefix(file, m.SelectedDir) {
		relPath = filepath.Join(m.SelectedDir, file)
	}
	wasSelected := m.isDeckSelected(file)
	newSelected := !wasSelected
	if wasSelected {
		delete(m.SelectedFiles, relPath)
		delete(m.SelectedFiles, file)
	} else {
		m.SelectedFiles[relPath] = true
	}
	m.saveDeckSelection(relPath, file, newSelected)
}

// LoadPersistedSelectedDecks scans deck files in BaseDir and marks any with selected: true in SelectedFiles.
func (m *Model) LoadPersistedSelectedDecks() {
	if m.SelectedFiles == nil {
		m.SelectedFiles = make(map[string]bool)
	}
	if m.BaseDir == "" {
		return
	}
	allFiles, err := deck.GetAllDeckFiles(m.BaseDir, "")
	if err != nil {
		return
	}
	for _, rel := range allFiles {
		if filepath.Base(rel) == "marked.yaml" {
			continue
		}
		fullPath := filepath.Join(m.BaseDir, rel)
		d, err := deck.LoadDeck(fullPath)
		if err != nil {
			continue
		}
		if d.IsSelected() {
			m.SelectedFiles[rel] = true
		}
	}
}

func (m Model) resolveDeckPath(relPath, file string) string {
	candidates := []string{
		relPath,
		filepath.Join(m.BaseDir, relPath),
		file,
		filepath.Join(m.BaseDir, file),
	}
	if m.SelectedDir != "" {
		candidates = append(candidates, filepath.Join(m.BaseDir, m.SelectedDir, file))
		candidates = append(candidates, filepath.Join(m.BaseDir, m.SelectedDir, relPath))
	}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func (m *Model) saveDeckSelection(relPath, file string, selected bool) {
	fullPath := m.resolveDeckPath(relPath, file)
	if fullPath == "" {
		return
	}
	d, err := deck.LoadDeck(fullPath)
	if err != nil {
		return
	}
	d.Selected = selected
	d.QuizSelected = false
	_ = deck.SaveDeck(fullPath, d)

	if m.Decks != nil {
		if cur, ok := m.Decks[fullPath]; ok {
			cur.Selected = selected
			cur.QuizSelected = false
			m.Decks[fullPath] = cur
		}
	}
}

// DeselectAll unselects all currently selected decks across all folders and updates the persisted YAML files.
func (m *Model) DeselectAll() {
	for key, isSelected := range m.SelectedFiles {
		if !isSelected {
			continue
		}
		fullPath := m.resolveDeckPath(key, key)
		if fullPath != "" {
			d, err := deck.LoadDeck(fullPath)
			if err == nil && d.IsSelected() {
				d.Selected = false
				d.QuizSelected = false
				_ = deck.SaveDeck(fullPath, d)
			}
			if m.Decks != nil {
				if cur, ok := m.Decks[fullPath]; ok {
					cur.Selected = false
					cur.QuizSelected = false
					m.Decks[fullPath] = cur
				}
			}
		}
	}
	m.SelectedFiles = make(map[string]bool)
	m.StatusMessage = "All decks deselected"
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

// GenerateQuizOptions generates random multiple-choice distractors for the current question,
// ensuring distractors are primarily pulled from the same marked cards or the same language deck.
func (m *Model) GenerateQuizOptions() {
	if m.CurrentIndex >= len(m.ActiveCards) {
		return
	}

	ref := m.ActiveCards[m.CurrentIndex]
	targetCard := m.Decks[ref.Filename].Cards[ref.OrigIdx]
	targetLang := m.DetectCardLanguage(targetCard)

	// 1. Gather all candidate cards from currently loaded decks that match targetLang
	var candidateCards []deck.Flashcard
	for _, d := range m.Decks {
		for _, c := range d.Cards {
			if targetLang == "" || m.DetectCardLanguage(c) == targetLang {
				candidateCards = append(candidateCards, c)
			}
		}
	}

	// 2. If fewer than 6 candidate cards are in loaded decks:
	if len(candidateCards) < 6 {
		if targetLang != "" {
			// Supplement strictly from the same language deck(s)
			preferredDeck := targetCard.Deck
			if preferredDeck == "" && ref.Filename != m.getMarkedDeckPath() {
				if rel, err := filepath.Rel(m.BaseDir, ref.Filename); err == nil {
					preferredDeck = filepath.ToSlash(rel)
				}
			}
			extraCards := m.getLanguageDistractorCards(targetLang, preferredDeck, 6-len(candidateCards))
			candidateCards = append(candidateCards, extraCards...)
		} else {
			// Language unknown: supplement from any available deck
			extraCards := m.getExtraDistractorCards(6 - len(candidateCards))
			candidateCards = append(candidateCards, extraCards...)
		}
	}

	m.QuizOptions, m.CorrectIndex = quiz.GenerateOptions(targetCard, candidateCards, 5)
}

// SetupArabicDrill prepares question sets for the chosen Arabic drill type.
func (m *Model) SetupArabicDrill(dType arabic.DrillType, stageIdx int) {
	m.ArabicDrillType = dType
	m.ArabicQuestionIdx = 0
	m.ArabicScore = 0
	m.ArabicShowFeedback = false
	m.ArabicIsCorrect = false

	switch dType {
	case arabic.DrillPositionalForms:
		m.ArabicQuestions = arabic.GeneratePositionalFormQuestions(10)
	case arabic.DrillLetterSounds:
		m.ArabicQuestions = arabic.GenerateLetterSoundQuestions(10)
	case arabic.DrillConnectors:
		m.ArabicQuestions = arabic.GenerateConnectorQuestions(10)
	case arabic.DrillSunMoon:
		m.ArabicQuestions = arabic.GenerateSunMoonQuestions(10)
	case arabic.DrillConfusables:
		m.ArabicQuestions = arabic.GenerateConfusableQuestions(10)
	case arabic.DrillStage:
		if stageIdx >= 0 && stageIdx < len(arabic.Stages) {
			m.ArabicQuestions = arabic.GenerateStageQuestions(arabic.Stages[stageIdx], 8)
		}
	}
}

// SetupMemorizeWord configures options for the active word in Memorize mode.
func (m *Model) SetupMemorizeWord() {
	if m.MemorizeActiveText == nil || m.MemorizeCurrentIndex >= len(m.MemorizeActiveText.Words) {
		m.State = StateMemorizeComplete
		return
	}
	item := m.MemorizeActiveText.Words[m.MemorizeCurrentIndex]
	m.MemorizeCurrentOptions, m.MemorizeCorrectIndex = memorize.ShuffleOptions(item)
	m.MemorizeShowFeedback = false
	m.MemorizeSelectedOption = -1
	m.MemorizeIsCorrect = false
}

// SelectMemorizeText loads the chosen text, extracts sections, and transitions to portion selection.
func (m *Model) SelectMemorizeText(filePath string) error {
	loaded, err := memorize.LoadText(filePath)
	if err != nil {
		return err
	}
	m.MemorizeActiveText = loaded
	m.MemorizeSections = loaded.GetSections()
	m.PortionStartSection = 0
	if len(m.MemorizeSections) > 0 {
		m.PortionEndSection = len(m.MemorizeSections) - 1
	} else {
		m.PortionEndSection = 0
	}
	m.PortionFocus = 0
	m.PortionCompletedMsg = ""
	m.State = StateMemorizePortionSelect
	return nil
}

// StartMemorizePortion begins practicing the continuous portion between startSec and endSec.
func (m *Model) StartMemorizePortion(startSec, endSec int) {
	if len(m.MemorizeSections) == 0 {
		return
	}
	if startSec < 0 {
		startSec = 0
	}
	if endSec >= len(m.MemorizeSections) {
		endSec = len(m.MemorizeSections) - 1
	}
	if startSec > endSec {
		startSec = endSec
	}

	m.PortionStartSection = startSec
	m.PortionEndSection = endSec
	m.PortionStartWordIdx = m.MemorizeSections[startSec].StartIdx
	m.PortionEndWordIdx = m.MemorizeSections[endSec].EndIdx

	m.MemorizeCurrentIndex = m.PortionStartWordIdx
	m.MemorizeMistakes = 0
	m.MemorizeStreak = 0
	m.MemorizeBestStreak = 0
	m.PortionCompletedMsg = ""
	m.State = StateMemorize
	m.SetupMemorizeWord()
}

// StartMemorizeSession loads the chosen text and starts the full session.
func (m *Model) StartMemorizeSession(filePath string) error {
	if err := m.SelectMemorizeText(filePath); err != nil {
		return err
	}
	m.StartMemorizePortion(0, len(m.MemorizeSections)-1)
	return nil
}
