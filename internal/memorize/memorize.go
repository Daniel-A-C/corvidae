// Package memorize provides word-by-word active recall memorization text processing, distractor generation, and portion selection.
package memorize

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	reVerseNumber = regexp.MustCompile(`\b(\d+)\b`)
	reVersePrefix = regexp.MustCompile(`(\b\d+\s+)`)
	reCleanWord   = regexp.MustCompile(`^[¡¿"']+|[.,;:!?"']+$`)
	rePunctAffix  = regexp.MustCompile(`^([¡¿"']*)(.*?)([.,;:!?"']*)$`)
)

// WordItem represents an individual word in a memorization text with its options.
type WordItem struct {
	Word    string   `yaml:"word" json:"word"`
	Options []string `yaml:"options" json:"options"`
	Prefix  string   `yaml:"prefix,omitempty" json:"prefix,omitempty"`
}

// Text represents a full text to be memorized, with metadata and items.
type Text struct {
	Title       string     `yaml:"title" json:"title"`
	Language    string     `yaml:"language,omitempty" json:"language,omitempty"`
	Description string     `yaml:"description,omitempty" json:"description,omitempty"`
	Progress    int        `yaml:"progress,omitempty" json:"progress,omitempty"`
	BestStreak  int        `yaml:"best_streak,omitempty" json:"best_streak,omitempty"`
	Mistakes    int        `yaml:"mistakes,omitempty" json:"mistakes,omitempty"`
	Words       []WordItem `yaml:"words" json:"words"`
	FilePath    string     `yaml:"-" json:"-"`
}

// TextHeader provides summary information for listing available memorization texts.
type TextHeader struct {
	Title     string
	Filename  string
	FullPath  string
	WordCount int
	Progress  int
	Language  string
}

// Section represents a continuous portion or unit of a text (e.g. verse or paragraph).
type Section struct {
	Index     int    `json:"index"`
	Label     string `json:"label"`     // e.g. "Verse 1" or "Section 1"
	Number    string `json:"number"`    // e.g. "1", "15"
	StartIdx  int    `json:"start_idx"` // inclusive word index in Words
	EndIdx    int    `json:"end_idx"`   // exclusive word index in Words
	WordCount int    `json:"word_count"`
	Preview   string `json:"preview"`   // first few words preview
}

// GetSections partitions a Text into logical continuous sections based on verse numbers, paragraph breaks, or chunks.
func (t *Text) GetSections() []Section {
	if len(t.Words) == 0 {
		return nil
	}

	type marker struct {
		startIdx int
		number   string
		label    string
	}
	var markers []marker

	for i, w := range t.Words {
		if reVerseNumber.MatchString(w.Prefix) {
			m := reVerseNumber.FindStringSubmatch(w.Prefix)
			num := m[1]
			markers = append(markers, marker{
				startIdx: i,
				number:   num,
				label:    "Verse " + num,
			})
		}
	}

	// If no verse numbers were detected, check for paragraph breaks
	if len(markers) == 0 {
		for i, w := range t.Words {
			if i == 0 {
				markers = append(markers, marker{
					startIdx: 0,
					number:   "1",
					label:    "Part 1",
				})
			} else if strings.Contains(w.Prefix, "\n") {
				partNum := fmt.Sprintf("%d", len(markers)+1)
				markers = append(markers, marker{
					startIdx: i,
					number:   partNum,
					label:    "Part " + partNum,
				})
			}
		}
	}

	// If still only 0 or 1 marker and text is long, divide into chunks of ~20 words
	if len(markers) <= 1 && len(t.Words) > 25 {
		markers = nil
		chunkSize := 20
		for i := 0; i < len(t.Words); i += chunkSize {
			partNum := fmt.Sprintf("%d", len(markers)+1)
			markers = append(markers, marker{
				startIdx: i,
				number:   partNum,
				label:    fmt.Sprintf("Part %s", partNum),
			})
		}
	}

	// If still no markers, one single section
	if len(markers) == 0 {
		markers = append(markers, marker{
			startIdx: 0,
			number:   "1",
			label:    "Full Text",
		})
	}

	var sections []Section
	for i, m := range markers {
		endIdx := len(t.Words)
		if i+1 < len(markers) {
			endIdx = markers[i+1].startIdx
		}

		// Generate preview
		var previewWords []string
		for j := m.startIdx; j < endIdx && j < m.startIdx+6; j++ {
			previewWords = append(previewWords, t.Words[j].Word)
		}
		preview := strings.Join(previewWords, " ")
		if endIdx-m.startIdx > 6 {
			preview += "..."
		}

		sections = append(sections, Section{
			Index:     i,
			Label:     m.label,
			Number:    m.number,
			StartIdx:  m.startIdx,
			EndIdx:    endIdx,
			WordCount: endIdx - m.startIdx,
			Preview:   preview,
		})
	}

	return sections
}

