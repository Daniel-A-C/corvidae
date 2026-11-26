package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"flashcards/internal/reading"

	tea "github.com/charmbracelet/bubbletea"
)

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

	if _, ok := m.ReadingSession.CurrentWeave.KeyToWord[key]; ok {
		m.ReadingSession.SelectWordByKey(key)
		m.ReadingWordFeedback = ""
		m.State = StateReadingWordDetail
		return m, nil
	}

	switch key {
	case "enter", " ", "n":
		m.ReadingSession.FinishSentence()
		m.Stats.RecordReadingSentence(time.Now())
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
