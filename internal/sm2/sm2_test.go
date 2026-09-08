package sm2

import (
	"testing"
	"time"

	"flashcards/internal/deck"
)

func TestCalculateSM2_InitialReview(t *testing.T) {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	card := deck.Flashcard{
		Character: "你好",
		Meaning:   "Hello",
	}

	// Grade Good (3) on first review
	updated := CalculateSM2(card, GradeGood, now)

	if updated.Ease != 2.36 { // 2.5 + (0.1 - (2)*(0.08 + 2*0.02)) = 2.5 + (0.1 - 0.24) = 2.36
		t.Errorf("expected ease 2.36, got %f", updated.Ease)
	}
	if updated.Reps != 1 {
		t.Errorf("expected reps 1, got %d", updated.Reps)
	}
	if updated.Interval != 1 {
		t.Errorf("expected interval 1, got %d", updated.Interval)
	}
	if updated.NextReview != "2026-09-09" {
		t.Errorf("expected next review 2026-09-09, got %s", updated.NextReview)
	}
}

func TestCalculateSM2_RepetitionProgression(t *testing.T) {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	card := deck.Flashcard{
		Reps:     1,
		Interval: 1,
		Ease:     2.5,
	}

	// Rep 2: Interval should become 6
	updated := CalculateSM2(card, GradeEasy, now)
	if updated.Reps != 2 {
		t.Errorf("expected reps 2, got %d", updated.Reps)
	}
	if updated.Interval != 6 {
		t.Errorf("expected interval 6, got %d", updated.Interval)
	}
	if updated.NextReview != "2026-09-14" {
		t.Errorf("expected next review 2026-09-14, got %s", updated.NextReview)
	}

	// Rep 3: Interval is calculated using current ease (2.5) before ease is updated:
	// Interval = round(6 * 2.5) = 15
	// Then ease is updated: 2.5 + (0.1 - 0) = 2.6
	updated2 := CalculateSM2(updated, GradePerfect, now)
	if updated2.Reps != 3 {
		t.Errorf("expected reps 3, got %d", updated2.Reps)
	}
	if updated2.Interval != 15 {
		t.Errorf("expected interval 15, got %d", updated2.Interval)
	}
	if updated2.Ease != 2.6 {
		t.Errorf("expected ease 2.6, got %f", updated2.Ease)
	}
}

func TestCalculateSM2_FailureReset(t *testing.T) {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	card := deck.Flashcard{
		Reps:     4,
		Interval: 25,
		Ease:     2.2,
	}

	// Failed review (GradeHard = 2 < 3)
	updated := CalculateSM2(card, GradeHard, now)
	if updated.Reps != 0 {
		t.Errorf("expected reps to reset to 0, got %d", updated.Reps)
	}
	if updated.Interval != 1 {
		t.Errorf("expected interval to reset to 1, got %d", updated.Interval)
	}
	if updated.NextReview != "2026-09-09" {
		t.Errorf("expected next review 2026-09-09, got %s", updated.NextReview)
	}
}

func TestCalculateSM2_MinEaseBoundary(t *testing.T) {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	card := deck.Flashcard{
		Ease: 1.35,
	}

	// GradeBlackout (0) decreases ease dramatically
	updated := CalculateSM2(card, GradeBlackout, now)
	if updated.Ease != MinEase {
		t.Errorf("expected ease clamped to MinEase %f, got %f", MinEase, updated.Ease)
	}
}
