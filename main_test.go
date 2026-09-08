package main

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestGetDeckDirectories(t *testing.T) {
	dirs, err := getDeckDirectories()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"Arabic", "French", "Mandarin", "Polish", "Spanish"}
	if len(dirs) != len(expected) {
		t.Fatalf("expected %d dirs, got %d: %v", len(expected), len(dirs), dirs)
	}

	for i, name := range expected {
		if dirs[i] != name {
			t.Errorf("expected dir %d to be %s, got %s", i, name, dirs[i])
		}
	}
}

func TestGetDeckFiles(t *testing.T) {
	mandarinFiles, err := getDeckFiles("Mandarin")
	if err != nil {
		t.Fatalf("failed to get Mandarin files: %v", err)
	}
	if len(mandarinFiles) != 12 {
		t.Errorf("expected 12 Mandarin decks, got %d", len(mandarinFiles))
	}

	spanishFiles, err := getDeckFiles("Spanish")
	if err != nil {
		t.Fatalf("failed to get Spanish files: %v", err)
	}
	if len(spanishFiles) != 0 {
		t.Errorf("expected 0 Spanish decks, got %d", len(spanishFiles))
	}
}

func TestCountDecksInDir(t *testing.T) {
	if count := countDecksInDir("Mandarin"); count != 12 {
		t.Errorf("expected 12 decks in Mandarin, got %d", count)
	}
	if count := countDecksInDir("Polish"); count != 0 {
		t.Errorf("expected 0 decks in Polish, got %d", count)
	}
}

func TestLoadDeck(t *testing.T) {
	path := filepath.Join("decks", "Mandarin", "basics.yaml")
	deck, err := loadDeck(path)
	if err != nil {
		t.Fatalf("failed to load %s: %v", path, err)
	}
	if len(deck.Cards) == 0 {
		t.Fatalf("expected cards in basics.yaml, got 0")
	}
}

func TestModelSelectionFlow(t *testing.T) {
	m := initialModel()
	if m.state != stateModeSelect {
		t.Fatalf("expected initial state to be stateModeSelect, got %d", m.state)
	}
}

func TestTeaNavigation(t *testing.T) {
	m := initialModel()

	// 1. Enter from ModeSelect -> DirSelect
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(model)
	if m.state != stateDirSelect {
		t.Fatalf("expected stateDirSelect, got %d", m.state)
	}

	// 2. Navigate to Mandarin (it's index 2: Arabic=0, French=1, Mandarin=2)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedM.(model)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedM.(model)
	if m.dirCursor != 2 || m.dirs[m.dirCursor] != "Mandarin" {
		t.Fatalf("expected cursor at Mandarin, got index %d (%s)", m.dirCursor, m.dirs[m.dirCursor])
	}

	// 3. Enter on Mandarin -> DeckSelect
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(model)
	if m.state != stateDeckSelect {
		t.Fatalf("expected stateDeckSelect, got %d", m.state)
	}
	if m.selectedDir != "Mandarin" {
		t.Fatalf("expected selectedDir to be Mandarin, got %s", m.selectedDir)
	}
	if len(m.deckFiles) != 12 {
		t.Fatalf("expected 12 decks, got %d", len(m.deckFiles))
	}

	// 4. Test Esc to go back to DirSelect
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(model)
	if m.state != stateDirSelect {
		t.Fatalf("expected stateDirSelect after Esc, got %d", m.state)
	}

	// 5. Test Esc to go back to ModeSelect
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(model)
	if m.state != stateModeSelect {
		t.Fatalf("expected stateModeSelect after Esc, got %d", m.state)
	}

	// 6. Navigate to DirSelect -> Spanish (empty dir) -> DeckSelect
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(model)
	// Spanish is at index 4 (Arabic=0, French=1, Mandarin=2, Polish=3, Spanish=4)
	m.dirCursor = 4
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(model)
	if m.state != stateDeckSelect {
		t.Fatalf("expected stateDeckSelect for Spanish, got %d", m.state)
	}
	if len(m.deckFiles) != 0 {
		t.Fatalf("expected 0 decks in Spanish, got %d", len(m.deckFiles))
	}
	// Enter on empty directory returns to DirSelect
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(model)
	if m.state != stateDirSelect {
		t.Fatalf("expected stateDirSelect after pressing enter on empty dir, got %d", m.state)
	}
}

func TestDeckSelectionAndSession(t *testing.T) {
	m := initialModel()

	// Enter DirSelect
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(model)

	// Cursor to Mandarin
	for i, d := range m.dirs {
		if d == "Mandarin" {
			m.dirCursor = i
			break
		}
	}

	// Enter DeckSelect
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(model)

	// Toggle selection for first deck
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updatedM.(model)
	firstDeck := m.deckFiles[0]
	if !m.selectedFiles[firstDeck] {
		t.Fatalf("expected %s to be selected", firstDeck)
	}

	// Confirm selection -> enters stateReview
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(model)
	if m.state != stateReview {
		t.Fatalf("expected stateReview, got %d", m.state)
	}
	if len(m.decks) != 1 {
		t.Fatalf("expected 1 loaded deck, got %d", len(m.decks))
	}
}

func TestViews(t *testing.T) {
	m := initialModel()
	m.width = 80
	m.height = 24

	// stateModeSelect view
	view := m.View()
	if view == "" {
		t.Fatalf("expected non-empty view for stateModeSelect")
	}

	// stateDirSelect view
	m.state = stateDirSelect
	view = m.View()
	if view == "" {
		t.Fatalf("expected non-empty view for stateDirSelect")
	}

	// stateDeckSelect view
	m.state = stateDeckSelect
	m.selectedDir = "Mandarin"
	m.deckFiles, _ = getDeckFiles("Mandarin")
	view = m.View()
	if view == "" {
		t.Fatalf("expected non-empty view for stateDeckSelect")
	}
}

