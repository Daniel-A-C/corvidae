package ui

import (
	"fmt"
	"strings"
	"time"

	"flashcards/internal/memorize"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) memorizeModePill() string {
	if m.MemorizeTypingMode {
		return SyncOnPillStyle.Render("Mode: First-Letter")
	}
	return ModePillStyle.Render(fmt.Sprintf("Options: %d", m.MemorizeOptionCount))
}

func (m Model) viewMemorizeSelectText() string {
	frameWidth := m.frameWidth()
	if len(m.MemorizeTexts) == 0 {
		empty := fmt.Sprintf("No memorization texts found in %s.\n\nPlace .yaml or .txt files in %s/\n\n%s",
			m.TextsDir, m.TextsDir, HintStyle.Render("[Esc] Return  •  [q] Quit"))
		return renderFrame(frameWidth, empty)
	}

	var b strings.Builder
	b.WriteString(renderSplitLine(MemorizeTitleStyle.Render("★ MEMORIZE TEXTS ★"), m.memorizeModePill(), frameWidth) + "\n")
	b.WriteString(MemorizeSubStyle.Render("Recite word by word via Options or First-Letter Typing with section SRS") + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n\n")

	for i, t := range m.MemorizeTexts {
		key := IndexToKey(i)
		keyBadge := ""
		if key != "" {
			keyBadge = KeyStyle.Render(fmt.Sprintf("[%s]", key)) + " "
		}
		cursor := "  "
		label := t.Title
		if m.MemorizeCursor == i {
			cursor = "> "
			label = CursorStyle.Render(label)
		}

		info := HintStyle.Render(fmt.Sprintf("%d words", t.WordCount))
		if t.TotalSections > 0 {
			if t.DueSections > 0 {
				info += "  •  " + DueBadgeStyle.Render(fmt.Sprintf("%d/%d due", t.DueSections, t.TotalSections))
			} else {
				info += "  •  " + CorrectStyle.Render("✓ caught up")
			}
		}
		leftPart := fmt.Sprintf("%s%s%s", CursorStyle.Render(cursor), keyBadge, label)
		b.WriteString(renderSplitLine(leftPart, info, frameWidth) + "\n")
	}

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	footerRaw := fmt.Sprintf("[%s] Select  •  [%s] Typing/Options  •  [%s] Options (%d)  •  [%s] Help  •  [%s] Back",
		KeyStyle.Render("key/Enter"), KeyStyle.Render("t"), KeyStyle.Render("o/2-5"), m.MemorizeOptionCount, KeyStyle.Render("?"), KeyStyle.Render("Esc"))
	b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))
	return renderFrame(frameWidth, b.String())
}

