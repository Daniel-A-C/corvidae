package forge

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

// ValidateStagedWords checks each staged word against the Lexicon and builds ProveChallenges
// for valid words. Invalid words are returned in invalidSlots.
func ValidateStagedWords(
	staged []StagedWord,
	hand []DeckCard,
	lex *Lexicon,
	prog *ProgressStore,
	boss BossBlindID,
	rng *rand.Rand,
) ([]ProveChallenge, map[int]string) {
	var challenges []ProveChallenge
	invalidSlots := make(map[int]string)

	mode := ProveMeaning
	if boss == BossToneTrial {
		mode = ProveTone
	} else if boss == BossPinyinVeil {
		mode = ProvePinyin
	}

	for slotIdx, sw := range staged {
		if len(sw.CardIndices) == 0 {
			continue
		}
		var sb strings.Builder
		for _, ci := range sw.CardIndices {
			if ci >= 0 && ci < len(hand) {
				sb.WriteString(hand[ci].Glyph)
			}
		}
		word := sb.String()
		runes := []rune(word)

		if len(runes) == 1 {
			h, ok := lex.CharsByGlyph[word]
			if !ok || !h.Standalone {
				invalidSlots[slotIdx] = word
				continue
			}
			ch := buildSoloChallenge(slotIdx, h, sw, mode, lex, prog, boss, rng)
			challenges = append(challenges, ch)
			continue
		}

		comp, ok := lex.CompoundsByWord[word]
		if !ok {
			invalidSlots[slotIdx] = word
			continue
		}
		ch := buildCompoundChallenge(slotIdx, comp, sw, mode, lex, prog, boss, rng)
		challenges = append(challenges, ch)
	}
	return challenges, invalidSlots
}

func buildSoloChallenge(
	slotIdx int,
	h *HanziEntry,
	sw StagedWord,
	mode ProveMode,
	lex *Lexicon,
	prog *ProgressStore,
	boss BossBlindID,
	rng *rand.Rand,
) ProveChallenge {
	ch := ProveChallenge{
		WordSlotIdx:        slotIdx,
		Word:               h.Char,
		Pinyin:             h.Pinyin,
		Meaning:            h.Meaning,
		Tones:              []int{h.Tone},
		Mode:               mode,
		IsVolatileCompound: prog != nil && prog.MaterialForChar(h.Char) == MaterialGlass,
		CardIndices:        append([]int(nil), sw.CardIndices...),
		IsPivot:            sw.IsPivot,
		IsSolo:             true,
	}
	switch mode {
	case ProveTone:
		opts := []string{"1", "2", "3", "4"}
		ch.Options = opts
		ch.CorrectIdx = h.Tone - 1
		if ch.CorrectIdx < 0 || ch.CorrectIdx >= 4 {
			ch.CorrectIdx = 3
		}
	case ProvePinyin:
		ch.Options, ch.CorrectIdx = buildHanziPinyinOptions(h, lex, rng)
	default:
		ch.Options, ch.CorrectIdx = buildHanziMeaningOptions(h, lex, rng)
	}
	return ch
}

func buildCompoundChallenge(
	slotIdx int,
	comp *CompoundEntry,
	sw StagedWord,
	mode ProveMode,
	lex *Lexicon,
	prog *ProgressStore,
	boss BossBlindID,
	rng *rand.Rand,
) ProveChallenge {
	ch := ProveChallenge{
		WordSlotIdx:        slotIdx,
		Word:               comp.Word,
		Pinyin:             comp.Pinyin,
		Meaning:            comp.Meaning,
		Tones:              append([]int(nil), comp.Tones...),
		Mode:               mode,
		IsVolatileCompound: prog != nil && prog.IsVolatileCompound(comp.Word),
		CardIndices:        append([]int(nil), sw.CardIndices...),
		IsPivot:            sw.IsPivot,
		IsSolo:             false,
	}
	switch mode {
	case ProveTone:
		ch.Options, ch.CorrectIdx = buildTonePatternOptions(comp.Tones, rng)
	case ProvePinyin:
		ch.Options, ch.CorrectIdx = buildCompoundPinyinOptions(comp, lex, rng)
	default:
		ch.Options, ch.CorrectIdx = buildCompoundMeaningOptions(comp, lex, boss == BossComponentTrap, rng)
	}
	return ch
}