func TestQuizModeCorrectAnswer(t *testing.T) {
	m := initialModel()
	m.width = 120
	m.height = 40
	m.mode = modeQuiz
	m.selectedDir = "Mandarin"
	m.selectedFiles["tech_engineering.yaml"] = true

	err := m.loadSelectedDecks()
	if err != nil {
		t.Fatalf("failed to load decks: %v", err)
	}

	m.setupQuiz()
	m.state = stateQuiz

	if len(m.activeCards) == 0 {
		t.Fatalf("expected active cards, got 0")
	}

	keys := []string{"d", "f", "g", "h", "j", "k"}
	correctKey := keys[m.correctIndex]

	ref := m.activeCards[m.currentIndex]
	card := m.decks[ref.filename].Cards[ref.origIdx]

	// Send the correct answer key
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(correctKey)})
	m = updatedM.(model)

	if !m.showFeedback {
		t.Fatalf("expected showFeedback to be true after answering")
	}
	if !m.isCorrect {
		t.Fatalf("expected isCorrect to be true for correct answer")
	}

	view := m.View()
	if !strings.Contains(view, "Correct!") {
		t.Errorf("expected view to contain 'Correct!', got:\n%s", view)
	}
	if !strings.Contains(view, card.Pinyin) {
		t.Errorf("expected view to contain pinyin '%s'", card.Pinyin)
	}
	if !strings.Contains(view, card.Meaning) {
		t.Errorf("expected view to contain meaning '%s'", card.Meaning)
	}
	if card.Explanation != "" {
		normalizedView := strings.Join(strings.Fields(view), " ")
		if !strings.Contains(normalizedView, card.Explanation) {
			t.Errorf("expected view to contain explanation '%s'", card.Explanation)
		}
	}
	if !strings.Contains(view, "Spacebar to continue") {
		t.Errorf("expected view to contain continue prompt")
	}

	// Press spacebar to advance
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updatedM.(model)

	if m.showFeedback {
		t.Fatalf("expected showFeedback to be false after spacebar")
	}
	if m.currentIndex != 1 {
		t.Fatalf("expected currentIndex to advance to 1, got %d", m.currentIndex)
	}
}

func TestQuizModeIncorrectAnswer(t *testing.T) {
	m := initialModel()
	m.width = 120
	m.height = 40
	m.mode = modeQuiz
	m.selectedDir = "Mandarin"
	m.selectedFiles["tech_engineering.yaml"] = true

	err := m.loadSelectedDecks()
	if err != nil {
		t.Fatalf("failed to load decks: %v", err)
	}

	m.setupQuiz()
	m.state = stateQuiz

	if len(m.activeCards) == 0 {
		t.Fatalf("expected active cards, got 0")
	}

	keys := []string{"d", "f", "g", "h", "j", "k"}
	wrongIdx := (m.correctIndex + 1) % len(m.quizOptions)
	wrongKey := keys[wrongIdx]

	ref := m.activeCards[m.currentIndex]
	card := m.decks[ref.filename].Cards[ref.origIdx]

	// Send an incorrect answer key
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(wrongKey)})
	m = updatedM.(model)

	if !m.showFeedback {
		t.Fatalf("expected showFeedback to be true after answering")
	}
	if m.isCorrect {
		t.Fatalf("expected isCorrect to be false for incorrect answer")
	}

	view := m.View()
	if !strings.Contains(view, "Incorrect!") {
		t.Errorf("expected view to contain 'Incorrect!', got:\n%s", view)
	}
	if !strings.Contains(view, "The correct answer was:") {
		t.Errorf("expected view to contain 'The correct answer was:'")
	}
	if card.Explanation != "" {
		normalizedView := strings.Join(strings.Fields(view), " ")
		if !strings.Contains(normalizedView, card.Explanation) {
			t.Errorf("expected view to contain explanation '%s'", card.Explanation)
		}
	}
	if !strings.Contains(view, "Spacebar to continue") {
		t.Errorf("expected view to contain continue prompt")
	}

	// Press Enter to advance
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(model)

	if m.showFeedback {
		t.Fatalf("expected showFeedback to be false after enter")
	}
	if m.currentIndex != 1 {
		t.Fatalf("expected currentIndex to advance to 1, got %d", m.currentIndex)
	}
}

func TestQuizModeWithoutExplanation(t *testing.T) {
	m := initialModel()
	m.width = 120
	m.height = 40
	m.mode = modeQuiz
	m.selectedDir = "Mandarin"
	m.selectedFiles["kitchen.yaml"] = true

	err := m.loadSelectedDecks()
	if err != nil {
		t.Fatalf("failed to load decks: %v", err)
	}

	m.setupQuiz()
	m.state = stateQuiz

	// Correct answer flow
	keys := []string{"d", "f", "g", "h", "j", "k"}
	correctKey := keys[m.correctIndex]

	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(correctKey)})
	m = updatedM.(model)

	if !m.showFeedback || !m.isCorrect {
		t.Fatalf("expected correct feedback state")
	}

	view := m.View()
	if !strings.Contains(view, "Correct!") {
		t.Errorf("expected view to contain 'Correct!', got:\n%s", view)
	}
}


