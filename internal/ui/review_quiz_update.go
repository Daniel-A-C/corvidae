package ui

import (
	"math/rand"
	"time"

	"flashcards/internal/deck"
	"flashcards/internal/sm2"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.State = StateModeSelect
		return m, nil
	}

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
			m.ReviewGrades = [6]int{}
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

		now := time.Now()
		d.Cards[ref.OrigIdx] = sm2.CalculateSM2(d.Cards[ref.OrigIdx], grade, now)
		m.Decks[ref.Filename] = d

		if err := deck.SaveDeck(ref.Filename, d); err != nil {
			m.Err = err
			return m, nil
		}
		m.Stats.RecordReview(now)

		if grade >= 0 && grade < len(m.ReviewGrades) {
			m.ReviewGrades[grade]++
		}
		m.CurrentIndex++
		m.ShowAnswer = false
		m.ShowExplanation = false
	}

	return m, nil
}

func (m Model) updateQuiz(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "q" || msg.String() == "esc" {
		m.ShowFeedback = false
		m.QuizLastCardSnapshot = nil
		m.State = StateModeSelect
		return m, nil
	}

	if msg.String() == "S" && m.Config != nil {
		m.Config.QuizSM2Sync = !m.Config.QuizSM2Sync
		m.saveConfig()
		return m, nil
	}

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
		case "k", "K", "w", "W":
			if !m.IsCorrect {
				m.QuizCorrect++
				m.QuizStreak = m.QuizPrevStreak + 1
				if m.QuizStreak > m.QuizBestStreak {
					m.QuizBestStreak = m.QuizStreak
				}
				if m.Config != nil && m.Config.QuizSM2Sync && m.QuizLastCardSnapshot != nil && m.CurrentIndex < len(m.ActiveCards) {
					m.restoreCardAndSyncSM2(m.ActiveCards[m.CurrentIndex], m.QuizLastCardSnapshot, sm2.GradeGood)
				}
				m.Stats.RecordQuizOverride(time.Now())
				m.QuizLastCardSnapshot = nil
				m.IsCorrect = true
				m.ShowFeedback = false
				m.QuizSelectedIdx = -1
				m.CurrentIndex++
				m.GenerateQuizOptions()
				return m, nil
			}
		case " ", "enter":
			m.ShowFeedback = false
			m.QuizSelectedIdx = -1
			m.QuizLastCardSnapshot = nil
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
	if choiceIdx != -1 && choiceIdx < len(m.QuizOptions) {
		m.QuizSelectedIdx = choiceIdx
		m.ShowFeedback = true
		m.IsCorrect = (choiceIdx == m.CorrectIndex)
		m.QuizLastCardSnapshot = nil
		if m.IsCorrect {
			m.QuizCorrect++
			m.QuizStreak++
			if m.QuizStreak > m.QuizBestStreak {
				m.QuizBestStreak = m.QuizStreak
			}
			if m.Config != nil && m.Config.QuizSM2Sync && m.CurrentIndex < len(m.ActiveCards) {
				m.QuizLastCardSnapshot = m.syncCardSM2(m.ActiveCards[m.CurrentIndex], sm2.GradeGood)
			}
		} else {
			m.QuizPrevStreak = m.QuizStreak
			m.QuizStreak = 0
			if m.Config != nil && m.Config.QuizSM2Sync && m.CurrentIndex < len(m.ActiveCards) {
				m.QuizLastCardSnapshot = m.syncCardSM2(m.ActiveCards[m.CurrentIndex], sm2.GradeWrong)
			}
		}
		m.Stats.RecordQuiz(m.IsCorrect, time.Now())
	}

	return m, nil
}
