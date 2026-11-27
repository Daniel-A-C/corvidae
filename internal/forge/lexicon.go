package forge

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

const DefaultLexiconFilename = "mandarin_lexicon.yaml"

type lexiconFile struct {
	Hanzi     []HanziEntry    `yaml:"hanzi"`
	Compounds []CompoundEntry `yaml:"compounds"`
}

// Lexicon holds the indexed Mandarin character and compound graph for Morpheme Forge.
type Lexicon struct {
	Hanzi           []HanziEntry
	Compounds       []CompoundEntry
	CharsByGlyph    map[string]*HanziEntry
	CompoundsByWord map[string]*CompoundEntry
	CompoundsByChar map[string][]*CompoundEntry
}

// LoadLexicon reads and indexes the Mandarin Forge lexicon YAML from path.
func LoadLexicon(path string) (*Lexicon, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read forge lexicon %s: %w", path, err)
	}
	return ParseLexicon(data)
}

// LoadLexiconFromDir loads mandarin_lexicon.yaml from forgeDir.
func LoadLexiconFromDir(forgeDir string) (*Lexicon, error) {
	if forgeDir == "" {
		forgeDir = "forge"
	}
	return LoadLexicon(filepath.Join(forgeDir, DefaultLexiconFilename))
}

// ParseLexicon unmarshals YAML bytes and constructs fast lookup indices.
func ParseLexicon(data []byte) (*Lexicon, error) {
	var raw lexiconFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse forge lexicon YAML: %w", err)
	}
	lex := &Lexicon{
		Hanzi:           raw.Hanzi,
		Compounds:       raw.Compounds,
		CharsByGlyph:    make(map[string]*HanziEntry, len(raw.Hanzi)),
		CompoundsByWord: make(map[string]*CompoundEntry, len(raw.Compounds)),
		CompoundsByChar: make(map[string][]*CompoundEntry, len(raw.Hanzi)),
	}
	for i := range lex.Hanzi {
		h := &lex.Hanzi[i]
		if h.Tone < 1 || h.Tone > 4 {
			h.Tone = 4
		}
		if h.Suit == "" {
			h.Suit = SuitCore
		}
		lex.CharsByGlyph[h.Char] = h
	}
	for i := range lex.Compounds {
		c := &lex.Compounds[i]
		lex.CompoundsByWord[c.Word] = c
		seen := make(map[string]bool)
		for _, r := range c.Word {
			g := string(r)
			if !seen[g] {
				seen[g] = true
				lex.CompoundsByChar[g] = append(lex.CompoundsByChar[g], c)
			}
		}
	}
	return lex, nil
}

// CountNewCompoundsWithDeck reports how many lexicon compounds containing glyph
// can be formed using glyph plus the existing glyphs in deckGlyphs.
func (l *Lexicon) CountNewCompoundsWithDeck(glyph string, deckGlyphs map[string]bool) int {
	if l == nil {
		return 0
	}
	count := 0
	for _, comp := range l.CompoundsByChar[glyph] {
		canForm := true
		for _, r := range comp.Word {
			g := string(r)
			if g != glyph && !deckGlyphs[g] {
				canForm = false
				break
			}
		}
		if canForm {
			count++
		}
	}
	return count
}

// StarterGlyphs returns the initial Hanzi glyphs for the given starter archetype ID.
// For "volatile_scholar", characters that are currently Glass (unmastered) in progress
// are prioritized while keeping high compound connectivity.
func (l *Lexicon) StarterGlyphs(archetypeID string, prog *ProgressStore) []string {
	if l == nil {
		return nil
	}
	if archetypeID == "volatile_scholar" && prog != nil {
		// Pick up to 32 high-degree chars, prioritizing Glass (unmastered/leech) chars.
		type scored struct {
			glyph string
			glass int
			deg   int
		}
		var list []scored
		for _, h := range l.Hanzi {
			g := 0
			if prog.MaterialForChar(h.Char) == MaterialGlass {
				g = 1
			}
			list = append(list, scored{
				glyph: h.Char,
				glass: g,
				deg:   len(l.CompoundsByChar[h.Char]),
			})
		}
		sort.Slice(list, func(i, j int) bool {
			// Balance connectivity with Glass priority: degree >= 4 Glass first, then highest degree.
			iGood := list[i].glass == 1 && list[i].deg >= 4
			jGood := list[j].glass == 1 && list[j].deg >= 4
			if iGood != jGood {
				return iGood
			}
			if list[i].deg != list[j].deg {
				return list[i].deg > list[j].deg
			}
			return list[i].glyph < list[j].glyph
		})
		n := 32
		if len(list) < n {
			n = len(list)
		}
		out := make([]string, 0, n)
		for i := 0; i < n; i++ {
			out = append(out, list[i].glyph)
		}
		return out
	}

	var glyphs []string
	for _, h := range l.Hanzi {
		for _, sd := range h.StarterDecks {
			if sd == archetypeID {
				glyphs = append(glyphs, h.Char)
				break
			}
		}
	}
	if len(glyphs) == 0 {
		// Fallback: take top 30 highest-degree characters.
		type degItem struct {
			g string
			d int
		}
		var all []degItem
		for _, h := range l.Hanzi {
			all = append(all, degItem{h.Char, len(l.CompoundsByChar[h.Char])})
		}
		sort.Slice(all, func(i, j int) bool {
			if all[i].d != all[j].d {
				return all[i].d > all[j].d
			}
			return all[i].g < all[j].g
		})
		n := 30
		if len(all) < n {
			n = len(all)
		}
		for i := 0; i < n; i++ {
			glyphs = append(glyphs, all[i].g)
		}
	}
	return glyphs
}
