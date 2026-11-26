package ui

import (
	"fmt"
	"time"

	"flashcards/internal/sm2"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateMatch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "q" || msg.String() == "esc" {
		m.State = StateModeSelect
		return m, nil
	}

	if msg.String() == "S" && m.Config != nil {
		m.Config.MatchSM2Sync = !m.Config.MatchSM2Sync
		m.saveConfig()
		if m.Config.MatchSM2Sync {
			m.MatchLastFeedback = "SM-2 Sync enabled — matched pairs will update card intervals"
		} else {
			m.MatchLastFeedback = "SM-2 Sync disabled"
		}
		m.MatchLastIsError = false
		return m, nil
	}

	if len(m.ActiveCards) == 0 || m.MatchRoundStart >= len(m.ActiveCards) {
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
			m.SetupMatch()
		case "enter", " ":
			m.State = StateModeSelect
		}
		return m, nil
	}

	m.MatchMismatchLeft = -1
	m.MatchMismatchRight = -1

	if msg.String() == "p" || msg.String() == "P" {
		m.MatchHidePinyin = !m.MatchHidePinyin
		if m.Config != nil {
			m.Config.MatchHidePinyin = m.MatchHidePinyin
			m.saveConfig()
		}
		if m.MatchHidePinyin {
			m.MatchLastFeedback = "Pinyin hidden — showing meanings only"
		} else {
			m.MatchLastFeedback = "Pinyin visible — showing Pinyin & meanings"
		}
		m.MatchLastIsError = false
		return m, nil
	}

	if msg.String() == "m" {
		if card, nowMarked, ok := m.toggleMatchCardMarked(); ok {
			if nowMarked {
				m.MatchLastFeedback = fmt.Sprintf("★ Marked card: %s (%s)", card.Character, card.FormatQuizOption())
			} else {
				m.MatchLastFeedback = fmt.Sprintf("Unmarked card: %s", card.Character)
			}
			m.MatchLastIsError = false
		}
		return m, nil
	}

	leftIdx := -1
	rightIdx := -1

	idx := KeyToIndex(msg.String())
	if idx >= 0 && idx < 5 {
		leftIdx = idx
	} else if idx >= 5 && idx < 10 {
		rightIdx = idx - 5
	} else {
		switch msg.String() {
		case "1":
			leftIdx = 0
		case "2":
			leftIdx = 1
		case "3":
			leftIdx = 2
		case "4":
			leftIdx = 3
		case "5":
			leftIdx = 4
		case "6":
			rightIdx = 0
		case "7":
			rightIdx = 1
		case "8":
			rightIdx = 2
		case "9":
			rightIdx = 3
		case "0":
			rightIdx = 4
		}
	}

	if leftIdx >= 0 && leftIdx < len(m.MatchLeftItems) {
		if m.MatchLeftItems[leftIdx].Matched {
			return m, nil
		}
		if m.MatchSelectedLeft == leftIdx {
			m.MatchSelectedLeft = -1
			return m, nil
		}
		m.MatchSelectedLeft = leftIdx
	} else if rightIdx >= 0 && rightIdx < len(m.MatchRightItems) {
		if m.MatchRightItems[rightIdx].Matched {
			return m, nil
		}
		if m.MatchSelectedRight == rightIdx {
			m.MatchSelectedRight = -1
			return m, nil
		}
		m.MatchSelectedRight = rightIdx
	} else {
		return m, nil
	}

	if m.MatchSelectedLeft >= 0 && m.MatchSelectedRight >= 0 {
		leftItem := m.MatchLeftItems[m.MatchSelectedLeft]
		rightItem := m.MatchRightItems[m.MatchSelectedRight]
		m.MatchAttempts++

		isMatch := leftItem.PairID == rightItem.PairID || SameCard(leftItem.Card, rightItem.Card)
		if isMatch {
			m.MatchLeftItems[m.MatchSelectedLeft].Matched = true
			m.MatchRightItems[m.MatchSelectedRight].Matched = true
			m.MatchPairsMatched++
			m.MatchStreak++
			if m.MatchStreak > m.MatchBestStreak {
				m.MatchBestStreak = m.MatchStreak
			}
			if m.Config != nil && m.Config.MatchSM2Sync {
				grade := sm2.GradeGood
				if m.MatchCardHadMistake[leftItem.PairID] {
					grade = sm2.GradeHard
				}
				m.syncCardSM2(leftItem.CardRef, grade)
			}
			m.Stats.RecordMatchPair(time.Now())

			refCopy := leftItem.CardRef
			m.MatchLastCard = &refCopy
			m.MatchLastFeedback = fmt.Sprintf("✓ Matched: %s  —  %s", leftItem.Card.Character, leftItem.Card.FormatQuizOption())
			m.MatchLastIsError = false
			m.MatchSelectedLeft = -1
			m.MatchSelectedRight = -1

			allMatched := true
			for _, item := range m.MatchLeftItems {
				if !item.Matched {
					allMatched = false
					break
				}
			}
			if allMatched {
				m.MatchRoundStart += len(m.MatchLeftItems)
				m.SetupMatchRound()
			}
		} else {
			if m.MatchCardHadMistake == nil {
				m.MatchCardHadMistake = make(map[int]bool)
			}
			m.MatchCardHadMistake[leftItem.PairID] = true
			m.MatchCardHadMistake[rightItem.PairID] = true
			m.MatchMistakes++
			m.MatchStreak = 0
			m.MatchMismatchLeft = m.MatchSelectedLeft
			m.MatchMismatchRight = m.MatchSelectedRight
			m.MatchLastFeedback = fmt.Sprintf("✗ Not a match: %s ≠ %s", leftItem.Card.Character, rightItem.Card.FormatMatchOption(m.MatchHidePinyin))
			m.MatchLastIsError = true
			m.MatchSelectedLeft = -1
			m.MatchSelectedRight = -1
		}
	}

	return m, nil
}
