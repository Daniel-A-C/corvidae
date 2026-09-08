package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// --- Structs ---

type Flashcard struct {
	Character   string  `yaml:"character"`
	Pinyin      string  `yaml:"pinyin"`
	Meaning     string  `yaml:"meaning"`
	Explanation string  `yaml:"explanation,omitempty"`
	Interval    int     `yaml:"interval,omitempty"`
	Ease        float64 `yaml:"ease,omitempty"`
	Reps        int     `yaml:"reps,omitempty"`
	NextReview  string  `yaml:"next_review,omitempty"`
}

type Deck struct {
	Cards []Flashcard `yaml:"cards"`
}

// Used to map a reviewed card back to its specific source file
type cardRef struct {
	filename string
	origIdx  int
}

// --- Styles ---

var (
	charStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7CCB"))
	pinyinStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#569CD6"))
	meaningStyle     = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#ebebeb"))
	explanationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")).Italic(true).Width(60).Align(lipgloss.Center)
	hintStyle        = lipgloss.NewStyle().Faint(false)
	keyStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4EC9B0"))
	cursorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7CCB")).Bold(true)
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)
	correctStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)
	logoStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9")).Width(60).Align(lipgloss.Left)
)

const asciiLogo = `                           .-.
                          ( o>
                          / ) \
                         ='---'=
                           m m
   ____                          _       _                
  / ___|   ___    _ __  __   __ (_)   __| |   __ _    ___ 
 | |      / _ \  | '__| \ \ / / | |  / _' |  / _' |  / _ \
 | |___  | (_) | | |     \ V /  | | | (_| | | (_| | |  __/
  \____|  \___/  |_|      \_/   |_|  \__,_|  \__,_|  \___|`


// --- App States ---

const (
	stateModeSelect = iota
	stateDirSelect
	stateDeckSelect
	stateReview
	stateQuiz
)

const (
	modeReview = iota
	modeQuiz
)

type model struct {
	state           int
	mode            int
	dirs            []string
	dirCursor       int
	selectedDir     string
	deckFiles       []string
	cursor          int
	selectedFiles   map[string]bool

	decks           map[string]Deck
	activeCards     []cardRef
	currentIndex    int
	showAnswer      bool
	showExplanation bool

	// Quiz state variables
	quizOptions  []string
	correctIndex int
	showFeedback bool

	err    error
	width  int
	height int
}

// --- Helpers ---

