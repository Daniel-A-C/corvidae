package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"flashcards/internal/arabic"
	"flashcards/internal/deck"
	"flashcards/internal/reading"
)

// View renders the terminal user interface according to the current state.
func (m Model) View() string {
	if m.Err != nil {
		return fmt.Sprintf("Error: %v\n", m.Err)
	}
	if m.ShowHelp {
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.viewHelpOverlay())
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
	case StateReadingSelectText:
		content = m.viewReadingSelectText()
	case StateReading:
		content = m.viewReading()
	case StateReadingWordDetail:
		content = m.viewReadingWordDetail()
	case StateReadingDiscussion:
		content = m.viewReadingDiscussion()
	case StateReadingComplete:
		content = m.viewReadingComplete()
	case StateMatch:
		content = m.viewMatch()
	}

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) modeDisplayName() string {
	switch m.Mode {
	case ModeReview:
		return "SM-2 Review"
	case ModeQuiz:
		return "Quiz"
	case ModeArabic:
		return "Arabic Academy"
	case ModeMemorize:
		return "Memorize"
	case ModeReading:
		return "Reading"
	case ModeMatch:
		return "Matching Game"
	default:
		return "Practice"
	}
}

func (m Model) viewModeSelect() string {
	frameWidth := m.frameWidth()

	logo := lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Center).Render(LogoStyle.Render(AsciiLogo))
	tagline := lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Center).Render(HintStyle.Render("Terminal Memory Workstation"))

	var b strings.Builder
	b.WriteString(logo + "\n")
	b.WriteString(tagline + "\n\n")

	// Status strip
	selectedCount := m.countSelectedDecks()
	markedCount := len(m.MarkedCards)
	var pills []string
	if selectedCount == 1 {
		pills = append(pills, DashboardPillStyle.Render("1 deck selected"))
	} else if selectedCount > 1 {
		pills = append(pills, DashboardPillStyle.Render(fmt.Sprintf("%d decks selected", selectedCount)))
	}
	if markedCount == 1 {
		pills = append(pills, DashboardPillStyle.Render(MarkedStarStyle.Render("★")+" 1 marked"))
	} else if markedCount > 1 {
		pills = append(pills, DashboardPillStyle.Render(MarkedStarStyle.Render("★")+fmt.Sprintf(" %d marked", markedCount)))
	}

	headerLeft := SectionTitleStyle.Render("Select Mode")
	headerRight := strings.Join(pills, "  ")
	b.WriteString(renderSplitLine(headerLeft, headerRight, frameWidth) + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n\n")

	modes := []struct {
		title string
		desc  string
	}{
		{"Spaced Repetition (SM-2)", "Daily due flashcards & scheduling"},
		{"Multiple Choice Quiz", "5-choice recall with live streak"},
		{"Arabic Alphabet Academy", "28 letters, cursive forms & drills"},
		{"Memorize by Options", "Word-by-word text recitation"},
		{"Learn by Reading", "Diglot weave immersion reader"},
		{"Matching Game", "Two-column home-row speed pairs"},
	}

	const titleColWidth = 27
	for i, mode := range modes {
		key := IndexToKey(i)
		keyBadge := KeyStyle.Render(fmt.Sprintf("[%s]", key))
		cursor := "  "
		paddedTitle := fmt.Sprintf("%-*s", titleColWidth, mode.title)
		label := paddedTitle
		if m.Mode == i {
			cursor = "> "
			label = CursorStyle.Render(paddedTitle)
		}
		if frameWidth >= 60 {
			desc := HintStyle.Render(truncateToWidth(mode.desc, frameWidth-titleColWidth-9))
			b.WriteString(fmt.Sprintf("%s%s  %s %s\n", CursorStyle.Render(cursor), keyBadge, label, desc))
		} else {
			b.WriteString(fmt.Sprintf("%s%s  %s\n", CursorStyle.Render(cursor), keyBadge, label))
		}
	}

	if markedCount > 0 {
		var cardWord string
		if markedCount == 1 {
			cardWord = "1 card"
		} else {
			cardWord = fmt.Sprintf("%d cards", markedCount)
		}
		b.WriteString("\n" + fmt.Sprintf("  %s  %s\n",
			KeyStyle.Render("[m]"),
			MarkedStarStyle.Render(fmt.Sprintf("★ Practice Marked Cards (%s)", cardWord)),
		))
	}

	if m.StatusMessage != "" {
		b.WriteString("\n" + QuizExplanationStyle.Width(frameWidth).Render(m.StatusMessage) + "\n")
	}

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	footerRaw := fmt.Sprintf("[%s] Select  •  [%s] Confirm  •  [%s] Help  •  [%s] Quit",
		KeyStyle.Render("a-h"), KeyStyle.Render("Enter"), KeyStyle.Render("?"), KeyStyle.Render("q"))
	b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))

	return renderFrame(frameWidth, b.String())
}

