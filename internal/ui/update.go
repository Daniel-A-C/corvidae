package ui

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"flashcards/internal/arabic"
	"flashcards/internal/deck"
	"flashcards/internal/memorize"
	"flashcards/internal/reading"
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
		case StateArabicMenu:
			return m.updateArabicMenu(msg)
		case StateArabicExplorer:
			return m.updateArabicExplorer(msg)
		case StateArabicDrill:
			return m.updateArabicDrill(msg)
		case StateArabicStages:
			return m.updateArabicStages(msg)
		case StateArabicStageView:
			return m.updateArabicStageView(msg)
		case StateMemorizeSelectText:
			return m.updateMemorizeSelectText(msg)
		case StateMemorizePortionSelect:
			return m.updateMemorizePortionSelect(msg)
		case StateMemorize:
			return m.updateMemorize(msg)
		case StateMemorizeComplete:
			return m.updateMemorizeComplete(msg)
		case StateReadingSelectText:
			return m.updateReadingSelectText(msg)
		case StateReading:
			return m.updateReading(msg)
		case StateReadingWordDetail:
			return m.updateReadingWordDetail(msg)
		case StateReadingDiscussion:
			return m.updateReadingDiscussion(msg)
		case StateReadingComplete:
			return m.updateReadingComplete(msg)
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
		return idx < 5
	case StateDirSelect:
		return idx < len(m.Dirs)
	case StateDeckSelect:
		return idx < len(m.DeckFiles)
	case StateQuiz:
		return !m.ShowFeedback && idx < len(m.QuizOptions)
	case StateArabicMenu:
		return idx < 7
	case StateArabicDrill:
		if m.ArabicQuestionIdx < len(m.ArabicQuestions) {
			return !m.ArabicShowFeedback && idx < len(m.ArabicQuestions[m.ArabicQuestionIdx].Options)
		}
	case StateArabicStages:
		return idx < len(arabic.Stages)
	case StateMemorizeSelectText:
		return idx < len(m.MemorizeTexts)
	case StateMemorize:
		return !m.MemorizeShowFeedback && idx < len(m.MemorizeCurrentOptions)
	case StateReadingSelectText:
		return idx < len(m.ReadingTexts)
	}
	return false
}

func (m Model) updateModeSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.StatusMessage != "" {
		m.StatusMessage = ""
	}

	switch msg.String() {
	case "m":
		return m.startMarkedPractice()
	case "a":
		m.Mode = ModeReview
		return m.confirmModeSelect()
	case "s":
		m.Mode = ModeQuiz
		return m.confirmModeSelect()
	case "d":
		m.Mode = ModeArabic
		return m.confirmModeSelect()
	case "f":
		m.Mode = ModeMemorize
		return m.confirmModeSelect()
	case "g":
		m.Mode = ModeReading
		return m.confirmModeSelect()
	case "up", "k":
		if m.Mode > 0 {
			m.Mode--
		}
	case "down", "j":
		if m.Mode < 4 {
			m.Mode++
		}
	case "enter", " ":
		return m.confirmModeSelect()
	}
	return m, nil
}

func (m Model) confirmModeSelect() (tea.Model, tea.Cmd) {
	if m.Mode == ModeArabic {
		m.ArabicMenuCursor = 0
		m.State = StateArabicMenu
		return m, nil
	}

	if m.Mode == ModeMemorize {
		texts, err := memorize.ListTexts(m.TextsDir)
		if err != nil {
			m.Err = err
			return m, nil
		}
		m.MemorizeTexts = texts
		m.MemorizeCursor = 0
		m.State = StateMemorizeSelectText
		return m, nil
	}

	if m.Mode == ModeReading {
		texts, err := reading.ListTexts(m.ReadingDir, m.ReadingProgress)
		if err != nil {
			m.Err = err
			return m, nil
		}
		m.ReadingTexts = texts
		m.ReadingCursor = 0
		m.State = StateReadingSelectText
		return m, nil
	}

	dirs, err := deck.GetDeckDirectories(m.BaseDir)
	if err != nil {
		m.Err = err
		return m, nil
	}
	m.Dirs = dirs
	m.DirCursor = 0
	m.SelectedDir = ""
	m.State = StateDirSelect
	return m, nil
}

