package ui

import (
	"fmt"
	"strings"

	"flashcards/internal/arabic"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewArabicMenu() string {
	frameWidth := m.frameWidth()
	var b strings.Builder
	b.WriteString(ArabicHeaderStyle.Render("★ ARABIC ALPHABET ACADEMY ★") + "\n")
	b.WriteString(ArabicSubHeaderStyle.Render("أكاديمية الحروف العربية — Master the Modern Standard Arabic Script") + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n\n")
	b.WriteString(SectionTitleStyle.Render("Select Learning Activity") + "\n\n")

	activities := []struct {
		title string
		desc  string
	}{
		{"Alphabet Explorer & Letter Inspector", "28-letter browser with 4 cursive shapes, sounds & examples"},
		{"Positional Forms Drill", "Master Isolated, Initial, Medial & Final cursive recognition"},
		{"Letter Sounds & Transliteration Quiz", "Connect letters, names, IPA sounds & transliterations"},
		{"Minimal Pairs & Confusable Letters", "Contrast tricky letters (ت/ط, د/ض, س/ص, ذ/ظ, ح/ه, ع/غ)"},
		{"Non-Connecting Letters Drill", "Drill the 6 non-connecting letters (ا, د, ذ, ر, ز, و)"},
		{"Sun & Moon Letters Drill", "Master definite article (الـ) assimilation & pronunciation"},
		{"Progressive Guided Lessons", "Step-by-step curriculum through 8 pedagogical stages"},
	}

	for i, act := range activities {
		key := IndexToKey(i)
		keyBadge := KeyStyle.Render(fmt.Sprintf("[%s]", key))
		cursor := "  "
		label := act.title
		if m.ArabicMenuCursor == i {
			cursor = "> "
			label = CursorStyle.Render(label)
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", CursorStyle.Render(cursor), keyBadge, label))
		b.WriteString(fmt.Sprintf("      %s\n", HintStyle.Render(truncateToWidth(act.desc, frameWidth-6))))
		if i == 0 || i == 3 || i == 5 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	footerRaw := fmt.Sprintf("[%s] Select  •  [%s] Confirm  •  [%s] Help  •  [%s] Main Menu  •  [%s] Quit",
		KeyStyle.Render("a-j"), KeyStyle.Render("Enter"), KeyStyle.Render("?"), KeyStyle.Render("Esc"), KeyStyle.Render("q"))
	b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))
	return renderFrame(frameWidth, b.String())
}

func (m Model) viewArabicExplorer() string {
	frameWidth := responsiveFrameWidth(m.Width, 78, 44)
	boxInnerWidth := frameWidth - 2
	if boxInnerWidth < 24 {
		boxInnerWidth = 24
	}
	contentWidth := boxInnerWidth - 4 // ArabicCardBox has Padding(1, 2)
	if contentWidth < 20 {
		contentWidth = 20
	}

	var b strings.Builder
	b.WriteString(ArabicHeaderStyle.Render("ALPHABET EXPLORER & LETTER INSPECTOR") + "\n")
	var tabBar string
	if m.ArabicExplorerTab == 0 {
		tabBar = CorrectStyle.Render("[ Tab 1: The 28 Letters ]") + "   " + HintStyle.Render("Tab 2: Auxiliary & Harakat (Tab)")
	} else {
		tabBar = HintStyle.Render("Tab 1: The 28 Letters (Tab)") + "   " + CorrectStyle.Render("[ Tab 2: Auxiliary & Harakat ]")
	}
	b.WriteString(tabBar + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n\n")

	if m.ArabicExplorerTab == 0 {
		cursor := m.ArabicLetterCursor
		if cursor < 0 || cursor >= len(arabic.Letters) {
			cursor = 0
		}
		l := arabic.Letters[cursor]

		var strip strings.Builder
		for idx, item := range arabic.Letters {
			glyph := item.Forms.Isolated
			if idx == cursor {
				strip.WriteString(CursorStyle.Render(fmt.Sprintf("[%s]", glyph)))
			} else {
				strip.WriteString(fmt.Sprintf(" %s ", glyph))
			}
			if idx == 13 {
				strip.WriteString("\n")
			}
		}
		b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Center).Render(strip.String()) + "\n\n")

		cellW := (contentWidth - 5) / 4
		if cellW > 15 {
			cellW = 15
		}
		if cellW < 8 {
			cellW = 8
		}
		bar := strings.Repeat("─", cellW)
		cellStyle := lipgloss.NewStyle().Width(cellW).Align(lipgloss.Center)
		posTable := fmt.Sprintf(
			"┌%s┬%s┬%s┬%s┐\n"+
				"│%s│%s│%s│%s│\n"+
				"│%s│%s│%s│%s│\n"+
				"└%s┴%s┴%s┴%s┘",
			bar, bar, bar, bar,
			cellStyle.Render("Isolated"),
			cellStyle.Render("Initial"),
			cellStyle.Render("Medial"),
			cellStyle.Render("Final"),
			cellStyle.Render(ArabicFormGlyph.Render(l.Forms.Isolated)),
			cellStyle.Render(ArabicFormGlyph.Render(l.Forms.Initial)),
			cellStyle.Render(ArabicFormGlyph.Render(l.Forms.Medial)),
			cellStyle.Render(ArabicFormGlyph.Render(l.Forms.Final)),
			bar, bar, bar, bar,
		)

		letterHeader := fmt.Sprintf("%s   %s (%s)",
			ArabicGlyphHuge.Render(l.Forms.Isolated),
			ArabicHeaderStyle.Render(l.ArabicName),
			KeyStyle.Render(l.EnglishName),
		)

		phonetics := fmt.Sprintf("Transliteration: %s   •   IPA: %s",
			PinyinStyle.Render(l.Translit),
			PronunciationStyle.Render(l.IPA),
		)

		soundLines := strings.Join(wrapLabeledText("", "Sound & Articulation: ", l.SoundDesc, lipgloss.NewStyle(), contentWidth), "\n")

		connBadge := BadgeConnectorStyle.Render("[✓ Dual-Connecting]")
		if !l.IsConnector {
			connBadge = BadgeNonConnectorStyle.Render("[⚠️ Non-Connecting]")
		}

		sunMoonBadge := BadgeMoonStyle.Render("[🌙 Moon Letter]")
		if l.IsSunLetter {
			sunMoonBadge = BadgeSunStyle.Render("[☀️ Sun Letter]")
		}

		emphaticBadge := ""
		if l.Emphatic {
			emphaticBadge = "  " + ErrorStyle.Render("[Emphatic]")
		}

		groupInfo := HintStyle.Render(fmt.Sprintf("Group %d: %s", l.PedGroup, l.GroupName))

		tips := ""
		if l.Tips != "" {
			tips = "\n" + strings.Join(wrapLabeledText("", "Tip: ", l.Tips, QuizExplanationStyle, contentWidth), "\n")
		}

		var exStr strings.Builder
		exStr.WriteString("Vocabulary Examples:\n")
		for _, ex := range l.Examples {
			exStr.WriteString(fmt.Sprintf("  • %s (%s) — %s [%s]\n",
				ArabicHeaderStyle.Render(ex.Arabic),
				PinyinStyle.Render(ex.Translit),
				MeaningStyle.Render(ex.Meaning),
				HintStyle.Render(ex.Position),
			))
		}

		cardContent := fmt.Sprintf(
			"%s\n%s\n%s\n\n%s\n\n%s  %s%s\n%s%s\n\n%s",
			letterHeader,
			phonetics,
			soundLines,
			posTable,
			connBadge, sunMoonBadge, emphaticBadge, groupInfo,
			tips,
			strings.TrimRight(exStr.String(), "\n"),
		)

		b.WriteString(ArabicCardBox.Width(boxInnerWidth).Render(cardContent) + "\n\n")
		footerRaw := fmt.Sprintf("[%s] Prev/Next  •  [%s] Switch Tab  •  [%s] Help  •  [%s] Back  •  [%s] Quit",
			KeyStyle.Render("←→/hl"), KeyStyle.Render("Tab"), KeyStyle.Render("?"), KeyStyle.Render("Esc"), KeyStyle.Render("q"))
		b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))
	} else {
		cursor := m.ArabicAuxCursor
		if cursor < 0 || cursor >= len(arabic.AuxiliaryItems) {
			cursor = 0
		}
		item := arabic.AuxiliaryItems[cursor]

		var strip strings.Builder
		for idx, a := range arabic.AuxiliaryItems {
			if idx == cursor {
				strip.WriteString(CursorStyle.Render(fmt.Sprintf("[%s]", a.Name)))
			} else {
				strip.WriteString(fmt.Sprintf(" %s ", a.Name))
			}
			if (idx+1)%3 == 0 && idx < len(arabic.AuxiliaryItems)-1 {
				strip.WriteString("\n")
			}
		}
		b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Center).Render(strip.String()) + "\n\n")

		descLines := strings.Join(wrapLabeledText("", "Description: ", item.Description, lipgloss.NewStyle(), contentWidth), "\n")
		cardContent := fmt.Sprintf(
			"%s   %s (%s)\n\nCategory: %s\nTransliteration: %s\n\n%s\n\nExample in Word: %s",
			ArabicGlyphHuge.Render(item.Arabic),
			ArabicHeaderStyle.Render(item.Name),
			KeyStyle.Render(fmt.Sprintf("%d of %d", cursor+1, len(arabic.AuxiliaryItems))),
			MeaningStyle.Render(item.Category),
			PinyinStyle.Render(item.Translit),
			descLines,
			CorrectStyle.Render(item.Example),
		)

		b.WriteString(ArabicCardBox.Width(boxInnerWidth).Render(cardContent) + "\n\n")
		footerRaw := fmt.Sprintf("[%s] Prev/Next  •  [%s] Switch Tab  •  [%s] Help  •  [%s] Back  •  [%s] Quit",
			KeyStyle.Render("←→/hl"), KeyStyle.Render("Tab"), KeyStyle.Render("?"), KeyStyle.Render("Esc"), KeyStyle.Render("q"))
		b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))
	}

	return renderFrame(frameWidth, b.String())
}

