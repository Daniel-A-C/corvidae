package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
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

// --- Styles ---

var (
	charStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7CCB"))
	pinyinStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#569CD6"))
	meaningStyle     = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#D4D4D4"))
	explanationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")).Italic(true).Width(60).Align(lipgloss.Center)
	hintStyle        = lipgloss.NewStyle().Faint(true)
	keyStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4EC9B0"))
	cursorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7CCB")).Bold(true)
)

// --- App States ---

const (
	stateDeckSelect = iota
	stateReview
)

type model struct {
	state           int
	deckFiles       []string
	cursor          int
	selectedFile    string

	deck            Deck
	dueIndices      []int
	currentIndex    int
	showAnswer      bool
	showExplanation bool

	err    error
	width  int
	height int
}

// --- Helpers ---

func getDeckFiles() ([]string, error) {
	var files []string
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			files = append(files, entry.Name())
		}
	}
	return files, nil
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
	files, err := getDeckFiles()
	return model{
		state:     stateDeckSelect,
		deckFiles: files,
		err:       err,
	}
}

func (m *model) setupReview() {
	today := time.Now().Format("2006-01-02")
	var due []int

	for i, card := range m.deck.Cards {
		if card.NextReview == "" || card.NextReview <= today {
			due = append(due, i)
		}
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(due), func(i, j int) {
		due[i], due[j] = due[j], due[i]
	})

	m.dueIndices = due
	m.currentIndex = 0
	m.showAnswer = false
	m.showExplanation = false
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

		if m.state == stateDeckSelect {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.deckFiles)-1 {
					m.cursor++
				}
			case "enter", " ":
				if len(m.deckFiles) > 0 {
					m.selectedFile = m.deckFiles[m.cursor]
					deck, err := loadDeck(m.selectedFile)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.deck = deck
					m.setupReview()
					m.state = stateReview
				}
			}
		} else if m.state == stateReview {
			// Catch state where no cards are due, or deck is finished
			if len(m.dueIndices) == 0 || m.currentIndex >= len(m.dueIndices) {
				if msg.String() == "r" {
					// Load every card in the deck for a force review
					var all []int
					for i := range m.deck.Cards {
						all = append(all, i)
					}
					rand.Seed(time.Now().UnixNano())
					rand.Shuffle(len(all), func(i, j int) {
						all[i], all[j] = all[j], all[i]
					})
					m.dueIndices = all
					m.currentIndex = 0
					m.showAnswer = false
					m.showExplanation = false
				}
				return m, nil
			}

			// Normal review logic
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
					realIndex := m.dueIndices[m.currentIndex]
					m.deck.Cards[realIndex] = calculateSM2(m.deck.Cards[realIndex], grade)
					
					saveDeck(m.selectedFile, m.deck)

					m.currentIndex++
					m.showAnswer = false
					m.showExplanation = false
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

	if m.state == stateDeckSelect {
		if len(m.deckFiles) == 0 {
			content = "No .yaml files found in this directory.\n(Press 'q' to quit)"
		} else {
			content = "Select a deck to practice:\n\n"
			for i, file := range m.deckFiles {
				cursor := "  "
				fileName := file
				if m.cursor == i {
					cursor = "> "
					fileName = cursorStyle.Render(file)
				}
				content += fmt.Sprintf("%s%s\n", cursorStyle.Render(cursor), fileName)
			}
			content += "\n" + hintStyle.Render("(Use j/k to move, Enter to select)")
		}
	} else if m.state == stateReview {
		if len(m.dueIndices) == 0 {
			content = fmt.Sprintf("No cards due in %s today!\n\n%s", m.selectedFile, hintStyle.Render("[r] Review all cards anyway  •  [q] Quit"))
		} else if m.currentIndex >= len(m.dueIndices) {
			content = fmt.Sprintf("Daily review complete! Great job.\n\n%s", hintStyle.Render("[r] Review all cards again  •  [q] Quit"))
		} else {
			realIndex := m.dueIndices[m.currentIndex]
			card := m.deck.Cards[realIndex]
			
			content = hintStyle.Render(fmt.Sprintf("Card %d of %d  •  %s", m.currentIndex+1, len(m.dueIndices), m.selectedFile)) + "\n\n"
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