func (m Model) viewDirSelect() string {
	frameWidth := m.frameWidth()
	if len(m.Dirs) == 0 && m.SelectedDir != "" {
		empty := fmt.Sprintf("No directories found in %s/%s.\n\n%s",
			m.BaseDir, m.SelectedDir, HintStyle.Render("[Esc] Back  •  [q] Quit"))
		return renderFrame(frameWidth, empty)
	}

	var breadcrumb string
	if m.SelectedDir == "" {
		breadcrumb = "Decks"
	} else {
		breadcrumb = "Decks › " + strings.ReplaceAll(filepath.ToSlash(m.SelectedDir), "/", " › ")
	}
	modePill := ModePillStyle.Render(m.modeDisplayName())
	availHeader := frameWidth - lipgloss.Width(modePill) - 2
	if availHeader < 12 {
		availHeader = 12
	}
	headerLine := renderSplitLine(BreadcrumbStyle.Render(truncateToWidth(breadcrumb, availHeader)), modePill, frameWidth)

	var b strings.Builder
	b.WriteString(headerLine + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n")

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
			countStr = HintStyle.Render("(1 deck)")
		} else {
			countStr = HintStyle.Render(fmt.Sprintf("(%d decks)", count))
		}

		selectedCount, totalDecks := m.countSelectedDecksInDir(dirRelPath)
		var indicator string
		if totalDecks > 0 {
			if selectedCount == totalDecks {
				indicator = "  " + FolderSelectedStyle.Render("[✓ all]")
			} else if selectedCount > 0 {
				indicator = "  " + FolderSubsetStyle.Render(fmt.Sprintf("[%d/%d selected]", selectedCount, totalDecks))
			}
		}

		if m.DirCursor == i {
			cursor = "> "
			label = CursorStyle.Render(label)
		}
		leftPart := fmt.Sprintf("%s%s%s", CursorStyle.Render(cursor), keyBadge, label)
		rightPart := countStr + indicator
		b.WriteString(renderSplitLine(leftPart, rightPart, frameWidth) + "\n")
		if (i+1)%5 == 0 && i < len(m.Dirs)-1 {
			b.WriteString("\n")
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
		b.WriteString(fmt.Sprintf("\n%s%s%s%s\n", CursorStyle.Render(cursor), keyBadge, label, countStr))
	}

	if m.StatusMessage != "" {
		b.WriteString("\n" + QuizExplanationStyle.Width(frameWidth).Render(m.StatusMessage) + "\n")
	}

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	selectedCount := m.countSelectedDecks()
	if selectedCount > 0 {
		var deckWord string
		if selectedCount == 1 {
			deckWord = "1 deck"
		} else {
			deckWord = fmt.Sprintf("%d decks", selectedCount)
		}
		b.WriteString(renderSplitLine(
			CorrectStyle.Render(fmt.Sprintf("%s selected across folders", deckWord)),
			CorrectStyle.Render("[Tab] Start  •  [u] Deselect All"),
			frameWidth,
		) + "\n")
		footerRaw := fmt.Sprintf("[%s] Open  •  [%s] Toggle folder  •  [%s] Start  •  [%s] Back  •  [%s] Quit",
			KeyStyle.Render("key"), KeyStyle.Render("Shift+key"), KeyStyle.Render("Tab"), KeyStyle.Render("Esc"), KeyStyle.Render("q"))
		b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))
	} else {
		footerRaw := fmt.Sprintf("[%s] Open  •  [%s] Toggle folder  •  [%s] Marked  •  [%s] Back  •  [%s] Quit",
			KeyStyle.Render("key"), KeyStyle.Render("Shift+key"), KeyStyle.Render("m"), KeyStyle.Render("Esc"), KeyStyle.Render("q"))
		b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))
	}
	return renderFrame(frameWidth, b.String())
}

