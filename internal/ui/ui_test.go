package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"flashcards/internal/deck"
	"flashcards/internal/reading"
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
	if !m.isDeckSelected("tech.yaml") {
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

	// Verify deck name is hidden while picking an answer
	questionView := m.View()
	if strings.Contains(questionView, "tech.yaml") {
		t.Errorf("expected question view to NOT contain deck name 'tech.yaml', got:\n%s", questionView)
	}
	if !strings.Contains(questionView, "Quiz: Question 1 of") {
		t.Errorf("expected question view to contain 'Quiz: Question 1 of', got:\n%s", questionView)
	}

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
	if !strings.Contains(view, "tech.yaml") {
		t.Errorf("expected feedback view to contain deck name 'tech.yaml', got:\n%s", view)
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

func TestArabicDeckWithPronunciation(t *testing.T) {
	tempDir := t.TempDir()
	arabicDir := filepath.Join(tempDir, "Arabic")
	if err := os.Mkdir(arabicDir, 0755); err != nil {
		t.Fatal(err)
	}
	arabicDeck := deck.Deck{
		Cards: []deck.Flashcard{
			{
				Character:     "مَرْحَبًا",
				Pronunciation: "marḥaban",
				Meaning:       "Hello",
				Explanation:   "Standard greeting",
			},
		},
	}
	if err := deck.SaveDeck(filepath.Join(arabicDir, "basics.yaml"), arabicDeck); err != nil {
		t.Fatal(err)
	}

	m := New(tempDir)
	m.SelectedDir = "Arabic"
	m.SelectedFiles["basics.yaml"] = true
	if err := m.LoadSelectedDecks(); err != nil {
		t.Fatal(err)
	}
	m.SetupReview()
	m.State = StateReview
	m.ShowAnswer = true

	view := m.View()
	if !strings.Contains(view, "Pronunciation: marḥaban") {
		t.Errorf("expected view to contain 'Pronunciation: marḥaban', got:\n%s", view)
	}
	if strings.Contains(view, "Pinyin:") {
		t.Errorf("arabic deck should not display 'Pinyin:', got:\n%s", view)
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
	files, err := deck.GetAllDeckFiles(decksPath, "Mandarin")
	if err != nil {
		t.Fatal(err)
	}
	m.DeckFiles = files
	if len(m.DeckFiles) < 10 {
		t.Skip("Mandarin directory has fewer than 10 files")
	}

	view := m.View()

	// Verify each deck file has its corresponding key badge (up to available Hummingbird keys)
	numKeyed := len(m.DeckFiles)
	if numKeyed > len(HummingbirdKeys) {
		numKeyed = len(HummingbirdKeys)
	}
	for i := 0; i < numKeyed; i++ {
		k := fmt.Sprintf("[%s]", HummingbirdKeys[i])
		if !strings.Contains(view, k) {
			t.Errorf("expected view to contain %s for item %d (%s)", k, i, m.DeckFiles[i])
		}
	}

	if len(m.DeckFiles) > 5 {
		// Verify spacing between group 1 and 2
		gIdx := strings.Index(view, "[g]")
		hIdx := strings.Index(view, "[h]")
		if gIdx != -1 && hIdx != -1 && gIdx < hIdx {
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
		}
	}

	if len(m.DeckFiles) > 10 {
		// Verify spacing between group 2 and 3
		semiIdx := strings.Index(view, "[;]")
		zIdx := strings.Index(view, "[z]")
		if semiIdx != -1 && zIdx != -1 && semiIdx < zIdx {
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
	}
}

func setupNestedTestDecks(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	subA := filepath.Join(tempDir, "Mandarin", "Balatro")
	subB := filepath.Join(tempDir, "Mandarin", "Disney")
	subC := filepath.Join(tempDir, "Spanish")
	if err := os.MkdirAll(subA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subB, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subC, 0755); err != nil {
		t.Fatal(err)
	}

	deckA := deck.Deck{Cards: []deck.Flashcard{{Character: "同花", Meaning: "Flush"}}}
	deckB := deck.Deck{Cards: []deck.Flashcard{{Character: "灰姑娘", Meaning: "Cinderella"}}}
	deckC := deck.Deck{Cards: []deck.Flashcard{{Character: "Hola", Meaning: "Hello"}}}

	if err := deck.SaveDeck(filepath.Join(subA, "balatro1.yaml"), deckA); err != nil {
		t.Fatal(err)
	}
	if err := deck.SaveDeck(filepath.Join(subB, "cinderella.yaml"), deckB); err != nil {
		t.Fatal(err)
	}
	if err := deck.SaveDeck(filepath.Join(subC, "basics.yaml"), deckC); err != nil {
		t.Fatal(err)
	}

	return tempDir
}

func TestMultiDirectorySelectionPersistence(t *testing.T) {
	decksDir := setupNestedTestDecks(t)
	m := New(decksDir)

	// ModeSelect -> Enter -> DirSelect (root)
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	if m.State != StateDirSelect || m.SelectedDir != "" {
		t.Fatalf("expected StateDirSelect at root, got state %d, dir %q", m.State, m.SelectedDir)
	}

	// Root dirs: Mandarin=0, Spanish=1
	// Enter Mandarin
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	if m.State != StateDirSelect || m.SelectedDir != "Mandarin" {
		t.Fatalf("expected StateDirSelect at Mandarin, got state %d, dir %q", m.State, m.SelectedDir)
	}
	if len(m.Dirs) != 2 || m.Dirs[0] != "Balatro" || m.Dirs[1] != "Disney" {
		t.Fatalf("expected [Balatro, Disney], got %v", m.Dirs)
	}

	// Enter Balatro -> DeckSelect
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	if m.State != StateDeckSelect || m.SelectedDir != filepath.Join("Mandarin", "Balatro") {
		t.Fatalf("expected StateDeckSelect in Mandarin/Balatro, got dir %q", m.SelectedDir)
	}

	// Toggle balatro1.yaml using Space
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updatedM.(Model)
	if !m.isDeckSelected("balatro1.yaml") {
		t.Fatalf("expected balatro1.yaml to be selected")
	}

	// Press Esc to go up a level to Mandarin
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)
	if m.State != StateDirSelect || m.SelectedDir != "Mandarin" {
		t.Fatalf("expected StateDirSelect in Mandarin, got dir %q", m.SelectedDir)
	}
	if m.countSelectedDecks() != 1 {
		t.Fatalf("expected 1 deck selected, got %d", m.countSelectedDecks())
	}

	// Move cursor down to Disney (index 1) and press Enter
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedM.(Model)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	if m.State != StateDeckSelect || m.SelectedDir != filepath.Join("Mandarin", "Disney") {
		t.Fatalf("expected StateDeckSelect in Mandarin/Disney, got dir %q", m.SelectedDir)
	}

	// Verify balatro1.yaml is NOT in current view, but cinderella.yaml is
	if len(m.DeckFiles) != 1 || m.DeckFiles[0] != "cinderella.yaml" {
		t.Fatalf("expected [cinderella.yaml], got %v", m.DeckFiles)
	}
	if m.isDeckSelected("cinderella.yaml") {
		t.Fatalf("cinderella.yaml should not be selected yet")
	}

	// Toggle cinderella.yaml
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updatedM.(Model)
	if !m.isDeckSelected("cinderella.yaml") {
		t.Fatalf("expected cinderella.yaml to be selected")
	}
	if m.countSelectedDecks() != 2 {
		t.Fatalf("expected 2 decks selected across folders, got %d", m.countSelectedDecks())
	}

	// Go back up to Mandarin, then up to root
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)
	if m.SelectedDir != "Mandarin" {
		t.Fatalf("expected Mandarin, got %q", m.SelectedDir)
	}

	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)
	if m.SelectedDir != "" {
		t.Fatalf("expected root dir '', got %q", m.SelectedDir)
	}

	// Enter Spanish -> DeckSelect
	// Move down to Spanish (cursor was on Mandarin=0, Spanish=1)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedM.(Model)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	if m.State != StateDeckSelect || m.SelectedDir != "Spanish" {
		t.Fatalf("expected StateDeckSelect in Spanish, got dir %q", m.SelectedDir)
	}

	// Toggle Spanish basics.yaml
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updatedM.(Model)
	if m.countSelectedDecks() != 3 {
		t.Fatalf("expected 3 decks selected across folders, got %d", m.countSelectedDecks())
	}

	// Press Enter to start practice
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	if m.State != StateReview {
		t.Fatalf("expected StateReview, got %d", m.State)
	}

	// Verify all 3 decks loaded into m.Decks and ActiveCards has cards from all 3
	if len(m.Decks) != 3 {
		t.Fatalf("expected 3 decks loaded, got %d: %v", len(m.Decks), m.Decks)
	}
	if len(m.ActiveCards) != 3 {
		t.Fatalf("expected 3 active cards, got %d", len(m.ActiveCards))
	}
}