func (m Model) updateDirSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.StatusMessage != "" {
		m.StatusMessage = ""
	}

	if (msg.String() == "u" || msg.String() == "U" || msg.String() == "ctrl+d" || msg.String() == "ctrl+u") && m.hasAnySelectedDeck() {
		m.DeselectAll()
		return m, nil
	}

	if msg.String() == "m" {
		return m.startMarkedPractice()
	}

	idx := KeyToIndex(msg.String())
	if idx >= 0 && idx < len(m.Dirs) {
		return m.selectDirectory(idx)
	}

	switch msg.String() {
	case "up", "k":
		if m.DirCursor > 0 {
			m.DirCursor--
		}
	case "down", "j":
		maxCursor := len(m.Dirs) - 1
		if m.SelectedDir == "" {
			maxCursor = len(m.Dirs)
		}
		if m.DirCursor < maxCursor {
			m.DirCursor++
		}
	case "tab":
		if m.hasAnySelectedDeck() {
			return m.startPractice()
		}
	case "esc", "b":
		return m.goUpLevel()
	case "enter", " ":
		if m.SelectedDir == "" && m.DirCursor == len(m.Dirs) {
			return m.startMarkedPractice()
		}
		if len(m.Dirs) > 0 {
			return m.selectDirectory(m.DirCursor)
		}
	}
	return m, nil
}

func (m Model) selectDirectory(idx int) (tea.Model, tea.Cmd) {
	m.DirCursor = idx
	dirName := m.Dirs[idx]
	var targetDir string
	if m.SelectedDir == "" {
		targetDir = dirName
	} else {
		targetDir = filepath.Join(m.SelectedDir, dirName)
	}

	fullDirPath := filepath.Join(m.BaseDir, targetDir)
	subdirs, err := deck.GetDeckDirectories(fullDirPath)
	if err != nil {
		m.Err = err
		return m, nil
	}
	deckFiles, err := deck.GetDeckFiles(m.BaseDir, targetDir)
	if err != nil {
		m.Err = err
		return m, nil
	}

	m.SelectedDir = targetDir
	if len(subdirs) > 0 {
		m.Dirs = subdirs
		m.DirCursor = 0
		m.DeckFiles = deckFiles
		m.State = StateDirSelect
	} else {
		m.DeckFiles = deckFiles
		m.Cursor = 0
		// Retain m.SelectedFiles so decks in other directories remain selected
		m.State = StateDeckSelect
	}
	return m, nil
}

func (m Model) updateDeckSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.StatusMessage != "" {
		m.StatusMessage = ""
	}

	if (msg.String() == "u" || msg.String() == "U" || msg.String() == "ctrl+d" || msg.String() == "ctrl+u") && m.hasAnySelectedDeck() {
		m.DeselectAll()
		return m, nil
	}

	idx := KeyToIndex(msg.String())
	if idx >= 0 && idx < len(m.DeckFiles) {
		m.Cursor = idx
		file := m.DeckFiles[idx]
		m.toggleDeck(file)
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
	case "tab":
		if m.hasAnySelectedDeck() {
			return m.startPractice()
		}
	case "esc", "b":
		return m.goUpLevel()
	case " ":
		if len(m.DeckFiles) > 0 {
			file := m.DeckFiles[m.Cursor]
			m.toggleDeck(file)
		}
	case "enter":
		if len(m.DeckFiles) == 0 {
			return m.goUpLevel()
		}
		if m.hasAnySelectedDeck() {
			return m.startPractice()
		}
	}
	return m, nil
}

