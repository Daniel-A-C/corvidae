package forge

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"flashcards/internal/deck"
)

func loadRealLexicon(t *testing.T) *Lexicon {
	t.Helper()
	lex, err := LoadLexicon(filepath.Join("..", "..", "forge", DefaultLexiconFilename))
	if err != nil {
		t.Fatalf("failed to load mandarin_lexicon.yaml: %v", err)
	}
	return lex
}

func TestLexiconIntegrityAndConnectivity(t *testing.T) {
	lex := loadRealLexicon(t)
	if len(lex.Hanzi) < 100 {
		t.Fatalf("expected >= 100 hanzi, got %d", len(lex.Hanzi))
	}
	if len(lex.Compounds) < 250 {
		t.Fatalf("expected >= 250 compounds, got %d", len(lex.Compounds))
	}

	// Every character in every compound must exist in lex.CharsByGlyph
	for _, c := range lex.Compounds {
		runes := []rune(c.Word)
		if len(runes) < 2 || len(runes) > 4 {
			t.Errorf("compound %q has unexpected length %d", c.Word, len(runes))
		}
		if len(c.Tones) != len(runes) {
			t.Errorf("compound %q has %d tones for %d runes", c.Word, len(c.Tones), len(runes))
		}
		for _, r := range runes {
			if _, ok := lex.CharsByGlyph[string(r)]; !ok {
				t.Errorf("compound %q references unknown hanzi %q", c.Word, string(r))
			}
		}
	}

	// Verify all 4 starter decks produce valid hands with high compound connectivity
	prog := NewEmptyProgress("")
	for _, arch := range StarterArchetypes {
		glyphs := lex.StarterGlyphs(arch.ID, prog)
		if len(glyphs) < 25 {
			t.Errorf("starter %s has too few glyphs: %d", arch.ID, len(glyphs))
		}
	}
}

func TestPivotChainClassificationAndScoring(t *testing.T) {
	lex := loadRealLexicon(t)
	prog := NewEmptyProgress("")
	run := NewRun(lex, prog, "harry_potter", 42)

	// Find two real compounds in lex that share a bridge char: A+B and B+C
	var w1, w2 *CompoundEntry
	var a, b, c string
	for i := range lex.Compounds {
		r1 := []rune(lex.Compounds[i].Word)
		if len(r1) != 2 {
			continue
		}
		bridge := string(r1[1])
		for _, cand := range lex.CompoundsByChar[bridge] {
			r2 := []rune(cand.Word)
			if len(r2) == 2 && string(r2[0]) == bridge && string(r2[1]) != string(r1[0]) && string(r2[1]) != bridge {
				w1 = &lex.Compounds[i]
				w2 = cand
				a, b, c = string(r1[0]), bridge, string(r2[1])
				break
			}
		}
		if w1 != nil {
			break
		}
	}
	if w1 == nil || w2 == nil {
		t.Fatalf("could not find a 2-char pivot chain pair in lexicon")
	}

	// Place a, b, c into hand positions 0, 1, 2
	run.Hand = []DeckCard{
		{ID: 1, Glyph: a},
		{ID: 2, Glyph: b},
		{ID: 3, Glyph: c},
	}
	run.ClearStage()
	if !run.StageCard(0) || !run.StageCard(1) {
		t.Fatalf("failed to stage first word %s", w1.Word)
	}
	if !run.StartPivotSlot() {
		t.Fatalf("StartPivotSlot should succeed after 2-char word")
	}
	if !run.StageCard(2) {
		t.Fatalf("failed to stage third char for pivot word %s", w2.Word)
	}

	if !run.CommitPlay() {
		t.Fatalf("CommitPlay returned false for valid pivot chain %s + %s", w1.Word, w2.Word)
	}
	if len(run.Challenges) != 2 {
		t.Fatalf("expected 2 prove challenges, got %d", len(run.Challenges))
	}

	now := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	// Answer both correctly
	passed1, more1 := run.AnswerProve(run.Challenges[0].CorrectIdx, now, false, "")
	if !passed1 || !more1 {
		t.Fatalf("expected first prove pass with more remaining, got passed=%v more=%v", passed1, more1)
	}
	passed2, more2 := run.AnswerProve(run.Challenges[1].CorrectIdx, now, false, "")
	if !passed2 || more2 {
		t.Fatalf("expected second prove pass and finish, got passed=%v more=%v", passed2, more2)
	}
	if run.LastPlayResult == nil {
		t.Fatalf("expected LastPlayResult to be populated")
	}
	if run.LastPlayResult.HandType != HandPivotChain {
		t.Errorf("expected HandPivotChain, got %v", run.LastPlayResult.HandType)
	}
	if run.LastPlayResult.TotalScore <= 0 {
		t.Errorf("expected positive TotalScore, got %d", run.LastPlayResult.TotalScore)
	}
}