func TestHierarchicalDirectoryNavigationEsc(t *testing.T) {
	decksDir := setupNestedTestDecks(t)
	m := New(decksDir)

	// ModeSelect -> Enter -> DirSelect (root)
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)

	// Enter Mandarin -> DirSelect (Mandarin)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	if m.SelectedDir != "Mandarin" {
		t.Fatalf("expected Mandarin, got %q", m.SelectedDir)
	}

	// Enter Balatro -> DeckSelect (Mandarin/Balatro)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	if m.State != StateDeckSelect || m.SelectedDir != filepath.Join("Mandarin", "Balatro") {
		t.Fatalf("expected DeckSelect in Mandarin/Balatro, got %d / %q", m.State, m.SelectedDir)
	}

	// Esc -> back to Mandarin (DirSelect)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)
	if m.State != StateDirSelect || m.SelectedDir != "Mandarin" {
		t.Fatalf("expected DirSelect in Mandarin, got %d / %q", m.State, m.SelectedDir)
	}
	if m.DirCursor != 0 || m.Dirs[m.DirCursor] != "Balatro" {
		t.Fatalf("expected cursor restored to Balatro (0), got %d (%s)", m.DirCursor, m.Dirs[m.DirCursor])
	}

	// Esc -> back to root (DirSelect)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)
	if m.State != StateDirSelect || m.SelectedDir != "" {
		t.Fatalf("expected DirSelect at root, got %d / %q", m.State, m.SelectedDir)
	}
	if m.DirCursor != 0 || m.Dirs[m.DirCursor] != "Mandarin" {
		t.Fatalf("expected cursor restored to Mandarin (0), got %d (%s)", m.DirCursor, m.Dirs[m.DirCursor])
	}

	// Esc -> back to ModeSelect
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)
	if m.State != StateModeSelect {
		t.Fatalf("expected StateModeSelect, got %d", m.State)
	}
}

func TestStartPracticeFromDirSelectWithTab(t *testing.T) {
	decksDir := setupNestedTestDecks(t)
	m := New(decksDir)

	// ModeSelect -> Enter -> DirSelect (root)
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)

	// Enter Mandarin -> Enter Balatro -> Space (select balatro1.yaml)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updatedM.(Model)

	// Esc back to Mandarin (StateDirSelect)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)
	if m.State != StateDirSelect {
		t.Fatalf("expected StateDirSelect, got %d", m.State)
	}

	// Press Tab to start practice directly from DirSelect
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedM.(Model)
	if m.State != StateReview {
		t.Fatalf("expected StateReview, got %d", m.State)
	}
	if len(m.ActiveCards) != 1 {
		t.Fatalf("expected 1 active card, got %d", len(m.ActiveCards))
	}
}

func TestArabicModeEntry(t *testing.T) {
	decksDir := setupTestDecks(t)
	m := New(decksDir)

	if m.State != StateModeSelect {
		t.Fatalf("expected StateModeSelect, got %d", m.State)
	}

	// Pressing 'd' selects ModeArabic and transitions to StateArabicMenu
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updatedM.(Model)
	if m.State != StateArabicMenu {
		t.Fatalf("expected StateArabicMenu after pressing 'd', got %d", m.State)
	}

	view := m.View()
	if !strings.Contains(view, "ARABIC ALPHABET ACADEMY") {
		t.Errorf("expected menu view to mention ARABIC ALPHABET ACADEMY, got: %s", view)
	}

	// Pressing 'esc' returns to StateModeSelect
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)
	if m.State != StateModeSelect {
		t.Fatalf("expected StateModeSelect after pressing Esc, got %d", m.State)
	}
}

func TestArabicExplorerFlow(t *testing.T) {
	decksDir := setupTestDecks(t)
	m := New(decksDir)
	m.State = StateArabicMenu

	// Select Alphabet Explorer (activity 0)
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updatedM.(Model)
	if m.State != StateArabicExplorer {
		t.Fatalf("expected StateArabicExplorer, got %d", m.State)
	}
	if m.ArabicLetterCursor != 0 {
		t.Fatalf("expected ArabicLetterCursor 0, got %d", m.ArabicLetterCursor)
	}

	// Right arrow / 'l' advances to letter 1 (Baa)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updatedM.(Model)
	if m.ArabicLetterCursor != 1 {
		t.Fatalf("expected ArabicLetterCursor 1, got %d", m.ArabicLetterCursor)
	}

	// Left arrow / 'h' wraps back to 0 (Alif)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = updatedM.(Model)
	if m.ArabicLetterCursor != 0 {
		t.Fatalf("expected ArabicLetterCursor 0, got %d", m.ArabicLetterCursor)
	}

	// Tab toggles to Tab 1 (Auxiliary & Harakat)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedM.(Model)
	if m.ArabicExplorerTab != 1 {
		t.Fatalf("expected ArabicExplorerTab 1, got %d", m.ArabicExplorerTab)
	}

	viewAux := m.View()
	if !strings.Contains(viewAux, "Auxiliary & Harakat") {
		t.Errorf("expected view to render Auxiliary & Harakat, got: %s", viewAux)
	}

	// Tab toggles back to Tab 0
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedM.(Model)
	if m.ArabicExplorerTab != 0 {
		t.Fatalf("expected ArabicExplorerTab 0, got %d", m.ArabicExplorerTab)
	}

	// Esc returns to StateArabicMenu
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)
	if m.State != StateArabicMenu {
		t.Fatalf("expected StateArabicMenu, got %d", m.State)
	}
}

