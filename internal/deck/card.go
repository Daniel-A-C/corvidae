package deck

import "fmt"

// Flashcard represents an individual flashcard item with spaced repetition metadata.
type Flashcard struct {
	Character   string  `yaml:"character"`
	Pinyin      string  `yaml:"pinyin,omitempty"`
	Meaning     string  `yaml:"meaning"`
	Explanation string  `yaml:"explanation,omitempty"`
	Interval    int     `yaml:"interval,omitempty"`
	Ease        float64 `yaml:"ease,omitempty"`
	Reps        int     `yaml:"reps,omitempty"`
	NextReview  string  `yaml:"next_review,omitempty"`
}

// Deck represents a collection of flashcards stored in a YAML deck file.
type Deck struct {
	Cards []Flashcard `yaml:"cards"`
}

// CardRef maps an active card back to its originating deck file and card index.
type CardRef struct {
	Filename string
	OrigIdx  int
}

// FormatQuizOption returns a formatted string representation suitable for multiple-choice quiz options.
func (f Flashcard) FormatQuizOption() string {
	if f.Pinyin != "" {
		return fmt.Sprintf("%s - %s", f.Pinyin, f.Meaning)
	}
	return f.Meaning
}
