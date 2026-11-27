// Package forge implements the Morpheme Forge Mandarin roguelike game engine.
package forge

import "fmt"

// RadicalSuit represents the semantic radical family (Suit) of a Hanzi card.
type RadicalSuit string

const (
	SuitNature RadicalSuit = "nature"
	SuitHuman  RadicalSuit = "human"
	SuitCraft  RadicalSuit = "craft"
	SuitCore   RadicalSuit = "core"
)

// Icon returns a compact terminal symbol for the RadicalSuit.
func (s RadicalSuit) Icon() string {
	switch s {
	case SuitNature:
		return "水"
	case SuitHuman:
		return "人"
	case SuitCraft:
		return "工"
	default:
		return "元"
	}
}

// Label returns a human-readable label for the RadicalSuit.
func (s RadicalSuit) Label() string {
	switch s {
	case SuitNature:
		return "Nature"
	case SuitHuman:
		return "Human"
	case SuitCraft:
		return "Craft"
	default:
		return "Core"
	}
}

// MaterialTier represents a card's in-game physics derived from its real SM-2 memory state.
type MaterialTier int

const (
	MaterialGlass MaterialTier = iota // New or Leech (reps == 0 or ease < 1.8): high xMult, shatters on miss
	MaterialFoil                      // Learning (reps >= 1, interval < 21): steady Chips + flat Mult
	MaterialGold                      // Mature (interval >= 21, ease >= 1.8): reliable Chips + coins
)

// Label returns the display name for a MaterialTier.
func (m MaterialTier) Label() string {
	switch m {
	case MaterialGlass:
		return "Glass"
	case MaterialFoil:
		return "Foil"
	case MaterialGold:
		return "Gold"
	default:
		return "Glass"
	}
}

// Badge returns a compact 1-rune symbol for card tiles.
func (m MaterialTier) Badge() string {
	switch m {
	case MaterialGlass:
		return "⚡"
	case MaterialFoil:
		return "◆"
	case MaterialGold:
		return "★"
	default:
		return "⚡"
	}
}

// CardEdition represents an optional run-scoped shop enhancement on a DeckCard.
type CardEdition int

const (
	EditionNone CardEdition = iota
	EditionFoil             // +30 Chips
	EditionHolo             // +6 Mult
	EditionPoly             // x1.5 Mult
)

// Label returns a short display label for a CardEdition.
func (e CardEdition) Label() string {
	switch e {
	case EditionFoil:
		return "Foil+"
	case EditionHolo:
		return "Holo"
	case EditionPoly:
		return "Poly"
	default:
		return ""
	}
}

// HanziEntry defines a single Mandarin character in the Forge lexicon.
type HanziEntry struct {
	Char         string      `yaml:"char"`
	Pinyin       string      `yaml:"pinyin"`
	Tone         int         `yaml:"tone"`
	Meaning      string      `yaml:"meaning"`
	Suit         RadicalSuit `yaml:"suit"`
	Standalone   bool        `yaml:"standalone"`
	StarterDecks []string    `yaml:"starter_decks,omitempty"`
}

// CompoundEntry defines a multi-character Mandarin word in the Forge lexicon.
type CompoundEntry struct {
	Word    string `yaml:"word"`
	Pinyin  string `yaml:"pinyin"`
	Tones   []int  `yaml:"tones"`
	Meaning string `yaml:"meaning"`
	Domain  string `yaml:"domain,omitempty"`
}

// DeckCard is an instance of a Hanzi character inside a player's active run deck.
type DeckCard struct {
	ID      int
	Glyph   string
	Edition CardEdition
}

// HandType classifies the forged lexical structure played in a turn.
type HandType int

const (
	HandNone HandType = iota
	HandSolo
	HandCompound
	HandDoubleForge
	HandTriad
	HandPivotChain
	HandGrandChain
	HandIdiom
)

// AllHandTypes lists all upgradeable hand types in display order.
var AllHandTypes = []HandType{
	HandSolo,
	HandCompound,
	HandDoubleForge,
	HandTriad,
	HandPivotChain,
	HandGrandChain,
	HandIdiom,
}

// Name returns the display name of the HandType.
func (h HandType) Name() string {
	switch h {
	case HandSolo:
		return "Solo"
	case HandCompound:
		return "Compound"
	case HandDoubleForge:
		return "Double Forge"
	case HandTriad:
		return "Triad"
	case HandPivotChain:
		return "Pivot Chain"
	case HandGrandChain:
		return "Grand Chain"
	case HandIdiom:
		return "Quartet / Idiom"
	default:
		return "Fizzle"
	}
}

// BaseStats returns (baseChips, baseMult, chipsPerLevel, multPerLevel) for a HandType.
func (h HandType) BaseStats() (int, float64, int, float64) {
	switch h {
	case HandSolo:
		return 15, 1, 10, 1
	case HandCompound:
		return 35, 2, 15, 1
	case HandDoubleForge:
		return 65, 4, 20, 2
	case HandTriad:
		return 80, 5, 25, 2
	case HandPivotChain:
		return 100, 6, 30, 3
	case HandGrandChain:
		return 140, 8, 40, 3
	case HandIdiom:
		return 160, 9, 45, 4
	default:
		return 0, 0, 0, 0
	}
}

