package ui

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"flashcards/internal/config"
	"flashcards/internal/deck"
	"flashcards/internal/memorize"

	tea "github.com/charmbracelet/bubbletea"
)

func setupTempDeckModel(t *testing.T) (Model, string) {
	t.Helper()
	tmp := t.TempDir()
	decksDir := filepath.Join(tmp, "decks")
	deckPath := filepath.Join(decksDir, "Mandarin", "sample.yaml")
	d := deck.Deck{
		Selected: true,
		Cards: []deck.Flashcard{
			{Character: "电脑", Pinyin: "diànnǎo", Meaning: "Computer"},
			{Character: "书", Pinyin: "shū", Meaning: "Book"},
			{Character: "水", Pinyin: "shuǐ", Meaning: "Water"},
			{Character: "火", Pinyin: "huǒ", Meaning: "Fire"},
			{Character: "风", Pinyin: "fēng", Meaning: "Wind"},
			{Character: "山", Pinyin: "shān", Meaning: "Mountain"},
		},
	}
	if err := deck.SaveDeck(deckPath, d); err != nil {
		t.Fatalf("SaveDeck failed: %v", err)
	}

	cfg := config.Default()
	cfg.FilePath = filepath.Join(tmp, "config.yaml")
	cfg.DecksDir = decksDir
	cfg.TextsDir = filepath.Join(tmp, "texts")
	cfg.ReadingDir = filepath.Join(tmp, "reading")
	cfg.StatsFile = filepath.Join(tmp, "stats.yaml")

	m := NewWithConfig(cfg)
	m.Width = 100
	m.Height = 40
	return m, deckPath
}

func TestSettingsScreenPersistsConfig(t *testing.T) {
	m, _ := setupTempDeckModel(t)

	// Open settings with 'c'
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(Model)
	if m.State != StateSettings {
		t.Fatalf("expected StateSettings, got %d", m.State)
	}

	// Toggle Quiz SM-2 Sync ([a]) and Match SM-2 Sync ([s])
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(Model)

	if !m.Config.QuizSM2Sync || !m.Config.MatchSM2Sync {
		t.Fatalf("expected QuizSM2Sync and MatchSM2Sync to be true")
	}

	reloaded, err := config.Load(m.Config.FilePath)
	if err != nil {
		t.Fatalf("config.Load failed: %v", err)
	}
	if !reloaded.QuizSM2Sync || !reloaded.MatchSM2Sync {
		t.Errorf("expected persisted SM-2 sync flags in config.yaml, got %+v", reloaded)
	}
}

func TestQuizSM2SyncAndMisclickRollback(t *testing.T) {
	m, deckPath := setupTempDeckModel(t)
	m.Config.QuizSM2Sync = true
	m.Mode = ModeQuiz
	if err := m.LoadSelectedDecks(); err != nil {
		t.Fatalf("LoadSelectedDecks failed: %v", err)
	}
	m.SetupQuiz()
	m.State = StateQuiz

	// Pick a wrong option deliberately
	wrongIdx := (m.CorrectIndex + 1) % len(m.QuizOptions)
	ref := m.ActiveCards[m.CurrentIndex]
	wrongKey := IndexToKey(wrongIdx)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(wrongKey)})
	m = updated.(Model)
	if !m.ShowFeedback || m.IsCorrect {
		t.Fatalf("expected wrong answer feedback")
	}

	// Card should have been graded GradeWrong (Reps=0, Ease < 2.5)
	d, err := deck.LoadDeck(deckPath)
	if err != nil {
		t.Fatalf("LoadDeck failed: %v", err)
	}
	if d.Cards[ref.OrigIdx].Reps != 0 || d.Cards[ref.OrigIdx].Ease >= 2.5 {
		t.Errorf("expected GradeWrong SM-2 update, got Reps=%d Ease=%.2f", d.Cards[ref.OrigIdx].Reps, d.Cards[ref.OrigIdx].Ease)
	}

	// Press 'k' ("I knew it") -> should roll back snapshot and apply GradeGood (Reps=1)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	d, err = deck.LoadDeck(deckPath)
	if err != nil {
		t.Fatalf("LoadDeck after rollback failed: %v", err)
	}
	if d.Cards[ref.OrigIdx].Reps != 1 {
		t.Errorf("expected Reps=1 after 'I knew it' rollback to GradeGood, got %d", d.Cards[ref.OrigIdx].Reps)
	}
	if d.Cards[ref.OrigIdx].NextReview == "" {
		t.Errorf("expected NextReview to be populated")
	}
}