func buildCompoundMeaningOptions(comp *CompoundEntry, lex *Lexicon, extraTraps bool, rng *rand.Rand) ([]string, int) {
	seen := map[string]bool{comp.Meaning: true}
	var distractors []string

	// 1. Component-trap distractors: other compounds sharing a Hanzi with comp.Word
	var sharedPool []*CompoundEntry
	for _, r := range comp.Word {
		for _, other := range lex.CompoundsByChar[string(r)] {
			if other.Word != comp.Word && !seen[other.Meaning] {
				sharedPool = append(sharedPool, other)
			}
		}
	}
	shuffleSlice(sharedPool, rng)
	trapTarget := 2
	if extraTraps {
		trapTarget = 4
	}
	for _, other := range sharedPool {
		if len(distractors) >= trapTarget {
			break
		}
		if !seen[other.Meaning] {
			seen[other.Meaning] = true
			distractors = append(distractors, other.Meaning)
		}
	}

	// 2. Domain-matched or general compounds
	var fallback []*CompoundEntry
	for i := range lex.Compounds {
		other := &lex.Compounds[i]
		if other.Word != comp.Word && !seen[other.Meaning] {
			fallback = append(fallback, other)
		}
	}
	shuffleSlice(fallback, rng)
	for _, other := range fallback {
		if len(distractors) >= 4 {
			break
		}
		if !seen[other.Meaning] {
			seen[other.Meaning] = true
			distractors = append(distractors, other.Meaning)
		}
	}

	return assembleOptions(comp.Meaning, distractors, rng)
}

func buildHanziMeaningOptions(h *HanziEntry, lex *Lexicon, rng *rand.Rand) ([]string, int) {
	seen := map[string]bool{h.Meaning: true}
	var pool []*HanziEntry
	for i := range lex.Hanzi {
		other := &lex.Hanzi[i]
		if other.Char != h.Char && !seen[other.Meaning] {
			pool = append(pool, other)
		}
	}
	shuffleSlice(pool, rng)
	var distractors []string
	for _, other := range pool {
		if len(distractors) >= 4 {
			break
		}
		if !seen[other.Meaning] {
			seen[other.Meaning] = true
			distractors = append(distractors, other.Meaning)
		}
	}
	return assembleOptions(h.Meaning, distractors, rng)
}

func buildCompoundPinyinOptions(comp *CompoundEntry, lex *Lexicon, rng *rand.Rand) ([]string, int) {
	seen := map[string]bool{comp.Pinyin: true}
	var distractors []string

	// Prefer compounds sharing a character (so one syllable matches!)
	var shared []*CompoundEntry
	for _, r := range comp.Word {
		for _, other := range lex.CompoundsByChar[string(r)] {
			if other.Word != comp.Word && !seen[other.Pinyin] && len([]rune(other.Word)) == len([]rune(comp.Word)) {
				shared = append(shared, other)
			}
		}
	}
	shuffleSlice(shared, rng)
	for _, other := range shared {
		if len(distractors) >= 2 {
			break
		}
		if !seen[other.Pinyin] {
			seen[other.Pinyin] = true
			distractors = append(distractors, other.Pinyin)
		}
	}

	var fallback []*CompoundEntry
	for i := range lex.Compounds {
		other := &lex.Compounds[i]
		if other.Word != comp.Word && !seen[other.Pinyin] {
			fallback = append(fallback, other)
		}
	}
	shuffleSlice(fallback, rng)
	for _, other := range fallback {
		if len(distractors) >= 4 {
			break
		}
		if !seen[other.Pinyin] {
			seen[other.Pinyin] = true
			distractors = append(distractors, other.Pinyin)
		}
	}
	return assembleOptions(comp.Pinyin, distractors, rng)
}

