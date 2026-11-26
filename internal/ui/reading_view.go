package ui

import (
	"fmt"
	"strings"

	"flashcards/internal/reading"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewReadingSelectText() string {
	frameWidth := responsiveFrameWidth(m.Width, 78, 44)
	var b strings.Builder
	b.WriteString(ReadingHeaderStyle.Render("★ Corvidae: Learn by Reading (Diglot Weave) ★") + "\n")
	b.WriteString(HintStyle.Render("Read literature and scripture with progressive target language vocabulary immersion.") + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n\n")

	if len(m.ReadingTexts) == 0 {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("No texts found in %s.", m.ReadingDir)) + "\n\n")
		b.WriteString(HintStyle.Render("[Esc] Return to Mode Select  •  [q] Quit"))
		return renderFrame(frameWidth, b.String())
	}

	b.WriteString(SectionTitleStyle.Render("Select a Text to Read") + "\n\n")
	for i, t := range m.ReadingTexts {
		key := IndexToKey(i)
		keyBadge := ""
		if key != "" {
			keyBadge = KeyStyle.Render(fmt.Sprintf("[%s]", key)) + " "
		}

		cursor := "  "
		label := t.Title
		if m.ReadingCursor == i {
			cursor = "> "
			label = CursorStyle.Render(label)
		}

		catBadge := ReadingCategoryBadge.Render(t.Category)
		langBadge := PinyinStyle.Render("[" + t.Language + "]")
		progressStr := HintStyle.Render(fmt.Sprintf("(%d sentences)", t.SentenceCount))
		if t.Progress > 0 && t.SentenceCount > 0 {
			pct := float64(t.Progress) / float64(t.SentenceCount) * 100
			progressStr = HintStyle.Render(fmt.Sprintf("(Sentence %d of %d • %.0f%%)", t.Progress+1, t.SentenceCount, pct))
		}

		leftPart := fmt.Sprintf("%s%s%s  %s %s", CursorStyle.Render(cursor), keyBadge, label, catBadge, langBadge)
		b.WriteString(renderSplitLine(leftPart, progressStr, frameWidth) + "\n")
	}

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	footerRaw := fmt.Sprintf("[%s] Select  •  [%s] Start Reading  •  [%s] Help  •  [%s] Back  •  [%s] Quit",
		KeyStyle.Render("a/s"), KeyStyle.Render("Enter/Space"), KeyStyle.Render("?"), KeyStyle.Render("Esc"), KeyStyle.Render("q"))
	b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))
	return renderFrame(frameWidth, b.String())
}

