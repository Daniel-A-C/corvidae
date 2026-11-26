package ui

import (
	"fmt"
	"time"
	"unicode"

	"flashcards/internal/config"
	"flashcards/internal/memorize"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateMemorizeSelectText(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.MemorizeTexts) == 0 {
		switch msg.String() {
		case "esc", "b":
			m.State = StateModeSelect
		}
		return m, nil
	}

	switch msg.String() {
	case "t", "T":
		m.ToggleMemorizeTypingMode()
		return m, nil
	case "o", "O":
		m.CycleMemorizeOptionCount()
		return m, nil
	case "2", "3", "4", "5":
		m.SetMemorizeOptionCount(int(msg.String()[0] - '0'))
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
		if m.PortionEndSection < numSecs-1 {
			m.PortionEndSection++
		} else if m.PortionStartSection > 0 {
			m.PortionStartSection--
		}

	case "c":
		if m.PortionEndSection > m.PortionStartSection {
			m.PortionEndSection--
		}

	case "n", ">":
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

	case "d", "D":
		if m.MemorizeActiveText != nil {
			if s, e, ok := m.MemorizeActiveText.NextDueRange(m.MemorizeSections, time.Now()); ok {
				m.PortionStartSection = s
				m.PortionEndSection = e
				m.PortionCompletedMsg = ""
			} else {
				m.PortionCompletedMsg = "✓ All sections are caught up — none due today!"
			}
		}

	case "a":
		m.PortionStartSection = 0
		m.PortionEndSection = numSecs - 1

	case "t", "T":
		m.ToggleMemorizeTypingMode()
		return m, nil

	case "o", "O":
		m.CycleMemorizeOptionCount()
		return m, nil

	case "2", "3", "4", "5":
		m.SetMemorizeOptionCount(int(msg.String()[0] - '0'))
		return m, nil

	case "1":
		m.PortionEndSection = m.PortionStartSection

	case "enter", " ":
		m.StartMemorizePortion(m.PortionStartSection, m.PortionEndSection)
		return m, nil

	case "esc", "b":
		// Refresh due section badges on text list when returning
		if texts, err := memorize.ListTexts(m.TextsDir); err == nil {
			m.MemorizeTexts = texts
		}
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
			m.MemorizeTypedWrong = ""
			m.MemorizeCurrentIndex++
			if m.MemorizeCurrentIndex >= m.PortionEndWordIdx {
				return m.finishPortion()
			}
			m.SetupMemorizeWord()
			return m, nil
		case "k", "K", "w", "W":
			if m.MemorizeMistakes > 0 {
				m.MemorizeMistakes--
			}
			m.MemorizeStreak = m.MemorizePrevStreak + 1
			if m.MemorizeStreak > m.MemorizeBestStreak {
				m.MemorizeBestStreak = m.MemorizeStreak
			}
			m.Stats.RecordMemorizeWord(time.Now())
			m.MemorizeShowFeedback = false
			m.MemorizeTypedWrong = ""
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
		case "r":
			m.MemorizeShowFeedback = false
			m.MemorizeTypedWrong = ""
			m.SetupMemorizeWord()
			return m, nil
		case "tab":
			m.ToggleMemorizeTypingMode()
			m.MemorizeShowFeedback = false
			m.SetupMemorizeWord()
			return m, nil
		case "o", "O":
			if !m.MemorizeTypingMode {
				m.CycleMemorizeOptionCount()
				m.MemorizeShowFeedback = false
				m.SetupMemorizeWord()
			}
			return m, nil
		case "esc", "b":
			m.State = StateMemorizePortionSelect
			return m, nil
		}
		return m, nil
	}

	// First-Letter Typing Recitation Mode
	if m.MemorizeTypingMode {
		switch msg.String() {
		case "esc":
			m.State = StateMemorizePortionSelect
			return m, nil
		case "tab":
			m.ToggleMemorizeTypingMode()
			m.SetupMemorizeWord()
			return m, nil
		case " ":
			// Peek at current word (counts as a mistake and shows feedback)
			m.MemorizePrevStreak = m.MemorizeStreak
			m.MemorizeStreak = 0
			m.MemorizeMistakes++
			m.MemorizeTypedWrong = ""
			m.MemorizeIsCorrect = false
			m.MemorizeShowFeedback = true
			return m, nil
		}

		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			r := msg.Runes[0]
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				targetWord := m.MemorizeActiveText.Words[m.MemorizeCurrentIndex].Word
				if memorize.MatchesFirstLetter(targetWord, r) {
					m.MemorizeStreak++
					if m.MemorizeStreak > m.MemorizeBestStreak {
						m.MemorizeBestStreak = m.MemorizeStreak
					}
					m.Stats.RecordMemorizeWord(time.Now())
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

				m.MemorizePrevStreak = m.MemorizeStreak
				m.MemorizeStreak = 0
				m.MemorizeMistakes++
				m.MemorizeTypedWrong = string(r)
				m.MemorizeIsCorrect = false
				m.MemorizeShowFeedback = true
				return m, nil
			}
		}
		return m, nil
	}

	// Multiple-Choice Options Mode
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
			m.Stats.RecordMemorizeWord(time.Now())
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
		m.MemorizePrevStreak = m.MemorizeStreak
		m.MemorizeStreak = 0
		m.MemorizeMistakes++
		m.MemorizeSelectedOption = choiceIdx
		m.MemorizeIsCorrect = false
		m.MemorizeShowFeedback = true
		return m, nil
	}

	switch msg.String() {
	case "t", "T", "tab":
		m.ToggleMemorizeTypingMode()
		m.SetupMemorizeWord()
		return m, nil
	case "o", "O":
		m.CycleMemorizeOptionCount()
		m.SetupMemorizeWord()
		return m, nil
	case "r":
		m.MemorizeCurrentIndex = m.PortionStartWordIdx
		m.MemorizeMistakes = 0
		m.MemorizeStreak = 0
		m.MemorizePrevStreak = 0
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

	// Update per-section SM-2 metadata and persist to disk
	if m.MemorizeActiveText != nil && len(m.MemorizeSections) > 0 {
		modeStr := config.MemorizeModeOpts
		if m.MemorizeTypingMode {
			modeStr = config.MemorizeModeTyping
		}
		m.MemorizeActiveText.RecordPortionReview(m.MemorizeSections, m.PortionStartSection, m.PortionEndSection, acc, modeStr, time.Now())
		if m.MemorizeActiveText.FilePath != "" {
			_ = memorize.SaveText(m.MemorizeActiveText.FilePath, m.MemorizeActiveText)
		}
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
	case "t", "T":
		m.ToggleMemorizeTypingMode()
		return m, nil
	case "o", "O":
		m.CycleMemorizeOptionCount()
		return m, nil
	case "2", "3", "4", "5":
		m.SetMemorizeOptionCount(int(msg.String()[0] - '0'))
		return m, nil
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