func getDeckDirectories() ([]string, error) {
	entries, err := os.ReadDir("decks")
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			dirs = append(dirs, entry.Name())
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

func getDeckFiles(dir string) ([]string, error) {
	dirPath := filepath.Join("decks", dir)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasPrefix(entry.Name(), ".") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func countDecksInDir(dir string) int {
	files, err := getDeckFiles(dir)
	if err != nil {
		return 0
	}
	return len(files)
}

func loadDeck(filename string) (Deck, error) {
	var deck Deck
	data, err := os.ReadFile(filename)
	if err != nil {
		return deck, err
	}
	err = yaml.Unmarshal(data, &deck)
	return deck, err
}

func saveDeck(filename string, d Deck) {
	data, err := yaml.Marshal(d)
	if err == nil {
		os.WriteFile(filename, data, 0644)
	}
}

func calculateSM2(card Flashcard, grade int) Flashcard {
	if card.Ease == 0 {
		card.Ease = 2.5
	}
	if grade >= 3 {
		if card.Reps == 0 {
			card.Interval = 1
		} else if card.Reps == 1 {
			card.Interval = 6
		} else {
			card.Interval = int(math.Round(float64(card.Interval) * card.Ease))
		}
		card.Reps++
	} else {
		card.Reps = 0
		card.Interval = 1
	}

	card.Ease = card.Ease + (0.1 - float64(5-grade)*(0.08+float64(5-grade)*0.02))
	if card.Ease < 1.3 {
		card.Ease = 1.3
	}

	next := time.Now().AddDate(0, 0, card.Interval)
	card.NextReview = next.Format("2006-01-02")
	return card
}

// --- Bubble Tea Interface ---

func initialModel() model {
	dirs, err := getDeckDirectories()
	return model{
		state:         stateModeSelect,
		mode:          modeReview,
		dirs:          dirs,
		selectedFiles: make(map[string]bool),
		err:           err,
	}
}

func (m *model) loadSelectedDecks() error {
	m.decks = make(map[string]Deck)
	for file, isSelected := range m.selectedFiles {
		if isSelected {
			fullPath := filepath.Join("decks", m.selectedDir, file)
			deck, err := loadDeck(fullPath)
			if err != nil {
				return err
			}
			m.decks[fullPath] = deck
		}
	}
	return nil
}

func (m *model) setupReview() {
	today := time.Now().Format("2006-01-02")
	var due []cardRef

	for file, deck := range m.decks {
		for i, card := range deck.Cards {
			if card.NextReview == "" || card.NextReview <= today {
				due = append(due, cardRef{filename: file, origIdx: i})
			}
		}
	}

	rand.Shuffle(len(due), func(i, j int) { due[i], due[j] = due[j], due[i] })

	m.activeCards = due
	m.currentIndex = 0
	m.showAnswer = false
	m.showExplanation = false
}

func (m *model) setupQuiz() {
	var all []cardRef
	for file, deck := range m.decks {
		for i := range deck.Cards {
			all = append(all, cardRef{filename: file, origIdx: i})
		}
	}

	rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
	m.activeCards = all
	m.currentIndex = 0
	m.showFeedback = false
	m.generateQuizOptions()
}

func (m *model) generateQuizOptions() {
	if m.currentIndex >= len(m.activeCards) {
		return
	}

	ref := m.activeCards[m.currentIndex]
	correctCard := m.decks[ref.filename].Cards[ref.origIdx]
	
	// Create combined Pinyin and Meaning strings for the options
	correctOption := fmt.Sprintf("%s - %s", correctCard.Pinyin, correctCard.Meaning)

	var allOptions []string
	for _, deck := range m.decks {
		for _, card := range deck.Cards {
			allOptions = append(allOptions, fmt.Sprintf("%s - %s", card.Pinyin, card.Meaning))
		}
	}

	uniquePool := make(map[string]bool)
	var pool []string
	for _, opt := range allOptions {
		if opt != correctOption && !uniquePool[opt] {
			uniquePool[opt] = true
			pool = append(pool, opt)
		}
	}

	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

	numOptions := 5
	if len(pool) < 5 {
		numOptions = len(pool)
	}

	options := []string{correctOption}
	options = append(options, pool[:numOptions]...)

	rand.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })

	m.quizOptions = options
	for i, opt := range options {
		if opt == correctOption {
			m.correctIndex = i
			break
		}
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

		switch m.state {
		case stateModeSelect:
			switch msg.String() {
			case "up", "k":
				m.mode = modeReview
			case "down", "j":
				m.mode = modeQuiz
			case "enter", " ":
				dirs, err := getDeckDirectories()
				if err != nil {
					m.err = err
					return m, nil
				}
				m.dirs = dirs
				m.dirCursor = 0
				m.state = stateDirSelect
			}

		case stateDirSelect:
			switch msg.String() {
			case "up", "k":
				if m.dirCursor > 0 {
					m.dirCursor--
				}
			case "down", "j":
				if m.dirCursor < len(m.dirs)-1 {
					m.dirCursor++
				}
			case "esc", "b":
				m.state = stateModeSelect
			case "enter", " ":
				if len(m.dirs) > 0 {
					m.selectedDir = m.dirs[m.dirCursor]
					files, err := getDeckFiles(m.selectedDir)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.deckFiles = files
					m.cursor = 0
					m.selectedFiles = make(map[string]bool)
					m.state = stateDeckSelect
				}
			}

		case stateDeckSelect:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.deckFiles)-1 {
					m.cursor++
				}
			case "esc", "b":
				m.state = stateDirSelect
			case " ":
				if len(m.deckFiles) > 0 {
					file := m.deckFiles[m.cursor]
					m.selectedFiles[file] = !m.selectedFiles[file]
				}
			case "enter":
				if len(m.deckFiles) == 0 {
					m.state = stateDirSelect
					return m, nil
				}

				hasSelection := false
				for _, selected := range m.selectedFiles {
					if selected {
						hasSelection = true
						break
					}
				}

				if hasSelection {
					err := m.loadSelectedDecks()
					if err != nil {
						m.err = err
						return m, nil
					}
					if m.mode == modeReview {
						m.setupReview()
						m.state = stateReview
					} else {
						m.setupQuiz()
						m.state = stateQuiz
					}
				}
			}

		case stateReview:
			if len(m.activeCards) == 0 || m.currentIndex >= len(m.activeCards) {
				if msg.String() == "r" {
					var all []cardRef
					for file, deck := range m.decks {
						for i := range deck.Cards {
							all = append(all, cardRef{filename: file, origIdx: i})
						}
					}
					rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
					m.activeCards = all
					m.currentIndex = 0
					m.showAnswer = false
					m.showExplanation = false
				} else if msg.String() == "enter" {
					// Return to mode select
					m.state = stateModeSelect
				}
				return m, nil
			}

			if !m.showAnswer {
				if msg.String() == " " {
					m.showAnswer = true
				}
			} else {
				if msg.String() == "e" {
					m.showExplanation = !m.showExplanation
					return m, nil
				}

				grade := -1
				switch msg.String() {
				case "d": grade = 0
				case "f": grade = 1
				case "g": grade = 2
				case "h": grade = 3
				case "j": grade = 4
				case "k": grade = 5
				}

				if grade != -1 {
					ref := m.activeCards[m.currentIndex]
					deck := m.decks[ref.filename]
					
					deck.Cards[ref.origIdx] = calculateSM2(deck.Cards[ref.origIdx], grade)
					m.decks[ref.filename] = deck 
					
					saveDeck(ref.filename, deck)

					m.currentIndex++
					m.showAnswer = false
					m.showExplanation = false
				}
			}

		case stateQuiz:
			if m.currentIndex >= len(m.activeCards) {
				if msg.String() == "enter" {
					m.state = stateModeSelect
				}
				return m, nil
			}

			if m.showFeedback {
				if msg.String() == " " || msg.String() == "enter" {
					m.showFeedback = false
					m.currentIndex++
					m.generateQuizOptions()
				}
				return m, nil
			}

			input := msg.String()
			choiceIdx := -1
			
			// Map home row keys to options 0-5
			switch input {
			case "d": choiceIdx = 0
			case "f": choiceIdx = 1
			case "g": choiceIdx = 2
			case "h": choiceIdx = 3
			case "j": choiceIdx = 4
			case "k": choiceIdx = 5
			}
			
			if choiceIdx != -1 && choiceIdx < len(m.quizOptions) {
				if choiceIdx == m.correctIndex {
					m.currentIndex++
					m.generateQuizOptions()
				} else {
					m.showFeedback = true
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	var content string

	switch m.state {
	case stateModeSelect:
		content = logoStyle.Render(asciiLogo) + "\n\n"
		content += "Select Practice Mode:\n\n"
		
		modes := []string{"Spaced Repetition (SM-2)", "Multiple Choice Quiz"}
		for i, label := range modes {
			cursor := "  "
			if m.mode == i {
				cursor = "> "
				label = cursorStyle.Render(label)
			}
			content += fmt.Sprintf("%s%s\n", cursorStyle.Render(cursor), label)
		}
		content += "\n" + hintStyle.Render("(Use j/k to move, Enter to select)")

	case stateDirSelect:
		if len(m.dirs) == 0 {
			content = "No language directories found in decks/.\n\n" + hintStyle.Render("(Press Esc to return, q to quit)")
		} else {
			content = "Select Deck Directory / Language:\n\n"
			for i, dir := range m.dirs {
				cursor := "  "
				label := dir
				count := countDecksInDir(dir)
				var countStr string
				if count == 1 {
					countStr = hintStyle.Render(" (1 deck)")
				} else {
					countStr = hintStyle.Render(fmt.Sprintf(" (%d decks)", count))
				}

				if m.dirCursor == i {
					cursor = "> "
					label = cursorStyle.Render(label)
				}
				content += fmt.Sprintf("%s%s%s\n", cursorStyle.Render(cursor), label, countStr)
			}
			content += "\n" + hintStyle.Render("(Use j/k to move, Enter to select, Esc to go back, q to quit)")
		}

	case stateDeckSelect:
		if len(m.deckFiles) == 0 {
			content = fmt.Sprintf("No .yaml files found in decks/%s.\n\n%s", m.selectedDir, hintStyle.Render("(Press Esc or Enter to go back, q to quit)"))
		} else {
			content = fmt.Sprintf("Select decks to practice (%s):\n\n", m.selectedDir)
			for i, file := range m.deckFiles {
				cursor := "  "
				if m.cursor == i {
					cursor = "> "
				}
				
				check := "[ ]"
				if m.selectedFiles[file] {
					check = "[x]"
				}

				line := fmt.Sprintf("%s %s %s", cursorStyle.Render(cursor), keyStyle.Render(check), file)
				if m.cursor == i {
					line = cursorStyle.Render(fmt.Sprintf("%s %s %s", cursor, check, file))
				}
				content += line + "\n"
			}
			content += "\n" + hintStyle.Render("(Space to toggle, Enter to confirm, Esc to go back, q to quit)")
		}

	case stateReview:
		if len(m.activeCards) == 0 {
			content = fmt.Sprintf("No cards due today!\n\n%s", hintStyle.Render("[r] Review all cards anyway  •  [Enter] Main Menu  •  [q] Quit"))
		} else if m.currentIndex >= len(m.activeCards) {
			content = fmt.Sprintf("Daily review complete! Great job.\n\n%s", hintStyle.Render("[r] Review all cards again  •  [Enter] Main Menu  •  [q] Quit"))
		} else {
			ref := m.activeCards[m.currentIndex]
			card := m.decks[ref.filename].Cards[ref.origIdx]
			displayName := strings.TrimPrefix(ref.filename, "decks/")

			content = hintStyle.Render(fmt.Sprintf("Card %d of %d  •  %s", m.currentIndex+1, len(m.activeCards), displayName)) + "\n\n"
			content += charStyle.Render(card.Character) + "\n\n"

			if !m.showAnswer {
				content += hintStyle.Render("[ Spacebar to reveal ]") + "\n"
				content += "\n" + hintStyle.Render("(Press 'q' to quit)")
			} else {
				content += fmt.Sprintf("Pinyin:  %s\n", pinyinStyle.Render(card.Pinyin))
				content += fmt.Sprintf("Meaning: %s\n\n", meaningStyle.Render(card.Meaning))

				if m.showExplanation && card.Explanation != "" {
					content += explanationStyle.Render(card.Explanation) + "\n\n"
				}

				content += "How well did you know this?\n"
				content += fmt.Sprintf("[%s] Blackout  [%s] Wrong  [%s] Hard\n",
					keyStyle.Render("d"), keyStyle.Render("f"), keyStyle.Render("g"))
				content += fmt.Sprintf("[%s] Good      [%s] Easy   [%s] Perfect\n",
					keyStyle.Render("h"), keyStyle.Render("j"), keyStyle.Render("k"))

				content += "\n" + hintStyle.Render("[e] Toggle Explanation  •  [q] Quit")
			}
		}
		
	case stateQuiz:
		if m.currentIndex >= len(m.activeCards) {
			content = fmt.Sprintf("Quiz complete!\n\n%s", hintStyle.Render("[Enter] Main Menu  •  [q] Quit"))
		} else {
			ref := m.activeCards[m.currentIndex]
			card := m.decks[ref.filename].Cards[ref.origIdx]
			displayName := strings.TrimPrefix(ref.filename, "decks/")

			content = hintStyle.Render(fmt.Sprintf("Quiz: Question %d of %d  •  %s", m.currentIndex+1, len(m.activeCards), displayName)) + "\n\n"
			content += charStyle.Render(card.Character) + "\n\n"
			
			if m.showFeedback {
				content += errorStyle.Render("Incorrect!") + "\n"
				content += fmt.Sprintf("The correct answer was: %s\n\n", correctStyle.Render(m.quizOptions[m.correctIndex]))
				content += hintStyle.Render("[ Spacebar to continue ]")
			} else {
				quizKeys := []string{"d", "f", "g", "h", "j", "k"}
				for i, opt := range m.quizOptions {
					content += fmt.Sprintf("[%s] %s\n", keyStyle.Render(quizKeys[i]), opt)
				}
				content += "\n" + hintStyle.Render("(Press d, f, g, h, j, k to select, q to quit)")
			}
		}
	}

	styledContent := lipgloss.NewStyle().Align(lipgloss.Center).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, styledContent)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

