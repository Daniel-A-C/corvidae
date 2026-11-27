package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"flashcards/internal/deck"

	"github.com/charmbracelet/lipgloss"
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
	case StateDashboard:
		content = m.viewDashboard()
	case StateSettings:
		content = m.viewSettings()
	case StateForgeMenu:
		content = m.viewForgeMenu()
	case StateForgePlay:
		content = m.viewForgePlay()
	case StateForgeProve:
		content = m.viewForgeProve()
	case StateForgeShop:
		content = m.viewForgeShop()
	case StateForgeRunOver:
		content = m.viewForgeRunOver()
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
	case ModeForge:
		return "Morpheme Forge"
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
		{"Memorize Texts", "Options or first-letter recitation + SRS"},
		{"Learn by Reading", "Diglot weave immersion reader"},
		{"Matching Game", "Two-column home-row speed pairs"},
		{"Morpheme Forge (Mandarin)", "Hanzi compound & pivot-chain roguelike"},
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

	b.WriteString("\n")
	if markedCount > 0 {
		var cardWord string
		if markedCount == 1 {
			cardWord = "1 card"
		} else {
			cardWord = fmt.Sprintf("%d cards", markedCount)
		}
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			KeyStyle.Render("[m]"),
			MarkedStarStyle.Render(fmt.Sprintf("★ Practice Marked Cards (%s)", cardWord)),
		))
	}
	b.WriteString(fmt.Sprintf("  %s  %s    %s  %s\n",
		KeyStyle.Render("[i]"),
		SectionTitleStyle.Render("Study Dashboard & Heatmap"),
		KeyStyle.Render("[c]"),
		HintStyle.Render("Settings (config.yaml)"),
	))

	if m.StatusMessage != "" {
		b.WriteString("\n" + QuizExplanationStyle.Width(frameWidth).Render(m.StatusMessage) + "\n")
	}

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	footerRaw := fmt.Sprintf("[%s] Select  •  [%s] Confirm  •  [%s] Dashboard  •  [%s] Settings  •  [%s] Help  •  [%s] Quit",
		KeyStyle.Render("a-j"), KeyStyle.Render("Enter"), KeyStyle.Render("i"), KeyStyle.Render("c"), KeyStyle.Render("?"), KeyStyle.Render("q"))
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
