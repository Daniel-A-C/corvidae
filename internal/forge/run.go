package forge

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

const (
	HandSize          = 9
	MaxRelics         = 5
	WinAnte           = 6
	DefaultHands      = 4
	DefaultDiscards   = 3
	RerollCost        = 4
	PruneCost         = 4
	DuplicateCost     = 5
	ScrollCost        = 4
	DraftCost         = 4
	StartingMoney     = 6
)

// DraftOption is one selectable Hanzi in the Shop's Hanzi Draft pack.
type DraftOption struct {
	Glyph           string
	UnlocksCount    int
	Bought          bool
}

// ShopRelicOffer is a Relic for sale in the Shop.
type ShopRelicOffer struct {
	Def    RelicDef
	Bought bool
}

// ShopScrollOffer upgrades a HandType permanently for this run.
type ShopScrollOffer struct {
	HandType HandType
	Cost     int
	Bought   bool
}

// ShopState holds the current Shop inventory between Blinds.
type ShopState struct {
	Relics          []ShopRelicOffer
	Draft           []DraftOption
	Scroll          ShopScrollOffer
	PrunedThisShop  bool
	DupedThisShop   bool
	StatusMessage   string
}

// Run encapsulates the complete live state of a Morpheme Forge roguelike run.
type Run struct {
	Lexicon         *Lexicon
	Progress        *ProgressStore
	RNG             *rand.Rand
	ArchetypeID     string
	ArchetypeName   string

	// Progression
	Ante            int // 1..6, then 7+ in Endless
	BlindKind       BlindKind
	EndlessMode     bool
	WonRun          bool
	GameOver        bool
	TargetScore     int
	BlindScore      int
	ActiveBoss      BossBlindID
	SilencedSuit    RadicalSuit
	TutorShieldUsed bool

	// Resources
	Money           int
	MaxHands        int
	MaxDiscards     int
	HandsLeft       int
	DiscardsLeft    int

	// Cards & Deck
	NextCardID      int
	FullDeck        []DeckCard
	DrawPile        []DeckCard
	Hand            []DeckCard
	DiscardPile     []DeckCard

	// Staging (current turn)
	Staged          []StagedWord

	// Prove phase
	PendingInvalid  map[int]string
	Challenges      []ProveChallenge
	ProveIdx        int
	PassedFlags     []bool
	LastProveMsg    string
	LastProveFailed bool
	LastPlayResult  *PlayResult

	// Build & Shop
	Relics          []RelicID
	HandLevels      map[HandType]int
	Shop            ShopState

	// Run telemetry
	WordsForged     int
	BestHandScore   int
	UsedBosses      map[BossBlindID]bool
}

// NewRun initializes a fresh Morpheme Forge run with the chosen starter archetype.
func NewRun(lex *Lexicon, prog *ProgressStore, archetypeID string, seed int64) *Run {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	if prog == nil {
		prog = NewEmptyProgress("")
	}
	rng := rand.New(rand.NewSource(seed))

	archName := "Hogwarts Core"
	for _, a := range StarterArchetypes {
		if a.ID == archetypeID {
			archName = a.Name
			break
		}
	}

	r := &Run{
		Lexicon:       lex,
		Progress:      prog,
		RNG:           rng,
		ArchetypeID:   archetypeID,
		ArchetypeName: archName,
		Ante:          1,
		BlindKind:     BlindSmall,
		Money:         StartingMoney,
		MaxHands:      DefaultHands,
		MaxDiscards:   DefaultDiscards,
		HandLevels:    make(map[HandType]int),
		UsedBosses:    make(map[BossBlindID]bool),
	}

	for _, ht := range AllHandTypes {
		r.HandLevels[ht] = 1
	}

	// Apply starter archetype passives
	switch archetypeID {
	case "harry_potter":
		r.MaxHands++
	case "mind_and_people":
		r.MaxDiscards++
	case "volatile_scholar":
		r.Relics = append(r.Relics, RelicForgivingTutor)
		r.Money += 4
	}

	// Build initial deck from starter glyphs
	glyphs := lex.StarterGlyphs(archetypeID, prog)
	for _, g := range glyphs {
		r.NextCardID++
		r.FullDeck = append(r.FullDeck, DeckCard{
			ID:    r.NextCardID,
			Glyph: g,
		})
	}

	prog.RunsPlayed++
	r.StartBlind()
	return r
}