// StatsAtLevel returns (chips, mult) for a HandType at 1-based level.
func (h HandType) StatsAtLevel(level int) (int, float64) {
	if level < 1 {
		level = 1
	}
	bc, bm, uc, um := h.BaseStats()
	return bc + (level-1)*uc, bm + float64(level-1)*um
}

// ProveMode specifies what the home-row recall check tests.
type ProveMode int

const (
	ProveMeaning ProveMode = iota
	ProveTone
	ProvePinyin
)

// Label returns a short prompt header for the ProveMode.
func (p ProveMode) Label() string {
	switch p {
	case ProveTone:
		return "Identify Tone Pattern"
	case ProvePinyin:
		return "Select Pinyin"
	default:
		return "Select Meaning"
	}
}

// StagedWord represents one word slot being assembled by the player from their hand.
type StagedWord struct {
	CardIndices []int // Indices into Run.Hand
	IsPivot     bool  // True if its first card bridges from the previous StagedWord's last card
}

// ProveChallenge is a single home-row multiple-choice verification for a valid forged word.
type ProveChallenge struct {
	WordSlotIdx        int
	Word               string
	Pinyin             string
	Meaning            string
	Tones              []int
	Mode               ProveMode
	Options            []string
	CorrectIdx         int
	IsVolatileCompound bool
	CardIndices        []int
	IsPivot            bool
	IsSolo             bool
}

// WordOutcome summarizes the validation and prove result of one staged word.
type WordOutcome struct {
	Word           string
	Pinyin         string
	Meaning        string
	Valid          bool
	Passed         bool
	GlassShattered bool
	ShieldSaved    bool
	CoinsEarned    int
	Notes          []string
}

// ScoreEvent is a single line item in the Chips × Mult scoring cascade.
type ScoreEvent struct {
	Label  string
	Detail string
}

// MissedHint represents a strong compound or Pivot Chain that was available in the player's
// hand on the just-completed turn, shown after the turn for learning.
type MissedHint struct {
	IsPivot bool
	Label   string // e.g. "发现 (fā xiàn)" or "回来 ⇄ 来到"
	Detail  string // e.g. "to discover" or "to return ⇄ to arrive"
}

// PlayResult holds the complete scoring breakdown of a committed hand.
type PlayResult struct {
	HandType       HandType
	HandLevel      int
	RadicalFlush   bool
	ToneFlush      bool
	ToneEscalation bool
	Chips          int
	Mult           float64
	TotalScore     int
	CoinsEarned    int
	GlassShattered bool
	Words          []WordOutcome
	Events         []ScoreEvent
	MissedHints    []MissedHint
}

// StarterArchetype describes one of the selectable starting decks for a Forge run.
type StarterArchetype struct {
	ID      string
	Name    string
	Desc    string
	Passive string
}

// StarterArchetypes lists all 4 starter decks in display order.
var StarterArchetypes = []StarterArchetype{
	{
		ID:      "harry_potter",
		Name:    "Hogwarts Core",
		Desc:    "30 high-connectivity Harry Potter hub Hanzi",
		Passive: "+1 Hand per Blind",
	},
	{
		ID:      "john_gospel",
		Name:    "Gospel of John",
		Desc:    "30 Gospel of John & shared literary Hanzi",
		Passive: "Radical Flush bonus starts upgraded (+30 Chips)",
	},
	{
		ID:      "mind_and_people",
		Name:    "Human & Mind",
		Desc:    "30 Human & Core radical Hanzi across both texts",
		Passive: "+1 Discard per Blind",
	},
	{
		ID:      "volatile_scholar",
		Name:    "Volatile Scholar",
		Desc:    "32 Hanzi biased toward unmastered/Glass cards",
		Passive: "Starts with Forgiving Tutor relic + $4",
	},
}

// RelicID identifies a passive roguelike modifier (Joker).
type RelicID string

// RelicDef describes a Relic's display info and shop cost.
type RelicDef struct {
	ID     RelicID
	Name   string
	Desc   string
	Cost   int
	Rarity string // "Common", "Uncommon", "Rare"
}

// BossBlindID identifies a Boss Blind rule modifier.
type BossBlindID string

// BossBlindDef describes a Boss Blind's name and rule-bending constraint.
type BossBlindDef struct {
	ID   BossBlindID
	Name string
	Desc string
}

// BlindKind distinguishes Small, Big, and Boss blinds within an Ante.
type BlindKind int

const (
	BlindSmall BlindKind = iota
	BlindBig
	BlindBoss
)

// Name returns the display title for a BlindKind.
func (b BlindKind) Name() string {
	switch b {
	case BlindSmall:
		return "Small Blind"
	case BlindBig:
		return "Big Blind"
	case BlindBoss:
		return "Boss Blind"
	default:
		return "Blind"
	}
}

// FormatTones formats a tone slice like [4, 4] as "4-4".
func FormatTones(tones []int) string {
	out := ""
	for i, t := range tones {
		if i > 0 {
			out += "-"
		}
		out += fmt.Sprintf("%d", t)
	}
	return out
}