func (m Model) viewArabicDrill() string {
	frameWidth := m.frameWidth()

	if len(m.ArabicQuestions) == 0 {
		controls := HintStyle.Render(fmt.Sprintf("[%s] Return to Menu", KeyStyle.Render("Enter / Esc")))
		return renderFrame(frameWidth, renderCompleteCard(frameWidth, "No questions loaded.", "", nil, controls))
	}

	if m.ArabicQuestionIdx >= len(m.ArabicQuestions) {
		total := len(m.ArabicQuestions)
		pct := (m.ArabicScore * 100) / total
		statusMsg := "Excellent! You've mastered this session!"
		if pct < 70 {
			statusMsg = "Good effort! Regular practice will make these letter shapes automatic."
		}
		barLen := 26
		if frameWidth < 46 {
			barLen = 16
		}
		bar := renderProgressBar(total, total, barLen)
		stats := []string{
			fmt.Sprintf("Final Score: %s / %s (%d%%)",
				CorrectStyle.Render(fmt.Sprintf("%d", m.ArabicScore)),
				KeyStyle.Render(fmt.Sprintf("%d", total)),
				pct),
			QuizExplanationStyle.Width(frameWidth - 8).Render(statusMsg),
		}
		controls := HintStyle.Render(fmt.Sprintf("[%s] Retry Drill  •  [%s] Academy Menu  •  [%s] Quit",
			KeyStyle.Render("r"), KeyStyle.Render("Enter / Esc"), KeyStyle.Render("q")))
		return renderFrame(frameWidth, renderCompleteCard(frameWidth, "★ DRILL COMPLETED ★", bar, stats, controls))
	}

	q := m.ArabicQuestions[m.ArabicQuestionIdx]
	total := len(m.ArabicQuestions)
	answered := m.ArabicQuestionIdx
	if m.ArabicShowFeedback {
		answered++
	}
	accStr := "100%"
	if answered > 0 {
		accStr = fmt.Sprintf("%.0f%%", float64(m.ArabicScore)/float64(answered)*100)
	}

	headerLine1 := HintStyle.Render(fmt.Sprintf("Arabic Drill: Question %d of %d  •  Score: %d/%d  •  Accuracy: %s",
		m.ArabicQuestionIdx+1, total, m.ArabicScore, answered, accStr))
	barLen := 26
	if frameWidth < 46 {
		barLen = 16
	}
	headerLine2 := renderProgressBar(answered, total, barLen)
	headerZone := headerLine1 + "\n" + headerLine2

	boxInnerWidth := frameWidth - 2
	if boxInnerWidth < 16 {
		boxInnerWidth = 16
	}
	var promptLines []string
	promptLines = append(promptLines, ArabicHeaderStyle.Render(q.Prompt))
	if q.SubPrompt != "" {
		promptLines = append(promptLines, HintStyle.Render(q.SubPrompt))
	}
	if q.Glyph != "" {
		promptLines = append(promptLines, "")
		promptLines = append(promptLines, ArabicGlyphHuge.Render(q.Glyph))
	}
	promptZone := QuizPromptBoxStyle.Width(boxInnerWidth).Render(strings.Join(promptLines, "\n"))

	const drillBodyHeight = 10
	var bodyLines []string
	if m.ArabicShowFeedback {
		if m.ArabicIsCorrect {
			bodyLines = append(bodyLines, "  "+CorrectStyle.Render("✓ Correct!"))
		} else {
			bodyLines = append(bodyLines, "  "+ErrorStyle.Render("✗ Incorrect!"))
			bodyLines = append(bodyLines, "")
			bodyLines = append(bodyLines, wrapLabeledText("  ", "The correct answer was: ", q.Options[q.CorrectIndex], CorrectStyle, frameWidth)...)
		}
		if q.Explanation != "" {
			bodyLines = append(bodyLines, "")
			expW := frameWidth - 4
			if expW < 20 {
				expW = 20
			}
			for _, el := range strings.Split(QuizExplanationStyle.Width(expW).Render(q.Explanation), "\n") {
				bodyLines = append(bodyLines, "  "+el)
			}
		}
	} else {
		for i, opt := range q.Options {
			key := IndexToKey(i)
			keyBadge := KeyStyle.Render(fmt.Sprintf("[%s]", key))
			bodyLines = append(bodyLines, fmt.Sprintf("  %s  %s", keyBadge, opt))
			if (i+1)%5 == 0 && i < len(q.Options)-1 {
				bodyLines = append(bodyLines, "")
			}
		}
	}
	bodyZone := fitBlockHeight(strings.Join(bodyLines, "\n"), drillBodyHeight)

	var footerRaw string
	if m.ArabicShowFeedback {
		footerRaw = fmt.Sprintf("[%s] Continue  •  [%s] Help  •  [%s] Menu",
			KeyStyle.Render("Spacebar / Enter"), KeyStyle.Render("?"), KeyStyle.Render("Esc"))
	} else {
		footerRaw = fmt.Sprintf("Press key to select  •  [%s] Help  •  [%s] Menu  •  [%s] Quit",
			KeyStyle.Render("?"), KeyStyle.Render("Esc"), KeyStyle.Render("q"))
	}
	footerZone := renderDivider(frameWidth) + "\n" +
		fitBlockHeight(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)), 2)

	assembled := lipgloss.JoinVertical(
		lipgloss.Left,
		headerZone,
		"",
		promptZone,
		"",
		bodyZone,
		footerZone,
	)
	return renderFrame(frameWidth, assembled)
}

