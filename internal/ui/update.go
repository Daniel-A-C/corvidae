package ui

import (
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"flashcards/internal/deck"
	"flashcards/internal/sm2"
)

// Update handles incoming terminal messages and user input events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "q" && !m.isSelectionKey("q") {
			return m, tea.Quit
		}

		switch m.State {
		case StateModeSelect:
			return m.updateModeSelect(msg)
		case StateDirSelect:
			return m.updateDirSelect(msg)
		case StateDeckSelect:
			return m.updateDeckSelect(msg)
		case StateReview:
			return m.updateReview(msg)
		case StateQuiz:
			return m.updateQuiz(msg)
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, nil
}

func (m Model) isSelectionKey(key string) bool {
	idx := KeyToIndex(key)
	if idx == -1 {
		return false
	}
	switch m.State {
	case StateModeSelect:
		return idx < 2
	case StateDirSelect:
		return idx < len(m.Dirs)
	case StateDeckSelect:
		return idx < len(m.DeckFiles)
	case StateQuiz:
		return !m.ShowFeedback && idx < len(m.QuizOptions)
	}
	return false
}

func (m Model) updateModeSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "a":
		m.Mode = ModeReview
		return m.confirmModeSelect()
	case "s":
		m.Mode = ModeQuiz
		return m.confirmModeSelect()
	case "up", "k":
		m.Mode = ModeReview
	case "down", "j":
		m.Mode = ModeQuiz
	case "enter", " ":
		return m.confirmModeSelect()
	}
	return m, nil
}

func (m Model) confirmModeSelect() (tea.Model, tea.Cmd) {
	dirs, err := deck.GetDeckDirectories(m.BaseDir)
	if err != nil {
		m.Err = err
		return m, nil
	}
	m.Dirs = dirs
	m.DirCursor = 0
	m.State = StateDirSelect
	return m, nil
}

func (m Model) updateDirSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	idx := KeyToIndex(msg.String())
	if idx >= 0 && idx < len(m.Dirs) {
		return m.selectDirectory(idx)
	}

	switch msg.String() {
	case "up":
		if m.DirCursor > 0 {
			m.DirCursor--
		}
	case "down":
		if m.DirCursor < len(m.Dirs)-1 {
			m.DirCursor++
		}
	case "k":
		if m.DirCursor > 0 {
			m.DirCursor--
		}
	case "j":
		if m.DirCursor < len(m.Dirs)-1 {
			m.DirCursor++
		}
	case "esc", "b":
		m.State = StateModeSelect
	case "enter", " ":
		if len(m.Dirs) > 0 {
			return m.selectDirectory(m.DirCursor)
		}
	}
	return m, nil
}

func (m Model) selectDirectory(idx int) (tea.Model, tea.Cmd) {
	m.DirCursor = idx
	m.SelectedDir = m.Dirs[idx]
	files, err := deck.GetDeckFiles(m.BaseDir, m.SelectedDir)
	if err != nil {
		m.Err = err
		return m, nil
	}
	m.DeckFiles = files
	m.Cursor = 0
	m.SelectedFiles = make(map[string]bool)
	m.State = StateDeckSelect
	return m, nil
}

func (m Model) updateDeckSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	idx := KeyToIndex(msg.String())
	if idx >= 0 && idx < len(m.DeckFiles) {
		m.Cursor = idx
		file := m.DeckFiles[idx]
		m.SelectedFiles[file] = !m.SelectedFiles[file]
		return m, nil
	}

	switch msg.String() {
	case "up":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down":
		if m.Cursor < len(m.DeckFiles)-1 {
			m.Cursor++
		}
	case "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "j":
		if m.Cursor < len(m.DeckFiles)-1 {
			m.Cursor++
		}
	case "esc", "b":
		m.State = StateDirSelect
	case " ":
		if len(m.DeckFiles) > 0 {
			file := m.DeckFiles[m.Cursor]
			m.SelectedFiles[file] = !m.SelectedFiles[file]
		}
	case "enter":
		if len(m.DeckFiles) == 0 {
			m.State = StateDirSelect
			return m, nil
		}

		hasSelection := false
		for _, selected := range m.SelectedFiles {
			if selected {
				hasSelection = true
				break
			}
		}

		if hasSelection {
			err := m.LoadSelectedDecks()
			if err != nil {
				m.Err = err
				return m, nil
			}
			if m.Mode == ModeReview {
				m.SetupReview()
				m.State = StateReview
			} else {
				m.SetupQuiz()
				m.State = StateQuiz
			}
		}
	}
	return m, nil
}

func (m Model) updateReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.ActiveCards) == 0 || m.CurrentIndex >= len(m.ActiveCards) {
		if msg.String() == "r" {
			var all []deck.CardRef
			for file, d := range m.Decks {
				for i := range d.Cards {
					all = append(all, deck.CardRef{Filename: file, OrigIdx: i})
				}
			}
			rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
			m.ActiveCards = all
			m.CurrentIndex = 0
			m.ShowAnswer = false
			m.ShowExplanation = false
		} else if msg.String() == "enter" {
			m.State = StateModeSelect
		}
		return m, nil
	}

	if !m.ShowAnswer {
		if msg.String() == " " {
			m.ShowAnswer = true
		}
		return m, nil
	}

	if msg.String() == "e" {
		m.ShowExplanation = !m.ShowExplanation
		return m, nil
	}

	grade := -1
	switch msg.String() {
	case "d":
		grade = sm2.GradeBlackout
	case "f":
		grade = sm2.GradeWrong
	case "g":
		grade = sm2.GradeHard
	case "h":
		grade = sm2.GradeGood
	case "j":
		grade = sm2.GradeEasy
	case "k":
		grade = sm2.GradePerfect
	}

	if grade != -1 {
		ref := m.ActiveCards[m.CurrentIndex]
		d := m.Decks[ref.Filename]

		d.Cards[ref.OrigIdx] = sm2.CalculateSM2(d.Cards[ref.OrigIdx], grade, time.Now())
		m.Decks[ref.Filename] = d

		if err := deck.SaveDeck(ref.Filename, d); err != nil {
			m.Err = err
			return m, nil
		}

		m.CurrentIndex++
		m.ShowAnswer = false
		m.ShowExplanation = false
	}

	return m, nil
}

func (m Model) updateQuiz(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.CurrentIndex >= len(m.ActiveCards) {
		if msg.String() == "enter" {
			m.State = StateModeSelect
		}
		return m, nil
	}

	if m.ShowFeedback {
		if msg.String() == " " || msg.String() == "enter" {
			m.ShowFeedback = false
			m.CurrentIndex++
			m.GenerateQuizOptions()
		}
		return m, nil
	}

	choiceIdx := KeyToIndex(msg.String())
	if choiceIdx != -1 && choiceIdx < len(m.QuizOptions) {
		m.ShowFeedback = true
		m.IsCorrect = (choiceIdx == m.CorrectIndex)
	}

	return m, nil
}