// HasRelic reports whether the player currently holds the given RelicID.
func (r *Run) HasRelic(id RelicID) bool {
	for _, held := range r.Relics {
		if held == id {
			return true
		}
	}
	return false
}

// BlindTarget computes the target score for an Ante and BlindKind.
func BlindTarget(ante int, kind BlindKind) int {
	// Smooth, exciting curve:
	// Ante 1: 180 / 270 / 380
	// Ante 2: 360 / 540 / 760
	// Ante 3: 700 / 1050 / 1450
	// Ante 4: 1350 / 2000 / 2800
	// Ante 5: 2500 / 3750 / 5200
	// Ante 6: 4600 / 6900 / 9600
	// Endless (7+): exponential scaling
	baseTable := []int{0, 180, 360, 700, 1350, 2500, 4600}
	var base float64
	if ante <= 6 {
		base = float64(baseTable[ante])
	} else {
		base = 4600.0 * math.Pow(1.85, float64(ante-6))
	}
	switch kind {
	case BlindBig:
		base *= 1.5
	case BlindBoss:
		base *= 2.1
	}
	return int(math.Round(base/10.0)) * 10
}

// StartBlind resets the deck, draws 9 cards, and configures Boss Blind modifiers if applicable.
func (r *Run) StartBlind() {
	r.BlindScore = 0
	r.HandsLeft = r.MaxHands
	r.DiscardsLeft = r.MaxDiscards
	r.TutorShieldUsed = false
	r.ActiveBoss = BossNone
	r.SilencedSuit = ""
	r.LastPlayResult = nil
	r.Staged = []StagedWord{{}}

	r.TargetScore = BlindTarget(r.Ante, r.BlindKind)

	if r.BlindKind == BlindBoss {
		r.ActiveBoss = r.pickBossBlind()
		switch r.ActiveBoss {
		case BossSilencedSuit:
			suits := []RadicalSuit{SuitNature, SuitHuman, SuitCraft, SuitCore}
			r.SilencedSuit = suits[r.RNG.Intn(len(suits))]
		case BossTightPurse:
			if r.DiscardsLeft > 2 {
				r.DiscardsLeft = 2
			}
		case BossHighBar:
			r.TargetScore = int(math.Round(float64(r.TargetScore)*1.5/10.0)) * 10
			r.HandsLeft++
		}
	}

	// Shuffle FullDeck into DrawPile
	r.DrawPile = append([]DeckCard(nil), r.FullDeck...)
	shuffleSlice(r.DrawPile, r.RNG)
	r.Hand = nil
	r.DiscardPile = nil
	r.DrawToFullHand()
}

func (r *Run) pickBossBlind() BossBlindID {
	var pool []BossBlindID
	for _, b := range BossBlindCatalog {
		if !r.UsedBosses[b.ID] {
			pool = append(pool, b.ID)
		}
	}
	if len(pool) == 0 {
		r.UsedBosses = make(map[BossBlindID]bool)
		for _, b := range BossBlindCatalog {
			pool = append(pool, b.ID)
		}
	}
	pick := pool[r.RNG.Intn(len(pool))]
	r.UsedBosses[pick] = true
	return pick
}

// DrawToFullHand draws cards from DrawPile (reshuffling DiscardPile if needed) up to HandSize.
func (r *Run) DrawToFullHand() {
	for len(r.Hand) < HandSize {
		if len(r.DrawPile) == 0 {
			if len(r.DiscardPile) == 0 {
				break
			}
			r.DrawPile = append([]DeckCard(nil), r.DiscardPile...)
			r.DiscardPile = nil
			shuffleSlice(r.DrawPile, r.RNG)
		}
		card := r.DrawPile[0]
		r.DrawPile = r.DrawPile[1:]
		r.Hand = append(r.Hand, card)
	}
}

// IsCardStaged reports whether hand index idx is already used in any staged word.
func (r *Run) IsCardStaged(idx int) bool {
	for _, sw := range r.Staged {
		for _, ci := range sw.CardIndices {
			if ci == idx {
				return true
			}
		}
	}
	return false
}