// ResolveTextsDir determines the effective directory for memorization texts.
// It checks the primary dir, and if not existing or empty, checks common fallbacks.
func ResolveTextsDir(primaryDir string) string {
	if primaryDir == "" {
		primaryDir = "memorizationTexts"
	}
	if info, err := os.Stat(primaryDir); err == nil && info.IsDir() {
		return primaryDir
	}
	// Fallback check
	fallbacks := []string{"memorizationTexts", "memory"}
	for _, fb := range fallbacks {
		if info, err := os.Stat(fb); err == nil && info.IsDir() {
			return fb
		}
	}
	return primaryDir
}

// ListTexts scans the given directory for memorization texts (.yaml, .json, .txt).
// If a raw .txt file is found without a corresponding .yaml, it auto-generates the .yaml file.
func ListTexts(dir string) ([]TextHeader, error) {
	effectiveDir := ResolveTextsDir(dir)
	entries, err := os.ReadDir(effectiveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []TextHeader{}, nil
		}
		return nil, err
	}

	// Group files by base stem
	fileMap := make(map[string][]string)
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".yaml" || ext == ".yml" || ext == ".json" || ext == ".txt" {
			stem := strings.TrimSuffix(name, ext)
			fileMap[stem] = append(fileMap[stem], name)
		}
	}

	var headers []TextHeader
	var stems []string
	for stem := range fileMap {
		stems = append(stems, stem)
	}
	sort.Strings(stems)

	for _, stem := range stems {
		files := fileMap[stem]
		// Choose preferred file: .yaml/.yml > .json > .txt
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
			// Auto-generate .yaml from .txt if not present
			txtPath := filepath.Join(effectiveDir, txtFile)
			yamlPath := filepath.Join(effectiveDir, stem+".yaml")
			t, err := GenerateFromTextFile(txtPath, stem)
			if err == nil {
				_ = SaveText(yamlPath, t)
				chosenFile = stem + ".yaml"
			} else {
				chosenFile = txtFile
			}
		} else if !hasYaml && chosenFile == "" && len(files) > 0 {
			chosenFile = files[0]
		}

		fullPath := filepath.Join(effectiveDir, chosenFile)
		loaded, err := LoadText(fullPath)
		if err != nil {
			headers = append(headers, TextHeader{
				Title:    stem,
				Filename: chosenFile,
				FullPath: fullPath,
			})
			continue
		}

		title := loaded.Title
		if title == "" {
			title = stem
		}

		headers = append(headers, TextHeader{
			Title:     title,
			Filename:  chosenFile,
			FullPath:  fullPath,
			WordCount: len(loaded.Words),
			Progress:  loaded.Progress,
			Language:  loaded.Language,
		})
	}

	return headers, nil
}

// LoadText reads and parses a memorization text file from disk.
func LoadText(filePath string) (*Text, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var t Text
	if ext == ".txt" {
		stem := strings.TrimSuffix(filepath.Base(filePath), ext)
		gen, err := GenerateFromPlainText(string(data), stem)
		if err != nil {
			return nil, err
		}
		gen.FilePath = filePath
		return gen, nil
	}

	err = yaml.Unmarshal(data, &t)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}
	t.FilePath = filePath
	return &t, nil
}

// SaveText serializes a Text struct and writes it to disk, creating parent directories if needed.
func SaveText(filePath string, t *Text) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(t)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// GenerateFromTextFile reads a plain text file and converts it into a structured Text.
func GenerateFromTextFile(txtPath string, defaultTitle string) (*Text, error) {
	data, err := os.ReadFile(txtPath)
	if err != nil {
		return nil, err
	}
	return GenerateFromPlainText(string(data), defaultTitle)
}

