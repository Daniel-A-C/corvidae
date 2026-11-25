package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"flashcards/internal/deck"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
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
	if m.Width <= 0 {
		return quizDefaultFrameWidth
	}
	w := m.Width - 4
	if w > quizDefaultFrameWidth {
		return quizDefaultFrameWidth
	}
	if w < quizMinFrameWidth {
		if m.Width < quizMinFrameWidth {
			if m.Width < 20 {
				return 20
			}
			return m.Width
		}
		return quizMinFrameWidth
	}
	return w
}

func (m Model) viewQuiz() string {
	frameWidth := m.quizFrameWidth()

	if m.CurrentIndex >= len(m.ActiveCards) {
		return m.renderQuizComplete(frameWidth)
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

	return lipgloss.NewStyle().
		Width(frameWidth).
		Align(lipgloss.Left).
		Render(assembled)
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
	bar := strings.Repeat("█", barLen)
	progressBar := QuizProgressStyle.Render("[" + bar + "]")
	summary := fmt.Sprintf("Score: %d of %d  •  Accuracy: %s", m.QuizCorrect, totalCards, accStr)
	streakLine := fmt.Sprintf("Best Streak: %d", m.QuizBestStreak)
	controls := HintStyle.Render("[r] Retry quiz  •  [Enter / q] Main Menu")

	inner := strings.Join([]string{
		CorrectStyle.Render("Quiz complete!"),
		"",
		progressBar,
		HintStyle.Render(summary),
		HintStyle.Render(streakLine),
		"",
		controls,
	}, "\n")

	boxInnerWidth := frameWidth - 2
	if boxInnerWidth < 20 {
		boxInnerWidth = 20
	}
	return QuizCompleteBoxStyle.Width(boxInnerWidth).Render(inner)
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

	var line1 string
	if isMarked {
		badge := MarkedBadgeStyle.Render("★ MARKED")
		leftW := lipgloss.Width(leftStats)
		badgeW := lipgloss.Width(badge)
		gap := frameWidth - leftW - badgeW
		if gap >= 1 {
			line1 = leftStats + strings.Repeat(" ", gap) + badge
		} else {
			line1 = leftStats + " " + badge
		}
	} else {
		line1 = leftStats
	}

	barLen := 26
	if frameWidth < 46 {
		barLen = 16
	}
	filled := 0
	if totalCards > 0 {
		filled = int(float64(barLen) * float64(current) / float64(totalCards))
	}
	if filled > barLen {
		filled = barLen
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barLen-filled)
	line2 := QuizProgressStyle.Render("["+bar+"]") + "  " + HintStyle.Render(fmt.Sprintf("Streak: %d", m.QuizStreak))

	return line1 + "\n" + line2
}

// renderQuizPromptCard renders Zone 2: a fixed 3-line rounded card box containing the target character.
func (m Model) renderQuizPromptCard(frameWidth int, card deck.Flashcard) string {
	boxInnerWidth := frameWidth - 2
	if boxInnerWidth < 10 {
		boxInnerWidth = 10
	}
	// Content area inside Padding(0, 2) is boxInnerWidth - 4.
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
	divider := QuizDividerStyle.Render(strings.Repeat("─", frameWidth))
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
		raw = fmt.Sprintf("%s  •  Press key to select  •  [%s] Main Menu", markHint, KeyStyle.Render("q"))
	}

	wrapped := lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(raw))
	return fitBlockHeight(wrapped, quizFooterHeight)
}

