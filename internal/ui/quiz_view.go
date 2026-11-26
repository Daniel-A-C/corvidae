package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"flashcards/internal/deck"

	"github.com/charmbracelet/lipgloss"
)

const (
	quizDefaultFrameWidth = 76
	quizMinFrameWidth     = 40
	quizBodyHeight        = 12
	quizSourceLines       = 2
	quizFooterHeight      = 2
)

// quizFrameWidth computes a fixed, responsive frame width for Quiz Mode.
func (m Model) quizFrameWidth() int {
	return responsiveFrameWidth(m.Width, quizDefaultFrameWidth, quizMinFrameWidth)
}

func (m Model) viewQuiz() string {
	frameWidth := m.quizFrameWidth()

	if m.CurrentIndex >= len(m.ActiveCards) {
		return renderFrame(frameWidth, m.renderQuizComplete(frameWidth))
	}

	ref := m.ActiveCards[m.CurrentIndex]
	card := m.Decks[ref.Filename].Cards[ref.OrigIdx]
	isMarked := m.isCardMarked(card)

	headerZone := m.renderQuizHeader(frameWidth, isMarked)
	promptZone := m.renderQuizPromptCard(frameWidth, card)

	var bodyZone string
	if m.ShowFeedback {
		bodyZone = m.renderQuizFeedback(frameWidth, card)
	} else {
		bodyZone = m.renderQuizOptions(frameWidth)
	}

	sourceZone := m.renderQuizSourceBlock(frameWidth, ref)
	footerZone := m.renderQuizFooter(frameWidth, isMarked)

	assembled := lipgloss.JoinVertical(
		lipgloss.Left,
		headerZone,
		"",
		promptZone,
		"",
		bodyZone,
		sourceZone,
		footerZone,
	)

	return renderFrame(frameWidth, assembled)
}

func (m Model) renderQuizComplete(frameWidth int) string {
	totalCards := len(m.ActiveCards)
	accStr := "100%"
	if totalCards > 0 {
		accStr = fmt.Sprintf("%.1f%%", float64(m.QuizCorrect)/float64(totalCards)*100)
	}
	barLen := 26
	if frameWidth < 46 {
		barLen = 16
	}
	bar := renderProgressBar(totalCards, totalCards, barLen)
	stats := []string{
		HintStyle.Render(fmt.Sprintf("Score: %d of %d  •  Accuracy: %s", m.QuizCorrect, totalCards, accStr)),
		HintStyle.Render(fmt.Sprintf("Best Streak: %d", m.QuizBestStreak)),
	}
	controls := HintStyle.Render(fmt.Sprintf("[%s] Retry quiz  •  [%s] Main Menu",
		KeyStyle.Render("r"), KeyStyle.Render("Enter / q")))
	return renderCompleteCard(frameWidth, "Quiz complete!", bar, stats, controls)
}

// renderQuizHeader renders Zone 1: a fixed 2-line header with stats, right-aligned mark badge, progress bar, and streak.
func (m Model) renderQuizHeader(frameWidth int, isMarked bool) string {
	totalCards := len(m.ActiveCards)
	current := m.CurrentIndex
	pct := 0.0
	if totalCards > 0 {
		pct = float64(current) / float64(totalCards) * 100
	}

	answered := current
	if m.ShowFeedback {
		answered = current + 1
	}
	accStr := "100%"
	if answered > 0 {
		accStr = fmt.Sprintf("%.1f%%", float64(m.QuizCorrect)/float64(answered)*100)
	}

	leftStats := HintStyle.Render(fmt.Sprintf("Quiz: Question %d of %d (%.1f%%)  •  Accuracy: %s",
		m.CurrentIndex+1, totalCards, pct, accStr))

	var badges []string
	if m.Config != nil && m.Config.QuizSM2Sync {
		badges = append(badges, SyncOnPillStyle.Render("SM-2: ON"))
	}
	if isMarked {
		badges = append(badges, MarkedBadgeStyle.Render("★ MARKED"))
	}
	line1 := renderSplitLine(leftStats, strings.Join(badges, " "), frameWidth)

	barLen := 26
	if frameWidth < 46 {
		barLen = 16
	}
	line2 := renderProgressBar(answered, totalCards, barLen) + "  " + formatStreakBadge(m.QuizStreak)

	return line1 + "\n" + line2
}

