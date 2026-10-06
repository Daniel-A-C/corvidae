package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"flashcards/internal/arabic"
	"flashcards/internal/deck"
)

// View renders the terminal user interface according to the current state.
func (m Model) View() string {
	if m.Err != nil {
		return fmt.Sprintf("Error: %v\n", m.Err)
	}

	var content string
	switch m.State {
	case StateModeSelect:
		content = m.viewModeSelect()
	case StateDirSelect:
		content = m.viewDirSelect()
	case StateDeckSelect:
		content = m.viewDeckSelect()
	case StateReview:
		content = m.viewReview()
	case StateQuiz:
		content = m.viewQuiz()
	case StateArabicMenu:
		content = m.viewArabicMenu()
	case StateArabicExplorer:
		content = m.viewArabicExplorer()
	case StateArabicDrill:
		content = m.viewArabicDrill()
	case StateArabicStages:
		content = m.viewArabicStages()
	case StateArabicStageView:
		content = m.viewArabicStageView()
	case StateMemorizeSelectText:
		content = m.viewMemorizeSelectText()
	case StateMemorizePortionSelect:
		content = m.viewMemorizePortionSelect()
	case StateMemorize:
		content = m.viewMemorize()
	case StateMemorizeComplete:
		content = m.viewMemorizeComplete()
	}

	styledContent := lipgloss.NewStyle().Align(lipgloss.Center).Render(content)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, styledContent)
}

func (m Model) viewModeSelect() string {
	content := LogoStyle.Render(AsciiLogo) + "\n\n"
	content += "Select Mode:\n\n"

	modes := []string{
		"Spaced Repetition (SM-2)",
		"Multiple Choice Quiz",
		"Arabic Alphabet Academy",
		"Memorize by Options",
	}
	for i, label := range modes {
		key := IndexToKey(i)
		keyBadge := KeyStyle.Render(fmt.Sprintf("[%s]", key))
		cursor := "  "
		if m.Mode == i {
			cursor = "> "
			label = CursorStyle.Render(label)
		}
		content += fmt.Sprintf("%s%s %s\n", CursorStyle.Render(cursor), keyBadge, label)
	}
	markedCount := len(m.MarkedCards)
	if markedCount > 0 {
		var cardWord string
		if markedCount == 1 {
			cardWord = "1 card"
		} else {
			cardWord = fmt.Sprintf("%d cards", markedCount)
		}
		content += "\n" + fmt.Sprintf("  %s %s\n",
			KeyStyle.Render("[m]"),
			MarkedStarStyle.Render(fmt.Sprintf("★ Practice Marked Cards (%s)", cardWord)),
		)
	}

	if m.StatusMessage != "" {
		content += "\n" + ExplanationStyle.Render(m.StatusMessage) + "\n"
	}

	content += "\n" + HintStyle.Render("(Press a/s/d/f to select, Enter to confirm, q to quit)")
	return content
}