func TestAntiCheeseInvalidFizzleAndGlassShatter(t *testing.T) {
	lex := loadRealLexicon(t)
	prog := NewEmptyProgress("")
	run := NewRun(lex, prog, "harry_potter", 7)
	initialHands := run.HandsLeft

	// Pick a guaranteed non-word pair
	run.Hand = []DeckCard{
		{ID: 1, Glyph: "镜"},
		{ID: 2, Glyph: "谎"},
	}
	run.ClearStage()
	run.StageCard(0)
	run.StageCard(1)
	ok := run.CommitPlay()
	if ok {
		t.Fatalf("expected invalid combo 镜谎 to fizzle immediately")
	}
	if run.HandsLeft != initialHands-1 {
		t.Errorf("expected fizzle to consume 1 hand, got %d (was %d)", run.HandsLeft, initialHands)
	}
	if run.LastPlayResult == nil || run.LastPlayResult.TotalScore != 0 {
		t.Errorf("expected 0 score on fizzle")
	}

	// Test Glass Shatter when missing a valid compound
	comp := lex.Compounds[0]
	r := []rune(comp.Word)
	run.Hand = nil
	for i, ch := range r {
		run.Hand = append(run.Hand, DeckCard{ID: 10 + i, Glyph: string(ch)})
	}
	run.ClearStage()
	for i := range r {
		run.StageCard(i)
	}
	if !run.CommitPlay() {
		t.Fatalf("expected valid compound %s to enter Prove phase", comp.Word)
	}
	wrongIdx := (run.Challenges[0].CorrectIdx + 1) % len(run.Challenges[0].Options)
	passed, more := run.AnswerProve(wrongIdx, time.Now(), false, "")
	if passed || more {
		t.Fatalf("expected miss and completion, got passed=%v more=%v", passed, more)
	}
	if !run.LastPlayResult.GlassShattered {
		t.Errorf("expected GlassShattered=true on missed new/volatile compound")
	}
	if run.LastPlayResult.TotalScore != 0 {
		t.Errorf("expected 0 score when only word was missed, got %d", run.LastPlayResult.TotalScore)
	}
}

func TestParallelProgressAndMandarinDeckSync(t *testing.T) {
	tmpDir := t.TempDir()
	progPath := filepath.Join(tmpDir, "forge_progress.yaml")
	prog, err := LoadProgress(progPath)
	if err != nil {
		t.Fatalf("LoadProgress failed: %v", err)
	}
	if prog.MaterialForChar("来") != MaterialGlass {
		t.Errorf("expected new char to be MaterialGlass")
	}

	now := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	prog.RecordOutcome("回来", true, now)
	if prog.MaterialForChar("来") != MaterialFoil {
		t.Errorf("expected char with 1 rep to become MaterialFoil, got %v", prog.MaterialForChar("来"))
	}
	if err := prog.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	reloaded, err := LoadProgress(progPath)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if reloaded.Compounds["回来"].Reps != 1 {
		t.Errorf("expected persisted compound reps=1, got %d", reloaded.Compounds["回来"].Reps)
	}

	// Test optional sync to decks/Mandarin/
	decksDir := filepath.Join(tmpDir, "decks")
	hpDir := filepath.Join(decksDir, "Mandarin", "HarryPotter")
	if err := os.MkdirAll(hpDir, 0755); err != nil {
		t.Fatal(err)
	}
	deckPath := filepath.Join(hpDir, "test.yaml")
	d := deck.Deck{
		Cards: []deck.Flashcard{
			{Character: "回来", Pinyin: "huí lái", Meaning: "To return"},
		},
	}
	if err := deck.SaveDeck(deckPath, d); err != nil {
		t.Fatal(err)
	}

	synced := SyncCompoundToMandarinDecks(decksDir, "回来", true, now)
	if synced != 1 {
		t.Fatalf("expected 1 synced card, got %d", synced)
	}
	updated, err := deck.LoadDeck(deckPath)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Cards[0].Reps != 1 || updated.Cards[0].Interval != 1 {
		t.Errorf("expected synced flashcard Reps=1 Interval=1, got %+v", updated.Cards[0])
	}
}
