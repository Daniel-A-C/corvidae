package ui

import (
	"fmt"
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

	correctKey := HummingbirdKeys[m.CorrectIndex]

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

func TestHummingbirdKeyMapping(t *testing.T) {
	if len(HummingbirdKeys) != 30 {
		t.Fatalf("expected 30 HummingbirdKeys, got %d", len(HummingbirdKeys))
	}

	expectedOrder := []string{
		// Home row left (5)
		"a", "s", "d", "f", "g",
		// Home row right (5)
		"h", "j", "k", "l", ";",
		// Bottom row left (5)
		"z", "x", "c", "v", "b",
		// Bottom row right (5)
		"n", "m", ",", ".", "/",
		// Top row left (5)
		"q", "w", "e", "r", "t",
		// Top row right (5)
		"y", "u", "i", "o", "p",
	}

	for i, expected := range expectedOrder {
		if HummingbirdKeys[i] != expected {
			t.Errorf("key index %d: expected %s, got %s", i, expected, HummingbirdKeys[i])
		}
		if idx := KeyToIndex(expected); idx != i {
			t.Errorf("KeyToIndex(%s): expected %d, got %d", expected, i, idx)
		}
		if key := IndexToKey(i); key != expected {
			t.Errorf("IndexToKey(%d): expected %s, got %s", i, expected, key)
		}
	}

	// Test case-insensitivity
	if KeyToIndex("A") != 0 || KeyToIndex("G") != 4 || KeyToIndex("H") != 5 {
		t.Errorf("KeyToIndex should be case-insensitive")
	}

	// Test out of bounds
	if KeyToIndex("unknown") != -1 {
		t.Errorf("expected -1 for unknown key")
	}
	if IndexToKey(30) != "" || IndexToKey(-1) != "" {
		t.Errorf("expected empty string for out of bounds IndexToKey")
	}
}

func TestHummingbirdModeSelect(t *testing.T) {
	decksDir := setupTestDecks(t)
	m := New(decksDir)

	// Direct select 's' -> Quiz mode
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	mQuiz := updatedM.(Model)
	if mQuiz.State != StateDirSelect {
		t.Fatalf("expected StateDirSelect, got %d", mQuiz.State)
	}
	if mQuiz.Mode != ModeQuiz {
		t.Fatalf("expected ModeQuiz, got %d", mQuiz.Mode)
	}

	// Direct select 'a' -> Review mode
	m = New(decksDir)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	mReview := updatedM.(Model)
	if mReview.State != StateDirSelect {
		t.Fatalf("expected StateDirSelect, got %d", mReview.State)
	}
	if mReview.Mode != ModeReview {
		t.Fatalf("expected ModeReview, got %d", mReview.Mode)
	}
}

func TestHummingbirdDirSelect(t *testing.T) {
	tempDir := t.TempDir()
	// Create 7 directories to cross the group of 5 boundary:
	// 0: Arabic ('a'), 1: Chinese ('s'), 2: French ('d'), 3: German ('f'), 4: Italian ('g')
	// 5: Japanese ('h'), 6: Russian ('j')
	dirNames := []string{"Arabic", "Chinese", "French", "German", "Italian", "Japanese", "Russian"}
	for _, d := range dirNames {
		if err := os.Mkdir(filepath.Join(tempDir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	m := New(tempDir)
	// Advance to DirSelect
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updatedM.(Model)

	if len(m.Dirs) != 7 {
		t.Fatalf("expected 7 dirs, got %d", len(m.Dirs))
	}

	// Direct select 'h' (index 5 -> Japanese)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m2 := updatedM.(Model)
	if m2.State != StateDeckSelect {
		t.Fatalf("expected StateDeckSelect, got %d", m2.State)
	}
	if m2.SelectedDir != "Japanese" {
		t.Fatalf("expected Japanese, got %s", m2.SelectedDir)
	}

	// Direct select 'j' (index 6 -> Russian)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m3 := updatedM.(Model)
	if m3.State != StateDeckSelect {
		t.Fatalf("expected StateDeckSelect, got %d", m3.State)
	}
	if m3.SelectedDir != "Russian" {
		t.Fatalf("expected Russian, got %s", m3.SelectedDir)
	}
}

func TestHummingbirdDeckSelect(t *testing.T) {
	m := Model{
		State:         StateDeckSelect,
		DeckFiles:     []string{"deck0.yaml", "deck1.yaml", "deck2.yaml", "deck3.yaml", "deck4.yaml", "deck5.yaml"},
		SelectedFiles: make(map[string]bool),
	}

	// Press 'a' (index 0) to toggle deck0.yaml
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updatedM.(Model)
	if !m.SelectedFiles["deck0.yaml"] {
		t.Fatalf("expected deck0.yaml to be selected")
	}
	if m.Cursor != 0 {
		t.Fatalf("expected cursor to be 0, got %d", m.Cursor)
	}

	// Press 'h' (index 5) to toggle deck5.yaml
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = updatedM.(Model)
	if !m.SelectedFiles["deck5.yaml"] {
		t.Fatalf("expected deck5.yaml to be selected")
	}
	if m.Cursor != 5 {
		t.Fatalf("expected cursor to be 5, got %d", m.Cursor)
	}

	// Press 'a' again to un-toggle deck0.yaml
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updatedM.(Model)
	if m.SelectedFiles["deck0.yaml"] {
		t.Fatalf("expected deck0.yaml to be unselected")
	}
}

func TestHummingbirdViewSpacing(t *testing.T) {
	m := Model{
		State:     StateDeckSelect,
		DeckFiles: []string{"d0.yaml", "d1.yaml", "d2.yaml", "d3.yaml", "d4.yaml", "d5.yaml"},
		SelectedFiles: make(map[string]bool),
	}

	view := m.View()

	// Check key badges
	if !strings.Contains(view, "[a]") || !strings.Contains(view, "[g]") || !strings.Contains(view, "[h]") {
		t.Errorf("expected view to contain [a], [g], and [h] badges, got:\n%s", view)
	}

	// Check that there is spacing (a blank line) between item 'g' (5th) and item 'h' (6th)
	gIndex := strings.Index(view, "[g]")
	hIndex := strings.Index(view, "[h]")
	if gIndex == -1 || hIndex == -1 || gIndex >= hIndex {
		t.Fatalf("expected [g] before [h]")
	}
	between := view[gIndex:hIndex]
	lines := strings.Split(between, "\n")
	hasBlankLine := false
	for _, l := range lines[1 : len(lines)-1] {
		if strings.TrimSpace(l) == "" {
			hasBlankLine = true
			break
		}
	}
	if !hasBlankLine {
		t.Errorf("expected empty line separator between group 1 and group 2, got:\n%q", between)
	}
}

func TestHummingbirdQuitOrKeyCollision(t *testing.T) {
	// When menu has < 21 items, 'q' quits
	m := Model{
		State:     StateDeckSelect,
		DeckFiles: []string{"deck0.yaml", "deck1.yaml"},
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatalf("expected tea.Quit when 'q' is pressed on small menu")
	}

	// When menu has 21 items, item 20 is 'q' (HummingbirdKeys[20] == "q")
	// Pressing 'q' should toggle item 20 instead of quitting
	deckFiles := make([]string, 21)
	for i := 0; i < 21; i++ {
		deckFiles[i] = fmt.Sprintf("deck%d.yaml", i)
	}
	m21 := Model{
		State:         StateDeckSelect,
		DeckFiles:     deckFiles,
		SelectedFiles: make(map[string]bool),
	}
	updatedM, cmd2 := m21.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd2 != nil {
		t.Fatalf("expected no quit command when 'q' matches an active menu item")
	}
	resM := updatedM.(Model)
	if !resM.SelectedFiles["deck20.yaml"] {
		t.Fatalf("expected deck20.yaml to be selected via 'q'")
	}
}

func TestHummingbirdRealDecksRendering(t *testing.T) {
	decksPath := filepath.Join("..", "..", "decks")
	if _, err := os.Stat(decksPath); os.IsNotExist(err) {
		t.Skip("decks directory not present at expected relative path")
	}

	m := New(decksPath)
	m.State = StateDeckSelect
	m.SelectedDir = "Mandarin"
	files, err := deck.GetDeckFiles(decksPath, "Mandarin")
	if err != nil {
		t.Fatal(err)
	}
	m.DeckFiles = files
	if len(m.DeckFiles) < 10 {
		t.Skip("Mandarin directory has fewer than 10 files")
	}

	view := m.View()

	// Verify group 1: keys a, s, d, f, g
	for _, k := range []string{"[a]", "[s]", "[d]", "[f]", "[g]"} {
		if !strings.Contains(view, k) {
			t.Errorf("expected view to contain %s", k)
		}
	}
	// Verify group 2: keys h, j, k, l, ;
	for _, k := range []string{"[h]", "[j]", "[k]", "[l]", "[;]"} {
		if !strings.Contains(view, k) {
			t.Errorf("expected view to contain %s", k)
		}
	}
	// Verify group 3: keys z, x, c, v
	for _, k := range []string{"[z]", "[x]", "[c]", "[v]"} {
		if !strings.Contains(view, k) {
			t.Errorf("expected view to contain %s", k)
		}
	}

	// Verify spacing between group 1 and 2
	gIdx := strings.Index(view, "[g]")
	hIdx := strings.Index(view, "[h]")
	between1 := view[gIdx:hIdx]
	lines1 := strings.Split(between1, "\n")
	hasBlank1 := false
	for _, l := range lines1[1 : len(lines1)-1] {
		if strings.TrimSpace(l) == "" {
			hasBlank1 = true
			break
		}
	}
	if !hasBlank1 {
		t.Errorf("expected blank line between [g] and [h]")
	}

	// Verify spacing between group 2 and 3
	semiIdx := strings.Index(view, "[;]")
	zIdx := strings.Index(view, "[z]")
	between2 := view[semiIdx:zIdx]
	lines2 := strings.Split(between2, "\n")
	hasBlank2 := false
	for _, l := range lines2[1 : len(lines2)-1] {
		if strings.TrimSpace(l) == "" {
			hasBlank2 = true
			break
		}
	}
	if !hasBlank2 {
		t.Errorf("expected blank line between [;] and [z]")
	}
}