func (m Model) viewDirSelect() string {
	if len(m.Dirs) == 0 && m.SelectedDir != "" {
		return fmt.Sprintf("No directories found in %s/%s.\n\n%s", m.BaseDir, m.SelectedDir, HintStyle.Render("(Press Esc to return, q to quit)"))
	}

	var content string
	if m.SelectedDir == "" {
		content = "Select Deck Directory / Language:\n\n"
	} else {
		content = fmt.Sprintf("Select Deck Directory (%s):\n\n", m.SelectedDir)
	}

	for i, dir := range m.Dirs {
		key := IndexToKey(i)
		keyBadge := ""
		if key != "" {
			keyBadge = KeyStyle.Render(fmt.Sprintf("[%s]", key)) + " "
		}
		cursor := "  "
		label := dir
		dirRelPath := dir
		if m.SelectedDir != "" {
			dirRelPath = filepath.Join(m.SelectedDir, dir)
		}
		count := deck.CountDecksInDir(m.BaseDir, dirRelPath)
		var countStr string
		if count == 1 {
			countStr = HintStyle.Render(" (1 deck)")
		} else {
			countStr = HintStyle.Render(fmt.Sprintf(" (%d decks)", count))
		}

		if m.DirCursor == i {
			cursor = "> "
			label = CursorStyle.Render(label)
		}
		content += fmt.Sprintf("%s%s%s%s\n", CursorStyle.Render(cursor), keyBadge, label, countStr)
		if (i+1)%5 == 0 && i < len(m.Dirs)-1 {
			content += "\n"
		}
	}

	if m.SelectedDir == "" {
		cursor := "  "
		if m.DirCursor == len(m.Dirs) {
			cursor = "> "
		}
		count := len(m.MarkedCards)
		var countStr string
		if count == 1 {
			countStr = HintStyle.Render(" (1 card)")
		} else {
			countStr = HintStyle.Render(fmt.Sprintf(" (%d cards)", count))
		}
		label := "Marked Cards"
		if count > 0 {
			label = MarkedStarStyle.Render("★ ") + label
		}
		if m.DirCursor == len(m.Dirs) {
			label = CursorStyle.Render(label)
		}
		keyBadge := KeyStyle.Render("[m]") + " "
		content += fmt.Sprintf("\n%s%s%s%s\n", CursorStyle.Render(cursor), keyBadge, label, countStr)
	}

	if m.StatusMessage != "" {
		content += "\n" + ExplanationStyle.Render(m.StatusMessage) + "\n"
	}

	selectedCount := m.countSelectedDecks()
	if selectedCount > 0 {
		var deckWord string
		if selectedCount == 1 {
			deckWord = "1 deck"
		} else {
			deckWord = fmt.Sprintf("%d decks", selectedCount)
		}
		content += "\n" + CorrectStyle.Render(fmt.Sprintf("%s selected across folders  •  [Tab] Start Practice", deckWord)) + "\n"
		content += "\n" + HintStyle.Render("(Press key to select, Tab to start, Esc to go back, q to quit)")
	} else {
		content += "\n" + HintStyle.Render("(Press key to select, [m] Marked Cards, Esc to go back, q to quit)")
	}
	return content
}

func (m Model) viewDeckSelect() string {
	if len(m.DeckFiles) == 0 {
		return fmt.Sprintf("No .yaml files found in %s/%s.\n\n%s", m.BaseDir, m.SelectedDir, HintStyle.Render("(Press Esc or Enter to go back, q to quit)"))
	}

	content := fmt.Sprintf("Select decks to practice (%s)", m.SelectedDir)
	selectedCount := m.countSelectedDecks()
	if selectedCount > 0 {
		var deckWord string
		if selectedCount == 1 {
			deckWord = "1 deck selected"
		} else {
			deckWord = fmt.Sprintf("%d decks selected", selectedCount)
		}
		content += fmt.Sprintf(" [%s]", deckWord)
	}
	content += ":\n\n"

	for i, file := range m.DeckFiles {
		key := IndexToKey(i)
		keyBadge := ""
		if key != "" {
			keyBadge = KeyStyle.Render(fmt.Sprintf("[%s]", key)) + " "
		}
		cursor := "  "
		if m.Cursor == i {
			cursor = "> "
		}

		check := "[ ]"
		if m.isDeckSelected(file) {
			check = "[x]"
		}

		displayFile := file
		if m.Cursor == i {
			displayFile = CursorStyle.Render(file)
		}
		line := fmt.Sprintf("%s%s%s %s", CursorStyle.Render(cursor), keyBadge, KeyStyle.Render(check), displayFile)
		content += line + "\n"
		if (i+1)%5 == 0 && i < len(m.DeckFiles)-1 {
			content += "\n"
		}
	}
	content += "\n" + HintStyle.Render("(Press key to toggle, Enter to confirm, Esc to go back, q to quit)")
	return content
}