func TestArabicDrillFlow(t *testing.T) {
	decksDir := setupTestDecks(t)
	m := New(decksDir)
	m.State = StateArabicMenu

	// Select Positional Forms Drill (activity 1, key 's')
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updatedM.(Model)
	if m.State != StateArabicDrill {
		t.Fatalf("expected StateArabicDrill, got %d", m.State)
	}
	if len(m.ArabicQuestions) != 10 {
		t.Fatalf("expected 10 questions, got %d", len(m.ArabicQuestions))
	}

	// Initial drill view
	view := m.View()
	if !strings.Contains(view, "Question 1 of 10") {
		t.Errorf("expected view to show Question 1 of 10, got: %s", view)
	}

	// Answer question with key 'a' (index 0)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updatedM.(Model)
	if !m.ArabicShowFeedback {
		t.Fatalf("expected ArabicShowFeedback true after answering")
	}

	// View shows feedback
	fbView := m.View()
	if !strings.Contains(fbView, "Correct") && !strings.Contains(fbView, "Incorrect") {
		t.Errorf("expected feedback in view, got: %s", fbView)
	}

	// Spacebar advances to Question 2
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updatedM.(Model)
	if m.ArabicShowFeedback {
		t.Fatalf("expected ArabicShowFeedback false after pressing Space")
	}
	if m.ArabicQuestionIdx != 1 {
		t.Fatalf("expected ArabicQuestionIdx 1, got %d", m.ArabicQuestionIdx)
	}

	// Complete all questions
	m.ArabicQuestionIdx = len(m.ArabicQuestions)
	completeView := m.View()
	if !strings.Contains(completeView, "DRILL COMPLETED") {
		t.Errorf("expected DRILL COMPLETED view, got: %s", completeView)
	}

	// Press 'r' to retry
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updatedM.(Model)
	if m.ArabicQuestionIdx != 0 {
		t.Fatalf("expected ArabicQuestionIdx 0 after retry, got %d", m.ArabicQuestionIdx)
	}

	// Esc returns to Arabic Menu
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)
	if m.State != StateArabicMenu {
		t.Fatalf("expected StateArabicMenu, got %d", m.State)
	}
}

func TestArabicStagesFlow(t *testing.T) {
	decksDir := setupTestDecks(t)
	m := New(decksDir)
	m.State = StateArabicMenu

	// Select Guided Lessons (activity 6, key 'j')
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedM.(Model)
	if m.State != StateArabicStages {
		t.Fatalf("expected StateArabicStages, got %d", m.State)
	}

	viewStages := m.View()
	if !strings.Contains(viewStages, "PROGRESSIVE GUIDED LESSONS") {
		t.Errorf("expected view to contain PROGRESSIVE GUIDED LESSONS, got: %s", viewStages)
	}

	// Select Stage 0 (Stage 1: The Anchors & Boat Letters)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	if m.State != StateArabicStageView {
		t.Fatalf("expected StateArabicStageView, got %d", m.State)
	}

	viewStageDetail := m.View()
	if !strings.Contains(viewStageDetail, "The Anchors & Boat Letters") {
		t.Errorf("expected stage details in view, got: %s", viewStageDetail)
	}

	// Press Enter to start Stage Drill
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	if m.State != StateArabicDrill {
		t.Fatalf("expected StateArabicDrill, got %d", m.State)
	}
	if len(m.ArabicQuestions) == 0 {
		t.Fatalf("expected stage questions loaded")
	}

	// Esc returns back to Stages
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)
	if m.State != StateArabicStages {
		t.Fatalf("expected StateArabicStages after esc, got %d", m.State)
	}

	// Esc returns back to Arabic Menu
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)
	if m.State != StateArabicMenu {
		t.Fatalf("expected StateArabicMenu after esc, got %d", m.State)
	}
}

func TestQuizMarkAndUnmarkCard(t *testing.T) {
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

	// Initially, no card is marked
	ref := m.ActiveCards[m.CurrentIndex]
	currentCard := m.Decks[ref.Filename].Cards[ref.OrigIdx]
	if m.isCardMarked(currentCard) {
		t.Fatalf("expected card to not be marked initially")
	}

	// Question screen: press 'm' to mark
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = updatedM.(Model)

	if !m.isCardMarked(currentCard) {
		t.Fatalf("expected card to be marked after pressing 'm'")
	}
	if len(m.MarkedCards) != 1 {
		t.Fatalf("expected 1 marked card, got %d", len(m.MarkedCards))
	}
	view := m.View()
	if !strings.Contains(view, "★ MARKED") {
		t.Errorf("expected view to contain '★ MARKED', got:\n%s", view)
	}
	if !strings.Contains(view, "[m] Unmark card") {
		t.Errorf("expected view to contain '[m] Unmark card', got:\n%s", view)
	}

	// Verify file was saved on disk
	markedFile := filepath.Join(decksDir, "marked.yaml")
	loaded, err := deck.LoadDeck(markedFile)
	if err != nil {
		t.Fatalf("failed to load marked deck from disk: %v", err)
	}
	if len(loaded.Cards) != 1 || loaded.Cards[0].Character != currentCard.Character {
		t.Fatalf("expected disk marked deck to have %s, got %+v", currentCard.Character, loaded.Cards)
	}

	// Answer the question
	correctKey := HummingbirdKeys[m.CorrectIndex]
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(correctKey)})
	m = updatedM.(Model)

	if !m.ShowFeedback {
		t.Fatalf("expected feedback view")
	}
	feedbackView := m.View()
	if !strings.Contains(feedbackView, "★ MARKED") {
		t.Errorf("expected feedback view to show '★ MARKED', got:\n%s", feedbackView)
	}
	if !strings.Contains(feedbackView, "Unmark") {
		t.Errorf("expected feedback view to contain 'Unmark', got:\n%s", feedbackView)
	}

	// Feedback screen: press 'u' to unmark
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updatedM.(Model)

	if m.isCardMarked(currentCard) {
		t.Fatalf("expected card to be unmarked after pressing 'u'")
	}
	if len(m.MarkedCards) != 0 {
		t.Fatalf("expected 0 marked cards, got %d", len(m.MarkedCards))
	}
	feedbackViewAfterUnmark := m.View()
	if strings.Contains(feedbackViewAfterUnmark, "★ MARKED") {
		t.Errorf("expected feedback view to NOT show '★ MARKED' after unmark, got:\n%s", feedbackViewAfterUnmark)
	}

	// Feedback screen: press 'm' to re-mark
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = updatedM.(Model)
	if !m.isCardMarked(currentCard) {
		t.Fatalf("expected card to be re-marked after pressing 'm'")
	}
}