func (m Model) viewReading() string {
	if m.ReadingSession == nil || m.ReadingSession.Text == nil {
		return m.viewReadingSelectText()
	}

	session := m.ReadingSession
	_, exists := session.CurrentSentence()
	if !exists {
		return m.viewReadingComplete()
	}

	frameWidth := responsiveFrameWidth(m.Width, 78, 44)
	boxInnerWidth := frameWidth - 2
	if boxInnerWidth < 24 {
		boxInnerWidth = 24
	}

	totalSentences := len(session.Text.Sentences)
	currentSentNum := session.SentenceIdx + 1
	pct := 0.0
	if totalSentences > 0 {
		pct = float64(currentSentNum) / float64(totalSentences) * 100
	}

	titleLine := fmt.Sprintf("📖 %s", session.Text.Title)
	if session.Text.Source != "" {
		titleLine += fmt.Sprintf(" (%s)", session.Text.Source)
	}
	langPill := ModePillStyle.Render(session.Text.Language)
	headerLine1 := renderSplitLine(ReadingHeaderStyle.Render(truncateToWidth(titleLine, frameWidth-lipgloss.Width(langPill)-2)), langPill, frameWidth)

	barLen := 20
	if frameWidth < 54 {
		barLen = 12
	}
	filled := session.Aggressiveness
	if filled > 5 {
		filled = 5
	}
	meter := strings.Repeat("█", filled) + strings.Repeat("░", 5-filled)
	aggLabel := reading.AggressivenessLabels[session.Aggressiveness]
	headerLine2 := renderProgressBar(currentSentNum, totalSentences, barLen) + "  " +
		HintStyle.Render(fmt.Sprintf("Sentence %d of %d (%.0f%%)", currentSentNum, totalSentences, pct))
	headerLine3 := HintStyle.Render("Aggressiveness: ") +
		ReadingAggressivenessStyle.Render("["+meter+"] "+aggLabel) +
		HintStyle.Render("  [-/+]")

	var sentenceRender strings.Builder
	for _, seg := range session.CurrentWeave.Segments {
		if !seg.IsSubstituted {
			sentenceRender.WriteString(MemorizeCompleted.Render(seg.Text))
		} else {
			keyTag := ReadingKeyBadge.Render("[" + seg.SubstitutedInfo.Key + "]")
			wordTag := ReadingWordSubstituted.Render(seg.SubstitutedInfo.Translation.Target)
			sentenceRender.WriteString(" " + keyTag + " " + wordTag + " ")
		}
	}

	readingBox := ReadingBoxStyle.Width(boxInnerWidth).Render(strings.TrimSpace(sentenceRender.String()))

	var legendLine string
	if len(session.CurrentWeave.SubstitutedWords) > 0 {
		var wordBadges []string
		for _, sw := range session.CurrentWeave.SubstitutedWords {
			badge := ReadingKeyBadge.Render(sw.Key) + ":" + ReadingWordSubstituted.Render(sw.Translation.Target)
			wordBadges = append(wordBadges, badge)
		}
		legendLine = HintStyle.Render("Inspect word: ") + strings.Join(wordBadges, "  ")
	} else {
		legendLine = HintStyle.Render("Pure English mode — press [+] to weave vocabulary")
	}

	footerRaw := fmt.Sprintf("[%s] Next  •  [%s] Grammar  •  [%s] Prev  •  [%s] Tune  •  [%s] Pure Eng  •  [%s] Texts",
		KeyStyle.Render("Space/Enter"), KeyStyle.Render("h"), KeyStyle.Render("p"),
		KeyStyle.Render("-/+"), KeyStyle.Render("0"), KeyStyle.Render("Esc"))
	footerZone := renderDivider(frameWidth) + "\n" +
		lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw))

	assembled := lipgloss.JoinVertical(
		lipgloss.Left,
		headerLine1,
		headerLine2,
		headerLine3,
		"",
		readingBox,
		"",
		lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(legendLine),
		"",
		footerZone,
	)
	return renderFrame(frameWidth, assembled)
}

func (m Model) viewReadingWordDetail() string {
	if m.ReadingSession == nil || m.ReadingSession.SelectedWord == nil {
		return m.viewReading()
	}

	frameWidth := responsiveFrameWidth(m.Width, 72, 42)
	boxInnerWidth := frameWidth - 2
	if boxInnerWidth < 24 {
		boxInnerWidth = 24
	}
	contentWidth := boxInnerWidth - 6 // Padding(1, 3)
	if contentWidth < 18 {
		contentWidth = 18
	}

	sw := m.ReadingSession.SelectedWord
	trans := sw.Translation

	var b strings.Builder
	b.WriteString(ReadingHeaderStyle.Render("★ VOCABULARY DETAIL & RECALL ★") + "\n")
	b.WriteString(renderDivider(contentWidth) + "\n\n")

	b.WriteString(CharStyle.Render(trans.Target) + "\n")
	if trans.Pinyin != "" {
		b.WriteString(PinyinStyle.Render(trans.Pinyin) + "\n")
	} else if trans.Pronunciation != "" {
		b.WriteString(PronunciationStyle.Render(trans.Pronunciation) + "\n")
	}

	b.WriteString(MeaningStyle.Render("“"+trans.Meaning+"”") + "\n\n")

	if trans.Explanation != "" {
		b.WriteString(QuizExplanationStyle.Width(contentWidth).Render(trans.Explanation) + "\n\n")
	}

	b.WriteString(HintStyle.Render(fmt.Sprintf("Original English: \"%s\"  •  Tracking: Marked as Learning", sw.Original)) + "\n")

	if m.ReadingWordFeedback != "" {
		b.WriteString("\n" + CorrectStyle.Render(m.ReadingWordFeedback) + "\n")
	}

	b.WriteString("\n" + renderDivider(contentWidth) + "\n")
	footerRaw := fmt.Sprintf("[%s] Add to Marked  •  [%s] I Know This  •  [%s] Return",
		KeyStyle.Render("m"), KeyStyle.Render("k"), KeyStyle.Render("Space / Esc"))
	b.WriteString(HintStyle.Render(footerRaw))

	return renderFrame(frameWidth, ReadingWordDetailCard.Width(boxInnerWidth).Render(b.String()))
}

