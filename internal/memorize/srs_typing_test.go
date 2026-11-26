package memorize

import (
	"path/filepath"
	"testing"
	"time"
)

func TestFirstLetterAndMatching(t *testing.T) {
	cases := []struct {
		word  string
		input rune
		want  bool
	}{
		{"Pablo,", 'p', true},
		{"Pablo,", 'P', true},
		{"Ésta", 'e', true},
		{"Ésta", 'é', true},
		{"¡Aleluya!", 'a', true},
		{"¿Quién?", 'q', true},
		{"\"gracia\"", 'g', true},
		{"ñandú", 'n', true},
		{"prisionero", 's', false},
		{"123", '1', true},
	}

	for _, tc := range cases {
		if got := MatchesFirstLetter(tc.word, tc.input); got != tc.want {
			t.Errorf("MatchesFirstLetter(%q, %q) = %v, want %v", tc.word, tc.input, got, tc.want)
		}
	}
}

func TestSectionSM2SchedulingAndPersistence(t *testing.T) {
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	txt := &Text{
		Title: "Test Passage",
		Words: []WordItem{
			{Word: "First", Prefix: "1 "},
			{Word: "verse"},
			{Word: "Second", Prefix: "\n\n2 "},
			{Word: "verse"},
			{Word: "Third", Prefix: "\n\n3 "},
			{Word: "verse"},
		},
	}
	sections := txt.GetSections()
	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}
	if due := txt.DueSectionCount(now); due != 3 {
		t.Fatalf("expected 3 due sections initially, got %d", due)
	}

	// Review sections 0..1 with 100% accuracy
	txt.RecordPortionReview(sections, 0, 1, 100.0, "typing", now)
	if due := txt.DueSectionCount(now); due != 1 {
		t.Errorf("expected 1 due section remaining today, got %d", due)
	}

	// NextDueRange should pick section 2
	s, e, ok := txt.NextDueRange(sections, now)
	if !ok || s != 2 || e != 2 {
		t.Errorf("expected NextDueRange (2, 2, true), got (%d, %d, %v)", s, e, ok)
	}

	// Save and reload round-trip
	path := filepath.Join(t.TempDir(), "passage.yaml")
	if err := SaveText(path, txt); err != nil {
		t.Fatalf("SaveText failed: %v", err)
	}
	loaded, err := LoadText(path)
	if err != nil {
		t.Fatalf("LoadText failed: %v", err)
	}
	rev, ok := loaded.SectionReviewFor(sections[0])
	if !ok {
		t.Fatalf("expected saved SectionReview for section 1")
	}
	if rev.Reps != 1 || rev.Interval != 1 || rev.NextReview != "2026-11-26" || rev.LastMode != "typing" {
		t.Errorf("unexpected SectionReview: %+v", rev)
	}
}