func TestSelectMarkedDeckFromMenus(t *testing.T) {
	decksDir := setupTestDecks(t)
	m := New(decksDir)

	// Add 2 marked cards to marked.yaml
	markedCards := []deck.Flashcard{
		{Character: "电脑", Meaning: "Computer"},
		{Character: "Bonjour", Meaning: "Hello"},
	}
	m.MarkedCards = markedCards
	if err := m.saveMarkedDeck(); err != nil {
		t.Fatal(err)
	}

	// Verify mode select view displays marked cards
	m2 := New(decksDir)
	modeView := m2.View()
	if !strings.Contains(modeView, "★ Practice Marked Cards (2 cards)") {
		t.Errorf("expected mode select to display marked cards, got:\n%s", modeView)
	}

	// Press 'm' from mode select to start practice
	updatedM, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m2 = updatedM.(Model)
	if m2.State != StateQuiz {
		t.Fatalf("expected StateQuiz after pressing 'm' in mode select, got %d", m2.State)
	}
	if len(m2.ActiveCards) != 2 {
		t.Fatalf("expected 2 active cards in marked quiz, got %d", len(m2.ActiveCards))
	}

	// Test selecting from DirSelect view
	m3 := New(decksDir)
	// Enter ModeQuiz
	updatedM, _ = m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m3 = updatedM.(Model)
	if m3.State != StateDirSelect {
		t.Fatalf("expected StateDirSelect, got %d", m3.State)
	}

	dirView := m3.View()
	if !strings.Contains(dirView, "★ Marked Cards (2 cards)") {
		t.Errorf("expected dir select to display marked cards, got:\n%s", dirView)
	}

	// Press 'm' from DirSelect
	updatedM, _ = m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m3 = updatedM.(Model)
	if m3.State != StateQuiz {
		t.Fatalf("expected StateQuiz after pressing 'm' in dir select, got %d", m3.State)
	}
	if len(m3.ActiveCards) != 2 {
		t.Fatalf("expected 2 active cards, got %d", len(m3.ActiveCards))
	}

	// Test navigating down with arrow keys to Marked Cards in DirSelect
	m4 := New(decksDir)
	updatedM, _ = m4.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m4 = updatedM.(Model)
	// In setupTestDecks, dirs are French (0), Mandarin (1).
	// Marked Cards is at index 2 (len(m.Dirs)).
	// Move down twice
	updatedM, _ = m4.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m4 = updatedM.(Model)
	updatedM, _ = m4.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m4 = updatedM.(Model)
	if m4.DirCursor != len(m4.Dirs) {
		t.Fatalf("expected DirCursor at %d, got %d", len(m4.Dirs), m4.DirCursor)
	}
	// Press Enter to start practice
	updatedM, _ = m4.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 = updatedM.(Model)
	if m4.State != StateQuiz {
		t.Fatalf("expected StateQuiz after pressing Enter on Marked Cards, got %d", m4.State)
	}
}

func TestEmptyMarkedCardsMessage(t *testing.T) {
	decksDir := setupTestDecks(t)
	m := New(decksDir)

	// Press 'm' when 0 marked cards
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = updatedM.(Model)

	if m.State != StateModeSelect {
		t.Fatalf("expected StateModeSelect, got %d", m.State)
	}
	if !strings.Contains(m.StatusMessage, "No marked cards yet") {
		t.Errorf("expected StatusMessage to explain no marked cards, got: %s", m.StatusMessage)
	}
	view := m.View()
	if !strings.Contains(view, "No marked cards yet") {
		t.Errorf("expected view to contain status message, got:\n%s", view)
	}
}

func TestPlayingMarkedDeckLoopingAndUnmarking(t *testing.T) {
	decksDir := setupTestDecks(t)
	m := New(decksDir)

	m.MarkedCards = []deck.Flashcard{
		{Character: "电脑", Meaning: "Computer"},
		{Character: "手机", Meaning: "Mobile phone"},
	}
	if err := m.saveMarkedDeck(); err != nil {
		t.Fatal(err)
	}

	// Start marked practice
	m.startMarkedPractice()
	if m.State != StateQuiz {
		t.Fatalf("expected StateQuiz, got %d", m.State)
	}
	if len(m.ActiveCards) != 2 {
		t.Fatalf("expected 2 active cards, got %d", len(m.ActiveCards))
	}

	// Question 1: Answer and unmark
	ref1 := m.ActiveCards[0]
	card1 := m.Decks[ref1.Filename].Cards[ref1.OrigIdx]

	correctKey := HummingbirdKeys[m.CorrectIndex]
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(correctKey)})
	m = updatedM.(Model)

	// Unmark card 1
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updatedM.(Model)
	if m.isCardMarked(card1) {
		t.Fatalf("expected card 1 to be unmarked")
	}
	if len(m.MarkedCards) != 1 {
		t.Fatalf("expected 1 marked card remaining, got %d", len(m.MarkedCards))
	}

	// Advance to Question 2
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updatedM.(Model)
	if m.CurrentIndex != 1 {
		t.Fatalf("expected CurrentIndex 1, got %d", m.CurrentIndex)
	}

	// Question 2: Answer, but leave it marked
	correctKey2 := HummingbirdKeys[m.CorrectIndex]
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(correctKey2)})
	m = updatedM.(Model)

	// Advance past Question 2 -> Quiz complete
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updatedM.(Model)
	if m.CurrentIndex < len(m.ActiveCards) {
		t.Fatalf("expected quiz complete")
	}

	completeView := m.View()
	if !strings.Contains(completeView, "Quiz complete!") {
		t.Errorf("expected 'Quiz complete!', got:\n%s", completeView)
	}
	if !strings.Contains(completeView, "[r] Retry quiz") {
		t.Errorf("expected '[r] Retry quiz', got:\n%s", completeView)
	}

	// Press 'r' to loop through remaining marked cards
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updatedM.(Model)

	// Now only 1 card should be in the active quiz!
	if len(m.ActiveCards) != 1 {
		t.Fatalf("expected 1 card remaining in looped quiz, got %d", len(m.ActiveCards))
	}
	if m.CurrentIndex != 0 {
		t.Fatalf("expected CurrentIndex 0 in restarted quiz, got %d", m.CurrentIndex)
	}

	// Answer the remaining card and unmark it
	correctKey3 := HummingbirdKeys[m.CorrectIndex]
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(correctKey3)})
	m = updatedM.(Model)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updatedM.(Model)

	if len(m.MarkedCards) != 0 {
		t.Fatalf("expected 0 marked cards remaining, got %d", len(m.MarkedCards))
	}

	// Advance
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updatedM.(Model)

	// Press 'r' when 0 marked cards left
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updatedM.(Model)
	if m.State != StateModeSelect {
		t.Fatalf("expected return to StateModeSelect when no marked cards left, got %d", m.State)
	}
}