// GenerateFromPlainText parses raw text and creates WordItems with plausible distractors.
func GenerateFromPlainText(raw string, defaultTitle string) (*Text, error) {
	lines := strings.Split(raw, "\n")
	type parsedVerse struct {
		num  string
		text string
	}
	var verses []parsedVerse

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := reVersePrefix.Split(line, -1)
		matches := reVersePrefix.FindAllString(line, -1)

		if len(matches) > 0 {
			for i, m := range matches {
				vNum := strings.TrimSpace(m)
				vText := ""
				if i+1 < len(parts) {
					vText = strings.TrimSpace(parts[i+1])
				}
				if vText != "" {
					verses = append(verses, parsedVerse{num: vNum, text: vText})
				}
			}
		} else {
			verses = append(verses, parsedVerse{num: "", text: line})
		}
	}

	if len(verses) == 0 {
		verses = append(verses, parsedVerse{num: "", text: strings.TrimSpace(raw)})
	}

	// Extract unique vocabulary
	var allWords []string
	for _, v := range verses {
		for _, w := range strings.Fields(v.text) {
			c := strings.ToLower(reCleanWord.ReplaceAllString(w, ""))
			if c != "" {
				allWords = append(allWords, c)
			}
		}
	}

	uniqueVocab := make(map[string]bool)
	var vocabList []string
	for _, w := range allWords {
		if !uniqueVocab[w] {
			uniqueVocab[w] = true
			vocabList = append(vocabList, w)
		}
	}

	var items []WordItem
	verseCount := 0
	for _, v := range verses {
		words := strings.Fields(v.text)
		for wIdx, w := range words {
			prefix := ""
			if wIdx == 0 && v.num != "" {
				if verseCount > 0 {
					prefix = "\n\n" + v.num + " "
				} else {
					prefix = v.num + " "
				}
			} else if wIdx == 0 && verseCount > 0 && v.num == "" {
				prefix = "\n\n"
			}

			opts := generateFallbackDistractors(w, vocabList)
			items = append(items, WordItem{
				Word:    w,
				Options: opts,
				Prefix:  prefix,
			})
		}
		verseCount++
	}

	title := defaultTitle
	if title == "" {
		title = "Memorization Text"
	}

	return &Text{
		Title: title,
		Words: items,
	}, nil
}

func generateFallbackDistractors(target string, vocab []string) []string {
	m := rePunctAffix.FindStringSubmatch(target)
	leading := ""
	core := target
	trailing := ""
	if len(m) == 4 {
		leading = m[1]
		core = m[2]
		trailing = m[3]
	}

	isCap := len(core) > 0 && strings.ToUpper(string(core[0])) == string(core[0])
	base := strings.ToLower(core)

	// Distractor selection based on length similarity
	var pool []string
	for _, v := range vocab {
		if v != base && abs(len(v)-len(base)) <= 2 {
			pool = append(pool, v)
		}
	}
	if len(pool) < 4 {
		for _, v := range vocab {
			if v != base && !contains(pool, v) {
				pool = append(pool, v)
			}
		}
	}

	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

	var distractors []string
	for _, p := range pool {
		if len(distractors) >= 4 {
			break
		}
		formatted := p
		if isCap && len(formatted) > 0 {
			formatted = strings.ToUpper(formatted[:1]) + formatted[1:]
		}
		formatted = leading + formatted + trailing
		if formatted != target && !contains(distractors, formatted) {
			distractors = append(distractors, formatted)
		}
	}

	fallbackWord := "otro"
	if isCap {
		fallbackWord = "Otro"
	}
	for len(distractors) < 4 {
		cand := leading + fallbackWord + trailing
		distractors = append(distractors, cand)
	}

	options := append([]string{target}, distractors...)
	return options
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// ShuffleOptions prepares a randomized slice of options for the given item,
// and returns the shuffled options along with the 0-based index of the correct option.
func ShuffleOptions(item WordItem) ([]string, int) {
	opts := make([]string, len(item.Options))
	copy(opts, item.Options)

	// Ensure target word is in options
	found := false
	for _, o := range opts {
		if o == item.Word {
			found = true
			break
		}
	}
	if !found {
		if len(opts) == 0 {
			opts = []string{item.Word}
		} else {
			opts[0] = item.Word
		}
	}

	rand.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })

	correctIdx := 0
	for i, o := range opts {
		if o == item.Word {
			correctIdx = i
			break
		}
	}
	return opts, correctIdx
}