func (m Model) goUpLevel() (tea.Model, tea.Cmd) {
	if m.State == StateDeckSelect {
		parentDir := filepath.Dir(m.SelectedDir)
		if parentDir == "." {
			parentDir = ""
		}
		prevName := filepath.Base(m.SelectedDir)

		subdirs, err := deck.GetDeckDirectories(filepath.Join(m.BaseDir, parentDir))
		if err != nil {
			m.Err = err
			return m, nil
		}
		m.SelectedDir = parentDir
		m.Dirs = subdirs
		m.DirCursor = 0
		for i, d := range m.Dirs {
			if d == prevName {
				m.DirCursor = i
				break
			}
		}
		m.State = StateDirSelect
		return m, nil
	}

	if m.State == StateDirSelect {
		if m.SelectedDir == "" {
			m.State = StateModeSelect
			return m, nil
		}

		parentDir := filepath.Dir(m.SelectedDir)
		if parentDir == "." {
			parentDir = ""
		}
		prevName := filepath.Base(m.SelectedDir)

		subdirs, err := deck.GetDeckDirectories(filepath.Join(m.BaseDir, parentDir))
		if err != nil {
			m.Err = err
			return m, nil
		}
		m.SelectedDir = parentDir
		m.Dirs = subdirs
		m.DirCursor = 0
		for i, d := range m.Dirs {
			if d == prevName {
				m.DirCursor = i
				break
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) startPractice() (tea.Model, tea.Cmd) {
	m.PlayingMarkedDeck = false
	err := m.LoadSelectedDecks()
	if err != nil {
		m.Err = err
		return m, nil
	}
	if len(m.Decks) == 0 {
		return m, nil
	}
	if m.Mode == ModeReview {
		m.SetupReview()
		m.State = StateReview
	} else {
		m.SetupQuiz()
		m.State = StateQuiz
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

	if msg.String() == "m" {
		m.toggleCurrentCardMarked()
		return m, nil
	}
	if msg.String() == "u" {
		m.unmarkCurrentCard()
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
		switch msg.String() {
		case "r":
			if m.isPlayingMarkedDeck() {
				m.LoadMarkedDeck()
				if len(m.MarkedCards) == 0 {
					m.State = StateModeSelect
					return m, nil
				}
				return m.startMarkedPractice()
			}
			m.SetupQuiz()
		case "enter":
			m.State = StateModeSelect
		}
		return m, nil
	}

	if m.ShowFeedback {
		switch msg.String() {
		case "m":
			m.toggleCurrentCardMarked()
			return m, nil
		case "u":
			m.unmarkCurrentCard()
			return m, nil
		case " ", "enter":
			m.ShowFeedback = false
			m.CurrentIndex++
			m.GenerateQuizOptions()
			return m, nil
		}
		return m, nil
	}

	if msg.String() == "m" {
		m.toggleCurrentCardMarked()
		return m, nil
	}
	if msg.String() == "u" {
		m.unmarkCurrentCard()
		return m, nil
	}

	choiceIdx := KeyToIndex(msg.String())
	if choiceIdx != -1 && choiceIdx < len(m.QuizOptions) {
		m.ShowFeedback = true
		m.IsCorrect = (choiceIdx == m.CorrectIndex)
	}

	return m, nil
}

func (m Model) updateArabicMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	idx := KeyToIndex(msg.String())
	if idx >= 0 && idx < 7 {
		m.ArabicMenuCursor = idx
		return m.selectArabicActivity(idx)
	}

	switch msg.String() {
	case "up", "k":
		if m.ArabicMenuCursor > 0 {
			m.ArabicMenuCursor--
		}
	case "down", "j":
		if m.ArabicMenuCursor < 6 {
			m.ArabicMenuCursor++
		}
	case "enter", " ":
		return m.selectArabicActivity(m.ArabicMenuCursor)
	case "esc", "b":
		m.State = StateModeSelect
	}
	return m, nil
}

func (m Model) selectArabicActivity(idx int) (tea.Model, tea.Cmd) {
	m.ArabicMenuCursor = idx
	switch idx {
	case 0:
		m.ArabicLetterCursor = 0
		m.ArabicAuxCursor = 0
		m.ArabicExplorerTab = 0
		m.State = StateArabicExplorer
	case 1:
		m.SetupArabicDrill(arabic.DrillPositionalForms, 0)
		m.State = StateArabicDrill
	case 2:
		m.SetupArabicDrill(arabic.DrillLetterSounds, 0)
		m.State = StateArabicDrill
	case 3:
		m.SetupArabicDrill(arabic.DrillConfusables, 0)
		m.State = StateArabicDrill
	case 4:
		m.SetupArabicDrill(arabic.DrillConnectors, 0)
		m.State = StateArabicDrill
	case 5:
		m.SetupArabicDrill(arabic.DrillSunMoon, 0)
		m.State = StateArabicDrill
	case 6:
		m.ArabicStageCursor = 0
		m.State = StateArabicStages
	}
	return m, nil
}

func (m Model) updateArabicExplorer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if m.ArabicExplorerTab == 0 {
			m.ArabicExplorerTab = 1
		} else {
			m.ArabicExplorerTab = 0
		}
	case "left", "h":
		if m.ArabicExplorerTab == 0 {
			m.ArabicLetterCursor = (m.ArabicLetterCursor - 1 + len(arabic.Letters)) % len(arabic.Letters)
		} else {
			m.ArabicAuxCursor = (m.ArabicAuxCursor - 1 + len(arabic.AuxiliaryItems)) % len(arabic.AuxiliaryItems)
		}
	case "right", "l":
		if m.ArabicExplorerTab == 0 {
			m.ArabicLetterCursor = (m.ArabicLetterCursor + 1) % len(arabic.Letters)
		} else {
			m.ArabicAuxCursor = (m.ArabicAuxCursor + 1) % len(arabic.AuxiliaryItems)
		}
	case "up", "k":
		if m.ArabicExplorerTab == 0 {
			m.ArabicLetterCursor = (m.ArabicLetterCursor - 14 + len(arabic.Letters)) % len(arabic.Letters)
		} else {
			m.ArabicAuxCursor = (m.ArabicAuxCursor - 3 + len(arabic.AuxiliaryItems)) % len(arabic.AuxiliaryItems)
		}
	case "down", "j":
		if m.ArabicExplorerTab == 0 {
			m.ArabicLetterCursor = (m.ArabicLetterCursor + 14) % len(arabic.Letters)
		} else {
			m.ArabicAuxCursor = (m.ArabicAuxCursor + 3) % len(arabic.AuxiliaryItems)
		}
	case "esc", "b":
		m.State = StateArabicMenu
	}
	return m, nil
}

func (m Model) updateArabicDrill(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ArabicQuestionIdx >= len(m.ArabicQuestions) {
		switch msg.String() {
		case "r":
			m.SetupArabicDrill(m.ArabicDrillType, m.ArabicStageCursor)
		case "enter", "esc", "b":
			if m.ArabicDrillType == arabic.DrillStage {
				m.State = StateArabicStages
			} else {
				m.State = StateArabicMenu
			}
		}
		return m, nil
	}

	if m.ArabicShowFeedback {
		if msg.String() == " " || msg.String() == "enter" {
			m.ArabicShowFeedback = false
			m.ArabicQuestionIdx++
		}
		return m, nil
	}

	if msg.String() == "esc" || msg.String() == "b" {
		if m.ArabicDrillType == arabic.DrillStage {
			m.State = StateArabicStages
		} else {
			m.State = StateArabicMenu
		}
		return m, nil
	}

	q := m.ArabicQuestions[m.ArabicQuestionIdx]
	choiceIdx := KeyToIndex(msg.String())
	if choiceIdx != -1 && choiceIdx < len(q.Options) {
		m.ArabicShowFeedback = true
		if choiceIdx == q.CorrectIndex {
			m.ArabicIsCorrect = true
			m.ArabicScore++
		} else {
			m.ArabicIsCorrect = false
		}
	}

	return m, nil
}

func (m Model) updateArabicStages(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	idx := KeyToIndex(msg.String())
	if idx >= 0 && idx < len(arabic.Stages) {
		m.ArabicStageCursor = idx
		m.State = StateArabicStageView
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.ArabicStageCursor > 0 {
			m.ArabicStageCursor--
		}
	case "down", "j":
		if m.ArabicStageCursor < len(arabic.Stages)-1 {
			m.ArabicStageCursor++
		}
	case "enter", " ":
		m.State = StateArabicStageView
	case "esc", "b":
		m.State = StateArabicMenu
	}
	return m, nil
}

func (m Model) updateArabicStageView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", " ":
		m.SetupArabicDrill(arabic.DrillStage, m.ArabicStageCursor)
		m.State = StateArabicDrill
	case "esc", "b":
		m.State = StateArabicStages
	}
	return m, nil
}