func (m Model) viewDeckSelect() string {
	frameWidth := m.frameWidth()
	if len(m.DeckFiles) == 0 {
		empty := fmt.Sprintf("No .yaml files found in %s/%s.\n\n%s",
			m.BaseDir, m.SelectedDir, HintStyle.Render("[Esc / Enter] Back  •  [q] Quit"))
		return renderFrame(frameWidth, empty)
	}

	breadcrumb := "Decks › " + strings.ReplaceAll(filepath.ToSlash(m.SelectedDir), "/", " › ")
	modePill := ModePillStyle.Render(m.modeDisplayName())
	availHeader := frameWidth - lipgloss.Width(modePill) - 2
	if availHeader < 12 {
		availHeader = 12
	}
	headerLine := renderSplitLine(BreadcrumbStyle.Render(truncateToWidth(breadcrumb, availHeader)), modePill, frameWidth)

	var b strings.Builder
	b.WriteString(headerLine + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n")

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

		check := HintStyle.Render("[ ]")
		if m.isDeckSelected(file) {
			check = FolderSelectedStyle.Render("[✓]")
		}

		displayFile := file
		if m.Cursor == i {
			displayFile = CursorStyle.Render(file)
		}
		line := fmt.Sprintf("%s%s%s %s", CursorStyle.Render(cursor), keyBadge, check, displayFile)
		b.WriteString(line + "\n")
		if (i+1)%5 == 0 && i < len(m.DeckFiles)-1 {
			b.WriteString("\n")
		}
	}

	if m.StatusMessage != "" {
		b.WriteString("\n" + QuizExplanationStyle.Width(frameWidth).Render(m.StatusMessage) + "\n")
	}

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	selectedCount := m.countSelectedDecks()
	if selectedCount > 0 {
		var deckWord string
		if selectedCount == 1 {
			deckWord = "1 deck"
		} else {
			deckWord = fmt.Sprintf("%d decks", selectedCount)
		}
		b.WriteString(renderSplitLine(
			CorrectStyle.Render(fmt.Sprintf("%s selected across folders", deckWord)),
			CorrectStyle.Render("[Enter] Start  •  [u] Deselect All"),
			frameWidth,
		) + "\n")
		footerRaw := fmt.Sprintf("[%s] Toggle  •  [%s] Toggle all  •  [%s] Start  •  [%s] Back  •  [%s] Quit",
			KeyStyle.Render("key/Space"), KeyStyle.Render("Shift+Enter"), KeyStyle.Render("Enter"), KeyStyle.Render("Esc"), KeyStyle.Render("q"))
		b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))
	} else {
		footerRaw := fmt.Sprintf("[%s] Toggle  •  [%s] Toggle all  •  [%s] Back  •  [%s] Quit",
			KeyStyle.Render("key/Space"), KeyStyle.Render("Shift+Enter"), KeyStyle.Render("Esc"), KeyStyle.Render("q"))
		b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))
	}
	return renderFrame(frameWidth, b.String())
}

// viewReview is implemented in review_view.go.
// viewQuiz is implemented in quiz_view.go.

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

	// Zone 1: Fixed 2-line Header
	headerLine1 := HintStyle.Render(fmt.Sprintf("Arabic Drill: Question %d of %d  •  Score: %d/%d  •  Accuracy: %s",
		m.ArabicQuestionIdx+1, total, m.ArabicScore, answered, accStr))
	barLen := 26
	if frameWidth < 46 {
		barLen = 16
	}
	headerLine2 := renderProgressBar(answered, total, barLen)
	headerZone := headerLine1 + "\n" + headerLine2

	// Zone 2: Prompt Card
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

	// Zone 3: Fixed-height Body (options or feedback)
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

	// Zone 4: Divider + Fixed Footer
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

func (m Model) viewMemorizeSelectText() string {
	frameWidth := m.frameWidth()
	if len(m.MemorizeTexts) == 0 {
		empty := fmt.Sprintf("No memorization texts found in %s.\n\nPlace .yaml or .txt files in %s/\n\n%s",
			m.TextsDir, m.TextsDir, HintStyle.Render("[Esc] Return  •  [q] Quit"))
		return renderFrame(frameWidth, empty)
	}

	var b strings.Builder
	optPill := ModePillStyle.Render(fmt.Sprintf("Options: %d", m.MemorizeOptionCount))
	b.WriteString(renderSplitLine(MemorizeTitleStyle.Render("★ MEMORIZE BY OPTIONS ★"), optPill, frameWidth) + "\n")
	b.WriteString(MemorizeSubStyle.Render("Work through a text word by word with plausible options") + "\n")
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

		info := fmt.Sprintf("%d words", t.WordCount)
		if t.Progress > 0 && t.WordCount > 0 {
			pct := float64(t.Progress) / float64(t.WordCount) * 100
			info += fmt.Sprintf("  •  %d/%d (%.0f%%)", t.Progress, t.WordCount, pct)
		}
		leftPart := fmt.Sprintf("%s%s%s", CursorStyle.Render(cursor), keyBadge, label)
		b.WriteString(renderSplitLine(leftPart, HintStyle.Render(info), frameWidth) + "\n")
	}

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	footerRaw := fmt.Sprintf("[%s] Select  •  [%s] Options (%d)  •  [%s] Help  •  [%s] Back  •  [%s] Quit",
		KeyStyle.Render("key / Enter"), KeyStyle.Render("o / 2-5"), m.MemorizeOptionCount, KeyStyle.Render("?"), KeyStyle.Render("Esc"), KeyStyle.Render("q"))
	b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))
	return renderFrame(frameWidth, b.String())
}