// renderQuizPromptCard renders Zone 2: a fixed 3-line rounded card box containing the target character.
func (m Model) renderQuizPromptCard(frameWidth int, card deck.Flashcard) string {
	boxInnerWidth := frameWidth - 2
	if boxInnerWidth < 10 {
		boxInnerWidth = 10
	}
	textWidth := boxInnerWidth - 4
	if textWidth < 4 {
		textWidth = 4
	}
	charText := truncateToWidth(card.Character, textWidth)
	return QuizPromptBoxStyle.Width(boxInnerWidth).Render(CharStyle.Render(charText))
}

// renderQuizOptions renders Zone 3 in question mode, padded to exact quizBodyHeight lines.
func (m Model) renderQuizOptions(frameWidth int) string {
	var lines []string
	for i, opt := range m.QuizOptions {
		key := IndexToKey(i)
		keyBadge := KeyStyle.Render(fmt.Sprintf("[%s]", key))
		optLines := wrapQuizOption("  ", keyBadge+"  ", opt, frameWidth)
		lines = append(lines, optLines...)
		if (i+1)%5 == 0 && i < len(m.QuizOptions)-1 {
			lines = append(lines, "")
		}
	}
	return fitBlockHeight(strings.Join(lines, "\n"), quizBodyHeight)
}

// renderQuizFeedback renders Zone 3 in feedback mode, padded to exact quizBodyHeight lines.
func (m Model) renderQuizFeedback(frameWidth int, card deck.Flashcard) string {
	var lines []string
	if m.IsCorrect {
		lines = append(lines, "  "+CorrectStyle.Render("✓ Correct!"))
		lines = append(lines, "")
		if card.Pinyin != "" {
			lines = append(lines, wrapLabeledText("  ", "Pinyin:   ", card.Pinyin, PinyinStyle, frameWidth)...)
		} else if card.Pronunciation != "" {
			lines = append(lines, wrapLabeledText("  ", "Pronunciation: ", card.Pronunciation, PronunciationStyle, frameWidth)...)
		}
		lines = append(lines, wrapLabeledText("  ", "Meaning:  ", card.Meaning, MeaningStyle, frameWidth)...)
	} else {
		lines = append(lines, "  "+ErrorStyle.Render("✗ Incorrect!"))
		lines = append(lines, "")
		if m.QuizSelectedIdx >= 0 && m.QuizSelectedIdx < len(m.QuizOptions) {
			lines = append(lines, wrapLabeledText("  ", "You selected:           ", m.QuizOptions[m.QuizSelectedIdx], ErrorStyle, frameWidth)...)
		}
		correctOpt := ""
		if m.CorrectIndex >= 0 && m.CorrectIndex < len(m.QuizOptions) {
			correctOpt = m.QuizOptions[m.CorrectIndex]
		}
		lines = append(lines, wrapLabeledText("  ", "The correct answer was: ", correctOpt, CorrectStyle, frameWidth)...)
	}

	if card.Explanation != "" {
		lines = append(lines, "")
		expWidth := frameWidth - 4
		if expWidth < 20 {
			expWidth = 20
		}
		renderedExp := QuizExplanationStyle.Width(expWidth).Render(card.Explanation)
		for _, expLine := range strings.Split(renderedExp, "\n") {
			lines = append(lines, "  "+expLine)
		}
	}

	return fitBlockHeight(strings.Join(lines, "\n"), quizBodyHeight)
}

// renderQuizSourceBlock renders Zone 4: a subtle divider + 2 fixed lines reserved for the wrapped deck origin path.
func (m Model) renderQuizSourceBlock(frameWidth int, ref deck.CardRef) string {
	divider := renderDivider(frameWidth)
	if !m.ShowFeedback {
		return divider + "\n" + fitBlockHeight("", quizSourceLines)
	}

	displayName, err := filepath.Rel(m.BaseDir, ref.Filename)
	if err != nil {
		displayName = ref.Filename
	}
	wrapped := wrapPathWithIndent("Source: ", displayName, frameWidth, quizSourceLines)
	return divider + "\n" + fitBlockHeight(wrapped, quizSourceLines)
}