func (m Model) viewReview() string {
	if len(m.ActiveCards) == 0 {
		return fmt.Sprintf("No cards due today!\n\n%s", HintStyle.Render("[r] Review all cards anyway  •  [Enter] Main Menu  •  [q] Quit"))
	}
	if m.CurrentIndex >= len(m.ActiveCards) {
		return fmt.Sprintf("Daily review complete! Great job.\n\n%s", HintStyle.Render("[r] Review all cards again  •  [Enter] Main Menu  •  [q] Quit"))
	}

	ref := m.ActiveCards[m.CurrentIndex]
	card := m.Decks[ref.Filename].Cards[ref.OrigIdx]
	displayName, err := filepath.Rel(m.BaseDir, ref.Filename)
	if err != nil {
		displayName = ref.Filename
	}

	header := fmt.Sprintf("Card %d of %d  •  %s", m.CurrentIndex+1, len(m.ActiveCards), displayName)
	if m.isCardMarked(card) {
		header += "  " + MarkedBadgeStyle.Render("★ MARKED")
	}

	content := HintStyle.Render(header) + "\n\n"
	content += CharStyle.Render(card.Character) + "\n\n"

	if !m.ShowAnswer {
		content += HintStyle.Render("[ Spacebar to reveal ]") + "\n"
		content += "\n" + HintStyle.Render("([m] Toggle Mark  •  Press 'q' to quit)")
		return content
	}

	if card.Pinyin != "" {
		content += fmt.Sprintf("Pinyin:  %s\n", PinyinStyle.Render(card.Pinyin))
	} else if card.Pronunciation != "" {
		content += fmt.Sprintf("Pronunciation: %s\n", PronunciationStyle.Render(card.Pronunciation))
	}
	content += fmt.Sprintf("Meaning: %s\n\n", MeaningStyle.Render(card.Meaning))

	if m.ShowExplanation && card.Explanation != "" {
		content += ExplanationStyle.Render(card.Explanation) + "\n\n"
	}

	content += "How well did you know this?\n"
	content += fmt.Sprintf("[%s] Blackout  [%s] Wrong  [%s] Hard\n",
		KeyStyle.Render("d"), KeyStyle.Render("f"), KeyStyle.Render("g"))
	content += fmt.Sprintf("[%s] Good      [%s] Easy   [%s] Perfect\n",
		KeyStyle.Render("h"), KeyStyle.Render("j"), KeyStyle.Render("k"))

	content += "\n" + HintStyle.Render("[e] Toggle Explanation  •  [m] Toggle Mark  •  [q] Quit")
	return content
}

func (m Model) viewQuiz() string {
	if m.CurrentIndex >= len(m.ActiveCards) {
		return fmt.Sprintf("Quiz complete!\n\n%s", HintStyle.Render("[r] Retry quiz  •  [Enter] Main Menu  •  [q] Quit"))
	}

	ref := m.ActiveCards[m.CurrentIndex]
	card := m.Decks[ref.Filename].Cards[ref.OrigIdx]
	isMarked := m.isCardMarked(card)

	header := fmt.Sprintf("Quiz: Question %d of %d", m.CurrentIndex+1, len(m.ActiveCards))
	if m.ShowFeedback {
		displayName, err := filepath.Rel(m.BaseDir, ref.Filename)
		if err != nil {
			displayName = ref.Filename
		}
		header += fmt.Sprintf("  •  %s", displayName)
	}
	if isMarked {
		header += "  " + MarkedBadgeStyle.Render("★ MARKED")
	}

	content := HintStyle.Render(header) + "\n\n"
	content += CharStyle.Render(card.Character) + "\n\n"

	if m.ShowFeedback {
		if m.IsCorrect {
			content += CorrectStyle.Render("Correct!") + "\n\n"
			if card.Pinyin != "" {
				content += fmt.Sprintf("Pinyin:  %s\n", PinyinStyle.Render(card.Pinyin))
			} else if card.Pronunciation != "" {
				content += fmt.Sprintf("Pronunciation: %s\n", PronunciationStyle.Render(card.Pronunciation))
			}
			content += fmt.Sprintf("Meaning: %s\n\n", MeaningStyle.Render(card.Meaning))
			if card.Explanation != "" {
				content += ExplanationStyle.Render(card.Explanation) + "\n\n"
			}
		} else {
			content += ErrorStyle.Render("Incorrect!") + "\n\n"
			content += fmt.Sprintf("The correct answer was: %s\n\n", CorrectStyle.Render(m.QuizOptions[m.CorrectIndex]))
			if card.Explanation != "" {
				content += ExplanationStyle.Render(card.Explanation) + "\n\n"
			}
		}

		var markHint string
		if isMarked {
			markHint = fmt.Sprintf("[%s] Unmark  •  ", KeyStyle.Render("m/u"))
		} else {
			markHint = fmt.Sprintf("[%s] Mark card  •  ", KeyStyle.Render("m"))
		}
		content += HintStyle.Render(fmt.Sprintf("[ %sSpacebar to continue ]", markHint))
		return content
	}

	for i, opt := range m.QuizOptions {
		key := IndexToKey(i)
		keyBadge := KeyStyle.Render(fmt.Sprintf("[%s]", key))
		content += fmt.Sprintf("%s %s\n", keyBadge, opt)
		if (i+1)%5 == 0 && i < len(m.QuizOptions)-1 {
			content += "\n"
		}
	}

	var markHint string
	if isMarked {
		markHint = fmt.Sprintf("[%s] Unmark card  •  ", KeyStyle.Render("m"))
	} else {
		markHint = fmt.Sprintf("[%s] Mark card  •  ", KeyStyle.Render("m"))
	}
	content += "\n" + HintStyle.Render(fmt.Sprintf("(%sPress key to select, q to quit)", markHint))
	return content
}

