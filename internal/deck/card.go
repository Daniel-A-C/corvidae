// Package deck provides flashcard and deck data models and disk persistence.
package deck

import "fmt"

// Flashcard represents an individual flashcard item with spaced repetition metadata.
type Flashcard struct {
	Character     string  `yaml:"character"`
	Pinyin        string  `yaml:"pinyin,omitempty"`
	Pronunciation string  `yaml:"pronunciation,omitempty"`
	Meaning       string  `yaml:"meaning"`
	Explanation   string  `yaml:"explanation,omitempty"`
	Interval      int     `yaml:"interval,omitempty"`
	Ease          float64 `yaml:"ease,omitempty"`
	Reps          int     `yaml:"reps,omitempty"`
	NextReview    string  `yaml:"next_review,omitempty"`
	Language      string  `yaml:"language,omitempty"`
	Deck          string  `yaml:"deck,omitempty"`
}

// Deck represents a collection of flashcards stored in a YAML deck file.
type Deck struct {
	Selected     bool        `yaml:"selected,omitempty"`
	QuizSelected bool        `yaml:"quiz_selected,omitempty"`
	Cards        []Flashcard `yaml:"cards"`
}

// IsSelected reports whether the deck has been marked as selected in YAML.
func (d Deck) IsSelected() bool {
	return d.Selected || d.QuizSelected
}

// CardRef maps an active card back to its originating deck file and card index.
type CardRef struct {
	Filename string
	OrigIdx  int
}

// FormatQuizOption returns a formatted string representation suitable for multiple-choice quiz options.
func (f Flashcard) FormatQuizOption() string {
	phonetic := f.Pinyin
	if phonetic == "" {
		phonetic = f.Pronunciation
	}
	if phonetic != "" {
		return fmt.Sprintf("%s - %s", phonetic, f.Meaning)
	}
	return f.Meaning
}
