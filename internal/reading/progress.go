package reading

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// WordMasteryStatus indicates the reader's recall status for a word.
type WordMasteryStatus string

const (
	StatusUnknown  WordMasteryStatus = "unknown"
	StatusLearning WordMasteryStatus = "learning"
	StatusKnown    WordMasteryStatus = "known"
)

// WordProgress tracks user knowledge and encounters for a specific vocabulary word.
type WordProgress struct {
	Target        string            `yaml:"target" json:"target"`
	Language      string            `yaml:"language" json:"language"`
	English       string            `yaml:"english" json:"english"`
	Pinyin        string            `yaml:"pinyin,omitempty" json:"pinyin,omitempty"`
	Status        WordMasteryStatus `yaml:"status" json:"status"`
	RecallSuccess int               `yaml:"recall_success" json:"recall_success"`
	RecallFailure int               `yaml:"recall_failure" json:"recall_failure"`
	LastSeen      time.Time         `yaml:"last_seen" json:"last_seen"`
}

// ReadingProgress persists overall reading state across sessions.
type ReadingProgress struct {
	Aggressiveness int                     `yaml:"aggressiveness" json:"aggressiveness"` // 0 to 4
	TextBookmarks  map[string]int          `yaml:"text_bookmarks" json:"text_bookmarks"` // textID -> sentence index
	Words          map[string]WordProgress `yaml:"words" json:"words"`                   // key: language:target
	FilePath       string                  `yaml:"-" json:"-"`
}

func wordKey(lang, target string) string {
	if lang == "" {
		lang = "Mandarin"
	}
	return lang + ":" + target
}

// DefaultProgress returns initialized empty reading progress.
func DefaultProgress() *ReadingProgress {
	return &ReadingProgress{
		Aggressiveness: 2, // Default to Balanced (40%)
		TextBookmarks:  make(map[string]int),
		Words:          make(map[string]WordProgress),
	}
}

// LoadProgress loads reading progress from the default or specified file.
func LoadProgress(filePath string) (*ReadingProgress, error) {
	if filePath == "" {
		filePath = filepath.Join("readingTranslationTexts", "reading_progress.yaml")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			p := DefaultProgress()
			p.FilePath = filePath
			return p, nil
		}
		return nil, fmt.Errorf("failed to read reading progress: %w", err)
	}

	p := DefaultProgress()
	if err := yaml.Unmarshal(data, p); err != nil {
		return nil, fmt.Errorf("failed to parse reading progress YAML: %w", err)
	}
	if p.TextBookmarks == nil {
		p.TextBookmarks = make(map[string]int)
	}
	if p.Words == nil {
		p.Words = make(map[string]WordProgress)
	}
	if p.Aggressiveness < 0 || p.Aggressiveness > 4 {
		p.Aggressiveness = 2
	}
	p.FilePath = filePath

	return p, nil
}

// Save persists the reading progress to disk.
func (p *ReadingProgress) Save() error {
	if p == nil || p.FilePath == "" {
		return nil
	}

	dir := filepath.Dir(p.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for reading progress: %w", err)
	}

	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal reading progress: %w", err)
	}

	return os.WriteFile(p.FilePath, data, 0644)
}

// GetWordStatus returns the mastery status for a given target word.
func (p *ReadingProgress) GetWordStatus(lang, target string) WordMasteryStatus {
	if p == nil || p.Words == nil {
		return StatusUnknown
	}
	k := wordKey(lang, target)
	wp, exists := p.Words[k]
	if !exists {
		return StatusUnknown
	}
	return wp.Status
}

// RecordWordHelpRequested records that the user did not know or inspected a word.
func (p *ReadingProgress) RecordWordHelpRequested(w WordTranslation) {
	if p == nil {
		return
	}
	if p.Words == nil {
		p.Words = make(map[string]WordProgress)
	}
	k := wordKey(w.Language, w.Target)
	wp, exists := p.Words[k]
	if !exists {
		wp = WordProgress{
			Target:        w.Target,
			Language:      w.Language,
			English:       w.English,
			Pinyin:        w.Pinyin,
			Status:        StatusLearning,
			RecallFailure: 1,
			LastSeen:      time.Now(),
		}
	} else {
		wp.RecallFailure++
		wp.Status = StatusLearning
		wp.LastSeen = time.Now()
	}
	p.Words[k] = wp
}

// RecordWordMastered manually marks a word as known.
func (p *ReadingProgress) RecordWordMastered(w WordTranslation) {
	if p == nil {
		return
	}
	if p.Words == nil {
		p.Words = make(map[string]WordProgress)
	}
	k := wordKey(w.Language, w.Target)
	wp, exists := p.Words[k]
	if !exists {
		wp = WordProgress{
			Target:        w.Target,
			Language:      w.Language,
			English:       w.English,
			Pinyin:        w.Pinyin,
			Status:        StatusKnown,
			RecallSuccess: 3,
			LastSeen:      time.Now(),
		}
	} else {
		wp.Status = StatusKnown
		wp.RecallSuccess += 2
		wp.LastSeen = time.Now()
	}
	p.Words[k] = wp
}

// RecordSentenceSuccess updates mastery for all substituted words in a sentence that the user knew without help.
func (p *ReadingProgress) RecordSentenceSuccess(words []WordTranslation) {
	if p == nil {
		return
	}
	if p.Words == nil {
		p.Words = make(map[string]WordProgress)
	}
	now := time.Now()
	for _, w := range words {
		k := wordKey(w.Language, w.Target)
		wp, exists := p.Words[k]
		if !exists {
			wp = WordProgress{
				Target:        w.Target,
				Language:      w.Language,
				English:       w.English,
				Pinyin:        w.Pinyin,
				Status:        StatusLearning,
				RecallSuccess: 1,
				LastSeen:      now,
			}
		} else {
			wp.RecallSuccess++
			wp.LastSeen = now
			// If recalled successfully 3+ times without failures overtaking, upgrade to known
			if wp.RecallSuccess >= 3 && wp.RecallSuccess >= wp.RecallFailure*2 {
				wp.Status = StatusKnown
			}
		}
		p.Words[k] = wp
	}
}

// SetBookmark updates the bookmark index for a text.
func (p *ReadingProgress) SetBookmark(textID string, sentenceIdx int) {
	if p == nil {
		return
	}
	if p.TextBookmarks == nil {
		p.TextBookmarks = make(map[string]int)
	}
	p.TextBookmarks[textID] = sentenceIdx
}

// GetBookmark retrieves the bookmark index for a text.
func (p *ReadingProgress) GetBookmark(textID string) int {
	if p == nil || p.TextBookmarks == nil {
		return 0
	}
	return p.TextBookmarks[textID]
}
