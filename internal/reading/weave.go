package reading

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
)

// Available selection keys for words in a sentence that avoid collisions with navigation keys.
var WordSelectionKeys = []string{
	"a", "s", "d", "f", "g",
	"z", "x", "c", "v",
	"1", "2", "3", "4", "5", "6", "7", "8", "9",
}

// AggressivenessLabels provides descriptive names for each substitution level.
var AggressivenessLabels = []string{
	"0% - Pure Reading (English Only)",
	"20% - Gentle (Review + 1 New)",
	"40% - Balanced (Moderate Immersion)",
	"70% - Intensive (High Vocabulary)",
	"100% - Full Immersion (All Words)",
}

// WeaveSentence performs Diglot Weave substitution on a Sentence given user progress and aggressiveness level.
func WeaveSentence(sentence Sentence, progress *ReadingProgress, aggressiveness int) WeavedSentence {
	if aggressiveness < 0 {
		aggressiveness = 0
	}
	if aggressiveness > 4 {
		aggressiveness = 4
	}

	result := WeavedSentence{
		OriginalSentence: sentence,
		Aggressiveness:   aggressiveness,
		KeyToWord:        make(map[string]SubstitutedWord),
	}

	// Aggressiveness 0: Pure English reading mode (no substitutions)
	if aggressiveness == 0 || len(sentence.Words) == 0 {
		result.Segments = []Segment{
			{
				IsSubstituted: false,
				Text:          sentence.English,
			},
		}
		return result
	}

	// 1. Determine how many words to substitute based on aggressiveness level
	totalAvailable := len(sentence.Words)
	var maxToSubstitute int

	switch aggressiveness {
	case 1: // Gentle: 20-25% (min 1, max 2)
		maxToSubstitute = int(math.Ceil(float64(totalAvailable) * 0.25))
		if maxToSubstitute < 1 {
			maxToSubstitute = 1
		}
	case 2: // Moderate: ~40-50%
		maxToSubstitute = int(math.Ceil(float64(totalAvailable) * 0.45))
		if maxToSubstitute < 1 {
			maxToSubstitute = 1
		}
	case 3: // Intensive: ~70-75%
		maxToSubstitute = int(math.Ceil(float64(totalAvailable) * 0.75))
		if maxToSubstitute < 1 {
			maxToSubstitute = 1
		}
	case 4: // Full: 100%
		maxToSubstitute = totalAvailable
	}

	// 2. Score and sort candidates:
	// Prioritize words the user already knows for gentle levels; introduce newer words as aggressiveness rises.
	type candidate struct {
		word     WordTranslation
		priority int
	}

	var candidates []candidate
	for _, w := range sentence.Words {
		status := StatusUnknown
		if progress != nil {
			status = progress.GetWordStatus(w.Language, w.Target)
		}

		prio := 10
		switch status {
		case StatusKnown:
			// In gentle modes, known words are preferred first for reinforcement
			prio = 50
		case StatusLearning:
			prio = 30
		case StatusUnknown:
			prio = 10
		}

		// Adjust priority for intensive modes (introduce unknowns more readily)
		if aggressiveness >= 3 {
			if status == StatusLearning {
				prio = 55
			} else if status == StatusUnknown {
				prio = 40
			}
		}

		candidates = append(candidates, candidate{
			word:     w,
			priority: prio,
		})
	}

	// Stable sort by priority descending, then by length of English phrase descending
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].priority != candidates[j].priority {
			return candidates[i].priority > candidates[j].priority
		}
		return len(candidates[i].word.English) > len(candidates[j].word.English)
	})

	if maxToSubstitute > len(candidates) {
		maxToSubstitute = len(candidates)
	}
	selectedWords := make(map[string]WordTranslation)
	for i := 0; i < maxToSubstitute; i++ {
		w := candidates[i].word
		selectedWords[strings.ToLower(w.English)] = w
	}

	// 3. Find non-overlapping occurrences of selected words in the English sentence
	type matchOccur struct {
		start int
		end   int
		orig  string
		trans WordTranslation
	}

	var matches []matchOccur

	// Sort selected words by English text length descending so longer phrases match first
	var phraseList []WordTranslation
	for _, w := range selectedWords {
		phraseList = append(phraseList, w)
	}
	sort.Slice(phraseList, func(i, j int) bool {
		return len(phraseList[i].English) > len(phraseList[j].English)
	})

	occupied := make([]bool, len(sentence.English))

	for _, w := range phraseList {
		// Use word boundary regex matching case-insensitively
		pattern := `(?i)\b` + regexp.QuoteMeta(w.English) + `\b`
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		locs := re.FindAllStringIndex(sentence.English, -1)
		for _, loc := range locs {
			start, end := loc[0], loc[1]
			// Check overlap
			overlaps := false
			for k := start; k < end; k++ {
				if occupied[k] {
					overlaps = true
					break
				}
			}
			if !overlaps {
				for k := start; k < end; k++ {
					occupied[k] = true
				}
				matches = append(matches, matchOccur{
					start: start,
					end:   end,
					orig:  sentence.English[start:end],
					trans: w,
				})
			}
		}
	}

	// Sort matches by start position in sentence
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].start < matches[j].start
	})

	// 4. Assemble Segments and assign keys
	var segments []Segment
	var substitutedList []SubstitutedWord
	cursor := 0
	keyIdx := 0

	for _, m := range matches {
		if m.start > cursor {
			segments = append(segments, Segment{
				IsSubstituted: false,
				Text:          sentence.English[cursor:m.start],
			})
		}

		key := ""
		if keyIdx < len(WordSelectionKeys) {
			key = WordSelectionKeys[keyIdx]
			keyIdx++
		} else {
			key = fmt.Sprintf("%d", keyIdx+1)
			keyIdx++
		}

		subWord := SubstitutedWord{
			Key:         key,
			Number:      keyIdx,
			Original:    m.orig,
			Translation: m.trans,
		}

		segments = append(segments, Segment{
			IsSubstituted:   true,
			Text:            m.trans.Target,
			SubstitutedInfo: subWord,
		})

		substitutedList = append(substitutedList, subWord)
		result.KeyToWord[key] = subWord
		cursor = m.end
	}

	if cursor < len(sentence.English) {
		segments = append(segments, Segment{
			IsSubstituted: false,
			Text:          sentence.English[cursor:],
		})
	}

	result.Segments = segments
	result.SubstitutedWords = substitutedList
	return result
}