// StageCard appends hand card idx to the active staged word slot.
func (r *Run) StageCard(idx int) bool {
	if idx < 0 || idx >= len(r.Hand) {
		return false
	}
	if len(r.Staged) == 0 {
		r.Staged = []StagedWord{{}}
	}
	cur := &r.Staged[len(r.Staged)-1]
	// Don't allow adding the exact same hand card twice in the same word slot,
	// or reusing a card already in an earlier slot (except via StartPivot).
	if r.IsCardStaged(idx) {
		return false
	}
	if len(cur.CardIndices) >= 4 {
		return false
	}
	cur.CardIndices = append(cur.CardIndices, idx)
	return true
}

// StartNewWordSlot starts a separate word slot (Double/Triple Forge).
func (r *Run) StartNewWordSlot() bool {
	if len(r.Staged) == 0 || len(r.Staged[len(r.Staged)-1].CardIndices) == 0 {
		return false
	}
	if len(r.Staged) >= 3 {
		return false
	}
	r.Staged = append(r.Staged, StagedWord{})
	return true
}

// StartPivotSlot starts a new word slot bridged by the last card of the previous word (Pivot Chain).
func (r *Run) StartPivotSlot() bool {
	if len(r.Staged) == 0 {
		return false
	}
	prev := r.Staged[len(r.Staged)-1]
	if len(prev.CardIndices) < 2 || len(r.Staged) >= 3 {
		return false
	}
	bridgeIdx := prev.CardIndices[len(prev.CardIndices)-1]
	r.Staged = append(r.Staged, StagedWord{
		CardIndices: []int{bridgeIdx},
		IsPivot:     true,
	})
	return true
}

// UndoStage pops the last staged character or empty word slot.
func (r *Run) UndoStage() {
	if len(r.Staged) == 0 {
		r.Staged = []StagedWord{{}}
		return
	}
	cur := &r.Staged[len(r.Staged)-1]
	if len(cur.CardIndices) == 0 {
		if len(r.Staged) > 1 {
			r.Staged = r.Staged[:len(r.Staged)-1]
		}
		return
	}
	// If this is a pivot slot with only the bridge char, remove the whole pivot slot
	if cur.IsPivot && len(cur.CardIndices) == 1 {
		r.Staged = r.Staged[:len(r.Staged)-1]
		return
	}
	cur.CardIndices = cur.CardIndices[:len(cur.CardIndices)-1]
}

// ClearStage resets all staged words.
func (r *Run) ClearStage() {
	r.Staged = []StagedWord{{}}
}

// HasStagedCards reports whether at least one card is staged.
func (r *Run) HasStagedCards() bool {
	for _, sw := range r.Staged {
		if len(sw.CardIndices) > 0 {
			return true
		}
	}
	return false
}

// DiscardStaged discards all currently staged hand cards and draws replacements.
func (r *Run) DiscardStaged() bool {
	if r.DiscardsLeft <= 0 || !r.HasStagedCards() {
		return false
	}
	toDrop := make(map[int]bool)
	for _, sw := range r.Staged {
		for _, ci := range sw.CardIndices {
			toDrop[ci] = true
		}
	}
	var kept []DeckCard
	for i, c := range r.Hand {
		if toDrop[i] {
			r.DiscardPile = append(r.DiscardPile, c)
		} else {
			kept = append(kept, c)
		}
	}
	r.Hand = kept
	r.DiscardsLeft--
	r.ClearStage()
	r.DrawToFullHand()
	return true
}