func (m Model) viewArabicStages() string {
	frameWidth := m.frameWidth()
	var b strings.Builder
	b.WriteString(ArabicHeaderStyle.Render("PROGRESSIVE GUIDED LESSONS") + "\n")
	b.WriteString(ArabicSubHeaderStyle.Render("8 Structured Stages from The Anchors to Auxiliary Marks") + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n\n")

	for i, st := range arabic.Stages {
		key := IndexToKey(i)
		keyBadge := KeyStyle.Render(fmt.Sprintf("[%s]", key))
		cursor := "  "
		name := fmt.Sprintf("Stage %d: %s", st.Number, st.Name)
		if m.ArabicStageCursor == i {
			cursor = "> "
			name = CursorStyle.Render(name)
		}

		var letterPreview string
		if len(st.LetterIDs) > 0 {
			var glyphs []string
			for _, id := range st.LetterIDs {
				l := arabic.GetLetterByID(id)
				if l != nil {
					glyphs = append(glyphs, l.Forms.Isolated)
				}
			}
			letterPreview = fmt.Sprintf("(%s)", strings.Join(glyphs, " "))
		} else {
			letterPreview = "(ء, ة, ى, َ, ُ, ِ, ْ, ّ, ً)"
		}

		leftPart := fmt.Sprintf("%s%s %s", CursorStyle.Render(cursor), keyBadge, name)
		b.WriteString(renderSplitLine(leftPart, CorrectStyle.Render(letterPreview), frameWidth) + "\n")
		b.WriteString(fmt.Sprintf("      %s\n\n", HintStyle.Render(truncateToWidth(st.Description, frameWidth-6))))
	}

	b.WriteString(renderDivider(frameWidth) + "\n")
	footerRaw := fmt.Sprintf("[%s] View Stage  •  [%s] Help  •  [%s] Back  •  [%s] Quit",
		KeyStyle.Render("key / Enter"), KeyStyle.Render("?"), KeyStyle.Render("Esc"), KeyStyle.Render("q"))
	b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))
	return renderFrame(frameWidth, b.String())
}

