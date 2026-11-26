package stats

import (
	"path/filepath"
	"testing"
	"time"

	"flashcards/internal/deck"
)

func TestStoreRecordAndPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stats.yaml")
	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load missing failed: %v", err)
	}

	day1 := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
	day3 := time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)

	s.RecordReview(day1)
	s.RecordQuiz(true, day1)
	s.RecordQuiz(false, day2)
	s.RecordQuizOverride(day2)
	s.RecordMatchPair(day2)
	s.RecordMemorizeWord(day3)
	s.RecordReadingSentence(day3)
	s.RecordArabicDrill(day3)

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}
	if got := reloaded.TotalForDate("2026-11-20"); got != 2 {
		t.Errorf("expected 2 actions on day1, got %d", got)
	}
	if got := reloaded.Days["2026-11-21"].QuizCorrect; got != 1 {
		t.Errorf("expected quiz override to increment QuizCorrect to 1, got %d", got)
	}
	if got := reloaded.CurrentStreak(day3); got != 3 {
		t.Errorf("expected current streak 3, got %d", got)
	}
	if got := reloaded.LongestStreak(); got != 3 {
		t.Errorf("expected longest streak 3, got %d", got)
	}
	// If today has no activity yet, streak ending yesterday still counts
	day4 := time.Date(2026, 11, 23, 10, 0, 0, 0, time.UTC)
	if got := reloaded.CurrentStreak(day4); got != 3 {
		t.Errorf("expected current streak 3 when yesterday was active, got %d", got)
	}
	// Gap of 2 days resets current streak
	day5 := time.Date(2026, 11, 24, 10, 0, 0, 0, time.UTC)
	if got := reloaded.CurrentStreak(day5); got != 0 {
		t.Errorf("expected current streak 0 after 2-day gap, got %d", got)
	}
}

func TestHeatmapGridDimensionsAndLevels(t *testing.T) {
	s := NewEmpty(filepath.Join(t.TempDir(), "stats.yaml"))
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC) // Wednesday
	for i := 0; i < 40; i++ {
		s.RecordReview(now)
	}

	grid := s.HeatmapGrid(12, now)
	if len(grid) != 7 {
		t.Fatalf("expected 7 rows, got %d", len(grid))
	}
	if len(grid[0]) != 12 {
		t.Fatalf("expected 12 columns, got %d", len(grid[0]))
	}
	// Wednesday is row 2 (Mon=0, Tue=1, Wed=2) in the last column (11)
	cell := grid[2][11]
	if cell.Count != 40 || cell.Level != 4 || cell.Future {
		t.Errorf("unexpected today cell: %+v", cell)
	}
	// Thursday (row 3, col 11) should be marked Future
	if !grid[3][11].Future {
		t.Errorf("expected Thursday in current week to be marked Future")
	}
}

func TestScanDeckMaturity(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	d := deck.Deck{
		Cards: []deck.Flashcard{
			{Character: "新", Meaning: "New"},
			{Character: "学", Meaning: "Learning", Reps: 1, Interval: 6, Ease: 2.5, NextReview: "2026-11-25"},
			{Character: "熟", Meaning: "Mature", Reps: 5, Interval: 30, Ease: 2.6, NextReview: "2026-12-20"},
			{Character: "难", Meaning: "Leech", Reps: 0, Interval: 1, Ease: 1.3, NextReview: "2026-11-24"},
		},
	}
	if err := deck.SaveDeck(filepath.Join(dir, "Mandarin", "test.yaml"), d); err != nil {
		t.Fatalf("SaveDeck failed: %v", err)
	}
	// marked.yaml should be ignored
	_ = deck.SaveDeck(filepath.Join(dir, "marked.yaml"), d)

	m := ScanDeckMaturity(dir, now)
	if m.TotalDecks != 1 || m.TotalCards != 4 {
		t.Errorf("expected 1 deck / 4 cards, got %+v", m)
	}
	if m.NewCards != 1 || m.LearningCards != 2 || m.MatureCards != 1 {
		t.Errorf("unexpected maturity breakdown: %+v", m)
	}
	if m.DueToday != 3 {
		t.Errorf("expected 3 due today (New + Learning due today + Leech overdue), got %d", m.DueToday)
	}
	if m.Leeches != 1 {
		t.Errorf("expected 1 leech, got %d", m.Leeches)
	}
}