func (m Model) updateMemorizeSelectText(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.MemorizeTexts) == 0 {
		switch msg.String() {
		case "esc", "b":
			m.State = StateModeSelect
		}
		return m, nil
	}

	idx := KeyToIndex(msg.String())
	if idx >= 0 && idx < len(m.MemorizeTexts) {
		m.MemorizeCursor = idx
		if err := m.SelectMemorizeText(m.MemorizeTexts[idx].FullPath); err != nil {
			m.Err = err
			return m, nil
		}
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.MemorizeCursor > 0 {
			m.MemorizeCursor--
		}
	case "down", "j":
		if m.MemorizeCursor < len(m.MemorizeTexts)-1 {
			m.MemorizeCursor++
		}
	case "enter", " ":
		if m.MemorizeCursor >= 0 && m.MemorizeCursor < len(m.MemorizeTexts) {
			if err := m.SelectMemorizeText(m.MemorizeTexts[m.MemorizeCursor].FullPath); err != nil {
				m.Err = err
				return m, nil
			}
		}
	case "esc", "b":
		m.State = StateModeSelect
	}
	return m, nil
}

func (m Model) updateMemorizePortionSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	numSecs := len(m.MemorizeSections)
	if numSecs == 0 {
		m.State = StateMemorizeSelectText
		return m, nil
	}

	switch msg.String() {
	case "tab", "up", "down", "j", "k":
		// Toggle focus between Start and End
		if m.PortionFocus == 0 {
			m.PortionFocus = 1
		} else {
			m.PortionFocus = 0
		}

	case "s":
		m.PortionFocus = 0
	case "e":
		m.PortionFocus = 1

	case "left", "h":
		if m.PortionFocus == 0 {
			if m.PortionStartSection > 0 {
				m.PortionStartSection--
			}
		} else {
			if m.PortionEndSection > m.PortionStartSection {
				m.PortionEndSection--
			}
		}

	case "right", "l":
		if m.PortionFocus == 0 {
			if m.PortionStartSection < m.PortionEndSection {
				m.PortionStartSection++
			}
		} else {
			if m.PortionEndSection < numSecs-1 {
				m.PortionEndSection++
			}
		}

	case "[":
		if m.PortionStartSection > 0 {
			m.PortionStartSection--
		}
	case "]":
		if m.PortionStartSection < m.PortionEndSection {
			m.PortionStartSection++
		}
	case "{", "-":
		if m.PortionEndSection > m.PortionStartSection {
			m.PortionEndSection--
		}
	case "}", "+", "=":
		if m.PortionEndSection < numSecs-1 {
			m.PortionEndSection++
		}

	case "x":
		// Expand portion: advance End by 1, or retreat Start by 1 if End is already at last
		if m.PortionEndSection < numSecs-1 {
			m.PortionEndSection++
		} else if m.PortionStartSection > 0 {
			m.PortionStartSection--
		}

	case "c":
		// Contract portion
		if m.PortionEndSection > m.PortionStartSection {
			m.PortionEndSection--
		}

	case "n", ">":
		// Advance portion to next chunk
		chunkLen := m.PortionEndSection - m.PortionStartSection + 1
		if m.PortionEndSection+1 < numSecs {
			newStart := m.PortionEndSection + 1
			newEnd := newStart + chunkLen - 1
			if newEnd >= numSecs {
				newEnd = numSecs - 1
			}
			m.PortionStartSection = newStart
			m.PortionEndSection = newEnd
		}

	case "p", "<":
		// Shift portion back to previous chunk
		chunkLen := m.PortionEndSection - m.PortionStartSection + 1
		if m.PortionStartSection > 0 {
			newEnd := m.PortionStartSection - 1
			newStart := newEnd - chunkLen + 1
			if newStart < 0 {
				newStart = 0
			}
			m.PortionStartSection = newStart
			m.PortionEndSection = newEnd
		}

	case "a":
		// Select All
		m.PortionStartSection = 0
		m.PortionEndSection = numSecs - 1

	case "1":
		// Collapse to single section
		m.PortionEndSection = m.PortionStartSection

	case "enter", " ":
		m.StartMemorizePortion(m.PortionStartSection, m.PortionEndSection)
		return m, nil

	case "esc", "b":
		m.State = StateMemorizeSelectText
		return m, nil
	}

	return m, nil
}