func TestMarkedCardsInReviewMode(t *testing.T) {
	decksDir := setupTestDecks(t)
	m := New(decksDir)
	m.SelectedDir = "Mandarin"
	m.SelectedFiles["tech.yaml"] = true
	if err := m.LoadSelectedDecks(); err != nil {
		t.Fatal(err)
	}
	m.SetupReview()
	m.State = StateReview

	ref := m.ActiveCards[0]
	card := m.Decks[ref.Filename].Cards[ref.OrigIdx]

	// Toggle mark in review mode with 'm'
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = updatedM.(Model)
	if !m.isCardMarked(card) {
		t.Fatalf("expected card to be marked in review mode")
	}
	view := m.View()
	if !strings.Contains(view, "★ MARKED") {
		t.Errorf("expected review view to show '★ MARKED', got:\n%s", view)
	}

	// Toggle unmark with 'u'
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updatedM.(Model)
	if m.isCardMarked(card) {
		t.Fatalf("expected card to be unmarked in review mode")
	}
}

func TestMarkedCardsDistractorSupplement(t *testing.T) {
	decksDir := setupTestDecks(t)
	m := New(decksDir)

	// Mark only 1 card
	m.MarkedCards = []deck.Flashcard{
		{Character: "电脑", Pinyin: "diànnǎo", Meaning: "Computer"},
	}
	if err := m.saveMarkedDeck(); err != nil {
		t.Fatal(err)
	}

	// Start marked practice
	m.startMarkedPractice()
	if m.State != StateQuiz {
		t.Fatalf("expected StateQuiz, got %d", m.State)
	}
	// Because other decks exist in setupTestDecks (tech.yaml with 手机, basics.yaml with Bonjour),
	// distractors should be supplemented so QuizOptions has more than 1 option.
	if len(m.QuizOptions) <= 1 {
		t.Fatalf("expected QuizOptions to have supplemented distractors, got %d options: %v", len(m.QuizOptions), m.QuizOptions)
	}
}

func TestMarkedCardsPersistenceAcrossRestarts(t *testing.T) {
	decksDir := setupTestDecks(t)

	// Session 1: mark a card and quit
	m1 := New(decksDir)
	m1.markCard(deck.Flashcard{Character: "测试", Pinyin: "cèshì", Meaning: "Test"})

	// Session 2: start new instance with same decksDir
	m2 := New(decksDir)
	if len(m2.MarkedCards) != 1 {
		t.Fatalf("expected 1 marked card in new session, got %d", len(m2.MarkedCards))
	}
	if m2.MarkedCards[0].Character != "测试" {
		t.Errorf("expected character '测试', got %s", m2.MarkedCards[0].Character)
	}

	// Unmark and verify disk file is cleaned up
	m2.unmarkCard(m2.MarkedCards[0])
	if len(m2.MarkedCards) != 0 {
		t.Fatalf("expected 0 marked cards, got %d", len(m2.MarkedCards))
	}

	// Session 3: verify 0 marked cards loaded
	m3 := New(decksDir)
	if len(m3.MarkedCards) != 0 {
		t.Fatalf("expected 0 marked cards in session 3, got %d", len(m3.MarkedCards))
	}
}

func TestMarkedCardsDistractorsSameLanguage(t *testing.T) {
	tempDir := t.TempDir()
	mandarinDir := filepath.Join(tempDir, "Mandarin")
	arabicDir := filepath.Join(tempDir, "Arabic")
	if err := os.Mkdir(mandarinDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(arabicDir, 0755); err != nil {
		t.Fatal(err)
	}

	mandarinDeck := deck.Deck{
		Cards: []deck.Flashcard{
			{Character: "电脑", Pinyin: "diànnǎo", Meaning: "Computer"},
			{Character: "手机", Pinyin: "shǒujī", Meaning: "Mobile phone"},
			{Character: "书", Pinyin: "shū", Meaning: "Book"},
			{Character: "水", Pinyin: "shuǐ", Meaning: "Water"},
			{Character: "茶", Pinyin: "chá", Meaning: "Tea"},
			{Character: "猫", Pinyin: "māo", Meaning: "Cat"},
			{Character: "狗", Pinyin: "gǒu", Meaning: "Dog"},
		},
	}
	if err := deck.SaveDeck(filepath.Join(mandarinDir, "vocab.yaml"), mandarinDeck); err != nil {
		t.Fatal(err)
	}

	arabicDeck := deck.Deck{
		Cards: []deck.Flashcard{
			{Character: "مَرْحَبًا", Pronunciation: "marḥaban", Meaning: "Hello"},
			{Character: "شُكْرًا", Pronunciation: "shukran", Meaning: "Thank you"},
			{Character: "كِتَاب", Pronunciation: "kitāb", Meaning: "Book"},
		},
	}
	if err := deck.SaveDeck(filepath.Join(arabicDir, "basics.yaml"), arabicDeck); err != nil {
		t.Fatal(err)
	}

	// Case 1: Mark only Mandarin cards (2 cards)
	m := New(tempDir)
	m.MarkedCards = []deck.Flashcard{
		{Character: "电脑", Pinyin: "diànnǎo", Meaning: "Computer", Language: "Mandarin", Deck: "Mandarin/vocab.yaml"},
		{Character: "手机", Pinyin: "shǒujī", Meaning: "Mobile phone", Language: "Mandarin", Deck: "Mandarin/vocab.yaml"},
	}
	if err := m.saveMarkedDeck(); err != nil {
		t.Fatal(err)
	}

	m.startMarkedPractice()
	if m.State != StateQuiz {
		t.Fatalf("expected StateQuiz, got %d", m.State)
	}

	// Verify that ALL options for the Mandarin question are Mandarin (contain Pinyin or Mandarin words, NO Arabic)
	if len(m.QuizOptions) < 2 {
		t.Fatalf("expected at least 2 options, got %d", len(m.QuizOptions))
	}
	for _, opt := range m.QuizOptions {
		if strings.Contains(opt, "marḥaban") || strings.Contains(opt, "shukran") || strings.Contains(opt, "kitāb") {
			t.Errorf("found Arabic distractor in Mandarin quiz option: %s", opt)
		}
	}

	// Case 2: Mixed marked cards (2 Mandarin + 2 Arabic)
	mMixed := New(tempDir)
	mMixed.MarkedCards = []deck.Flashcard{
		{Character: "电脑", Pinyin: "diànnǎo", Meaning: "Computer", Language: "Mandarin"},
		{Character: "手机", Pinyin: "shǒujī", Meaning: "Mobile phone", Language: "Mandarin"},
		{Character: "مَرْحَبًا", Pronunciation: "marḥaban", Meaning: "Hello", Language: "Arabic"},
		{Character: "شُكْرًا", Pronunciation: "shukran", Meaning: "Thank you", Language: "Arabic"},
	}
	if err := mMixed.saveMarkedDeck(); err != nil {
		t.Fatal(err)
	}
	mMixed.startMarkedPractice()

	// Check each question in the mixed quiz
	for i := 0; i < len(mMixed.ActiveCards); i++ {
		mMixed.CurrentIndex = i
		mMixed.GenerateQuizOptions()
		ref := mMixed.ActiveCards[i]
		target := mMixed.Decks[ref.Filename].Cards[ref.OrigIdx]
		targetLang := mMixed.DetectCardLanguage(target)

		if targetLang == "Mandarin" {
			for _, opt := range mMixed.QuizOptions {
				if strings.Contains(opt, "marḥaban") || strings.Contains(opt, "shukran") || strings.Contains(opt, "kitāb") {
					t.Errorf("Mandarin question %s received Arabic distractor: %s", target.Character, opt)
				}
			}
		} else if targetLang == "Arabic" {
			for _, opt := range mMixed.QuizOptions {
				if strings.Contains(opt, "diànnǎo") || strings.Contains(opt, "shǒujī") || strings.Contains(opt, "Computer") {
					t.Errorf("Arabic question %s received Mandarin distractor: %s", target.Character, opt)
				}
			}
		}
	}
}

func TestWorkspaceMarkedDeck(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", "..", "decks", "marked.yaml")); os.IsNotExist(err) {
		t.Skip("decks/marked.yaml not present")
	}

	m := New(filepath.Join("..", "..", "decks"))
	if len(m.MarkedCards) == 0 {
		t.Skip("no marked cards in workspace decks/marked.yaml")
	}

	m.startMarkedPractice()
	if m.State != StateQuiz {
		t.Fatalf("expected StateQuiz, got %d", m.State)
	}

	for i := 0; i < len(m.ActiveCards); i++ {
		m.CurrentIndex = i
		m.GenerateQuizOptions()
		ref := m.ActiveCards[i]
		target := m.Decks[ref.Filename].Cards[ref.OrigIdx]
		targetLang := m.DetectCardLanguage(target)

		if targetLang == "Mandarin" {
			for _, opt := range m.QuizOptions {
				// Arabic transliterations / words must never appear
				if strings.Contains(opt, "marḥaban") || strings.Contains(opt, "shukran") || strings.Contains(opt, "Hello") && !strings.Contains(opt, "nǐ") {
					t.Errorf("Mandarin question '%s' received non-Mandarin distractor: %s", target.Character, opt)
				}
			}
		}
	}
}

