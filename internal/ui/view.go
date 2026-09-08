package ui

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
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
	}

	styledContent := lipgloss.NewStyle().Align(lipgloss.Center).Render(content)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, styledContent)
}

func (m Model) viewModeSelect() string {
	content := LogoStyle.Render(AsciiLogo) + "\n\n"
	content += "Select Practice Mode:\n\n"

	modes := []string{"Spaced Repetition (SM-2)", "Multiple Choice Quiz"}
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
	content += "\n" + HintStyle.Render("(Press a/s to select, Enter to confirm, q to quit)")
	return content
}

func (m Model) viewDirSelect() string {
	if len(m.Dirs) == 0 {
		return fmt.Sprintf("No language directories found in %s/.\n\n%s", m.BaseDir, HintStyle.Render("(Press Esc to return, q to quit)"))
	}

	content := "Select Deck Directory / Language:\n\n"
	for i, dir := range m.Dirs {
		key := IndexToKey(i)
		keyBadge := ""
		if key != "" {
			keyBadge = KeyStyle.Render(fmt.Sprintf("[%s]", key)) + " "
		}
		cursor := "  "
		label := dir
		count := deck.CountDecksInDir(m.BaseDir, dir)
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
	content += "\n" + HintStyle.Render("(Press key to select, Esc to go back, q to quit)")
	return content
}

func (m Model) viewDeckSelect() string {
	if len(m.DeckFiles) == 0 {
		return fmt.Sprintf("No .yaml files found in %s/%s.\n\n%s", m.BaseDir, m.SelectedDir, HintStyle.Render("(Press Esc or Enter to go back, q to quit)"))
	}

	content := fmt.Sprintf("Select decks to practice (%s):\n\n", m.SelectedDir)
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
		if m.SelectedFiles[file] {
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

	content := HintStyle.Render(fmt.Sprintf("Card %d of %d  •  %s", m.CurrentIndex+1, len(m.ActiveCards), displayName)) + "\n\n"
	content += CharStyle.Render(card.Character) + "\n\n"

	if !m.ShowAnswer {
		content += HintStyle.Render("[ Spacebar to reveal ]") + "\n"
		content += "\n" + HintStyle.Render("(Press 'q' to quit)")
		return content
	}

	if card.Pinyin != "" {
		content += fmt.Sprintf("Pinyin:  %s\n", PinyinStyle.Render(card.Pinyin))
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

	content += "\n" + HintStyle.Render("[e] Toggle Explanation  •  [q] Quit")
	return content
}

func (m Model) viewQuiz() string {
	if m.CurrentIndex >= len(m.ActiveCards) {
		return fmt.Sprintf("Quiz complete!\n\n%s", HintStyle.Render("[Enter] Main Menu  •  [q] Quit"))
	}

	ref := m.ActiveCards[m.CurrentIndex]
	card := m.Decks[ref.Filename].Cards[ref.OrigIdx]
	displayName, err := filepath.Rel(m.BaseDir, ref.Filename)
	if err != nil {
		displayName = ref.Filename
	}

	content := HintStyle.Render(fmt.Sprintf("Quiz: Question %d of %d  •  %s", m.CurrentIndex+1, len(m.ActiveCards), displayName)) + "\n\n"
	content += CharStyle.Render(card.Character) + "\n\n"

	if m.ShowFeedback {
		if m.IsCorrect {
			content += CorrectStyle.Render("Correct!") + "\n\n"
			if card.Pinyin != "" {
				content += fmt.Sprintf("Pinyin:  %s\n", PinyinStyle.Render(card.Pinyin))
			}
			content += fmt.Sprintf("Meaning: %s\n\n", MeaningStyle.Render(card.Meaning))
			if card.Explanation != "" {
				content += ExplanationStyle.Render(card.Explanation) + "\n\n"
			}
			content += HintStyle.Render("[ Spacebar to continue ]")
		} else {
			content += ErrorStyle.Render("Incorrect!") + "\n\n"
			content += fmt.Sprintf("The correct answer was: %s\n\n", CorrectStyle.Render(m.QuizOptions[m.CorrectIndex]))
			if card.Explanation != "" {
				content += ExplanationStyle.Render(card.Explanation) + "\n\n"
			}
			content += HintStyle.Render("[ Spacebar to continue ]")
		}
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
	content += "\n" + HintStyle.Render("(Press key to select, q to quit)")
	return content
}
