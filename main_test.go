package main

import (
	"path/filepath"
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