func TestMemorizeModeNavigation(t *testing.T) {
	decksDir := setupTestDecks(t)
	textsDir := filepath.Join("..", "..", "memorizationTexts")

	m := NewWithOptions(decksDir, textsDir)
	if m.State != StateModeSelect {
		t.Fatalf("expected StateModeSelect, got %d", m.State)
	}

	// Press 'f' to enter Memorize by Options mode
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = updatedM.(Model)

	if m.State != StateMemorizeSelectText {
		t.Fatalf("expected StateMemorizeSelectText, got %d", m.State)
	}
	if len(m.MemorizeTexts) == 0 {
		t.Fatalf("expected at least 1 memorization text, got %d", len(m.MemorizeTexts))
	}

	// First text should be Filemón
	if !strings.Contains(strings.ToLower(m.MemorizeTexts[0].Title), "filemón") {
		t.Errorf("expected Filemón text, got '%s'", m.MemorizeTexts[0].Title)
	}

	// Enter on the selected text -> StateMemorizePortionSelect
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)

	if m.State != StateMemorizePortionSelect {
		t.Fatalf("expected StateMemorizePortionSelect, got %d", m.State)
	}
	if len(m.MemorizeSections) != 25 {
		t.Fatalf("expected 25 verse sections in Filemón, got %d", len(m.MemorizeSections))
	}

	// Enter on portion select -> StateMemorize
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)

	if m.State != StateMemorize {
		t.Fatalf("expected StateMemorize, got %d", m.State)
	}
	if m.MemorizeActiveText == nil {
		t.Fatalf("expected MemorizeActiveText to be loaded")
	}
	if len(m.MemorizeActiveText.Words) != 423 {
		t.Errorf("expected 423 words in Filemón, got %d", len(m.MemorizeActiveText.Words))
	}
	if len(m.MemorizeCurrentOptions) != 5 {
		t.Errorf("expected 5 options, got %d", len(m.MemorizeCurrentOptions))
	}

	// Verify correct option
	targetWord := m.MemorizeActiveText.Words[0].Word
	if m.MemorizeCurrentOptions[m.MemorizeCorrectIndex] != targetWord {
		t.Errorf("expected correct option to be '%s', got '%s'", targetWord, m.MemorizeCurrentOptions[m.MemorizeCorrectIndex])
	}
}

func TestMemorizeWordProgressionAndFeedback(t *testing.T) {
	decksDir := setupTestDecks(t)
	textsDir := filepath.Join("..", "..", "memorizationTexts")

	m := NewWithOptions(decksDir, textsDir)
	// Press 'f', then Enter (select text), then Enter (start portion)
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = updatedM.(Model)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)

	if m.State != StateMemorize {
		t.Fatalf("expected StateMemorize, got %d", m.State)
	}

	// 1. Submit the CORRECT answer
	correctKey := IndexToKey(m.MemorizeCorrectIndex)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(correctKey)})
	m = updatedM.(Model)

	if m.MemorizeCurrentIndex != 1 {
		t.Errorf("expected index 1 after correct answer, got %d", m.MemorizeCurrentIndex)
	}
	if m.MemorizeStreak != 1 {
		t.Errorf("expected streak 1, got %d", m.MemorizeStreak)
	}
	if m.MemorizeMistakes != 0 {
		t.Errorf("expected 0 mistakes, got %d", m.MemorizeMistakes)
	}
	if m.MemorizeShowFeedback {
		t.Errorf("expected immediate advance without feedback pause on correct answer")
	}

	// 2. Submit an INCORRECT answer
	wrongIdx := (m.MemorizeCorrectIndex + 1) % len(m.MemorizeCurrentOptions)
	wrongKey := IndexToKey(wrongIdx)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(wrongKey)})
	m = updatedM.(Model)

	if !m.MemorizeShowFeedback {
		t.Errorf("expected MemorizeShowFeedback to be true after incorrect answer")
	}
	if m.MemorizeCurrentIndex != 1 {
		t.Errorf("expected index to stay at 1 during feedback, got %d", m.MemorizeCurrentIndex)
	}
	if m.MemorizeStreak != 0 {
		t.Errorf("expected streak reset to 0, got %d", m.MemorizeStreak)
	}
	if m.MemorizeMistakes != 1 {
		t.Errorf("expected mistakes to be 1, got %d", m.MemorizeMistakes)
	}

	// 3. Press Spacebar to acknowledge error and continue
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updatedM.(Model)

	if m.MemorizeShowFeedback {
		t.Errorf("expected feedback dismissed after spacebar")
	}
	if m.MemorizeCurrentIndex != 2 {
		t.Errorf("expected index to advance to 2, got %d", m.MemorizeCurrentIndex)
	}

	// 4. Test View rendering in Memorize mode
	rendered := m.View()
	if !strings.Contains(rendered, "Portion Word 3 of 423") {
		t.Errorf("expected view to contain 'Portion Word 3 of 423', got: %s", rendered)
	}
	if !strings.Contains(rendered, "[a]") {
		t.Errorf("expected view to contain option keys [a], got: %s", rendered)
	}
}