// CommitPlay validates staged words and prepares the Prove cascade.
// Returns true if there are valid words to Prove, or false if all staged words fizzled immediately.
func (r *Run) CommitPlay() bool {
	if r.HandsLeft <= 0 || !r.HasStagedCards() {
		return false
	}
	// Trim trailing empty slot if player hit Space then Enter
	for len(r.Staged) > 1 && len(r.Staged[len(r.Staged)-1].CardIndices) == 0 {
		r.Staged = r.Staged[:len(r.Staged)-1]
	}

	r.Challenges, r.PendingInvalid = ValidateStagedWords(
		r.Staged, r.Hand, r.Lexicon, r.Progress, r.ActiveBoss, r.RNG,
	)
	r.ProveIdx = 0
	r.PassedFlags = make([]bool, len(r.Challenges))
	r.LastProveMsg = ""
	r.LastProveFailed = false

	if len(r.Challenges) == 0 {
		// All staged words were invalid -> immediate fizzle, costs 1 Hand
		res := ScorePlay(r, r.Staged, r.Challenges, r.PendingInvalid, r.PassedFlags)
		r.LastPlayResult = &res
		r.finishCommittedHand()
		return false
	}
	return true
}

// AnswerProve records the player's home-row option choice for the current ProveChallenge.
// Returns (passed, moreChallengesRemaining).
func (r *Run) AnswerProve(optionIdx int, now time.Time, syncDecks bool, baseDecksDir string) (bool, bool) {
	if r.ProveIdx >= len(r.Challenges) {
		return false, false
	}
	ch := r.Challenges[r.ProveIdx]
	passed := optionIdx == ch.CorrectIdx
	r.PassedFlags[r.ProveIdx] = passed

	// Update parallel SM-2 progress
	r.Progress.RecordOutcome(ch.Word, passed, now)
	if passed {
		r.WordsForged++
	}
	if syncDecks && len([]rune(ch.Word)) >= 2 {
		SyncCompoundToMandarinDecks(baseDecksDir, ch.Word, passed, now)
	}

	if passed {
		r.LastProveFailed = false
		r.LastProveMsg = fmt.Sprintf("✓ %s (%s) — %s", ch.Word, ch.Pinyin, ch.Meaning)
	} else {
		r.LastProveFailed = true
		r.LastProveMsg = fmt.Sprintf("✗ %s (%s) — %s", ch.Word, ch.Pinyin, ch.Meaning)
	}

	r.ProveIdx++
	if r.ProveIdx >= len(r.Challenges) {
		res := ScorePlay(r, r.Staged, r.Challenges, r.PendingInvalid, r.PassedFlags)
		r.LastPlayResult = &res
		r.BlindScore += res.TotalScore
		r.Money += res.CoinsEarned
		if res.TotalScore > r.BestHandScore {
			r.BestHandScore = res.TotalScore
		}
		if res.TotalScore > r.Progress.BestScore {
			r.Progress.BestScore = res.TotalScore
		}
		r.finishCommittedHand()
		return passed, false
	}
	return passed, true
}

func (r *Run) finishCommittedHand() {
	// Move played hand cards to DiscardPile and redraw
	used := make(map[int]bool)
	for _, sw := range r.Staged {
		for _, ci := range sw.CardIndices {
			used[ci] = true
		}
	}
	var kept []DeckCard
	for i, c := range r.Hand {
		if used[i] {
			r.DiscardPile = append(r.DiscardPile, c)
		} else {
			kept = append(kept, c)
		}
	}
	r.Hand = kept
	r.HandsLeft--
	r.ClearStage()
	r.DrawToFullHand()

	if r.BlindScore >= r.TargetScore {
		if r.Ante > r.Progress.BestAnte {
			r.Progress.BestAnte = r.Ante
		}
		if r.Ante == WinAnte && r.BlindKind == BlindBoss && !r.WonRun {
			r.WonRun = true
			r.Progress.RunsWon++
		}
	} else if r.HandsLeft <= 0 {
		r.GameOver = true
	}
}

// BlindCleared reports whether the player reached the current Blind's TargetScore.
func (r *Run) BlindCleared() bool {
	return r.BlindScore >= r.TargetScore
}

// ClaimBlindReward awards end-of-blind coins and generates the Shop.
func (r *Run) ClaimBlindReward() int {
	baseReward := 4
	if r.BlindKind == BlindBig {
		baseReward = 5
	} else if r.BlindKind == BlindBoss {
		baseReward = 7
	}
	handBonus := r.HandsLeft
	interest := r.Money / 5
	if interest > 5 {
		interest = 5
	}
	total := baseReward + handBonus + interest
	r.Money += total
	r.GenerateShop()
	return total
}