func (m Model) viewMemorizePortionSelect() string {
	frameWidth := m.frameWidth()
	if m.MemorizeActiveText == nil || len(m.MemorizeSections) == 0 {
		return renderFrame(frameWidth, fmt.Sprintf("No sections found for this text.\n\n%s", HintStyle.Render("[Esc] Back")))
	}

	now := time.Now()
	startSec := m.MemorizeSections[m.PortionStartSection]
	endSec := m.MemorizeSections[m.PortionEndSection]
	portionWords := endSec.EndIdx - startSec.StartIdx
	totalWords := len(m.MemorizeActiveText.Words)
	numSections := m.PortionEndSection - m.PortionStartSection + 1
	dueCount := m.MemorizeActiveText.DueSectionCount(now)

	var b strings.Builder
	b.WriteString(renderSplitLine(MemorizeTitleStyle.Render("★ PORTION SELECTOR ★"), m.memorizeModePill(), frameWidth) + "\n")
	duePill := CorrectStyle.Render("✓ All sections caught up")
	if dueCount > 0 {
		duePill = DueBadgeStyle.Render(fmt.Sprintf("%d of %d sections due", dueCount, len(m.MemorizeSections)))
	}
	b.WriteString(renderSplitLine(MemorizeHeaderStyle.Render(m.MemorizeActiveText.Title), duePill, frameWidth) + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n\n")

	if m.PortionCompletedMsg != "" {
		b.WriteString(CorrectStyle.Render(m.PortionCompletedMsg) + "\n\n")
	}

	rangeLabel := startSec.Label
	if m.PortionStartSection != m.PortionEndSection {
		rangeLabel = fmt.Sprintf("%s to %s", startSec.Label, endSec.Label)
	}

	pctRange := 0.0
	if totalWords > 0 {
		pctRange = float64(portionWords) / float64(totalWords) * 100
	}
	summary := fmt.Sprintf("Selected: %s  •  %d sections  •  %d of %d words (%.1f%%)",
		rangeLabel, numSections, portionWords, totalWords, pctRange)
	b.WriteString(MemorizeSubStyle.Render(summary) + "\n")

	rangeBarLen := 26
	if frameWidth < 46 {
		rangeBarLen = 16
	}
	totalSecs := len(m.MemorizeSections)
	var rangeBar strings.Builder
	for pos := 0; pos < rangeBarLen; pos++ {
		secIdx := pos * totalSecs / rangeBarLen
		if secIdx >= m.PortionStartSection && secIdx <= m.PortionEndSection {
			rangeBar.WriteString("█")
		} else {
			rangeBar.WriteString("░")
		}
	}
	b.WriteString(MemorizeProgressStyle.Render("["+rangeBar.String()+"]") + "\n\n")

	cursorStart := "  "
	cursorEnd := "  "
	startBadge := "Start"
	endBadge := "End"

	if m.PortionFocus == 0 {
		cursorStart = "> "
		startBadge = CursorStyle.Render("▶ START")
	} else {
		cursorEnd = "> "
		endBadge = CursorStyle.Render("▶ END")
	}

	startVal := KeyStyle.Render(fmt.Sprintf("< [ %s ] >", startSec.Label))
	endVal := KeyStyle.Render(fmt.Sprintf("< [ %s ] >", endSec.Label))
	previewMax := frameWidth - 34
	if previewMax < 12 {
		previewMax = 12
	}

	b.WriteString(fmt.Sprintf("%s[%s]: %s  %s\n",
		cursorStart, startBadge, startVal, HintStyle.Render("\""+truncateToWidth(startSec.Preview, previewMax)+"\"")))
	b.WriteString(fmt.Sprintf("%s[%s]:   %s  %s\n\n",
		cursorEnd, endBadge, endVal, HintStyle.Render("\""+truncateToWidth(endSec.Preview, previewMax)+"\"")))

	b.WriteString(SectionTitleStyle.Render("Section Overview") + "\n")
	windowStart := m.PortionStartSection - 2
	if windowStart < 0 {
		windowStart = 0
	}
	windowEnd := windowStart + 7
	if windowEnd >= len(m.MemorizeSections) {
		windowEnd = len(m.MemorizeSections) - 1
		windowStart = windowEnd - 7
		if windowStart < 0 {
			windowStart = 0
		}
	}

	for i := windowStart; i <= windowEnd; i++ {
		sec := m.MemorizeSections[i]
		inRange := (i >= m.PortionStartSection && i <= m.PortionEndSection)
		prefix := "    "

		srsBadge := ""
		if rev, ok := m.MemorizeActiveText.SectionReviewFor(sec); ok && rev.NextReview != "" {
			if m.MemorizeActiveText.IsSectionDue(sec, now) {
				srsBadge = DueBadgeStyle.Render("[DUE]")
			} else {
				srsBadge = CorrectStyle.Render("[" + formatIntervalDays(rev.Interval) + "]")
			}
		} else {
			srsBadge = HintStyle.Render("[New]")
		}

		secPreview := truncateToWidth(sec.Preview, frameWidth-42)
		line := fmt.Sprintf("%-10s (%2d words): %s", sec.Label, sec.WordCount, secPreview)
		tag := ""
		if i == m.PortionStartSection && i == m.PortionEndSection {
			tag = " " + MarkedBadgeStyle.Render("PORTION")
		} else if i == m.PortionStartSection {
			tag = " " + KeyStyle.Render("[START]")
		} else if i == m.PortionEndSection {
			tag = " " + KeyStyle.Render("[END]")
		}

		if inRange {
			prefix = "  ► "
			b.WriteString(renderSplitLine(
				fmt.Sprintf("%s%s%s", CursorStyle.Render(prefix), MemorizeCompleted.Render(line), tag),
				srsBadge, frameWidth) + "\n")
		} else {
			b.WriteString(renderSplitLine(
				fmt.Sprintf("%s%s", prefix, HintStyle.Render(line)),
				srsBadge, frameWidth) + "\n")
		}
	}

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	b.WriteString(HintStyle.Render("[Tab/↑↓] Start/End  •  [←→/hl] Shift  •  [x] Expand  •  [n] Next  •  [d] Next Due") + "\n")
	b.WriteString(HintStyle.Render(fmt.Sprintf("[t] Typing/Options  •  [o/2-5] Options: %d  •  [a] All  •  [Enter] Practice  •  [Esc] Back", m.MemorizeOptionCount)))
	return renderFrame(frameWidth, b.String())
}

func (m Model) viewMemorize() string {
	frameWidth := m.frameWidth()
	if m.MemorizeActiveText == nil || m.MemorizeCurrentIndex >= m.PortionEndWordIdx {
		return m.viewMemorizeComplete()
	}

	totalPortionWords := m.PortionEndWordIdx - m.PortionStartWordIdx
	currentInPortion := m.MemorizeCurrentIndex - m.PortionStartWordIdx
	pct := 0.0
	if totalPortionWords > 0 {
		pct = float64(currentInPortion) / float64(totalPortionWords) * 100
	}
	attempts := currentInPortion + m.MemorizeMistakes
	accStr := "100%"
	if attempts > 0 {
		accStr = fmt.Sprintf("%.1f%%", float64(currentInPortion)/float64(attempts)*100)
	}

	barLen := 26
	if frameWidth < 46 {
		barLen = 16
	}

	var portionTitle string
	if len(m.MemorizeSections) > 0 && m.PortionStartSection < len(m.MemorizeSections) && m.PortionEndSection < len(m.MemorizeSections) {
		startLabel := m.MemorizeSections[m.PortionStartSection].Label
		endLabel := m.MemorizeSections[m.PortionEndSection].Label
		if m.PortionStartSection == m.PortionEndSection {
			portionTitle = fmt.Sprintf(" (%s)", startLabel)
		} else {
			portionTitle = fmt.Sprintf(" (%s–%s)", startLabel, endLabel)
		}
	}

	optPill := m.memorizeModePill()
	titleLeft := MemorizeHeaderStyle.Render(fmt.Sprintf("★ %s%s", m.MemorizeActiveText.Title, portionTitle))
	headerLine1 := renderSplitLine(titleLeft, optPill, frameWidth)
	headerLine2 := HintStyle.Render(fmt.Sprintf("Portion Word %d of %d (%.1f%%)  •  Accuracy: %s",
		currentInPortion+1, totalPortionWords, pct, accStr))
	headerLine3 := renderProgressBar(currentInPortion, totalPortionWords, barLen) + "  " + formatStreakBadge(m.MemorizeStreak)

	maxContentLines := 4
	boxWidth := frameWidth - 2
	if boxWidth > 66 {
		boxWidth = 66
	}
	if boxWidth < 24 {
		boxWidth = 24
	}
	padX := 2
	padY := 1
	boxHeight := maxContentLines + (padY * 2)

	startIdx := 0
	if m.MemorizeCurrentIndex > 30 {
		startIdx = m.MemorizeCurrentIndex - 30
	}
	var passage strings.Builder
	for i := startIdx; i < m.MemorizeCurrentIndex; i++ {
		item := m.MemorizeActiveText.Words[i]
		pfx := strings.TrimSpace(strings.ReplaceAll(item.Prefix, "\n", " "))
		if pfx != "" {
			if passage.Len() > 0 {
				passage.WriteString(" ")
			}
			passage.WriteString(pfx)
		}
		if passage.Len() > 0 {
			passage.WriteString(" ")
		}
		passage.WriteString(item.Word)
	}

	activeItem := m.MemorizeActiveText.Words[m.MemorizeCurrentIndex]
	activePfx := strings.TrimSpace(strings.ReplaceAll(activeItem.Prefix, "\n", " "))
	if activePfx != "" {
		if passage.Len() > 0 {
			passage.WriteString(" ")
		}
		passage.WriteString(activePfx)
	}
	if passage.Len() > 0 {
		passage.WriteString(" ")
	}
	passage.WriteString("[TARGET_BLANK]")

	blankLabel := "  ?  "
	if m.MemorizeTypingMode {
		blankLabel = "  _  "
	}
	parts := strings.Split(passage.String(), "[TARGET_BLANK]")
	renderedPassage := MemorizeCompleted.Render(parts[0]) + MemorizeTargetBlank.Render(blankLabel)
	if len(parts) > 1 {
		renderedPassage += MemorizeCompleted.Render(parts[1])
	}

	cardBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#BD93F9")).
		Padding(padY, padX).
		Width(boxWidth).
		Height(boxHeight).
		Align(lipgloss.Left)

	renderedBox := cardBox.Render(renderedPassage)
	lines := strings.Split(renderedBox, "\n")

	targetTotal := maxContentLines + (padY * 2) + 2
	if len(lines) > targetTotal {
		topBorder := lines[0]
		topPad := lines[1]
		botPad := lines[len(lines)-2]
		botBorder := lines[len(lines)-1]

		contentLines := lines[2 : len(lines)-2]
		keptContent := contentLines[len(contentLines)-maxContentLines:]

		result := []string{topBorder, topPad}
		result = append(result, keptContent...)
		result = append(result, botPad, botBorder)
		renderedBox = strings.Join(result, "\n")
	}

	const memorizeBodyHeight = 8
	var bodyLines []string
	var footerRaw string
	if m.MemorizeShowFeedback {
		if m.MemorizeTypingMode {
			expectedRune := memorize.FirstLetter(activeItem.Word)
			if m.MemorizeTypedWrong != "" {
				bodyLines = append(bodyLines, "  "+MemorizeWrongStyle.Render(fmt.Sprintf("✗ You typed '%s' — expected '%c'", m.MemorizeTypedWrong, expectedRune)))
			} else {
				bodyLines = append(bodyLines, "  "+HintStyle.Render("Peeked word:"))
			}
			bodyLines = append(bodyLines, "  "+MemorizeSuccessStyle.Render(fmt.Sprintf("✓ Word: \"%s\"", activeItem.Word)))
			footerRaw = fmt.Sprintf("[%s] Continue  •  [%s] I knew it (typo)  •  [%s] Retry word  •  [%s] Switch to Options",
				KeyStyle.Render("Space/Enter"), KeyStyle.Render("k"), KeyStyle.Render("r"), KeyStyle.Render("Tab"))
		} else {
			wrongWord := ""
			if m.MemorizeSelectedOption >= 0 && m.MemorizeSelectedOption < len(m.MemorizeCurrentOptions) {
				wrongWord = m.MemorizeCurrentOptions[m.MemorizeSelectedOption]
			}
			bodyLines = append(bodyLines, "  "+MemorizeWrongStyle.Render(fmt.Sprintf("✗ Incorrect: \"%s\"", wrongWord)))
			bodyLines = append(bodyLines, "  "+MemorizeSuccessStyle.Render(fmt.Sprintf("✓ The correct word was: \"%s\"", activeItem.Word)))
			footerRaw = fmt.Sprintf("[%s] Continue  •  [%s] I knew it (wrong button)  •  [%s] Retry word  •  [%s] Options",
				KeyStyle.Render("Spacebar / Enter"), KeyStyle.Render("k"), KeyStyle.Render("r"), KeyStyle.Render("o"))
		}
	} else if m.MemorizeTypingMode {
		bodyLines = append(bodyLines, SectionTitleStyle.Render("First-Letter Recitation"))
		bodyLines = append(bodyLines, "")
		bodyLines = append(bodyLines, "  "+MemorizeCompleted.Render("Type the first letter of the next word"))
		bodyLines = append(bodyLines, "  "+HintStyle.Render("(Case & accent insensitive — e.g. 'e' matches \"Ésta\")"))
		footerRaw = fmt.Sprintf("[%s] Type first letter  •  [%s] Peek  •  [%s] Switch to Options  •  [%s] Back",
			KeyStyle.Render("a-z"), KeyStyle.Render("Space"), KeyStyle.Render("Tab"), KeyStyle.Render("Esc"))
	} else {
		bodyLines = append(bodyLines, HintStyle.Render(fmt.Sprintf("Choose the next word (%d options):", len(m.MemorizeCurrentOptions))))
		bodyLines = append(bodyLines, "")
		for i, opt := range m.MemorizeCurrentOptions {
			key := IndexToKey(i)
			keyBadge := KeyStyle.Render(fmt.Sprintf("[%s]", key))
			numBadge := HintStyle.Render(fmt.Sprintf("(%d)", i+1))
			optRender := MemorizeCompleted.Render(opt)
			bodyLines = append(bodyLines, fmt.Sprintf("  %s %s  %s", keyBadge, numBadge, optRender))
		}
		var keyList []string
		for i := 0; i < len(m.MemorizeCurrentOptions); i++ {
			keyList = append(keyList, IndexToKey(i))
		}
		keyStr := strings.Join(keyList, "/")
		numStr := fmt.Sprintf("1-%d", len(m.MemorizeCurrentOptions))
		footerRaw = fmt.Sprintf("(Press key [%s] or [%s] to select  •  [t] Typing  •  [o] Options (%d)  •  [r] Restart  •  [Esc] Back)",
			keyStr, numStr, len(m.MemorizeCurrentOptions))
	}
	bodyZone := fitBlockHeight(strings.Join(bodyLines, "\n"), memorizeBodyHeight)
	footerZone := renderDivider(frameWidth) + "\n" +
		fitBlockHeight(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)), 2)

	assembled := lipgloss.JoinVertical(
		lipgloss.Left,
		headerLine1,
		headerLine2,
		headerLine3,
		"",
		renderedBox,
		"",
		bodyZone,
		footerZone,
	)
	return renderFrame(frameWidth, assembled)
}