func (m Model) viewArabicMenu() string {
	var b strings.Builder
	title := ArabicHeaderStyle.Render("★ ARABIC ALPHABET ACADEMY ★") + "\n"
	sub := ArabicSubHeaderStyle.Render("أكاديمية الحروف العربية — Master the Modern Standard Arabic Script") + "\n\n"
	b.WriteString(title)
	b.WriteString(sub)
	b.WriteString("Select Learning Activity:\n\n")

	activities := []struct {
		title string
		desc  string
	}{
		{"Alphabet Explorer & Letter Inspector", "Interactive 28-letter browser with 4 cursive shapes, sounds, & examples"},
		{"Positional Forms Drill", "Master Isolated, Initial, Medial, and Final cursive recognition"},
		{"Letter Sounds & Transliteration Quiz", "Connect letters, names, IPA sounds, and transliterations"},
		{"Minimal Pairs & Confusable Letters", "Contrast tricky letters (ت/ط, د/ض, س/ص, ذ/ظ, ح/ه, ع/غ)"},
		{"Non-Connecting Letters Drill", "Drill the 6 non-connecting letters (ا, د, ذ, ر, ز, و)"},
		{"Sun & Moon Letters Drill", "Master definite article (الـ) assimilation and pronunciation"},
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
		descStr := HintStyle.Render(" — " + act.desc)
		b.WriteString(fmt.Sprintf("%s%s %s%s\n", CursorStyle.Render(cursor), keyBadge, label, descStr))
		if i == 0 || i == 3 || i == 5 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n" + HintStyle.Render("(Press key to select, Enter to confirm, Esc to return to Main Menu, q to quit)"))
	return b.String()
}

func (m Model) viewArabicExplorer() string {
	var b strings.Builder
	header := ArabicHeaderStyle.Render("ALPHABET EXPLORER & LETTER INSPECTOR") + "\n"
	var tabBar string
	if m.ArabicExplorerTab == 0 {
		tabBar = CorrectStyle.Render("[ Tab 1: The 28 Letters (Active) ]") + "   " + HintStyle.Render("[ Tab 2: Auxiliary & Harakat (Tab to switch) ]")
	} else {
		tabBar = HintStyle.Render("[ Tab 1: The 28 Letters (Tab to switch) ]") + "   " + CorrectStyle.Render("[ Tab 2: Auxiliary & Harakat (Active) ]")
	}
	b.WriteString(header)
	b.WriteString(tabBar + "\n\n")

	if m.ArabicExplorerTab == 0 {
		cursor := m.ArabicLetterCursor
		if cursor < 0 || cursor >= len(arabic.Letters) {
			cursor = 0
		}
		l := arabic.Letters[cursor]

		// Alphabet Strip
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
		b.WriteString(strip.String() + "\n\n")

		cellStyle := lipgloss.NewStyle().Width(15).Align(lipgloss.Center)
		posTable := fmt.Sprintf(
			"┌───────────────┬───────────────┬───────────────┬───────────────┐\n"+
				"│%s│%s│%s│%s│\n"+
				"│%s│%s│%s│%s│\n"+
				"└───────────────┴───────────────┴───────────────┴───────────────┘",
			cellStyle.Render("Isolated"),
			cellStyle.Render("Initial"),
			cellStyle.Render("Medial"),
			cellStyle.Render("Final"),
			cellStyle.Render(ArabicFormGlyph.Render(l.Forms.Isolated)),
			cellStyle.Render(ArabicFormGlyph.Render(l.Forms.Initial)),
			cellStyle.Render(ArabicFormGlyph.Render(l.Forms.Medial)),
			cellStyle.Render(ArabicFormGlyph.Render(l.Forms.Final)),
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

		soundInfo := fmt.Sprintf("Sound & Articulation: %s", l.SoundDesc)

		connBadge := BadgeConnectorStyle.Render("[✓ Dual-Connecting Letter]")
		if !l.IsConnector {
			connBadge = BadgeNonConnectorStyle.Render("[⚠️ Non-Connecting Letter (Never connects to left)]")
		}

		sunMoonBadge := BadgeMoonStyle.Render("[🌙 Moon Letter (al-)]")
		if l.IsSunLetter {
			sunMoonBadge = BadgeSunStyle.Render("[☀️ Sun Letter (assimilates)]")
		}

		emphaticBadge := ""
		if l.Emphatic {
			emphaticBadge = "   " + ErrorStyle.Render("[Emphatic / Heavy]")
		}

		groupInfo := HintStyle.Render(fmt.Sprintf("Group %d: %s", l.PedGroup, l.GroupName))

		tips := ""
		if l.Tips != "" {
			tips = "\nMnemonic / Tip: " + ExplanationStyle.Render(l.Tips)
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
			"%s\n%s\n%s\n\n%s\n\n%s   %s%s   •   %s%s\n\n%s",
			letterHeader,
			phonetics,
			soundInfo,
			posTable,
			connBadge, sunMoonBadge, emphaticBadge, groupInfo,
			tips,
			exStr.String(),
		)

		b.WriteString(ArabicCardBox.Render(cardContent) + "\n\n")
		b.WriteString(HintStyle.Render("(←/h Prev Letter • →/l Next Letter • Tab Toggle Auxiliary • Esc/b Return to Menu • q Quit)"))
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
		b.WriteString(strip.String() + "\n\n")

		cardContent := fmt.Sprintf(
			"%s\n\n%s (%s)\n\nCategory: %s\nTransliteration: %s\n\nDescription: %s\n\nExample in Word: %s\n",
			ArabicGlyphHuge.Render(item.Arabic),
			ArabicHeaderStyle.Render(item.Name),
			KeyStyle.Render(fmt.Sprintf("%d of %d", cursor+1, len(arabic.AuxiliaryItems))),
			MeaningStyle.Render(item.Category),
			PinyinStyle.Render(item.Translit),
			item.Description,
			CorrectStyle.Render(item.Example),
		)

		b.WriteString(ArabicCardBox.Render(cardContent) + "\n\n")
		b.WriteString(HintStyle.Render("(←/h Prev Item • →/l Next Item • Tab Toggle 28 Letters • Esc/b Return to Menu • q Quit)"))
	}

	return b.String()
}

func (m Model) viewArabicDrill() string {
	if len(m.ArabicQuestions) == 0 {
		return fmt.Sprintf("No questions loaded.\n\n%s", HintStyle.Render("[Enter / Esc] Return to Menu"))
	}

	if m.ArabicQuestionIdx >= len(m.ArabicQuestions) {
		pct := (m.ArabicScore * 100) / len(m.ArabicQuestions)
		statusMsg := "Excellent! You've mastered this session!"
		if pct < 70 {
			statusMsg = "Good effort! Regular practice will make these letter shapes automatic."
		}

		result := fmt.Sprintf(
			"%s\n\nFinal Score: %s / %s (%d%%)\n\n%s\n\n%s",
			ArabicHeaderStyle.Render("★ DRILL COMPLETED ★"),
			CorrectStyle.Render(fmt.Sprintf("%d", m.ArabicScore)),
			KeyStyle.Render(fmt.Sprintf("%d", len(m.ArabicQuestions))),
			pct,
			ExplanationStyle.Render(statusMsg),
			HintStyle.Render("[r] Retry Drill  •  [Enter / Esc] Academy Menu  •  [q] Quit"),
		)
		return result
	}

	q := m.ArabicQuestions[m.ArabicQuestionIdx]
	header := fmt.Sprintf("Question %d of %d   •   Score: %d/%d",
		m.ArabicQuestionIdx+1, len(m.ArabicQuestions), m.ArabicScore, m.ArabicQuestionIdx)

	var b strings.Builder
	b.WriteString(HintStyle.Render(header) + "\n\n")
	b.WriteString(ArabicHeaderStyle.Render(q.Prompt) + "\n")
	if q.SubPrompt != "" {
		b.WriteString(HintStyle.Render(q.SubPrompt) + "\n")
	}
	b.WriteString("\n")

	if q.Glyph != "" {
		b.WriteString(ArabicGlyphHuge.Render(q.Glyph) + "\n\n")
	}

	if m.ArabicShowFeedback {
		if m.ArabicIsCorrect {
			b.WriteString(CorrectStyle.Render("✓ Correct!") + "\n\n")
		} else {
			b.WriteString(ErrorStyle.Render("✗ Incorrect!") + "\n\n")
			b.WriteString(fmt.Sprintf("The correct answer was: %s\n\n", CorrectStyle.Render(q.Options[q.CorrectIndex])))
		}
		if q.Explanation != "" {
			b.WriteString(ExplanationStyle.Render(q.Explanation) + "\n\n")
		}
		b.WriteString(HintStyle.Render("[ Spacebar or Enter to continue ]"))
		return b.String()
	}

	for i, opt := range q.Options {
		key := IndexToKey(i)
		keyBadge := KeyStyle.Render(fmt.Sprintf("[%s]", key))
		b.WriteString(fmt.Sprintf("%s %s\n", keyBadge, opt))
		if (i+1)%5 == 0 && i < len(q.Options)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n" + HintStyle.Render("(Press key to select, Esc to return to menu, q to quit)"))
	return b.String()
}

func (m Model) viewArabicStages() string {
	var b strings.Builder
	b.WriteString(ArabicHeaderStyle.Render("PROGRESSIVE GUIDED LESSONS") + "\n")
	b.WriteString(ArabicSubHeaderStyle.Render("8 Structured Stages from The Anchors to Auxiliary Marks") + "\n\n")
	b.WriteString("Select Stage to Study & Drill:\n\n")

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
			letterPreview = fmt.Sprintf(" (%s)", strings.Join(glyphs, " "))
		} else {
			letterPreview = " (ء, ة, ى, َ, ُ, ِ, ْ, ّ, ً)"
		}

		b.WriteString(fmt.Sprintf("%s%s %s%s\n", CursorStyle.Render(cursor), keyBadge, name, CorrectStyle.Render(letterPreview)))
		b.WriteString(fmt.Sprintf("     %s\n\n", HintStyle.Render(st.Description)))
	}

	b.WriteString(HintStyle.Render("(Press key / Enter to view stage, Esc to return to menu, q to quit)"))
	return b.String()
}

func (m Model) viewArabicStageView() string {
	cursor := m.ArabicStageCursor
	if cursor < 0 || cursor >= len(arabic.Stages) {
		cursor = 0
	}
	st := arabic.Stages[cursor]

	var b strings.Builder
	b.WriteString(ArabicHeaderStyle.Render(fmt.Sprintf("STAGE %d: %s", st.Number, st.Name)) + "\n")
	b.WriteString(HintStyle.Render(st.Description) + "\n\n")

	if len(st.LetterIDs) > 0 {
		b.WriteString("Letters in this stage:\n\n")
		for _, id := range st.LetterIDs {
			l := arabic.GetLetterByID(id)
			if l == nil {
				continue
			}
			connBadge := "[dual-connector]"
			if !l.IsConnector {
				connBadge = "[non-connector ⚠️]"
			}
			sunMoonBadge := "[moon]"
			if l.IsSunLetter {
				sunMoonBadge = "[sun]"
			}
			b.WriteString(fmt.Sprintf("  %s  %s (%s)  •  Forms: [%s | %s | %s | %s]  •  %s %s\n",
				ArabicGlyphHuge.Render(l.Forms.Isolated),
				ArabicHeaderStyle.Render(l.ArabicName),
				KeyStyle.Render(l.EnglishName),
				l.Forms.Isolated, l.Forms.Initial, l.Forms.Medial, l.Forms.Final,
				BadgeConnectorStyle.Render(connBadge),
				BadgeMoonStyle.Render(sunMoonBadge),
			))
			b.WriteString(fmt.Sprintf("      Sound: %s  •  Tip: %s\n\n",
				PinyinStyle.Render(l.SoundDesc),
				ExplanationStyle.Render(l.Tips),
			))
		}
	} else {
		b.WriteString("Auxiliary characters and diacritics in this stage:\n\n")
		for _, a := range arabic.AuxiliaryItems {
			b.WriteString(fmt.Sprintf("  %s  %s  •  %s\n      %s  •  Example: %s\n\n",
				ArabicGlyphHuge.Render(a.Arabic),
				ArabicHeaderStyle.Render(a.Name),
				MeaningStyle.Render(a.Category),
				HintStyle.Render(a.Description),
				CorrectStyle.Render(a.Example),
			))
		}
	}

	b.WriteString(HintStyle.Render("[Space / Enter] Start Stage Drill  •  [Esc / b] Back to Stages List  •  [q] Quit"))
	return b.String()
}

func (m Model) viewMemorizeSelectText() string {
	if len(m.MemorizeTexts) == 0 {
		return fmt.Sprintf("No memorization texts found in %s.\n\nPlace .yaml or .txt files in %s/\n\n%s",
			m.TextsDir, m.TextsDir, HintStyle.Render("(Press Esc to return, q to quit)"))
	}

	content := MemorizeTitleStyle.Render("★ MEMORIZE BY OPTIONS ★") + "\n"
	content += MemorizeSubStyle.Render("Work through a text word by word with plausible options:") + "\n\n"

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

		info := fmt.Sprintf(" (%d words)", t.WordCount)
		if t.Progress > 0 && t.WordCount > 0 {
			pct := float64(t.Progress) / float64(t.WordCount) * 100
			info += fmt.Sprintf(" • In progress: %d/%d (%.0f%%)", t.Progress, t.WordCount, pct)
		}
		infoStr := HintStyle.Render(info)

		content += fmt.Sprintf("%s%s%s%s\n", CursorStyle.Render(cursor), keyBadge, label, infoStr)
	}

	content += "\n" + HintStyle.Render("(Press key or Enter to select text, Esc to return, q to quit)")
	return content
}