// AdvanceToNextBlind steps Small -> Big -> Boss -> Next Ante Small and starts the Blind.
func (r *Run) AdvanceToNextBlind() {
	switch r.BlindKind {
	case BlindSmall:
		r.BlindKind = BlindBig
	case BlindBig:
		r.BlindKind = BlindBoss
	case BlindBoss:
		r.BlindKind = BlindSmall
		r.Ante++
		if r.Ante > WinAnte {
			r.EndlessMode = true
		}
	}
	r.StartBlind()
}

// GenerateShop populates 2 Relics, 3 Hanzi Draft picks, and 1 Scroll upgrade.
func (r *Run) GenerateShop() {
	r.Shop = ShopState{}

	// 1. Pick 2 unowned Relics
	var relicPool []RelicDef
	for _, def := range RelicCatalog {
		if !r.HasRelic(def.ID) {
			relicPool = append(relicPool, def)
		}
	}
	shuffleSlice(relicPool, r.RNG)
	for i := 0; i < len(relicPool) && i < 2; i++ {
		r.Shop.Relics = append(r.Shop.Relics, ShopRelicOffer{Def: relicPool[i]})
	}

	// 2. Pick 3 Hanzi Draft candidates not yet in FullDeck (or best connectivity)
	deckGlyphs := r.deckGlyphSet()
	type cand struct {
		glyph   string
		unlocks int
	}
	var cands []cand
	for _, h := range r.Lexicon.Hanzi {
		if !deckGlyphs[h.Char] {
			u := r.Lexicon.CountNewCompoundsWithDeck(h.Char, deckGlyphs)
			if u > 0 {
				cands = append(cands, cand{glyph: h.Char, unlocks: u})
			}
		}
	}
	shuffleSlice(cands, r.RNG)
	// Bias slightly so at least 1 high-connectivity option appears
	if len(cands) > 3 {
		bestIdx := 0
		for i := 1; i < len(cands); i++ {
			if cands[i].unlocks > cands[bestIdx].unlocks {
				bestIdx = i
			}
		}
		cands[0], cands[bestIdx] = cands[bestIdx], cands[0]
		cands = cands[:3]
	}
	for _, c := range cands {
		r.Shop.Draft = append(r.Shop.Draft, DraftOption{
			Glyph:        c.glyph,
			UnlocksCount: c.unlocks,
		})
	}

	// 3. Pick 1 Scroll (upgrade a HandType)
	ht := AllHandTypes[r.RNG.Intn(len(AllHandTypes))]
	r.Shop.Scroll = ShopScrollOffer{
		HandType: ht,
		Cost:     ScrollCost,
	}
}

func (r *Run) deckGlyphSet() map[string]bool {
	s := make(map[string]bool, len(r.FullDeck))
	for _, c := range r.FullDeck {
		s[c.Glyph] = true
	}
	return s
}

// BuyRelic purchases shop relic offer idx.
func (r *Run) BuyRelic(idx int) bool {
	if idx < 0 || idx >= len(r.Shop.Relics) {
		return false
	}
	offer := &r.Shop.Relics[idx]
	if offer.Bought {
		r.Shop.StatusMessage = "Already purchased."
		return false
	}
	if len(r.Relics) >= MaxRelics {
		r.Shop.StatusMessage = fmt.Sprintf("Relic slots full (%d/%d).", len(r.Relics), MaxRelics)
		return false
	}
	if r.Money < offer.Def.Cost {
		r.Shop.StatusMessage = "Not enough coins."
		return false
	}
	r.Money -= offer.Def.Cost
	r.Relics = append(r.Relics, offer.Def.ID)
	offer.Bought = true
	r.Shop.StatusMessage = fmt.Sprintf("Bought %s!", offer.Def.Name)
	return true
}

// BuyDraft adds a drafted Hanzi to FullDeck.
func (r *Run) BuyDraft(idx int) bool {
	if idx < 0 || idx >= len(r.Shop.Draft) {
		return false
	}
	opt := &r.Shop.Draft[idx]
	if opt.Bought {
		r.Shop.StatusMessage = "Already drafted."
		return false
	}
	if r.Money < DraftCost {
		r.Shop.StatusMessage = "Not enough coins."
		return false
	}
	r.Money -= DraftCost
	r.NextCardID++
	r.FullDeck = append(r.FullDeck, DeckCard{ID: r.NextCardID, Glyph: opt.Glyph})
	opt.Bought = true
	r.Shop.StatusMessage = fmt.Sprintf("Drafted %s into your deck (+%d compounds)!", opt.Glyph, opt.UnlocksCount)
	return true
}