func (m Model) viewArabicStageView() string {
	frameWidth := responsiveFrameWidth(m.Width, 78, 44)
	cursor := m.ArabicStageCursor
	if cursor < 0 || cursor >= len(arabic.Stages) {
		cursor = 0
	}
	st := arabic.Stages[cursor]

	var b strings.Builder
	b.WriteString(ArabicHeaderStyle.Render(fmt.Sprintf("STAGE %d: %s", st.Number, st.Name)) + "\n")
	b.WriteString(HintStyle.Render(st.Description) + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n\n")

	if len(st.LetterIDs) > 0 {
		for _, id := range st.LetterIDs {
			l := arabic.GetLetterByID(id)
			if l == nil {
				continue
			}
			connBadge := BadgeConnectorStyle.Render("[dual-connector]")
			if !l.IsConnector {
				connBadge = BadgeNonConnectorStyle.Render("[non-connector ⚠️]")
			}
			sunMoonBadge := BadgeMoonStyle.Render("[moon]")
			if l.IsSunLetter {
				sunMoonBadge = BadgeSunStyle.Render("[sun]")
			}
			left := fmt.Sprintf("  %s  %s (%s)  •  [%s | %s | %s | %s]",
				ArabicGlyphHuge.Render(l.Forms.Isolated),
				ArabicHeaderStyle.Render(l.ArabicName),
				KeyStyle.Render(l.EnglishName),
				l.Forms.Isolated, l.Forms.Initial, l.Forms.Medial, l.Forms.Final,
			)
			b.WriteString(renderSplitLine(left, connBadge+" "+sunMoonBadge, frameWidth) + "\n")
			for _, sl := range wrapLabeledText("      ", "Sound: ", l.SoundDesc, PinyinStyle, frameWidth) {
				b.WriteString(sl + "\n")
			}
			if l.Tips != "" {
				for _, tl := range wrapLabeledText("      ", "Tip:   ", l.Tips, QuizExplanationStyle, frameWidth) {
					b.WriteString(tl + "\n")
				}
			}
			b.WriteString("\n")
		}
	} else {
		for _, a := range arabic.AuxiliaryItems {
			b.WriteString(fmt.Sprintf("  %s  %s  •  %s\n",
				ArabicGlyphHuge.Render(a.Arabic),
				ArabicHeaderStyle.Render(a.Name),
				MeaningStyle.Render(a.Category),
			))
			for _, dl := range wrapPlainText(a.Description+"  •  Example: "+a.Example, frameWidth-6) {
				b.WriteString("      " + HintStyle.Render(dl) + "\n")
			}
			b.WriteString("\n")
		}
	}

	b.WriteString(renderDivider(frameWidth) + "\n")
	footerRaw := fmt.Sprintf("[%s] Start Stage Drill  •  [%s] Back to Stages  •  [%s] Quit",
		KeyStyle.Render("Space / Enter"), KeyStyle.Render("Esc / b"), KeyStyle.Render("q"))
	b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))
	return renderFrame(frameWidth, b.String())
}
