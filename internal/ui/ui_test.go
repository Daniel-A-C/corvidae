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

	// Verify each deck file has its corresponding key badge
	for i := 0; i < len(m.DeckFiles); i++ {
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