func (m Model) updateMemorize(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.MemorizeActiveText == nil || m.MemorizeCurrentIndex >= m.PortionEndWordIdx {
		return m.finishPortion()
	}

	if m.MemorizeShowFeedback {
		switch msg.String() {
		case " ", "enter":
			m.MemorizeShowFeedback = false
			m.MemorizeCurrentIndex++
			if m.MemorizeCurrentIndex >= m.PortionEndWordIdx {
				return m.finishPortion()
			}
			m.SetupMemorizeWord()
			return m, nil
		case "r":
			m.MemorizeShowFeedback = false
			m.SetupMemorizeWord()
			return m, nil
		case "esc", "b":
			m.State = StateMemorizePortionSelect
			return m, nil
		}
		return m, nil
	}

	choiceIdx := KeyToIndex(msg.String())
	if choiceIdx == -1 {
		switch msg.String() {
		case "1":
			choiceIdx = 0
		case "2":
			choiceIdx = 1
		case "3":
			choiceIdx = 2
		case "4":
			choiceIdx = 3
		case "5":
			choiceIdx = 4
		}
	}

	if choiceIdx >= 0 && choiceIdx < len(m.MemorizeCurrentOptions) {
		if choiceIdx == m.MemorizeCorrectIndex {
			m.MemorizeStreak++
			if m.MemorizeStreak > m.MemorizeBestStreak {
				m.MemorizeBestStreak = m.MemorizeStreak
			}
			m.MemorizeCurrentIndex++
			if m.MemorizeCurrentIndex >= m.PortionEndWordIdx {
				return m.finishPortion()
			}
			m.SetupMemorizeWord()
			if m.MemorizeCurrentIndex%5 == 0 && m.MemorizeActiveText.FilePath != "" {
				m.MemorizeActiveText.Progress = m.MemorizeCurrentIndex
				_ = memorize.SaveText(m.MemorizeActiveText.FilePath, m.MemorizeActiveText)
			}
			return m, nil
		}

		// Wrong choice
		m.MemorizeStreak = 0
		m.MemorizeMistakes++
		m.MemorizeSelectedOption = choiceIdx
		m.MemorizeIsCorrect = false
		m.MemorizeShowFeedback = true
		return m, nil
	}

	switch msg.String() {
	case "r":
		m.MemorizeCurrentIndex = m.PortionStartWordIdx
		m.MemorizeMistakes = 0
		m.MemorizeStreak = 0
		m.SetupMemorizeWord()
	case "esc", "b":
		m.State = StateMemorizePortionSelect
	}
	return m, nil
}