func (m Model) viewReadingDiscussion() string {
	if m.ReadingSession == nil {
		return m.viewReading()
	}

	session := m.ReadingSession
	sent, exists := session.CurrentSentence()
	if !exists {
		return m.viewReadingComplete()
	}

	frameWidth := responsiveFrameWidth(m.Width, 78, 44)
	boxInnerWidth := frameWidth - 2
	if boxInnerWidth < 24 {
		boxInnerWidth = 24
	}
	contentWidth := boxInnerWidth - 6 // Padding(1, 3)
	if contentWidth < 18 {
		contentWidth = 18
	}

	var b strings.Builder
	b.WriteString(ReadingHeaderStyle.Render("★ SENTENCE GRAMMAR & TRANSLATION DISCUSSION ★") + "\n")
	b.WriteString(renderDivider(contentWidth) + "\n\n")

	b.WriteString(KeyStyle.Render("English:") + "\n")
	b.WriteString(lipgloss.NewStyle().Width(contentWidth).Render(MemorizeCompleted.Render(sent.English)) + "\n\n")

	if sent.NaturalTarget != "" {
		b.WriteString(ReadingWordSubstituted.Render(fmt.Sprintf("Natural %s (Word Order):", session.Text.Language)) + "\n")
		b.WriteString(CharStyle.Render(sent.NaturalTarget) + "\n")
		if sent.TargetPinyin != "" {
			b.WriteString(ReadingWordPinyin.Render(sent.TargetPinyin) + "\n")
		}
		b.WriteString("\n")
	}

	if sent.GrammarNote != "" {
		b.WriteString(ReadingAggressivenessStyle.Render("Grammar & Contrast Notes:") + "\n")
		b.WriteString(QuizExplanationStyle.Width(contentWidth).Render(sent.GrammarNote) + "\n\n")
	}

	if len(sent.Words) > 0 {
		b.WriteString(ReadingSourceStyle.Render("Sentence Vocabulary:") + "\n")
		var vocabLines []string
		for _, w := range sent.Words {
			pinyinPart := ""
			if w.Pinyin != "" {
				pinyinPart = fmt.Sprintf(" (%s)", w.Pinyin)
			}
			line := fmt.Sprintf("• %s%s : %s",
				ReadingWordSubstituted.Render(w.Target),
				PinyinStyle.Render(pinyinPart),
				MeaningStyle.Render(w.Meaning),
			)
			vocabLines = append(vocabLines, line)
		}
		b.WriteString(strings.Join(vocabLines, "\n") + "\n\n")
	}

	b.WriteString(renderDivider(contentWidth) + "\n")
	footerRaw := fmt.Sprintf("[%s] Next Sentence  •  [%s] Back to Reading  •  [%s] Prev  •  [%s] Aggressiveness",
		KeyStyle.Render("Space/Enter/n"), KeyStyle.Render("h/Esc"), KeyStyle.Render("p"), KeyStyle.Render("-/+"))
	b.WriteString(HintStyle.Render(footerRaw))

	return renderFrame(frameWidth, ReadingDiscussionCard.Width(boxInnerWidth).Render(b.String()))
}

func (m Model) viewReadingComplete() string {
	frameWidth := m.frameWidth()
	totalSentences := 0
	textTitle := ""
	if m.ReadingSession != nil && m.ReadingSession.Text != nil {
		totalSentences = len(m.ReadingSession.Text.Sentences)
		textTitle = m.ReadingSession.Text.Title
	}
	barLen := 26
	if frameWidth < 46 {
		barLen = 16
	}
	bar := renderProgressBar(totalSentences, totalSentences, barLen)
	var stats []string
	if textTitle != "" {
		stats = append(stats, ReadingHeaderStyle.Render(textTitle))
	}
	stats = append(stats,
		HintStyle.Render(fmt.Sprintf("Total Sentences Completed: %d", totalSentences)),
		HintStyle.Render("Your vocabulary recall and progress have been saved."),
	)
	controls := HintStyle.Render(fmt.Sprintf("[%s] Read Again  •  [%s] Select Another Text  •  [%s] Quit",
		KeyStyle.Render("r"), KeyStyle.Render("Enter / Space"), KeyStyle.Render("q")))

	return renderFrame(frameWidth, renderCompleteCard(frameWidth, "★ TEXT READING COMPLETE! ★", bar, stats, controls))
}