func TestMatchSM2SyncCleanVsMistake(t *testing.T) {
	m, deckPath := setupTempDeckModel(t)
	m.Config.MatchSM2Sync = true
	m.Mode = ModeMatch
	if err := m.LoadSelectedDecks(); err != nil {
		t.Fatalf("LoadSelectedDecks failed: %v", err)
	}
	m.SetupMatch()
	m.State = StateMatch

	// Match left[0] cleanly with its matching right item
	targetPairID := m.MatchLeftItems[0].PairID
	targetRef := m.MatchLeftItems[0].CardRef
	rightIdx := -1
	for i, item := range m.MatchRightItems {
		if item.PairID == targetPairID {
			rightIdx = i
			break
		}
	}
	if rightIdx == -1 {
		t.Fatalf("matching right item not found")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(IndexToKey(0))})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(IndexToKey(rightIdx + 5))})
	m = updated.(Model)

	d, err := deck.LoadDeck(deckPath)
	if err != nil {
		t.Fatalf("LoadDeck failed: %v", err)
	}
	if d.Cards[targetRef.OrigIdx].Reps != 1 {
		t.Errorf("expected clean match to apply GradeGood (Reps=1), got Reps=%d", d.Cards[targetRef.OrigIdx].Reps)
	}
}

func TestMemorizeFirstLetterTypingAndSectionSRS(t *testing.T) {
	m, _ := setupTempDeckModel(t)
	textPath := filepath.Join(m.TextsDir, "short.yaml")
	txt := &memorize.Text{
		Title:    "Short Test",
		Language: "Spanish",
		Words: []memorize.WordItem{
			{Word: "¿Quién", Prefix: "1 ", Options: []string{"¿Quién", "Qué"}},
			{Word: "Ésta", Options: []string{"Ésta", "Eso"}},
			{Word: "paz.", Prefix: "\n\n2 ", Options: []string{"paz.", "luz."}},
		},
	}
	if err := memorize.SaveText(textPath, txt); err != nil {
		t.Fatalf("SaveText failed: %v", err)
	}

	if err := m.SelectMemorizeText(textPath); err != nil {
		t.Fatalf("SelectMemorizeText failed: %v", err)
	}

	// Toggle to First-Letter Typing with 't', collapse to verse 1 with '1', start with Enter
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = updated.(Model)
	if !m.MemorizeTypingMode {
		t.Fatalf("expected MemorizeTypingMode=true after 't'")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	// Word 0 is "¿Quién" — pressing 'q' must NOT quit, it should match and advance!
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(Model)
	if cmd != nil {
		t.Fatalf("pressing 'q' in typing mode should not trigger tea.Quit")
	}
	if m.MemorizeCurrentIndex != 1 {
		t.Fatalf("expected to advance to word 1 after typing 'q' for \"¿Quién\", got idx=%d", m.MemorizeCurrentIndex)
	}

	// Word 1 is "Ésta" — pressing 'e' (diacritic-insensitive) completes Verse 1 and records SM-2!
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(Model)
	if m.State != StateMemorizePortionSelect {
		t.Fatalf("expected return to StateMemorizePortionSelect on portion complete, got %d", m.State)
	}

	reloaded, err := memorize.LoadText(textPath)
	if err != nil {
		t.Fatalf("reload text failed: %v", err)
	}
	sections := reloaded.GetSections()
	rev, ok := reloaded.SectionReviewFor(sections[0])
	if !ok || rev.Reps != 1 || rev.LastMode != config.MemorizeModeTyping {
		t.Errorf("expected section 1 SRS saved with Reps=1 and mode=typing, got %+v (ok=%v)", rev, ok)
	}

	// Press 'd' in Portion Select -> should select next due section (Verse 2, index 1)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(Model)
	if m.PortionStartSection != 1 || m.PortionEndSection != 1 {
		t.Errorf("expected [d] to select Verse 2 (1..1), got %d..%d", m.PortionStartSection, m.PortionEndSection)
	}
}

func TestDashboardRenderingAndStatsRecording(t *testing.T) {
	m, _ := setupTempDeckModel(t)
	now := time.Now()
	m.Stats.RecordReview(now)
	m.Stats.RecordQuiz(true, now)
	m.Stats.RecordMatchPair(now)
	m.Stats.RecordMemorizeWord(now)

	// Open dashboard with 'i' from Main Menu
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = updated.(Model)
	if m.State != StateDashboard {
		t.Fatalf("expected StateDashboard, got %d", m.State)
	}
	if m.DashboardMaturity.TotalCards != 6 {
		t.Errorf("expected 6 cards scanned in maturity, got %d", m.DashboardMaturity.TotalCards)
	}

	view := m.View()
	if !strings.Contains(view, "STUDY DASHBOARD") || !strings.Contains(view, "Activity Heatmap") {
		t.Errorf("expected dashboard view to contain header and heatmap, got:\n%s", view)
	}
	if !strings.Contains(view, "Flashcard Maturity") || !strings.Contains(view, "30-Day Breakdown") {
		t.Errorf("expected dashboard view to contain maturity and 30-day breakdown")
	}
}