func (m Model) finishPortion() (tea.Model, tea.Cmd) {
	totalPortionWords := m.PortionEndWordIdx - m.PortionStartWordIdx
	attempts := totalPortionWords + m.MemorizeMistakes
	acc := 100.0
	if attempts > 0 {
		acc = float64(totalPortionWords) / float64(attempts) * 100
	}

	startLabel := fmt.Sprintf("Section %d", m.PortionStartSection+1)
	endLabel := fmt.Sprintf("Section %d", m.PortionEndSection+1)
	if m.PortionStartSection < len(m.MemorizeSections) {
		startLabel = m.MemorizeSections[m.PortionStartSection].Label
	}
	if m.PortionEndSection < len(m.MemorizeSections) {
		endLabel = m.MemorizeSections[m.PortionEndSection].Label
	}

	var rangeLabel string
	if m.PortionStartSection == m.PortionEndSection {
		rangeLabel = startLabel
	} else {
		rangeLabel = fmt.Sprintf("%s to %s", startLabel, endLabel)
	}

	m.PortionCompletedMsg = fmt.Sprintf("★ Portion Complete: %s (%d words • Accuracy: %.1f%% • Mistakes: %d)",
		rangeLabel, totalPortionWords, acc, m.MemorizeMistakes)
	m.State = StateMemorizePortionSelect
	return m, nil
}

func (m Model) updateMemorizeComplete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		m.MemorizeCurrentIndex = m.PortionStartWordIdx
		m.MemorizeMistakes = 0
		m.MemorizeStreak = 0
		m.State = StateMemorize
		m.SetupMemorizeWord()
	case "enter", "esc", "b", " ":
		m.State = StateMemorizePortionSelect
	}
	return m, nil
}

func (m Model) updateReadingSelectText(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.StatusMessage != "" {
		m.StatusMessage = ""
	}

	idx := KeyToIndex(msg.String())
	if idx >= 0 && idx < len(m.ReadingTexts) {
		m.ReadingCursor = idx
		return m.startReadingSession()
	}

	switch msg.String() {
	case "up", "k":
		if m.ReadingCursor > 0 {
			m.ReadingCursor--
		}
	case "down", "j":
		if m.ReadingCursor < len(m.ReadingTexts)-1 {
			m.ReadingCursor++
		}
	case "enter", " ":
		return m.startReadingSession()
	case "esc", "b":
		m.State = StateModeSelect
	}
	return m, nil
}

