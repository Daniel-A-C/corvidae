package ui

import (
	"fmt"
	"strings"

	"flashcards/internal/deck"

	"github.com/charmbracelet/lipgloss"
)

const (
	matchDefaultFrameWidth = 88
	matchMinFrameWidth     = 44
	matchFeedbackHeight    = 2
	matchFooterHeight      = 2
)

func (m Model) matchFrameWidth() int {
	return responsiveFrameWidth(m.Width, matchDefaultFrameWidth, matchMinFrameWidth)
}

func (m Model) viewMatch() string {
	frameWidth := m.matchFrameWidth()

	if len(m.ActiveCards) == 0 {
		controls := HintStyle.Render(fmt.Sprintf("[%s] Main Menu", KeyStyle.Render("Enter / q")))
		card := renderCompleteCard(frameWidth, "No cards available to match!", "", nil, controls)
		return renderFrame(frameWidth, card)
	}

	totalCards := len(m.ActiveCards)
	if m.MatchRoundStart >= totalCards {
		accStr := "100%"
		if m.MatchAttempts > 0 {
			accStr = fmt.Sprintf("%.1f%%", float64(m.MatchPairsMatched)/float64(m.MatchAttempts)*100)
		}
		barLen := 26
		if frameWidth < 48 {
			barLen = 16
		}
		bar := renderProgressBar(totalCards, totalCards, barLen)
		stats := []string{
			HintStyle.Render(fmt.Sprintf("Matched: %d of %d pairs  •  Accuracy: %s", m.MatchPairsMatched, totalCards, accStr)),
			HintStyle.Render(fmt.Sprintf("Best Streak: %d  •  Mistakes: %d", m.MatchBestStreak, m.MatchMistakes)),
		}
		controls := HintStyle.Render(fmt.Sprintf("[%s] Play again  •  [%s] Main Menu",
			KeyStyle.Render("r"), KeyStyle.Render("Enter / q")))
		card := renderCompleteCard(frameWidth, "Matching Game complete!", bar, stats, controls)
		return renderFrame(frameWidth, card)
	}

	totalRounds := (totalCards + MatchBatchSize - 1) / MatchBatchSize
	currentRound := (m.MatchRoundStart / MatchBatchSize) + 1

	accStr := "100%"
	if m.MatchAttempts > 0 {
		accStr = fmt.Sprintf("%.1f%%", float64(m.MatchPairsMatched)/float64(m.MatchAttempts)*100)
	}

	// Zone 1: Fixed 2-line Header
	leftStats := HintStyle.Render(fmt.Sprintf("Matching Game: Round %d of %d (%d/%d matched)  •  Accuracy: %s",
		currentRound, totalRounds, m.MatchPairsMatched, totalCards, accStr))
	pinyinPillText := "Pinyin: ON"
	if m.MatchHidePinyin {
		pinyinPillText = "Pinyin: OFF"
	}
	headerLine1 := renderSplitLine(leftStats, ModePillStyle.Render(pinyinPillText), frameWidth)

	barLen := 26
	if frameWidth < 50 {
		barLen = 16
	}
	headerLine2 := renderProgressBar(m.MatchPairsMatched, totalCards, barLen) + "  " + formatStreakBadge(m.MatchStreak)
	headerZone := headerLine1 + "\n" + headerLine2

	// Zone 2: Deterministic fixed-width 2-column board (never shifts across rounds or overflows terminal)
	gapWidth := 2
	if frameWidth >= 76 {
		gapWidth = 4
	}
	// Each bordered box adds 2 columns for its left/right border.
	availInner := frameWidth - gapWidth - 4
	if availInner < 24 {
		availInner = 24
	}
	leftInnerWidth := availInner * 36 / 100
	if leftInnerWidth < 12 {
		leftInnerWidth = 12
	}
	rightInnerWidth := availInner - leftInnerWidth
	if rightInnerWidth < 12 {
		rightInnerWidth = 12
	}

	// Content area inside Padding(0, 2) is innerWidth - 4.
	leftContentMax := leftInnerWidth - 4
	if leftContentMax < 8 {
		leftContentMax = 8
	}
	rightContentMax := rightInnerWidth - 4
	if rightContentMax < 8 {
		rightContentMax = 8
	}

	var leftRows []string
	leftHeader := MatchColumnHeaderStyle.Width(leftInnerWidth + 2).Align(lipgloss.Center).Render("Characters [a-g]")
	leftRows = append(leftRows, leftHeader)

	for i, item := range m.MatchLeftItems {
		key := IndexToKey(i)
		star := ""
		starW := 0
		if m.isCardMarked(item.Card) {
			star = MarkedStarStyle.Render(" ★")
			starW = 2
		}
		// Prefix "[a]   " or "[✓]  " takes 6 visual cols.
		charText := truncateToWidth(item.Card.Character, leftContentMax-6-starW)
		var rowContent string
		var boxStyle lipgloss.Style

		switch {
		case item.Matched:
			rowContent = fmt.Sprintf("[✓]  %s%s", charText, star)
			boxStyle = MatchCardMatchedBox
		case m.MatchSelectedLeft == i:
			rowContent = fmt.Sprintf("%s ▶ %s%s", KeyStyle.Render(fmt.Sprintf("[%s]", key)), CursorStyle.Render(charText), star)
			boxStyle = MatchCardSelectedBox
		case m.MatchMismatchLeft == i:
			rowContent = fmt.Sprintf("%s ✗ %s%s", ErrorStyle.Render(fmt.Sprintf("[%s]", key)), ErrorStyle.Render(charText), star)
			boxStyle = MatchCardMismatchBox
		default:
			rowContent = fmt.Sprintf("%s   %s%s", KeyStyle.Render(fmt.Sprintf("[%s]", key)), CharStyle.Render(charText), star)
			boxStyle = MatchCardBox
		}

		leftRows = append(leftRows, boxStyle.Width(leftInnerWidth).Align(lipgloss.Left).Render(rowContent))
	}

	var rightRows []string
	rightTitle := "Pinyin / Meaning [h-;]"
	if m.MatchHidePinyin {
		rightTitle = "Meaning (Pinyin Hidden) [h-;]"
	}
	rightHeader := MatchColumnHeaderStyle.Width(rightInnerWidth + 2).Align(lipgloss.Center).Render(rightTitle)
	rightRows = append(rightRows, rightHeader)

	for i, item := range m.MatchRightItems {
		key := IndexToKey(i + 5)
		rawOpt := item.Card.FormatMatchOption(m.MatchHidePinyin)
		optText := truncateToWidth(rawOpt, rightContentMax-6)
		var rowContent string
		var boxStyle lipgloss.Style

		switch {
		case item.Matched:
			rowContent = fmt.Sprintf("[✓]  %s", optText)
			boxStyle = MatchCardMatchedBox
		case m.MatchSelectedRight == i:
			rowContent = fmt.Sprintf("%s ▶ %s", KeyStyle.Render(fmt.Sprintf("[%s]", key)), CursorStyle.Render(optText))
			boxStyle = MatchCardSelectedBox
		case m.MatchMismatchRight == i:
			rowContent = fmt.Sprintf("%s ✗ %s", ErrorStyle.Render(fmt.Sprintf("[%s]", key)), ErrorStyle.Render(optText))
			boxStyle = MatchCardMismatchBox
		default:
			styledOpt := m.formatStyledMatchOptionTruncated(item.Card, rightContentMax-6)
			rowContent = fmt.Sprintf("%s   %s", KeyStyle.Render(fmt.Sprintf("[%s]", key)), styledOpt)
			boxStyle = MatchCardBox
		}

		rightRows = append(rightRows, boxStyle.Width(rightInnerWidth).Align(lipgloss.Left).Render(rowContent))
	}

	leftCol := lipgloss.JoinVertical(lipgloss.Left, leftRows...)
	rightCol := lipgloss.JoinVertical(lipgloss.Left, rightRows...)
	board := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, strings.Repeat(" ", gapWidth), rightCol)

	// Zone 3: Divider + Fixed 2-line Feedback Zone
	var feedbackStyled string
	if m.MatchLastFeedback != "" {
		fbLines := wrapPlainText(m.MatchLastFeedback, frameWidth)
		var styledLines []string
		for _, fl := range fbLines {
			if m.MatchLastIsError {
				styledLines = append(styledLines, ErrorStyle.Render(fl))
			} else {
				styledLines = append(styledLines, CorrectStyle.Render(fl))
			}
		}
		feedbackStyled = strings.Join(styledLines, "\n")
	} else {
		feedbackStyled = HintStyle.Render("Select a character [a-g] and its matching meaning [h-;]")
	}
	feedbackZone := renderDivider(frameWidth) + "\n" + fitBlockHeight(feedbackStyled, matchFeedbackHeight)

	// Zone 4: Fixed 2-line Footer
	pinyinToggleLabel := "Hide Pinyin"
	if m.MatchHidePinyin {
		pinyinToggleLabel = "Show Pinyin"
	}
	controlsRaw := fmt.Sprintf("[%s] Left  •  [%s] Right  •  [%s] %s  •  [%s] Mark  •  [%s] Help  •  [%s] Menu",
		KeyStyle.Render("a-g"), KeyStyle.Render("h-;"), KeyStyle.Render("p"), pinyinToggleLabel,
		KeyStyle.Render("m"), KeyStyle.Render("?"), KeyStyle.Render("q"))
	footerWrapped := lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(controlsRaw))
	footerZone := fitBlockHeight(footerWrapped, matchFooterHeight)

	assembled := lipgloss.JoinVertical(
		lipgloss.Left,
		headerZone,
		"",
		board,
		"",
		feedbackZone,
		footerZone,
	)

	return renderFrame(frameWidth, assembled)
}

func (m Model) formatStyledMatchOptionTruncated(card deck.Flashcard, maxWidth int) string {
	raw := card.FormatMatchOption(m.MatchHidePinyin)
	truncated := truncateToWidth(raw, maxWidth)
	if m.MatchHidePinyin {
		return MeaningStyle.Render(truncated)
	}
	if phonetic, meaning, ok := strings.Cut(truncated, " - "); ok {
		if card.Pinyin != "" {
			return fmt.Sprintf("%s - %s", PinyinStyle.Render(phonetic), MeaningStyle.Render(meaning))
		}
		return fmt.Sprintf("%s - %s", PronunciationStyle.Render(phonetic), MeaningStyle.Render(meaning))
	}
	if card.Pinyin != "" {
		return PinyinStyle.Render(truncated)
	}
	return MeaningStyle.Render(truncated)
}
