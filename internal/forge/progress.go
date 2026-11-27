package forge

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"flashcards/internal/deck"
	"flashcards/internal/sm2"
	"gopkg.in/yaml.v3"
)

const DefaultProgressFilename = "progress.yaml"

// SM2Item stores spaced-repetition scheduling parameters for a single Hanzi or Compound.
type SM2Item struct {
	Interval   int     `yaml:"interval,omitempty"`
	Ease       float64 `yaml:"ease,omitempty"`
	Reps       int     `yaml:"reps,omitempty"`
	NextReview string  `yaml:"next_review,omitempty"`
	Passes     int     `yaml:"passes,omitempty"`
	Misses     int     `yaml:"misses,omitempty"`
}

// ProgressStore persists parallel SM-2 progress and lifetime run stats for Morpheme Forge.
type ProgressStore struct {
	Characters   map[string]SM2Item `yaml:"characters,omitempty"`
	Compounds    map[string]SM2Item `yaml:"compounds,omitempty"`
	RunsPlayed   int                `yaml:"runs_played,omitempty"`
	RunsWon      int                `yaml:"runs_won,omitempty"`
	BestScore    int                `yaml:"best_score,omitempty"`
	BestAnte     int                `yaml:"best_ante,omitempty"`
	TotalForged  int                `yaml:"total_forged,omitempty"`
	FilePath     string             `yaml:"-"`
}

// NewEmptyProgress creates an in-memory ProgressStore pointing at filePath.
func NewEmptyProgress(filePath string) *ProgressStore {
	return &ProgressStore{
		Characters: make(map[string]SM2Item),
		Compounds:  make(map[string]SM2Item),
		FilePath:   filePath,
	}
}

// LoadProgress reads forge/progress.yaml. If missing, returns an empty ProgressStore.
func LoadProgress(path string) (*ProgressStore, error) {
	p := NewEmptyProgress(path)
	if path == "" {
		return p, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return p, nil
		}
		return p, fmt.Errorf("failed to read forge progress %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, p); err != nil {
		return p, fmt.Errorf("failed to parse forge progress %s: %w", path, err)
	}
	if p.Characters == nil {
		p.Characters = make(map[string]SM2Item)
	}
	if p.Compounds == nil {
		p.Compounds = make(map[string]SM2Item)
	}
	p.FilePath = path
	return p, nil
}

// Save writes the ProgressStore to disk if FilePath is set.
func (p *ProgressStore) Save() error {
	if p == nil || p.FilePath == "" {
		return nil
	}
	dir := filepath.Dir(p.FilePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	data, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(p.FilePath, data, 0644)
}

// MaterialForChar derives the card's in-game MaterialTier from its SM-2 memory state.
func (p *ProgressStore) MaterialForChar(glyph string) MaterialTier {
	if p == nil {
		return MaterialGlass
	}
	item, ok := p.Characters[glyph]
	if !ok || item.Reps == 0 || (item.Ease > 0 && item.Ease < 1.8) {
		return MaterialGlass
	}
	if item.Interval >= 21 {
		return MaterialGold
	}
	return MaterialFoil
}

// IsVolatileCompound reports whether a forged compound is New or a Leech in parallel SM-2.
func (p *ProgressStore) IsVolatileCompound(word string) bool {
	if p == nil {
		return true
	}
	item, ok := p.Compounds[word]
	if !ok || item.Reps == 0 || (item.Ease > 0 && item.Ease < 1.8) {
		return true
	}
	return false
}

// RecordOutcome updates parallel SM-2 scheduling for a forged word and its constituent Hanzi.
func (p *ProgressStore) RecordOutcome(word string, passed bool, now time.Time) {
	if p == nil || word == "" {
		return
	}
	grade := sm2.GradeWrong
	if passed {
		grade = sm2.GradeGood
	}
	// Update compound entry (for multi-char words)
	if len([]rune(word)) >= 2 {
		cItem := p.Compounds[word]
		cItem.Interval, cItem.Reps, cItem.Ease, cItem.NextReview = sm2.Schedule(
			cItem.Interval, cItem.Reps, cItem.Ease, grade, now,
		)
		if passed {
			cItem.Passes++
			p.TotalForged++
		} else {
			cItem.Misses++
		}
		p.Compounds[word] = cItem
	} else if passed {
		p.TotalForged++
	}

	// Update constituent Hanzi entries
	seen := make(map[string]bool)
	for _, r := range word {
		g := string(r)
		if seen[g] {
			continue
		}
		seen[g] = true
		hItem := p.Characters[g]
		hItem.Interval, hItem.Reps, hItem.Ease, hItem.NextReview = sm2.Schedule(
			hItem.Interval, hItem.Reps, hItem.Ease, grade, now,
		)
		if passed {
			hItem.Passes++
		} else {
			hItem.Misses++
		}
		p.Characters[g] = hItem
	}
}

// SyncCompoundToMandarinDecks scans baseDecksDir/Mandarin/**/*.yaml for any flashcard
// matching word and updates its SM-2 schedule in place. Returns the number of synced cards.
func SyncCompoundToMandarinDecks(baseDecksDir, word string, passed bool, now time.Time) int {
	if baseDecksDir == "" || word == "" {
		return 0
	}
	files, err := deck.GetAllDeckFiles(baseDecksDir, "Mandarin")
	if err != nil || len(files) == 0 {
		return 0
	}
	grade := sm2.GradeWrong
	if passed {
		grade = sm2.GradeGood
	}
	synced := 0
	for _, rel := range files {
		fullPath := filepath.Join(baseDecksDir, "Mandarin", rel)
		d, err := deck.LoadDeck(fullPath)
		if err != nil {
			continue
		}
		modified := false
		for i := range d.Cards {
			if d.Cards[i].Character == word {
				d.Cards[i] = sm2.CalculateSM2(d.Cards[i], grade, now)
				modified = true
				synced++
			}
		}
		if modified {
			_ = deck.SaveDeck(fullPath, d)
		}
	}
	return synced
}