// BuyScroll permanently levels up the shop's HandType.
func (r *Run) BuyScroll() bool {
	if r.Shop.Scroll.Bought {
		r.Shop.StatusMessage = "Scroll already used."
		return false
	}
	if r.Money < r.Shop.Scroll.Cost {
		r.Shop.StatusMessage = "Not enough coins."
		return false
	}
	r.Money -= r.Shop.Scroll.Cost
	r.HandLevels[r.Shop.Scroll.HandType]++
	r.Shop.Scroll.Bought = true
	r.Shop.StatusMessage = fmt.Sprintf("%s upgraded to Lv.%d!", r.Shop.Scroll.HandType.Name(), r.HandLevels[r.Shop.Scroll.HandType])
	return true
}

// RerollShop refreshes shop offers for RerollCost.
func (r *Run) RerollShop() bool {
	if r.Money < RerollCost {
		r.Shop.StatusMessage = "Not enough coins to reroll."
		return false
	}
	r.Money -= RerollCost
	r.GenerateShop()
	r.Shop.StatusMessage = "Shop rerolled!"
	return true
}

// PruneLowestConnectivityCard removes the lowest-connectivity card from FullDeck for PruneCost.
func (r *Run) PruneLowestConnectivityCard() bool {
	if r.Shop.PrunedThisShop {
		r.Shop.StatusMessage = "Already pruned a card this shop."
		return false
	}
	if len(r.FullDeck) <= HandSize+3 {
		r.Shop.StatusMessage = "Deck is already as lean as allowed."
		return false
	}
	if r.Money < PruneCost {
		r.Shop.StatusMessage = "Not enough coins to prune."
		return false
	}
	deckGlyphs := r.deckGlyphSet()
	worstIdx := 0
	worstScore := 1 << 30
	for i, c := range r.FullDeck {
		u := r.Lexicon.CountNewCompoundsWithDeck(c.Glyph, deckGlyphs)
		if u < worstScore {
			worstScore = u
			worstIdx = i
		}
	}
	removed := r.FullDeck[worstIdx].Glyph
	r.FullDeck = append(r.FullDeck[:worstIdx], r.FullDeck[worstIdx+1:]...)
	r.Money -= PruneCost
	r.Shop.PrunedThisShop = true
	r.Shop.StatusMessage = fmt.Sprintf("Pruned %s (only %d deck compounds).", removed, worstScore)
	return true
}

// DuplicateBestHubCard duplicates the highest-connectivity hub card in FullDeck (and gives it Foil+ Edition).
func (r *Run) DuplicateBestHubCard() bool {
	if r.Shop.DupedThisShop {
		r.Shop.StatusMessage = "Already duplicated a card this shop."
		return false
	}
	if r.Money < DuplicateCost {
		r.Shop.StatusMessage = "Not enough coins to duplicate."
		return false
	}
	deckGlyphs := r.deckGlyphSet()
	bestIdx := 0
	bestScore := -1
	for i, c := range r.FullDeck {
		u := r.Lexicon.CountNewCompoundsWithDeck(c.Glyph, deckGlyphs)
		if u > bestScore {
			bestScore = u
			bestIdx = i
		}
	}
	hub := r.FullDeck[bestIdx]
	r.NextCardID++
	editions := []CardEdition{EditionFoil, EditionHolo, EditionPoly}
	ed := editions[r.RNG.Intn(len(editions))]
	r.FullDeck = append(r.FullDeck, DeckCard{
		ID:      r.NextCardID,
		Glyph:   hub.Glyph,
		Edition: ed,
	})
	r.Money -= DuplicateCost
	r.Shop.DupedThisShop = true
	r.Shop.StatusMessage = fmt.Sprintf("Duplicated hub %s [%s] into your deck!", hub.Glyph, ed.Label())
	return true
}
