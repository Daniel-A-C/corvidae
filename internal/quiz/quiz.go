package quiz

import (
	"math/rand"

	"flashcards/internal/deck"
)

// GenerateOptions generates multiple-choice options for a given target card using a pool of all available cards.
// It returns the shuffled slice of options and the index of the correct answer within that slice.
func GenerateOptions(targetCard deck.Flashcard, allCards []deck.Flashcard, maxDistractors int) ([]string, int) {
	correctOption := targetCard.FormatQuizOption()

	uniquePool := make(map[string]bool)
	var pool []string
	for _, card := range allCards {
		opt := card.FormatQuizOption()
		if opt != correctOption && !uniquePool[opt] {
			uniquePool[opt] = true
			pool = append(pool, opt)
		}
	}

	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

	numDistractors := maxDistractors
	if len(pool) < maxDistractors {
		numDistractors = len(pool)
	}

	options := []string{correctOption}
	options = append(options, pool[:numDistractors]...)

	rand.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })

	correctIndex := 0
	for i, opt := range options {
		if opt == correctOption {
			correctIndex = i
			break
		}
	}

	return options, correctIndex
}