// fitBlockHeight ensures s has exactly height lines by padding with empty lines or clamping excess lines.
func fitBlockHeight(s string, height int) string {
	if height <= 0 {
		return ""
	}
	var lines []string
	if s != "" {
		lines = strings.Split(s, "\n")
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

// wrapPathWithIndent wraps a file path into at most maxLines lines within width,
// preferring to break after '/' separators and applying a hanging indent under prefix.
func wrapPathWithIndent(prefix, path string, width, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	prefixW := runewidth.StringWidth(prefix)
	avail := width - prefixW
	if avail < 10 {
		avail = 10
	}

	pathLines := splitPathIntoLines(path, avail, maxLines)
	indent := strings.Repeat(" ", prefixW)

	out := make([]string, 0, len(pathLines))
	for i, line := range pathLines {
		if i == 0 {
			out = append(out, QuizSourceLabelStyle.Render(prefix)+QuizSourcePathStyle.Render(line))
		} else {
			out = append(out, indent+QuizSourcePathStyle.Render(line))
		}
	}
	return strings.Join(out, "\n")
}

// splitPathIntoLines breaks path into up to maxLines chunks of at most lineWidth visual columns,
// preferring to break after '/' characters.
func splitPathIntoLines(path string, lineWidth, maxLines int) []string {
	if lineWidth <= 0 || maxLines <= 0 {
		return nil
	}
	// Tokenize keeping trailing '/' attached to each directory segment.
	parts := strings.Split(path, "/")
	tokens := make([]string, 0, len(parts))
	for i, p := range parts {
		if i < len(parts)-1 {
			tokens = append(tokens, p+"/")
		} else if p != "" {
			tokens = append(tokens, p)
		}
	}

	var lines []string
	current := ""
	currentW := 0

	flush := func() {
		if current != "" {
			lines = append(lines, current)
			current = ""
			currentW = 0
		}
	}

	for _, tok := range tokens {
		tokW := runewidth.StringWidth(tok)
		if currentW > 0 && currentW+tokW > lineWidth {
			flush()
		}
		if tokW <= lineWidth {
			current += tok
			currentW += tokW
			continue
		}
		// Single segment longer than lineWidth: split by rune width.
		for _, r := range tok {
			rw := runewidth.RuneWidth(r)
			if currentW > 0 && currentW+rw > lineWidth {
				flush()
			}
			current += string(r)
			currentW += rw
		}
	}
	flush()

	if len(lines) == 0 {
		return []string{""}
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		last := lines[maxLines-1]
		lines[maxLines-1] = truncateToWidth(last, lineWidth)
	}
	return lines
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

	// Determine if option has a "phonetic - meaning" split so we can style them cleanly.
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

	// Render the styled text once if it fits on a single line (the common case).
	if len(wrappedPlain) == 1 {
		styled := fmt.Sprintf("%s - %s", PinyinStyle.Render(phonetic), MeaningStyle.Render(meaning))
		return []string{prefix + styled}
	}

	// Multi-line wrapped option: color the phonetic prefix on the first line(s) and meaning after " - ".
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
			// Line crosses the " - " boundary.
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
		// Advance pos past the line and any single space separator consumed during wrapping.
		pos = lineEnd
		if pos < len(opt) && opt[pos] == ' ' {
			pos++
		}
	}
	return out
}

// wrapLabeledText wraps a labeled feedback line (e.g., "Meaning: ...") with hanging indent under the label.
func wrapLabeledText(leadIndent, label, value string, valueStyle lipgloss.Style, frameWidth int) []string {
	prefix := leadIndent + label
	prefixW := runewidth.StringWidth(prefix)
	avail := frameWidth - prefixW
	if avail < 12 {
		avail = 12
	}

	wrapped := wrapPlainText(value, avail)
	hangIndent := strings.Repeat(" ", prefixW)
	out := make([]string, len(wrapped))
	for i, line := range wrapped {
		if i == 0 {
			out[i] = prefix + valueStyle.Render(line)
		} else {
			out[i] = hangIndent + valueStyle.Render(line)
		}
	}
	return out
}

// wrapPlainText wraps plain text to maxWidth visual columns on word boundaries,
// falling back to rune-level splits for words longer than maxWidth.
func wrapPlainText(text string, maxWidth int) []string {
	if maxWidth <= 0 || text == "" {
		return []string{text}
	}
	if runewidth.StringWidth(text) <= maxWidth {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	current := ""
	currentW := 0

	for _, w := range words {
		ww := runewidth.StringWidth(w)
		if currentW == 0 {
			if ww <= maxWidth {
				current = w
				currentW = ww
			} else {
				// Hard-split oversized token by rune width.
				for _, r := range w {
					rw := runewidth.RuneWidth(r)
					if currentW > 0 && currentW+rw > maxWidth {
						lines = append(lines, current)
						current = ""
						currentW = 0
					}
					current += string(r)
					currentW += rw
				}
			}
			continue
		}
		if currentW+1+ww <= maxWidth {
			current += " " + w
			currentW += 1 + ww
		} else {
			lines = append(lines, current)
			current = ""
			currentW = 0
			if ww <= maxWidth {
				current = w
				currentW = ww
			} else {
				for _, r := range w {
					rw := runewidth.RuneWidth(r)
					if currentW > 0 && currentW+rw > maxWidth {
						lines = append(lines, current)
						current = ""
						currentW = 0
					}
					current += string(r)
					currentW += rw
				}
			}
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// truncateToWidth truncates s to at most maxWidth visual columns, appending "…" if truncated.
func truncateToWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}
	if maxWidth == 1 {
		return "…"
	}
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > maxWidth-1 {
			break
		}
		b.WriteRune(r)
		w += rw
	}
	b.WriteRune('…')
	return b.String()
}
