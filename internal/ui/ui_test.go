package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"flashcards/internal/deck"
)

func setupTestDecks(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	mandarinDir := filepath.Join(tempDir, "Mandarin")
	frenchDir := filepath.Join(tempDir, "French")
	if err := os.Mkdir(mandarinDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(frenchDir, 0755); err != nil {
		t.Fatal(err)
	}

	mandarinDeck := deck.Deck{
		Cards: []deck.Flashcard{
			{
				Character:   "电脑",
				Pinyin:      "diànnǎo",
				Meaning:     "Computer",
				Explanation: "Electric brain",
			},
			{
				Character:   "手机",
				Pinyin:      "shǒujī",
				Meaning:     "Mobile phone",
				Explanation: "Hand machine",
			},
		},
	}
	if err := deck.SaveDeck(filepath.Join(mandarinDir, "tech.yaml"), mandarinDeck); err != nil {
		t.Fatal(err)
	}

	frenchDeck := deck.Deck{
		Cards: []deck.Flashcard{
			{
				Character: "Bonjour",
				Meaning:   "Hello",
			},
		},
	}
	if err := deck.SaveDeck(filepath.Join(frenchDir, "basics.yaml"), frenchDeck); err != nil {
		t.Fatal(err)
	}

	return tempDir
}

func TestNavigationFlow(t *testing.T) {
	decksDir := setupTestDecks(t)
	m := New(decksDir)

	if m.State != StateModeSelect {
		t.Fatalf("expected StateModeSelect, got %d", m.State)
	}

	// 1. Enter from ModeSelect -> DirSelect
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	if m.State != StateDirSelect {
		t.Fatalf("expected StateDirSelect, got %d", m.State)
	}

	// French=0, Mandarin=1
	// Down arrow to Mandarin
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedM.(Model)
	if m.DirCursor != 1 || m.Dirs[m.DirCursor] != "Mandarin" {
		t.Fatalf("expected cursor at Mandarin, got index %d (%s)", m.DirCursor, m.Dirs[m.DirCursor])
	}

	// Enter on Mandarin -> DeckSelect
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	if m.State != StateDeckSelect {
		t.Fatalf("expected StateDeckSelect, got %d", m.State)
	}
	if len(m.DeckFiles) != 1 || m.DeckFiles[0] != "tech.yaml" {
		t.Fatalf("expected [tech.yaml], got %v", m.DeckFiles)
	}

	// Space to toggle tech.yaml
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updatedM.(Model)
	if !m.SelectedFiles["tech.yaml"] {
		t.Fatalf("expected tech.yaml to be selected")
	}

	// Confirm selection -> enters StateReview
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	if m.State != StateReview {
		t.Fatalf("expected StateReview, got %d", m.State)
	}
	if len(m.ActiveCards) != 2 {
		t.Fatalf("expected 2 active cards, got %d", len(m.ActiveCards))
	}
}

func TestReviewGradingFlow(t *testing.T) {
	decksDir := setupTestDecks(t)
	m := New(decksDir)
	m.SelectedDir = "Mandarin"
	m.SelectedFiles["tech.yaml"] = true
	if err := m.LoadSelectedDecks(); err != nil {
		t.Fatal(err)
	}
	m.SetupReview()
	m.State = StateReview

	// Reveal answer
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updatedM.(Model)
	if !m.ShowAnswer {
		t.Fatalf("expected ShowAnswer to be true")
	}

	// Toggle explanation
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updatedM.(Model)
	if !m.ShowExplanation {
		t.Fatalf("expected ShowExplanation to be true")
	}

	// Grade card with 'h' (GradeGood)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = updatedM.(Model)
	if m.CurrentIndex != 1 {
		t.Fatalf("expected CurrentIndex to advance to 1, got %d", m.CurrentIndex)
	}
	if m.ShowAnswer {
		t.Fatalf("expected ShowAnswer to be reset to false")
	}
}

func TestQuizFlowAndFeedback(t *testing.T) {
	decksDir := setupTestDecks(t)
	m := New(decksDir)
	m.Mode = ModeQuiz
	m.SelectedDir = "Mandarin"
	m.SelectedFiles["tech.yaml"] = true
	if err := m.LoadSelectedDecks(); err != nil {
		t.Fatal(err)
	}
	m.SetupQuiz()
	m.State = StateQuiz

	keys := []string{"d", "f", "g", "h", "j", "k"}
	correctKey := keys[m.CorrectIndex]

	// Answer correctly
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(correctKey)})
	m = updatedM.(Model)

	if !m.ShowFeedback || !m.IsCorrect {
		t.Fatalf("expected correct feedback state")
	}

	view := m.View()
	if !strings.Contains(view, "Correct!") {
		t.Errorf("expected view to contain 'Correct!', got:\n%s", view)
	}

	// Advance
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updatedM.(Model)
	if m.ShowFeedback {
		t.Fatalf("expected ShowFeedback to be false")
	}
	if m.CurrentIndex != 1 {
		t.Fatalf("expected CurrentIndex to be 1, got %d", m.CurrentIndex)
	}
}

func TestFrenchDeckWithoutPinyin(t *testing.T) {
	decksDir := setupTestDecks(t)
	m := New(decksDir)
	m.SelectedDir = "French"
	m.SelectedFiles["basics.yaml"] = true
	if err := m.LoadSelectedDecks(); err != nil {
		t.Fatal(err)
	}
	m.SetupReview()
	m.State = StateReview
	m.ShowAnswer = true

	view := m.View()
	if strings.Contains(view, "Pinyin:") {
		t.Errorf("deck without pinyin should not display 'Pinyin:', got:\n%s", view)
	}
	if !strings.Contains(view, "Meaning: Hello") {
		t.Errorf("expected view to contain 'Meaning: Hello', got:\n%s", view)
	}
}
