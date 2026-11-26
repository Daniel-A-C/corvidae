// Package sm2 implements the SuperMemo-2 (SM-2) spaced repetition scheduling algorithm.
package sm2

import (
	"math"
	"time"

	"flashcards/internal/deck"
)

const (
	DefaultEase = 2.5
	MinEase     = 1.3

	GradeBlackout = 0
	GradeWrong    = 1
	GradeHard     = 2
	GradeGood     = 3
	GradeEasy     = 4
	GradePerfect  = 5
)

// Schedule computes updated SM-2 parameters (interval, reps, ease, nextReview) for any spaced-repetition item.
// If now is zero, time.Now() is used.
func Schedule(interval, reps int, ease float64, grade int, now time.Time) (int, int, float64, string) {
	if now.IsZero() {
		now = time.Now()
	}

	if ease == 0 {
		ease = DefaultEase
	}

	if grade >= GradeGood {
		if reps == 0 {
			interval = 1
		} else if reps == 1 {
			interval = 6
		} else {
			interval = int(math.Round(float64(interval) * ease))
		}
		reps++
	} else {
		reps = 0
		interval = 1
	}

	ease = ease + (0.1 - float64(5-grade)*(0.08+float64(5-grade)*0.02))
	if ease < MinEase {
		ease = MinEase
	}

	next := now.AddDate(0, 0, interval)
	return interval, reps, ease, next.Format("2006-01-02")
}

// CalculateSM2 updates card spaced repetition parameters (interval, ease, reps, next_review)
// according to the SuperMemo-2 (SM-2) algorithm. If now is zero, time.Now() is used.
func CalculateSM2(card deck.Flashcard, grade int, now time.Time) deck.Flashcard {
	card.Interval, card.Reps, card.Ease, card.NextReview = Schedule(card.Interval, card.Reps, card.Ease, grade, now)
	return card
}