func TestMemorizePortionSelectionAndShifting(t *testing.T) {
	decksDir := setupTestDecks(t)
	textsDir := filepath.Join("..", "..", "memorizationTexts")

	m := NewWithOptions(decksDir, textsDir)
	// 'f' -> Enter into portion select
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = updatedM.(Model)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)

	if m.State != StateMemorizePortionSelect {
		t.Fatalf("expected StateMemorizePortionSelect, got %d", m.State)
	}

	// Default: all verses (0 to 24)
	if m.PortionStartSection != 0 || m.PortionEndSection != 24 {
		t.Errorf("expected default portion 0..24, got %d..%d", m.PortionStartSection, m.PortionEndSection)
	}

	// Shift Start to Verse 15 (index 14) and End to Verse 20 (index 19)
	m.PortionStartSection = 14
	m.PortionEndSection = 19

	// Test shifting start independently:
	// '[' decreases start
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	m = updatedM.(Model)
	if m.PortionStartSection != 13 {
		t.Errorf("expected PortionStartSection 13 after '[', got %d", m.PortionStartSection)
	}

	// ']' increases start
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	m = updatedM.(Model)
	if m.PortionStartSection != 14 {
		t.Errorf("expected PortionStartSection 14 after ']', got %d", m.PortionStartSection)
	}

	// Test shifting end independently:
	// '}' increases end
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'}'}})
	m = updatedM.(Model)
	if m.PortionEndSection != 20 {
		t.Errorf("expected PortionEndSection 20 after '}', got %d", m.PortionEndSection)
	}

	// '{' decreases end
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'{'}})
	m = updatedM.(Model)
	if m.PortionEndSection != 19 {
		t.Errorf("expected PortionEndSection 19 after '{', got %d", m.PortionEndSection)
	}

	// Test Expand ('x')
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = updatedM.(Model)
	if m.PortionEndSection != 20 {
		t.Errorf("expected PortionEndSection 20 after 'x', got %d", m.PortionEndSection)
	}

	// Test Contract ('c')
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updatedM.(Model)
	if m.PortionEndSection != 19 {
		t.Errorf("expected PortionEndSection 19 after 'c', got %d", m.PortionEndSection)
	}

	// Set exactly to verses 15–20 (indices 14 to 19)
	m.PortionStartSection = 14
	m.PortionEndSection = 19

	// Verify view rendering of portion selector
	viewStr := m.View()
	if !strings.Contains(viewStr, "Verse 15 to Verse 20") {
		t.Errorf("expected view to contain 'Verse 15 to Verse 20', got: %s", viewStr)
	}
	if !strings.Contains(viewStr, "[START]") || !strings.Contains(viewStr, "[END]") {
		t.Errorf("expected view to contain [START] and [END] markers, got: %s", viewStr)
	}

	// Start practicing verses 15–20
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)

	if m.State != StateMemorize {
		t.Fatalf("expected StateMemorize, got %d", m.State)
	}
	sec15 := m.MemorizeSections[14]
	sec20 := m.MemorizeSections[19]
	if m.MemorizeCurrentIndex != sec15.StartIdx {
		t.Errorf("expected start word index %d, got %d", sec15.StartIdx, m.MemorizeCurrentIndex)
	}
	if m.PortionEndWordIdx != sec20.EndIdx {
		t.Errorf("expected end word index %d, got %d", sec20.EndIdx, m.PortionEndWordIdx)
	}

	// Complete the portion: simulate reaching last word of portion
	m.MemorizeCurrentIndex = m.PortionEndWordIdx - 1
	m.SetupMemorizeWord()

	correctKey := IndexToKey(m.MemorizeCorrectIndex)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(correctKey)})
	m = updatedM.(Model)

	// User MUST be sent back to portion selector!
	if m.State != StateMemorizePortionSelect {
		t.Fatalf("expected return to StateMemorizePortionSelect, got %d", m.State)
	}
	if !strings.Contains(m.PortionCompletedMsg, "Portion Complete") {
		t.Errorf("expected PortionCompletedMsg to report completion, got: '%s'", m.PortionCompletedMsg)
	}

	// Now test 'n' (advance): moves from 14..19 to next chunk
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updatedM.(Model)
	if m.PortionStartSection != 20 {
		t.Errorf("expected advanced start section 20, got %d", m.PortionStartSection)
	}
}

func TestMemorizeBackEscNavigation(t *testing.T) {
	decksDir := setupTestDecks(t)
	textsDir := filepath.Join("..", "..", "memorizationTexts")

	m := NewWithOptions(decksDir, textsDir)
	// ModeSelect -> 'f' -> SelectText -> Enter -> PortionSelect -> Enter -> Memorize
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = updatedM.(Model)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)

	if m.State != StateMemorize {
		t.Fatalf("expected StateMemorize, got %d", m.State)
	}

	// Esc from StateMemorize -> back to StateMemorizePortionSelect
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)

	if m.State != StateMemorizePortionSelect {
		t.Fatalf("expected StateMemorizePortionSelect after Esc, got %d", m.State)
	}

	// Esc from StateMemorizePortionSelect -> back to StateMemorizeSelectText
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)

	if m.State != StateMemorizeSelectText {
		t.Fatalf("expected StateMemorizeSelectText after Esc, got %d", m.State)
	}

	// Esc from StateMemorizeSelectText -> back to StateModeSelect
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)

	if m.State != StateModeSelect {
		t.Fatalf("expected StateModeSelect after Esc, got %d", m.State)
	}
}