func (m Model) viewMemorizeComplete() string {
	frameWidth := m.frameWidth()
	totalWords := 0
	textTitle := ""
	if m.MemorizeActiveText != nil {
		totalWords = len(m.MemorizeActiveText.Words)
		textTitle = m.MemorizeActiveText.Title
	}
	attempts := totalWords + m.MemorizeMistakes
	acc := 100.0
	if attempts > 0 {
		acc = float64(totalWords) / float64(attempts) * 100
	}
	barLen := 26
	if frameWidth < 46 {
		barLen = 16
	}
	bar := renderProgressBar(totalWords, totalWords, barLen)
	var stats []string
	if textTitle != "" {
		stats = append(stats, MemorizeHeaderStyle.Render(textTitle))
	}
	stats = append(stats,
		HintStyle.Render(fmt.Sprintf("Total Words: %d  •  First-Try Acc: %.1f%%", totalWords, acc)),
		HintStyle.Render(fmt.Sprintf("Mistakes: %d  •  Best Streak: %d words  •  Options: %d", m.MemorizeMistakes, m.MemorizeBestStreak, m.MemorizeOptionCount)),
	)
	controls := HintStyle.Render(fmt.Sprintf("[%s] Options (%d)  •  [%s] Again  •  [%s] Texts  •  [%s] Quit",
		KeyStyle.Render("o / 2-5"), m.MemorizeOptionCount, KeyStyle.Render("r"), KeyStyle.Render("Enter"), KeyStyle.Render("q")))

	return renderFrame(frameWidth, renderCompleteCard(frameWidth, "★ TEXT MEMORIZATION COMPLETE! ★", bar, stats, controls))
}