func (m Model) viewMemorizePortionSelect() string {
	if m.MemorizeActiveText == nil || len(m.MemorizeSections) == 0 {
		return fmt.Sprintf("No sections found for this text.\n\n%s", HintStyle.Render("[Esc] Back"))
	}

	startSec := m.MemorizeSections[m.PortionStartSection]
	endSec := m.MemorizeSections[m.PortionEndSection]
	portionWords := endSec.EndIdx - startSec.StartIdx
	totalWords := len(m.MemorizeActiveText.Words)
	numSections := m.PortionEndSection - m.PortionStartSection + 1

	var b strings.Builder
	b.WriteString(MemorizeTitleStyle.Render("★ PORTION SELECTOR ★") + "\n")
	b.WriteString(MemorizeHeaderStyle.Render(m.MemorizeActiveText.Title) + "\n\n")

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
	b.WriteString(MemorizeSubStyle.Render(summary) + "\n\n")

	// Selector rows with indicators
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

	b.WriteString(fmt.Sprintf("%s[%s]: %s  %s\n",
		cursorStart, startBadge, startVal, HintStyle.Render("\""+startSec.Preview+"\"")))
	b.WriteString(fmt.Sprintf("%s[%s]:   %s  %s\n\n",
		cursorEnd, endBadge, endVal, HintStyle.Render("\""+endSec.Preview+"\"")))

	// Window preview of sections around the selection
	b.WriteString("Section Overview:\n")
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
		line := fmt.Sprintf("%-10s (%2d words): %s", sec.Label, sec.WordCount, sec.Preview)
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
			b.WriteString(fmt.Sprintf("%s%s%s\n", CursorStyle.Render(prefix), MemorizeCompleted.Render(line), tag))
		} else {
			b.WriteString(fmt.Sprintf("%s%s\n", prefix, HintStyle.Render(line)))
		}
	}

	b.WriteString("\n" + HintStyle.Render("[Tab / ↑↓] Switch Start/End  •  [←→ / hl] Shift  •  [x] Expand  •  [n] Advance") + "\n")
	b.WriteString(HintStyle.Render("[a] Select All  •  [Enter / Space] Practice Portion  •  [Esc] Back"))
	return b.String()
}

