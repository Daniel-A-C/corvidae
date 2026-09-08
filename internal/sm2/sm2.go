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

// CalculateSM2 updates card spaced repetition parameters (interval, ease, reps, next_review)
// according to the SuperMemo-2 (SM-2) algorithm. If now is zero, time.Now() is used.
func CalculateSM2(card deck.Flashcard, grade int, now time.Time) deck.Flashcard {
	if now.IsZero() {
		now = time.Now()
	}

	if card.Ease == 0 {
		card.Ease = DefaultEase
	}

	if grade >= GradeGood {
		if card.Reps == 0 {
			card.Interval = 1
		} else if card.Reps == 1 {
			card.Interval = 6
		} else {
			card.Interval = int(math.Round(float64(card.Interval) * card.Ease))
		}
		card.Reps++
	} else {
		card.Reps = 0
		card.Interval = 1
	}

	card.Ease = card.Ease + (0.1 - float64(5-grade)*(0.08+float64(5-grade)*0.02))
	if card.Ease < MinEase {
		card.Ease = MinEase
	}

	next := now.AddDate(0, 0, card.Interval)
	card.NextReview = next.Format("2006-01-02")
	return card
}
