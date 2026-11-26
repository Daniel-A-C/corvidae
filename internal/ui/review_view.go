package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"flashcards/internal/deck"
	"flashcards/internal/sm2"

	"github.com/charmbracelet/lipgloss"
)

const (
	reviewBodyHeight   = 12
	reviewSourceLines  = 2
	reviewFooterHeight = 2
)

func (m Model) viewReview() string {
	frameWidth := m.frameWidth()

	if len(m.ActiveCards) == 0 {
		controls := HintStyle.Render(fmt.Sprintf("[%s] Review all cards anyway  •  [%s] Main Menu  •  [%s] Quit",
			KeyStyle.Render("r"), KeyStyle.Render("Enter"), KeyStyle.Render("q")))
		card := renderCompleteCard(frameWidth, "No cards due today!", "", []string{
			HintStyle.Render("All scheduled cards in your selected decks are caught up."),
		}, controls)
		return renderFrame(frameWidth, card)
	}

	if m.CurrentIndex >= len(m.ActiveCards) {
		return renderFrame(frameWidth, m.renderReviewComplete(frameWidth))
	}

	ref := m.ActiveCards[m.CurrentIndex]
	card := m.Decks[ref.Filename].Cards[ref.OrigIdx]
	isMarked := m.isCardMarked(card)

	headerZone := m.renderReviewHeader(frameWidth, card, isMarked)
	promptZone := m.renderReviewPromptCard(frameWidth, card)
	bodyZone := m.renderReviewBody(frameWidth, card)
	sourceZone := m.renderReviewSourceBlock(frameWidth, ref)
	footerZone := m.renderReviewFooter(frameWidth, isMarked)

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

func (m Model) renderReviewComplete(frameWidth int) string {
	total := len(m.ActiveCards)
	barLen := 26
	if frameWidth < 46 {
		barLen = 16
	}
	bar := renderProgressBar(total, total, barLen)

	passed := m.ReviewGrades[sm2.GradeGood] + m.ReviewGrades[sm2.GradeEasy] + m.ReviewGrades[sm2.GradePerfect]
	missed := m.ReviewGrades[sm2.GradeBlackout] + m.ReviewGrades[sm2.GradeWrong] + m.ReviewGrades[sm2.GradeHard]

	stats := []string{
		HintStyle.Render(fmt.Sprintf("Reviewed: %d cards", total)),
	}
	if passed+missed > 0 {
		stats = append(stats, fmt.Sprintf("%s  •  %s",
			CorrectStyle.Render(fmt.Sprintf("Passed (Good–Perfect): %d", passed)),
			GradeHardStyle.Render(fmt.Sprintf("Needs work (Blackout–Hard): %d", missed)),
		))
	}
	controls := HintStyle.Render(fmt.Sprintf("[%s] Review all cards again  •  [%s] Main Menu  •  [%s] Quit",
		KeyStyle.Render("r"), KeyStyle.Render("Enter"), KeyStyle.Render("q")))

	return renderCompleteCard(frameWidth, "Daily review complete! Great job.", bar, stats, controls)
}

func (m Model) renderReviewHeader(frameWidth int, card deck.Flashcard, isMarked bool) string {
	total := len(m.ActiveCards)
	pct := 0.0
	if total > 0 {
		pct = float64(m.CurrentIndex) / float64(total) * 100
	}

	leftStats := HintStyle.Render(fmt.Sprintf("Review: Card %d of %d (%.1f%%)", m.CurrentIndex+1, total, pct))
	badge := ""
	if isMarked {
		badge = MarkedBadgeStyle.Render("★ MARKED")
	}
	line1 := renderSplitLine(leftStats, badge, frameWidth)

	barLen := 26
	if frameWidth < 46 {
		barLen = 16
	}
	ease := card.Ease
	if ease == 0 {
		ease = sm2.DefaultEase
	}
	meta := HintStyle.Render(fmt.Sprintf("Reps: %d  •  Ease: %.2f", card.Reps, ease))
	line2 := renderProgressBar(m.CurrentIndex, total, barLen) + "  " + meta

	return line1 + "\n" + line2
}

func (m Model) renderReviewPromptCard(frameWidth int, card deck.Flashcard) string {
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

func (m Model) renderReviewBody(frameWidth int, card deck.Flashcard) string {
	var lines []string
	if !m.ShowAnswer {
		lines = append(lines, "")
		lines = append(lines, "  "+KeyStyle.Render("[ Spacebar ]")+" "+HintStyle.Render("Reveal answer"))
		lines = append(lines, "")
		if card.Reps == 0 {
			lines = append(lines, "  "+HintStyle.Render("Status: New card"))
		} else {
			lines = append(lines, "  "+HintStyle.Render(fmt.Sprintf("Status: Last interval %s  •  Due %s", formatIntervalDays(card.Interval), card.NextReview)))
		}
		return fitBlockHeight(strings.Join(lines, "\n"), reviewBodyHeight)
	}

	if card.Pinyin != "" {
		lines = append(lines, wrapLabeledText("  ", "Pinyin: ", card.Pinyin, PinyinStyle, frameWidth)...)
	} else if card.Pronunciation != "" {
		lines = append(lines, wrapLabeledText("  ", "Pronunciation: ", card.Pronunciation, PronunciationStyle, frameWidth)...)
	}
	lines = append(lines, wrapLabeledText("  ", "Meaning: ", card.Meaning, MeaningStyle, frameWidth)...)

	if m.ShowExplanation && card.Explanation != "" {
		lines = append(lines, "")
		expWidth := frameWidth - 4
		if expWidth < 20 {
			expWidth = 20
		}
		renderedExp := QuizExplanationStyle.Width(expWidth).Render(card.Explanation)
		expLines := strings.Split(renderedExp, "\n")
		maxExpLines := 4
		if len(expLines) > maxExpLines {
			expLines = expLines[:maxExpLines]
		}
		for _, expLine := range expLines {
			lines = append(lines, "  "+expLine)
		}
	}

	lines = append(lines, "")
	lines = append(lines, "  "+HintStyle.Render("How well did you know this?"))

	now := time.Now()
	gradeChip := func(key, label string, grade int, style lipgloss.Style) string {
		next := sm2.CalculateSM2(card, grade, now)
		ivl := HintStyle.Render("(" + formatIntervalDays(next.Interval) + ")")
		return fmt.Sprintf("[%s] %s %s", KeyStyle.Render(key), style.Render(label), ivl)
	}

	row1 := "  " +
		padRightVisual(gradeChip("d", "Blackout", sm2.GradeBlackout, GradeBlackoutStyle), 22) +
		padRightVisual(gradeChip("f", "Wrong", sm2.GradeWrong, GradeWrongStyle), 20) +
		gradeChip("g", "Hard", sm2.GradeHard, GradeHardStyle)

	row2 := "  " +
		padRightVisual(gradeChip("h", "Good", sm2.GradeGood, GradeGoodStyle), 22) +
		padRightVisual(gradeChip("j", "Easy", sm2.GradeEasy, GradeEasyStyle), 20) +
		gradeChip("k", "Perfect", sm2.GradePerfect, GradePerfectStyle)

	lines = append(lines, row1)
	lines = append(lines, row2)

	return fitBlockHeight(strings.Join(lines, "\n"), reviewBodyHeight)
}

func (m Model) renderReviewSourceBlock(frameWidth int, ref deck.CardRef) string {
	divider := renderDivider(frameWidth)
	displayName, err := filepath.Rel(m.BaseDir, ref.Filename)
	if err != nil {
		displayName = ref.Filename
	}
	wrapped := wrapPathWithIndent("Source: ", displayName, frameWidth, reviewSourceLines)
	return divider + "\n" + fitBlockHeight(wrapped, reviewSourceLines)
}

func (m Model) renderReviewFooter(frameWidth int, isMarked bool) string {
	markAction := "Mark"
	if isMarked {
		markAction = "Unmark"
	}
	var raw string
	if !m.ShowAnswer {
		raw = fmt.Sprintf("[%s] Reveal  •  [%s] %s  •  [%s] Help  •  [%s] Menu  •  [%s] Quit",
			KeyStyle.Render("Space"), KeyStyle.Render("m"), markAction, KeyStyle.Render("?"), KeyStyle.Render("Esc"), KeyStyle.Render("q"))
	} else {
		raw = fmt.Sprintf("[%s] Explanation  •  [%s] %s  •  [%s] Help  •  [%s] Menu  •  [%s] Quit",
			KeyStyle.Render("e"), KeyStyle.Render("m"), markAction, KeyStyle.Render("?"), KeyStyle.Render("Esc"), KeyStyle.Render("q"))
	}
	wrapped := lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(raw))
	return fitBlockHeight(wrapped, reviewFooterHeight)
}

func formatIntervalDays(days int) string {
	if days <= 0 {
		return "1d"
	}
	if days < 14 {
		return fmt.Sprintf("%dd", days)
	}
	if days < 60 {
		return fmt.Sprintf("%dw", (days+3)/7)
	}
	if days < 365 {
		return fmt.Sprintf("%dmo", (days+15)/30)
	}
	return fmt.Sprintf("%.1fy", float64(days)/365.0)
}

func padRightVisual(s string, targetWidth int) string {
	w := lipgloss.Width(s)
	if w >= targetWidth {
		return s + " "
	}
	return s + strings.Repeat(" ", targetWidth-w)
}