func (m Model) viewMemorize() string {
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
	filled := 0
	if totalPortionWords > 0 {
		filled = int(float64(barLen) * float64(currentInPortion) / float64(totalPortionWords))
	}
	if filled > barLen {
		filled = barLen
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barLen-filled)

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

	var b strings.Builder
	b.WriteString(MemorizeHeaderStyle.Render(fmt.Sprintf("★ %s%s ★", m.MemorizeActiveText.Title, portionTitle)) + "\n")
	b.WriteString(HintStyle.Render(fmt.Sprintf("Portion Word %d of %d (%.1f%%)  •  Accuracy: %s  •  Streak: %d",
		currentInPortion+1, totalPortionWords, pct, accStr, m.MemorizeStreak)) + "\n")
	b.WriteString(MemorizeProgressStyle.Render("["+bar+"]") + "\n\n")

	// Build rolling text window (last ~25 words)
	startIdx := 0
	if m.MemorizeCurrentIndex > 25 {
		startIdx = m.MemorizeCurrentIndex - 25
	}
	var passage strings.Builder
	if startIdx > 0 {
		passage.WriteString("... ")
	}
	for i := startIdx; i < m.MemorizeCurrentIndex; i++ {
		item := m.MemorizeActiveText.Words[i]
		if item.Prefix != "" {
			passage.WriteString(item.Prefix)
		} else if i > startIdx {
			passage.WriteString(" ")
		}
		passage.WriteString(item.Word)
	}

	activeItem := m.MemorizeActiveText.Words[m.MemorizeCurrentIndex]
	if activeItem.Prefix != "" {
		passage.WriteString(activeItem.Prefix)
	} else if m.MemorizeCurrentIndex > 0 {
		passage.WriteString(" ")
	}
	passage.WriteString("[TARGET_BLANK]")

	parts := strings.Split(passage.String(), "[TARGET_BLANK]")
	renderedPassage := MemorizeCompleted.Render(parts[0]) + MemorizeTargetBlank.Render("  ?  ")
	if len(parts) > 1 {
		renderedPassage += MemorizeCompleted.Render(parts[1])
	}

	cardBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#BD93F9")).
		Padding(1, 2).
		Width(66).
		Align(lipgloss.Left)

	b.WriteString(cardBox.Render(renderedPassage) + "\n\n")

	if m.MemorizeShowFeedback {
		wrongWord := ""
		if m.MemorizeSelectedOption >= 0 && m.MemorizeSelectedOption < len(m.MemorizeCurrentOptions) {
			wrongWord = m.MemorizeCurrentOptions[m.MemorizeSelectedOption]
		}
		b.WriteString(MemorizeWrongStyle.Render(fmt.Sprintf("✗ Incorrect: \"%s\"", wrongWord)) + "\n")
		b.WriteString(MemorizeSuccessStyle.Render(fmt.Sprintf("✓ The correct word was: \"%s\"", activeItem.Word)) + "\n\n")
		b.WriteString(HintStyle.Render("[ Spacebar / Enter to continue  •  [r] Retry this word ]"))
		return b.String()
	}

	b.WriteString("Choose the next word:\n\n")
	for i, opt := range m.MemorizeCurrentOptions {
		key := IndexToKey(i)
		keyBadge := KeyStyle.Render(fmt.Sprintf("[%s]", key))
		numBadge := HintStyle.Render(fmt.Sprintf("(%d)", i+1))
		optRender := MemorizeCompleted.Render(opt)
		b.WriteString(fmt.Sprintf("  %s %s  %s\n", keyBadge, numBadge, optRender))
	}

	b.WriteString("\n" + HintStyle.Render("(Press key [a/s/d/f/g] or [1-5] to select  •  [r] Restart  •  [Esc] Back  •  [q] Quit)"))
	return b.String()
}