// renderQuizFooter renders Zone 5: a fixed 2-line left-aligned controls bar.
func (m Model) renderQuizFooter(frameWidth int, isMarked bool) string {
	var raw string
	if m.ShowFeedback {
		var markHint string
		if isMarked {
			markHint = fmt.Sprintf("[%s] Unmark", KeyStyle.Render("m/u"))
		} else {
			markHint = fmt.Sprintf("[%s] Mark card", KeyStyle.Render("m"))
		}
		if !m.IsCorrect {
			overrideHint := fmt.Sprintf("[%s] I knew it (wrong button)", KeyStyle.Render("k"))
			raw = fmt.Sprintf("%s  •  %s  •  [%s] Continue", markHint, overrideHint, KeyStyle.Render("Spacebar"))
		} else {
			raw = fmt.Sprintf("%s  •  [%s] Continue", markHint, KeyStyle.Render("Spacebar"))
		}
	} else {
		var markHint string
		if isMarked {
			markHint = fmt.Sprintf("[%s] Unmark card", KeyStyle.Render("m"))
		} else {
			markHint = fmt.Sprintf("[%s] Mark card", KeyStyle.Render("m"))
		}
		raw = fmt.Sprintf("%s  •  Press key to select  •  [%s] Help  •  [%s] Main Menu",
			markHint, KeyStyle.Render("?"), KeyStyle.Render("q"))
	}

	wrapped := lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(raw))
	return fitBlockHeight(wrapped, quizFooterHeight)
}

// wrapQuizOption formats a quiz option ("pinyin - meaning" or "meaning") with a styled key badge
// and hanging indentation when the option wraps across multiple lines.
func wrapQuizOption(leadIndent, renderedBadge, opt string, frameWidth int) []string {
	prefix := leadIndent + renderedBadge
	prefixW := lipgloss.Width(prefix)
	avail := frameWidth - prefixW
	if avail < 12 {
		avail = 12
	}

	wrappedPlain := wrapPlainText(opt, avail)
	hangIndent := strings.Repeat(" ", prefixW)

	phonetic, meaning, hasSplit := strings.Cut(opt, " - ")
	if !hasSplit {
		out := make([]string, len(wrappedPlain))
		for i, line := range wrappedPlain {
			if i == 0 {
				out[i] = prefix + MeaningStyle.Render(line)
			} else {
				out[i] = hangIndent + MeaningStyle.Render(line)
			}
		}
		return out
	}

	if len(wrappedPlain) == 1 {
		styled := fmt.Sprintf("%s - %s", PinyinStyle.Render(phonetic), MeaningStyle.Render(meaning))
		return []string{prefix + styled}
	}

	sep := " - "
	phoneticLen := len(phonetic)
	sepEnd := phoneticLen + len(sep)
	pos := 0
	out := make([]string, len(wrappedPlain))
	for i, line := range wrappedPlain {
		lineStart := pos
		lineEnd := pos + len(line)
		var styledLine string
		switch {
		case lineEnd <= phoneticLen:
			styledLine = PinyinStyle.Render(line)
		case lineStart >= sepEnd:
			styledLine = MeaningStyle.Render(line)
		default:
			var b strings.Builder
			if lineStart < phoneticLen {
				b.WriteString(PinyinStyle.Render(line[:phoneticLen-lineStart]))
			}
			midStart := max(0, phoneticLen-lineStart)
			midEnd := min(len(line), sepEnd-lineStart)
			if midStart < midEnd {
				b.WriteString(line[midStart:midEnd])
			}
			if lineEnd > sepEnd {
				b.WriteString(MeaningStyle.Render(line[sepEnd-lineStart:]))
			}
			styledLine = b.String()
		}
		if i == 0 {
			out[i] = prefix + styledLine
		} else {
			out[i] = hangIndent + styledLine
		}
		pos = lineEnd
		if pos < len(opt) && opt[pos] == ' ' {
			pos++
		}
	}
	return out
}