func (m Model) viewMemorizePortionSelect() string {
	frameWidth := m.frameWidth()
	if m.MemorizeActiveText == nil || len(m.MemorizeSections) == 0 {
		return renderFrame(frameWidth, fmt.Sprintf("No sections found for this text.\n\n%s", HintStyle.Render("[Esc] Back")))
	}

	startSec := m.MemorizeSections[m.PortionStartSection]
	endSec := m.MemorizeSections[m.PortionEndSection]
	portionWords := endSec.EndIdx - startSec.StartIdx
	totalWords := len(m.MemorizeActiveText.Words)
	numSections := m.PortionEndSection - m.PortionStartSection + 1

	var b strings.Builder
	optPill := ModePillStyle.Render(fmt.Sprintf("Options: %d", m.MemorizeOptionCount))
	b.WriteString(renderSplitLine(MemorizeTitleStyle.Render("★ PORTION SELECTOR ★"), optPill, frameWidth) + "\n")
	b.WriteString(MemorizeHeaderStyle.Render(m.MemorizeActiveText.Title) + "\n")
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

	// Compact visual range bar across the full text
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
		secPreview := truncateToWidth(sec.Preview, frameWidth-34)
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
			b.WriteString(fmt.Sprintf("%s%s%s\n", CursorStyle.Render(prefix), MemorizeCompleted.Render(line), tag))
		} else {
			b.WriteString(fmt.Sprintf("%s%s\n", prefix, HintStyle.Render(line)))
		}
	}

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	b.WriteString(HintStyle.Render("[Tab / ↑↓] Switch Start/End  •  [←→ / hl] Shift  •  [x] Expand  •  [n] Advance") + "\n")
	b.WriteString(HintStyle.Render(fmt.Sprintf("[o / 2-5] Options: %d  •  [a] Select All  •  [Enter / Space] Practice  •  [Esc] Back", m.MemorizeOptionCount)))
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

	optPill := ModePillStyle.Render(fmt.Sprintf("Options: %d", len(m.MemorizeCurrentOptions)))
	titleLeft := MemorizeHeaderStyle.Render(fmt.Sprintf("★ %s%s", m.MemorizeActiveText.Title, portionTitle))
	headerLine1 := renderSplitLine(titleLeft, optPill, frameWidth)
	headerLine2 := HintStyle.Render(fmt.Sprintf("Portion Word %d of %d (%.1f%%)  •  Accuracy: %s",
		currentInPortion+1, totalPortionWords, pct, accStr))
	headerLine3 := renderProgressBar(currentInPortion, totalPortionWords, barLen) + "  " + formatStreakBadge(m.MemorizeStreak)

	// Build rolling text window (last ~30 words) in a responsive fixed-size card box
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
	boxHeight := maxContentLines + (padY * 2) // 6 inner height: 4 content + 2 vertical pad

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

	parts := strings.Split(passage.String(), "[TARGET_BLANK]")
	renderedPassage := MemorizeCompleted.Render(parts[0]) + MemorizeTargetBlank.Render("  ?  ")
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

	targetTotal := maxContentLines + (padY * 2) + 2 // 8 lines
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

	// Fixed-height body zone (8 lines) so feedback <-> options never shifts vertically
	const memorizeBodyHeight = 8
	var bodyLines []string
	var footerRaw string
	if m.MemorizeShowFeedback {
		wrongWord := ""
		if m.MemorizeSelectedOption >= 0 && m.MemorizeSelectedOption < len(m.MemorizeCurrentOptions) {
			wrongWord = m.MemorizeCurrentOptions[m.MemorizeSelectedOption]
		}
		bodyLines = append(bodyLines, "  "+MemorizeWrongStyle.Render(fmt.Sprintf("✗ Incorrect: \"%s\"", wrongWord)))
		bodyLines = append(bodyLines, "  "+MemorizeSuccessStyle.Render(fmt.Sprintf("✓ The correct word was: \"%s\"", activeItem.Word)))
		footerRaw = fmt.Sprintf("[%s] Continue  •  [%s] I knew it (wrong button)  •  [%s] Retry word  •  [%s] Options",
			KeyStyle.Render("Spacebar / Enter"), KeyStyle.Render("k"), KeyStyle.Render("r"), KeyStyle.Render("o"))
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
		footerRaw = fmt.Sprintf("(Press key [%s] or [%s] to select  •  [o] Options (%d)  •  [r] Restart  •  [?] Help  •  [Esc] Back)",
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

// viewMatch is implemented in match_view.go.