func (m Model) viewMemorizeComplete() string {
	title := "★ TEXT MEMORIZATION COMPLETE! ★"
	totalWords := 0
	if m.MemorizeActiveText != nil {
		totalWords = len(m.MemorizeActiveText.Words)
	}
	attempts := totalWords + m.MemorizeMistakes
	acc := 100.0
	if attempts > 0 {
		acc = float64(totalWords) / float64(attempts) * 100
	}

	var b strings.Builder
	b.WriteString(MemorizeSuccessStyle.Render(title) + "\n\n")
	if m.MemorizeActiveText != nil {
		b.WriteString(MemorizeHeaderStyle.Render(m.MemorizeActiveText.Title) + "\n\n")
	}
	b.WriteString(fmt.Sprintf("Total Words:    %d\n", totalWords))
	b.WriteString(fmt.Sprintf("First-Try Acc:  %.1f%%\n", acc))
	b.WriteString(fmt.Sprintf("Mistakes:       %d\n", m.MemorizeMistakes))
	b.WriteString(fmt.Sprintf("Best Streak:    %d words\n\n", m.MemorizeBestStreak))

	b.WriteString(HintStyle.Render("[r] Memorize again  •  [Enter] Select another text  •  [q] Quit"))

	cardBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#50FA7B")).
		Padding(1, 4).
		Align(lipgloss.Center)
	return cardBox.Render(b.String())
}

