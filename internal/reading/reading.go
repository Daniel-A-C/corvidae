// Package reading provides Diglot Weave immersion reading, progressive vocabulary substitution, grammar discussions, and mastery tracking.
package reading

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var reSentenceSplit = regexp.MustCompile(`[^.!?—\n]+[.!?—\n]+`)

// WordTranslation represents the translation and grammatical info for a word or phrase.
type WordTranslation struct {
	English       string `yaml:"english" json:"english"`
	Target        string `yaml:"target" json:"target"`
	Pinyin        string `yaml:"pinyin,omitempty" json:"pinyin,omitempty"`
	Pronunciation string `yaml:"pronunciation,omitempty" json:"pronunciation,omitempty"`
	Meaning       string `yaml:"meaning" json:"meaning"`
	Explanation   string `yaml:"explanation,omitempty" json:"explanation,omitempty"`
	Language      string `yaml:"language,omitempty" json:"language,omitempty"`
	Difficulty    int    `yaml:"difficulty,omitempty" json:"difficulty,omitempty"` // 1 (beginner) to 5 (advanced)
}

// Sentence represents a single sentence with natural target translation and grammar contrast notes.
type Sentence struct {
	ID            string            `yaml:"id" json:"id"`
	English       string            `yaml:"english" json:"english"`
	NaturalTarget string            `yaml:"natural_target" json:"natural_target"`
	TargetPinyin  string            `yaml:"target_pinyin,omitempty" json:"target_pinyin,omitempty"`
	GrammarNote   string            `yaml:"grammar_note,omitempty" json:"grammar_note,omitempty"`
	Words         []WordTranslation `yaml:"words" json:"words"`
}

// Text represents a full reading text with sentences and metadata.
type Text struct {
	ID          string     `yaml:"id" json:"id"`
	Title       string     `yaml:"title" json:"title"`
	Source      string     `yaml:"source,omitempty" json:"source,omitempty"`
	Language    string     `yaml:"language" json:"language"`
	Description string     `yaml:"description,omitempty" json:"description,omitempty"`
	Category    string     `yaml:"category,omitempty" json:"category,omitempty"`
	Sentences   []Sentence `yaml:"sentences" json:"sentences"`
	FilePath    string     `yaml:"-" json:"-"`
}

// TextHeader provides summary information for listing available reading texts.
type TextHeader struct {
	ID            string
	Title         string
	Filename      string
	FullPath      string
	Language      string
	Category      string
	SentenceCount int
	Progress      int // index of current sentence
}

// SubstitutedWord represents a word substituted in a sentence with its assigned shortcut key.
type SubstitutedWord struct {
	Key         string
	Number      int
	Original    string
	Translation WordTranslation
}

// Segment represents a piece of the rendered sentence (either plain text or substituted word).
type Segment struct {
	IsSubstituted   bool
	Text            string
	SubstitutedInfo SubstitutedWord
}

// WeavedSentence represents a sentence that has had words substituted based on aggressiveness.
type WeavedSentence struct {
	OriginalSentence Sentence
	Aggressiveness   int
	Segments         []Segment
	SubstitutedWords []SubstitutedWord
	KeyToWord        map[string]SubstitutedWord
}

// ResolveReadingDir returns the effective directory for reading texts.
func ResolveReadingDir(primaryDir string) string {
	if primaryDir == "" {
		primaryDir = "readingTranslationTexts"
	}
	if info, err := os.Stat(primaryDir); err == nil && info.IsDir() {
		return primaryDir
	}
	fallbacks := []string{"readingTranslationTexts", "readingTexts", "texts"}
	for _, fb := range fallbacks {
		if info, err := os.Stat(fb); err == nil && info.IsDir() {
			return fb
		}
	}
	return primaryDir
}

// LoadText loads a Text from a YAML or JSON file.
func LoadText(filePath string) (*Text, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read text file %s: %w", filePath, err)
	}

	var text Text
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".json" {
		if err := json.Unmarshal(data, &text); err != nil {
			return nil, fmt.Errorf("failed to parse JSON from %s: %w", filePath, err)
		}
	} else {
		if err := yaml.Unmarshal(data, &text); err != nil {
			return nil, fmt.Errorf("failed to parse YAML from %s: %w", filePath, err)
		}
	}

	if text.ID == "" {
		stem := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		text.ID = stem
	}
	if text.Language == "" {
		text.Language = "Mandarin"
	}
	text.FilePath = filePath

	return &text, nil
}