func (m Model) startReadingSession() (tea.Model, tea.Cmd) {
	if len(m.ReadingTexts) == 0 || m.ReadingCursor >= len(m.ReadingTexts) {
		return m, nil
	}

	header := m.ReadingTexts[m.ReadingCursor]
	var text *reading.Text
	var err error

	if strings.HasSuffix(header.Filename, ".yaml") || strings.HasSuffix(header.Filename, ".yml") || strings.HasSuffix(header.Filename, ".json") {
		text, err = reading.LoadText(header.FullPath)
	} else {
		// Raw .txt file
		raw, readErr := os.ReadFile(header.FullPath)
		if readErr != nil {
			err = readErr
		} else {
			sentences := reading.SplitIntoSentences(string(raw))
			var sentList []reading.Sentence
			for i, s := range sentences {
				sentList = append(sentList, reading.Sentence{
					ID:      fmt.Sprintf("%s-%d", header.ID, i+1),
					English: s,
				})
			}
			text = &reading.Text{
				ID:        header.ID,
				Title:     header.Title,
				Language:  "Mandarin",
				Category:  "Reading",
				Sentences: sentList,
				FilePath:  header.FullPath,
			}
		}
	}

	if err != nil {
		m.Err = err
		return m, nil
	}

	m.ReadingSession = reading.NewSession(text, m.ReadingProgress)
	m.ReadingWordFeedback = ""
	m.State = StateReading
	return m, nil
}

func (m Model) updateReading(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ReadingSession == nil {
		m.State = StateReadingSelectText
		return m, nil
	}

	key := msg.String()

	// Check if key matches a substituted word
	if _, ok := m.ReadingSession.CurrentWeave.KeyToWord[key]; ok {
		m.ReadingSession.SelectWordByKey(key)
		m.ReadingWordFeedback = ""
		m.State = StateReadingWordDetail
		return m, nil
	}

	switch key {
	case "enter", " ", "n":
		m.ReadingSession.FinishSentence()
		m.State = StateReadingDiscussion
		return m, nil

	case "h", "?", "d":
		m.State = StateReadingDiscussion
		return m, nil

	case "+", "=", "]":
		m.ReadingSession.CycleAggressiveness(1)
		return m, nil

	case "-", "_", "[":
		m.ReadingSession.CycleAggressiveness(-1)
		return m, nil

	case "0":
		m.ReadingSession.SetAggressiveness(0)
		return m, nil

	case "p", "left":
		m.ReadingSession.Prev()
		return m, nil

	case "r":
		m.ReadingSession.Restart()
		return m, nil

	case "esc", "b":
		m.State = StateReadingSelectText
		return m, nil
	}

	return m, nil
}

func (m Model) updateReadingWordDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ReadingSession == nil || m.ReadingSession.SelectedWord == nil {
		m.State = StateReading
		return m, nil
	}

	switch msg.String() {
	case "m":
		added := m.MarkWordAsFlashcard(m.ReadingSession.SelectedWord.Translation)
		if added {
			m.ReadingWordFeedback = "★ Added to Marked Cards deck!"
		} else {
			m.ReadingWordFeedback = "★ Already in Marked Cards deck."
		}
		return m, nil

	case "k":
		m.ReadingSession.MarkSelectedWordKnown()
		m.ReadingWordFeedback = "✓ Marked as Known!"
		return m, nil

	case "esc", "enter", " ", "b":
		m.ReadingSession.CloseWordDetail()
		m.ReadingWordFeedback = ""
		m.State = StateReading
		return m, nil
	}

	return m, nil
}

func (m Model) updateReadingDiscussion(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ReadingSession == nil {
		m.State = StateReadingSelectText
		return m, nil
	}

	switch msg.String() {
	case "enter", " ", "n":
		advanced := m.ReadingSession.Next()
		if !advanced {
			m.State = StateReadingComplete
		} else {
			m.State = StateReading
		}
		return m, nil

	case "h", "esc", "b":
		m.State = StateReading
		return m, nil

	case "p", "left":
		m.ReadingSession.Prev()
		m.State = StateReading
		return m, nil

	case "+", "=", "]":
		m.ReadingSession.CycleAggressiveness(1)
		return m, nil

	case "-", "_", "[":
		m.ReadingSession.CycleAggressiveness(-1)
		return m, nil
	}

	return m, nil
}

func (m Model) updateReadingComplete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		if m.ReadingSession != nil {
			m.ReadingSession.Restart()
			m.State = StateReading
		}
	case "enter", "esc", "b", " ":
		m.State = StateReadingSelectText
	}
	return m, nil
}


