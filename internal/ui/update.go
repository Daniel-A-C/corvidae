package ui

import (
	"path/filepath"
	"strings"

	"flashcards/internal/arabic"
	"flashcards/internal/deck"
	"flashcards/internal/memorize"
	"flashcards/internal/reading"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles incoming terminal messages and user input events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.ShowHelp {
			switch msg.String() {
			case "esc", "?", "f1", " ", "enter", "q":
				m.ShowHelp = false
			}
			return m, nil
		}
		if msg.String() == "f1" || (msg.String() == "?" && m.canToggleHelpWithQuestionMark()) {
			m.ShowHelp = true
			return m, nil
		}
		inMemorizeTyping := m.State == StateMemorize && m.MemorizeTypingMode && !m.MemorizeShowFeedback
		inForge := m.State == StateForgeMenu || m.State == StateForgePlay || m.State == StateForgeProve || m.State == StateForgeShop || m.State == StateForgeRunOver
		if msg.String() == "q" && !inMemorizeTyping && !inForge && !m.isSelectionKey("q") && m.State != StateQuiz && m.State != StateMatch {
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
		case StateMatch:
			return m.updateMatch(msg)
		case StateDashboard:
			return m.updateDashboard(msg)
		case StateSettings:
			return m.updateSettings(msg)
		case StateForgeMenu:
			return m.updateForgeMenu(msg)
		case StateForgePlay:
			return m.updateForgePlay(msg)
		case StateForgeProve:
			return m.updateForgeProve(msg)
		case StateForgeShop:
			return m.updateForgeShop(msg)
		case StateForgeRunOver:
			return m.updateForgeRunOver(msg)
		}

	case tea.MouseMsg:
		switch m.State {
		case StateDirSelect:
			return m.updateDirSelectMouse(msg)
		case StateDeckSelect:
			return m.updateDeckSelectMouse(msg)
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
		return idx < 7
	case StateDirSelect:
		return idx < len(m.Dirs)
	case StateDeckSelect:
		return idx < len(m.DeckFiles)
	case StateQuiz:
		if key == "q" {
			return false
		}
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
		if m.MemorizeTypingMode {
			return !m.MemorizeShowFeedback
		}
		return !m.MemorizeShowFeedback && idx < len(m.MemorizeCurrentOptions)
	case StateReadingSelectText:
		return idx < len(m.ReadingTexts)
	case StateMatch:
		if key == "q" {
			return false
		}
		if idx >= 0 && idx < len(m.MatchLeftItems) {
			return true
		}
		if idx >= 5 && idx-5 < len(m.MatchRightItems) {
			return true
		}
	case StateSettings:
		return idx < 6
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
	case "i":
		m.openDashboard()
		return m, nil
	case "c":
		m.SettingsCursor = 0
		m.State = StateSettings
		return m, nil
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
	case "h":
		m.Mode = ModeMatch
		return m.confirmModeSelect()
	case "j":
		m.Mode = ModeForge
		return m.confirmModeSelect()
	case "up", "k":
		if m.Mode > 0 {
			m.Mode--
		}
	case "down":
		if m.Mode < 6 {
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

	if m.Mode == ModeForge {
		m.openForgeMenu()
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

	shiftIdx := ShiftKeyToIndex(msg.String())
	if shiftIdx >= 0 && shiftIdx < len(m.Dirs) {
		return m.toggleDirectorySelection(shiftIdx)
	}
	if msg.String() == "shift+enter" || msg.String() == "shift+space" {
		if m.DirCursor >= 0 && m.DirCursor < len(m.Dirs) {
			return m.toggleDirectorySelection(m.DirCursor)
		}
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
		m.State = StateDeckSelect
	}
	return m, nil
}

func (m Model) updateDeckSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.StatusMessage != "" {
		m.StatusMessage = ""
	}

	if msg.String() == "shift+enter" || msg.String() == "shift+space" {
		return m.toggleDirectorySelectionByPath(m.SelectedDir)
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
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
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

func (m Model) updateDirSelectMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Button != tea.MouseButtonLeft && msg.Type != tea.MouseLeft {
		return m, nil
	}
	if msg.Action != tea.MouseActionPress && msg.Action != 0 {
		return m, nil
	}

	idx := m.mouseYToDirIndex(msg.Y)
	if idx >= 0 && idx < len(m.Dirs) {
		if msg.Shift {
			return m.toggleDirectorySelection(idx)
		}
		return m.selectDirectory(idx)
	}
	return m, nil
}

func (m Model) updateDeckSelectMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Button != tea.MouseButtonLeft && msg.Type != tea.MouseLeft {
		return m, nil
	}
	if msg.Action != tea.MouseActionPress && msg.Action != 0 {
		return m, nil
	}

	idx := m.mouseYToDeckIndex(msg.Y)
	if idx >= 0 && idx < len(m.DeckFiles) {
		m.Cursor = idx
		if msg.Shift {
			return m.toggleDirectorySelectionByPath(m.SelectedDir)
		}
		file := m.DeckFiles[idx]
		m.toggleDeck(file)
		return m, nil
	}
	return m, nil
}

func (m Model) mouseYToDirIndex(y int) int {
	content := m.viewDirSelect()
	contentLines := strings.Count(content, "\n") + 1
	topOffset := 0
	if m.Height > contentLines {
		topOffset = (m.Height - contentLines) / 2
	}
	relLine := y - topOffset - 2
	if relLine < 0 {
		return -1
	}
	group := relLine / 6
	offset := relLine % 6
	if offset >= 5 {
		return -1
	}
	idx := group*5 + offset
	if idx >= 0 && idx < len(m.Dirs) {
		return idx
	}
	return -1
}

func (m Model) mouseYToDeckIndex(y int) int {
	content := m.viewDeckSelect()
	contentLines := strings.Count(content, "\n") + 1
	topOffset := 0
	if m.Height > contentLines {
		topOffset = (m.Height - contentLines) / 2
	}
	relLine := y - topOffset - 2
	if relLine < 0 {
		return -1
	}
	group := relLine / 6
	offset := relLine % 6
	if offset >= 5 {
		return -1
	}
	idx := group*5 + offset
	if idx >= 0 && idx < len(m.DeckFiles) {
		return idx
	}
	return -1
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
	} else if m.Mode == ModeMatch {
		m.SetupMatch()
		m.State = StateMatch
	} else {
		m.SetupQuiz()
		m.State = StateQuiz
	}
	return m, nil
}
