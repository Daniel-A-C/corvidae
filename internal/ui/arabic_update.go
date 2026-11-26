package ui

import (
	"time"

	"flashcards/internal/arabic"

	tea "github.com/charmbracelet/bubbletea"
)

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
		m.Stats.RecordArabicDrill(time.Now())
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