func buildHanziPinyinOptions(h *HanziEntry, lex *Lexicon, rng *rand.Rand) ([]string, int) {
	seen := map[string]bool{h.Pinyin: true}
	var pool []*HanziEntry
	for i := range lex.Hanzi {
		other := &lex.Hanzi[i]
		if other.Char != h.Char && !seen[other.Pinyin] {
			pool = append(pool, other)
		}
	}
	shuffleSlice(pool, rng)
	var distractors []string
	for _, other := range pool {
		if len(distractors) >= 4 {
			break
		}
		if !seen[other.Pinyin] {
			seen[other.Pinyin] = true
			distractors = append(distractors, other.Pinyin)
		}
	}
	return assembleOptions(h.Pinyin, distractors, rng)
}

func buildTonePatternOptions(tones []int, rng *rand.Rand) ([]string, int) {
	correct := FormatTones(tones)
	seen := map[string]bool{correct: true}
	var distractors []string
	n := len(tones)
	if n == 0 {
		n = 2
	}
	// Generate plausible 1-tone-perturbed patterns first
	for attempts := 0; attempts < 40 && len(distractors) < 3; attempts++ {
		cand := append([]int(nil), tones...)
		pos := rng.Intn(len(cand))
		cand[pos] = 1 + rng.Intn(4)
		s := FormatTones(cand)
		if !seen[s] {
			seen[s] = true
			distractors = append(distractors, s)
		}
	}
	return assembleOptions(correct, distractors, rng)
}

func assembleOptions(correct string, distractors []string, rng *rand.Rand) ([]string, int) {
	opts := make([]string, 0, len(distractors)+1)
	opts = append(opts, correct)
	opts = append(opts, distractors...)
	shuffleSlice(opts, rng)
	correctIdx := 0
	for i, o := range opts {
		if o == correct {
			correctIdx = i
			break
		}
	}
	return opts, correctIdx
}

func shuffleSlice[T any](s []T, rng *rand.Rand) {
	if len(s) <= 1 {
		return
	}
	rng.Shuffle(len(s), func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})
}

// ClassifyHand determines the HandType from the list of proven words.
func ClassifyHand(proven []ProveChallenge) HandType {
	if len(proven) == 0 {
		return HandNone
	}
	hasPivot := false
	maxLen := 0
	compoundCount := 0
	for _, p := range proven {
		l := len([]rune(p.Word))
		if l > maxLen {
			maxLen = l
		}
		if l >= 2 {
			compoundCount++
		}
		if p.IsPivot {
			hasPivot = true
		}
	}
	if maxLen >= 4 {
		return HandIdiom
	}
	if len(proven) >= 3 || (hasPivot && maxLen >= 3) {
		return HandGrandChain
	}
	if hasPivot {
		return HandPivotChain
	}
	if maxLen == 3 {
		return HandTriad
	}
	if compoundCount >= 2 {
		return HandDoubleForge
	}
	if compoundCount == 1 {
		return HandCompound
	}
	return HandSolo
}