// SaveText writes a Text structure out to a YAML file.
func SaveText(filePath string, text *Text) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(text)
	if err != nil {
		return fmt.Errorf("failed to marshal text to YAML: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// ListTexts scans the given directory for available reading texts.
func ListTexts(dir string, progress *ReadingProgress) ([]TextHeader, error) {
	effectiveDir := ResolveReadingDir(dir)
	entries, err := os.ReadDir(effectiveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []TextHeader{}, nil
		}
		return nil, err
	}

	fileMap := make(map[string][]string)
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		name := entry.Name()
		if name == "reading_progress.yaml" || name == "reading_progress.json" {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".yaml" || ext == ".yml" || ext == ".json" || ext == ".txt" {
			stem := strings.TrimSuffix(name, ext)
			fileMap[stem] = append(fileMap[stem], name)
		}
	}

	var stems []string
	for stem := range fileMap {
		stems = append(stems, stem)
	}
	sort.Strings(stems)

	var headers []TextHeader
	for _, stem := range stems {
		files := fileMap[stem]
		var chosenFile string
		var hasYaml bool
		var txtFile string

		for _, f := range files {
			ext := strings.ToLower(filepath.Ext(f))
			if ext == ".yaml" || ext == ".yml" {
				chosenFile = f
				hasYaml = true
				break
			} else if ext == ".json" && chosenFile == "" {
				chosenFile = f
			} else if ext == ".txt" {
				txtFile = f
			}
		}

		if chosenFile == "" && txtFile != "" {
			// Plain text without yaml - can be parsed on load
			chosenFile = txtFile
		} else if !hasYaml && chosenFile == "" && len(files) > 0 {
			chosenFile = files[0]
		}

		fullPath := filepath.Join(effectiveDir, chosenFile)
		var title = stem
		var lang = "Mandarin"
		var category = "Reading"
		var sentenceCount = 0

		if strings.HasSuffix(chosenFile, ".yaml") || strings.HasSuffix(chosenFile, ".yml") || strings.HasSuffix(chosenFile, ".json") {
			t, err := LoadText(fullPath)
			if err == nil {
				if t.Title != "" {
					title = t.Title
				}
				if t.Language != "" {
					lang = t.Language
				}
				if t.Category != "" {
					category = t.Category
				}
				sentenceCount = len(t.Sentences)
			}
		} else if strings.HasSuffix(chosenFile, ".txt") {
			// Estimate sentences from raw text
			raw, _ := os.ReadFile(fullPath)
			sentences := SplitIntoSentences(string(raw))
			sentenceCount = len(sentences)
		}

		var curProgress int
		if progress != nil && progress.TextBookmarks != nil {
			curProgress = progress.TextBookmarks[stem]
		}

		headers = append(headers, TextHeader{
			ID:            stem,
			Title:         title,
			Filename:      chosenFile,
			FullPath:      fullPath,
			Language:      lang,
			Category:      category,
			SentenceCount: sentenceCount,
			Progress:      curProgress,
		})
	}

	return headers, nil
}

// SplitIntoSentences splits a block of English text into individual sentences.
func SplitIntoSentences(text string) []string {
	// Normalize line breaks
	text = strings.ReplaceAll(text, "\r\n", "\n")
	paras := strings.Split(text, "\n\n")

	var result []string

	for _, p := range paras {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		matches := reSentenceSplit.FindAllString(p, -1)
		if len(matches) == 0 {
			result = append(result, p)
		} else {
			var remainder = p
			for _, m := range matches {
				mTrim := strings.TrimSpace(m)
				if mTrim != "" {
					result = append(result, mTrim)
				}
				remainder = strings.TrimPrefix(remainder, m)
			}
			remTrim := strings.TrimSpace(remainder)
			if remTrim != "" {
				result = append(result, remTrim)
			}
		}
	}

	return result
}