func TestReadingModeSelectAndNavigate(t *testing.T) {
	decksDir := setupTestDecks(t)
	textsDir := filepath.Join("..", "..", "memorizationTexts")
	readingDir := filepath.Join("..", "..", "readingTranslationTexts")

	m := NewWithOptions(decksDir, textsDir, readingDir)
	if m.State != StateModeSelect {
		t.Fatalf("expected StateModeSelect, got %d", m.State)
	}

	// Press 'g' to enter Reading mode
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = updatedM.(Model)

	if m.State != StateReadingSelectText {
		t.Fatalf("expected StateReadingSelectText, got %d", m.State)
	}
	if len(m.ReadingTexts) == 0 {
		t.Fatalf("expected at least 1 reading text loaded, got 0")
	}

	// View output should list texts
	view := m.View()
	if !strings.Contains(view, "Learn by Reading") {
		t.Errorf("expected view to contain 'Learn by Reading', got: %s", view)
	}

	// Navigate down with 'j'
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedM.(Model)
	if len(m.ReadingTexts) > 1 && m.ReadingCursor != 1 {
		t.Errorf("expected cursor at 1, got %d", m.ReadingCursor)
	}

	// Press Enter to start reading
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)

	if m.State != StateReading {
		t.Fatalf("expected StateReading, got %d", m.State)
	}
	if m.ReadingSession == nil {
		t.Fatalf("expected non-nil ReadingSession")
	}

	// Reading view should show sentence and title
	readingView := m.View()
	if !strings.Contains(readingView, "Sentence 1 of") {
		t.Errorf("expected view to contain sentence progress, got: %s", readingView)
	}
	if !strings.Contains(readingView, "Aggressiveness:") {
		t.Errorf("expected view to contain Aggressiveness indicator, got: %s", readingView)
	}
}

func TestReadingFlowAndWordDetail(t *testing.T) {
	decksDir := setupTestDecks(t)
	textsDir := filepath.Join("..", "..", "memorizationTexts")
	readingDir := filepath.Join("..", "..", "readingTranslationTexts")

	m := NewWithOptions(decksDir, textsDir, readingDir)
	m.ReadingProgress = reading.DefaultProgress()

	// Enter reading mode
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = updatedM.(Model)
	// Select first text
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	m.ReadingSession.SentenceIdx = 0
	m.ReadingSession.Restart()

	if m.State != StateReading {
		t.Fatalf("expected StateReading, got %d", m.State)
	}

	// Ensure there are substituted words
	session := m.ReadingSession
	if len(session.CurrentWeave.SubstitutedWords) == 0 {
		session.SetAggressiveness(4)
	}
	if len(session.CurrentWeave.SubstitutedWords) == 0 {
		t.Fatalf("expected substituted words in sentence at aggressiveness 4")
	}

	firstKey := session.CurrentWeave.SubstitutedWords[0].Key
	runes := []rune(firstKey)

	// Press the shortcut key for the first substituted word
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: runes})
	m = updatedM.(Model)

	if m.State != StateReadingWordDetail {
		t.Fatalf("expected StateReadingWordDetail after pressing word key, got %d", m.State)
	}
	if m.ReadingSession.SelectedWord == nil {
		t.Fatalf("expected SelectedWord to be set")
	}

	detailView := m.View()
	if !strings.Contains(detailView, "VOCABULARY DETAIL") {
		t.Errorf("expected detail view to contain header, got: %s", detailView)
	}

	// Press 'm' to add word to Marked Cards deck
	initialMarkedCount := len(m.MarkedCards)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = updatedM.(Model)

	if len(m.MarkedCards) != initialMarkedCount+1 {
		t.Errorf("expected marked cards count to increase to %d, got %d", initialMarkedCount+1, len(m.MarkedCards))
	}
	if !strings.Contains(m.ReadingWordFeedback, "Marked Cards") {
		t.Errorf("expected feedback message about marked cards, got %s", m.ReadingWordFeedback)
	}

	// Press 'k' to mark as known
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updatedM.(Model)
	if !strings.Contains(m.ReadingWordFeedback, "Marked as Known") {
		t.Errorf("expected feedback message about known, got %s", m.ReadingWordFeedback)
	}

	// Press Esc to return to reading sentence
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedM.(Model)

	if m.State != StateReading {
		t.Fatalf("expected StateReading after Esc, got %d", m.State)
	}
}

func TestReadingAggressivenessAndPureReadingMode(t *testing.T) {
	decksDir := setupTestDecks(t)
	textsDir := filepath.Join("..", "..", "memorizationTexts")
	readingDir := filepath.Join("..", "..", "readingTranslationTexts")

	m := NewWithOptions(decksDir, textsDir, readingDir)
	m.ReadingProgress = reading.DefaultProgress()
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = updatedM.(Model)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)
	m.ReadingSession.SentenceIdx = 0
	m.ReadingSession.Restart()

	// Press '0' for Pure Reading mode (English only)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})
	m = updatedM.(Model)

	if m.ReadingSession.Aggressiveness != 0 {
		t.Errorf("expected Aggressiveness 0, got %d", m.ReadingSession.Aggressiveness)
	}
	if len(m.ReadingSession.CurrentWeave.SubstitutedWords) != 0 {
		t.Errorf("expected 0 substituted words in pure reading mode, got %d", len(m.ReadingSession.CurrentWeave.SubstitutedWords))
	}

	view := m.View()
	if !strings.Contains(view, "Pure Reading") {
		t.Errorf("expected view to reflect Pure Reading mode, got: %s", view)
	}

	// Increase aggressiveness with '+'
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	m = updatedM.(Model)
	if m.ReadingSession.Aggressiveness != 1 {
		t.Errorf("expected Aggressiveness 1 after +, got %d", m.ReadingSession.Aggressiveness)
	}
}

func TestReadingDiscussionBoxAndAdvance(t *testing.T) {
	decksDir := setupTestDecks(t)
	textsDir := filepath.Join("..", "..", "memorizationTexts")
	readingDir := filepath.Join("..", "..", "readingTranslationTexts")

	m := NewWithOptions(decksDir, textsDir, readingDir)
	m.ReadingProgress = reading.DefaultProgress()

	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = updatedM.(Model)
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)

	// Explicitly start from sentence 0 for deterministic test
	m.ReadingSession.SentenceIdx = 0
	m.ReadingSession.Restart()

	// Press Enter to complete sentence -> shows discussion box
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)

	if m.State != StateReadingDiscussion {
		t.Fatalf("expected StateReadingDiscussion, got %d", m.State)
	}

	discussView := m.View()
	if !strings.Contains(discussView, "GRAMMAR") {
		t.Errorf("expected grammar discussion header, got: %s", discussView)
	}

	// Press Enter in discussion box to advance to next sentence
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)

	if m.State != StateReading {
		t.Fatalf("expected StateReading after advancing from discussion, got %d", m.State)
	}
	if m.ReadingSession.SentenceIdx != 1 {
		t.Errorf("expected SentenceIdx 1, got %d", m.ReadingSession.SentenceIdx)
	}
}