// ScorePlay computes the final Chips × Mult breakdown after all Prove challenges have been answered.
func ScorePlay(
	run *Run,
	staged []StagedWord,
	challenges []ProveChallenge,
	invalidSlots map[int]string,
	passedFlags []bool,
) PlayResult {
	res := PlayResult{}

	// Map challenges by slotIdx for ordered outcome reporting
	chalBySlot := make(map[int]int)
	for i, c := range challenges {
		chalBySlot[c.WordSlotIdx] = i
	}

	var proven []ProveChallenge
	glassShattered := false

	for slotIdx, sw := range staged {
		if len(sw.CardIndices) == 0 {
			continue
		}
		if badWord, isInv := invalidSlots[slotIdx]; isInv {
			res.Words = append(res.Words, WordOutcome{
				Word:  badWord,
				Valid: false,
				Notes: []string{"Not in lexicon (Fizzled)"},
			})
			continue
		}
		ci, ok := chalBySlot[slotIdx]
		if !ok {
			continue
		}
		ch := challenges[ci]
		passed := ci < len(passedFlags) && passedFlags[ci]

		// Heavy Forge Boss Blind: Solo words score 0
		if run.ActiveBoss == BossHeavyForge && ch.IsSolo {
			res.Words = append(res.Words, WordOutcome{
				Word:    ch.Word,
				Pinyin:  ch.Pinyin,
				Meaning: ch.Meaning,
				Valid:   true,
				Passed:  false,
				Notes:   []string{"Blocked by The Heavy Forge (Solo = 0)"},
			})
			continue
		}

		shieldSaved := false
		if !passed && run.HasRelic(RelicForgivingTutor) && !run.TutorShieldUsed {
			run.TutorShieldUsed = true
			passed = true
			shieldSaved = true
		}

		wo := WordOutcome{
			Word:        ch.Word,
			Pinyin:      ch.Pinyin,
			Meaning:     ch.Meaning,
			Valid:       true,
			Passed:      passed,
			ShieldSaved: shieldSaved,
		}
		if shieldSaved {
			wo.Notes = append(wo.Notes, "Saved by Forgiving Tutor!")
		}

		if passed {
			proven = append(proven, ch)
		} else {
			// Check if any card in this missed word was Glass -> shatter!
			for _, cardIdx := range ch.CardIndices {
				if cardIdx >= 0 && cardIdx < len(run.Hand) {
					if run.Progress.MaterialForChar(run.Hand[cardIdx].Glyph) == MaterialGlass {
						glassShattered = true
						wo.GlassShattered = true
					}
				}
			}
			if ch.IsVolatileCompound {
				glassShattered = true
				wo.GlassShattered = true
			}
			wo.Notes = append(wo.Notes, "Missed recall (Fizzled)")
		}
		res.Words = append(res.Words, wo)
	}

	res.GlassShattered = glassShattered
	if len(proven) == 0 {
		res.HandType = HandNone
		res.Events = append(res.Events, ScoreEvent{Label: "Fizzle", Detail: "No proven words — 0 pts"})
		return res
	}

	// 1. Base Hand Type
	res.HandType = ClassifyHand(proven)
	res.HandLevel = run.HandLevels[res.HandType]
	if res.HandLevel < 1 {
		res.HandLevel = 1
	}
	baseChips, baseMult := res.HandType.StatsAtLevel(res.HandLevel)
	chips := baseChips
	mult := baseMult
	xMult := 1.0

	res.Events = append(res.Events, ScoreEvent{
		Label:  fmt.Sprintf("%s (Lv.%d)", res.HandType.Name(), res.HandLevel),
		Detail: fmt.Sprintf("+%d Chips, +%.0f Mult", baseChips, baseMult),
	})

	// Collect unique played hand cards (and track tones/suits across proven words)
	var playedCards []DeckCard
	seenCardIdx := make(map[int]bool)
	var allTones []int
	var allSuits []RadicalSuit

	for _, ch := range proven {
		allTones = append(allTones, ch.Tones...)
		// Volatile compound discovery surge
		if ch.IsVolatileCompound && !ch.IsSolo {
			surge := 1.5
			if run.HasRelic(RelicGlassCannon) {
				surge = 2.0
			}
			xMult *= surge
			res.Events = append(res.Events, ScoreEvent{
				Label:  fmt.Sprintf("Volatile Discovery (%s)", ch.Word),
				Detail: fmt.Sprintf("x%.2f Mult", surge),
			})
		}

		// Harmonic Pair relic: all tones in a multi-char word match
		if run.HasRelic(RelicHarmonicPair) && len(ch.Tones) >= 2 {
			same := true
			for _, t := range ch.Tones[1:] {
				if t != ch.Tones[0] {
					same = false
					break
				}
			}
			if same {
				mult += 8
				res.Events = append(res.Events, ScoreEvent{
					Label:  fmt.Sprintf("Harmonic Pair (%s)", ch.Word),
					Detail: "+8 Mult",
				})
			}
		}

		// Triad Smith relic
		if run.HasRelic(RelicTriadSmith) && len([]rune(ch.Word)) >= 3 {
			xMult *= 1.6
			res.Events = append(res.Events, ScoreEvent{
				Label:  fmt.Sprintf("Triad Smith (%s)", ch.Word),
				Detail: "x1.6 Mult",
			})
		}

		for _, ci := range ch.CardIndices {
			if ci >= 0 && ci < len(run.Hand) && !seenCardIdx[ci] {
				seenCardIdx[ci] = true
				playedCards = append(playedCards, run.Hand[ci])
			}
		}
	}

	// 2. Per-card Material, Edition, Suit & Tone Triggers
	for _, card := range playedCards {
		h, ok := run.Lexicon.CharsByGlyph[card.Glyph]
		if !ok {
			continue
		}
		allSuits = append(allSuits, h.Suit)

		if run.ActiveBoss == BossSilencedSuit && h.Suit == run.SilencedSuit {
			res.Events = append(res.Events, ScoreEvent{
				Label:  fmt.Sprintf("%s (Silenced %s)", card.Glyph, h.Suit.Label()),
				Detail: "0 Chips (Boss Blind)",
			})
			continue
		}

		// Material physics
		mat := run.Progress.MaterialForChar(card.Glyph)
		switch mat {
		case MaterialGlass:
			chips += 15
			gm := 1.25
			if run.HasRelic(RelicVolatileScholar) {
				chips += 20
				gm = 1.5
			}
			xMult *= gm
			res.Events = append(res.Events, ScoreEvent{
				Label:  fmt.Sprintf("%s [Glass]", card.Glyph),
				Detail: fmt.Sprintf("+%d Chips, x%.2f Mult", 15+boolInt(run.HasRelic(RelicVolatileScholar))*20, gm),
			})
		case MaterialFoil:
			cAdd := 25
			if run.HasRelic(RelicDeepMemory) {
				cAdd += 30
			}
			chips += cAdd
			mult += 3
			res.Events = append(res.Events, ScoreEvent{
				Label:  fmt.Sprintf("%s [Foil]", card.Glyph),
				Detail: fmt.Sprintf("+%d Chips, +3 Mult", cAdd),
			})
		case MaterialGold:
			chips += 40
			coins := 1
			if run.HasRelic(RelicOldReliable) {
				coins += 2
			}
			res.CoinsEarned += coins
			res.Events = append(res.Events, ScoreEvent{
				Label:  fmt.Sprintf("%s [Gold]", card.Glyph),
				Detail: fmt.Sprintf("+40 Chips, +$%d", coins),
			})
		}

		// Shop Edition
		switch card.Edition {
		case EditionFoil:
			chips += 30
			res.Events = append(res.Events, ScoreEvent{Label: card.Glyph + " (Foil+)", Detail: "+30 Chips"})
		case EditionHolo:
			mult += 6
			res.Events = append(res.Events, ScoreEvent{Label: card.Glyph + " (Holo)", Detail: "+6 Mult"})
		case EditionPoly:
			xMult *= 1.5
			res.Events = append(res.Events, ScoreEvent{Label: card.Glyph + " (Poly)", Detail: "x1.5 Mult"})
		}

		// Suit Relics
		if h.Suit == SuitNature && run.HasRelic(RelicNatureFlow) {
			chips += 30
			res.Events = append(res.Events, ScoreEvent{Label: "Nature Flow (" + card.Glyph + ")", Detail: "+30 Chips"})
		}
		if h.Suit == SuitHuman && run.HasRelic(RelicHumanResonance) {
			mult += 4
			res.Events = append(res.Events, ScoreEvent{Label: "Human Resonance (" + card.Glyph + ")", Detail: "+4 Mult"})
		}
		if h.Suit == SuitCraft && run.HasRelic(RelicCraftAnvil) {
			chips += 25
			mult += 2
			res.Events = append(res.Events, ScoreEvent{Label: "Craft Anvil (" + card.Glyph + ")", Detail: "+25 Chips, +2 Mult"})
		}

		// Tone Relics
		if h.Tone == 4 && run.HasRelic(RelicFourthToneEcho) {
			mult += 5
			res.Events = append(res.Events, ScoreEvent{Label: "4th-Tone Echo (" + card.Glyph + ")", Detail: "+5 Mult"})
		}
		if h.Tone == 2 && run.HasRelic(RelicRisingChorus) {
			chips += 35
			res.Events = append(res.Events, ScoreEvent{Label: "Rising Chorus (" + card.Glyph + ")", Detail: "+35 Chips"})
		}
	}

	// 3. Orthogonal Pattern Bonuses (>= 3 played characters)
	if len(allSuits) >= 3 && allEqualSuit(allSuits) {
		res.RadicalFlush = true
		rfChips := 40
		if run.ArchetypeID == "john_gospel" {
			rfChips += 30
		}
		chips += rfChips
		xMult *= 2.0
		res.Events = append(res.Events, ScoreEvent{
			Label:  fmt.Sprintf("Radical Flush (%s)", allSuits[0].Label()),
			Detail: fmt.Sprintf("+%d Chips, x2.0 Mult", rfChips),
		})
		if run.HasRelic(RelicRadicalPurist) {
			xMult *= 1.75
			res.Events = append(res.Events, ScoreEvent{Label: "Radical Purist", Detail: "x1.75 Mult"})
		}
	}

	if len(allTones) >= 3 && allEqualInt(allTones) {
		res.ToneFlush = true
		chips += 30
		xMult *= 1.8
		res.Events = append(res.Events, ScoreEvent{
			Label:  fmt.Sprintf("Tone Flush (Tone %d)", allTones[0]),
			Detail: "+30 Chips, x1.8 Mult",
		})
	}

	if len(allTones) >= 3 && isStrictlyAscending(allTones) {
		res.ToneEscalation = true
		chips += 50
		xMult *= 2.0
		res.Events = append(res.Events, ScoreEvent{
			Label:  "Tone Escalation (" + FormatTones(allTones) + ")",
			Detail: "+50 Chips, x2.0 Mult",
		})
		if run.HasRelic(RelicToneLadder) {
			xMult *= 1.5
			res.Events = append(res.Events, ScoreEvent{Label: "Tone Ladder", Detail: "x1.5 Mult"})
		}
	}

	// 4. Hand-Type & Economy Relics
	if (res.HandType == HandPivotChain || res.HandType == HandGrandChain) && run.HasRelic(RelicPivotMaster) {
		chips += 60
		xMult *= 1.5
		res.Events = append(res.Events, ScoreEvent{Label: "Pivot Master", Detail: "+60 Chips, x1.5 Mult"})
	}
	if res.HandType == HandDoubleForge && run.HasRelic(RelicTwinForge) {
		chips += 70
		mult += 6
		res.Events = append(res.Events, ScoreEvent{Label: "Twin Anvil", Detail: "+70 Chips, +6 Mult"})
	}
	if res.HandType == HandSolo && run.HasRelic(RelicLoneWolf) {
		chips += 45
		mult += 4
		res.Events = append(res.Events, ScoreEvent{Label: "Lone Wolf", Detail: "+45 Chips, +4 Mult"})
	}
	if run.HasRelic(RelicMerchantsInk) && run.Money > 0 {
		bonus := run.Money * 2
		chips += bonus
		res.Events = append(res.Events, ScoreEvent{
			Label:  "Merchant's Ink",
			Detail: fmt.Sprintf("+%d Chips ($%d)", bonus, run.Money),
		})
	}

	// 5. Glass Shatter penalty: zeroes xMult bonus back to 1.0
	if glassShattered && xMult > 1.0 {
		res.Events = append(res.Events, ScoreEvent{
			Label:  "Glass Shattered!",
			Detail: fmt.Sprintf("x%.2f Mult lost from missed Volatile word", xMult),
		})
		xMult = 1.0
	}

	finalMult := math.Round(mult*xMult*10) / 10
	res.Chips = chips
	res.Mult = finalMult
	res.TotalScore = int(math.Round(float64(chips) * finalMult))
	return res
}

func allEqualSuit(s []RadicalSuit) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}

func allEqualInt(s []int) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}

func isStrictlyAscending(s []int) bool {
	for i := 1; i < len(s); i++ {
		if s[i] <= s[i-1] {
			return false
		}
	}
	return true
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
