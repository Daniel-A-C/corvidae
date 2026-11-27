package forge

const (
	// Tone Archetype Relics
	RelicFourthToneEcho RelicID = "fourth_tone_echo"
	RelicToneLadder     RelicID = "tone_ladder"
	RelicRisingChorus   RelicID = "rising_chorus"
	RelicHarmonicPair   RelicID = "harmonic_pair"

	// Radical Suit Archetype Relics
	RelicNatureFlow     RelicID = "nature_flow"
	RelicHumanResonance RelicID = "human_resonance"
	RelicCraftAnvil     RelicID = "craft_anvil"
	RelicRadicalPurist  RelicID = "radical_purist"

	// Structure & Pivot Chain Relics
	RelicPivotMaster RelicID = "pivot_master"
	RelicTwinForge   RelicID = "twin_forge"
	RelicTriadSmith  RelicID = "triad_smith"
	RelicLoneWolf    RelicID = "lone_wolf"

	// SM-2 Memory & Economy Relics
	RelicVolatileScholar RelicID = "volatile_scholar"
	RelicGlassCannon     RelicID = "glass_cannon"
	RelicOldReliable     RelicID = "old_reliable"
	RelicForgivingTutor  RelicID = "forgiving_tutor"
	RelicDeepMemory      RelicID = "deep_memory"
	RelicMerchantsInk    RelicID = "merchants_ink"
)

// RelicCatalog contains all 18 Relics available in Morpheme Forge.
var RelicCatalog = []RelicDef{
	// Tone builds
	{ID: RelicFourthToneEcho, Name: "4th-Tone Echo", Desc: "Each 4th-tone Hanzi played grants +5 Mult", Cost: 5, Rarity: "Common"},
	{ID: RelicToneLadder, Name: "Tone Ladder", Desc: "Tone Escalation grants an extra x1.5 Mult", Cost: 7, Rarity: "Rare"},
	{ID: RelicRisingChorus, Name: "Rising Chorus", Desc: "Each 2nd-tone Hanzi played grants +35 Chips", Cost: 5, Rarity: "Common"},
	{ID: RelicHarmonicPair, Name: "Harmonic Pair", Desc: "Each word with matching tones (e.g. 4-4) grants +8 Mult", Cost: 6, Rarity: "Uncommon"},

	// Radical Suit builds
	{ID: RelicNatureFlow, Name: "Nature Flow", Desc: "Each Nature (水) Hanzi played grants +30 Chips", Cost: 5, Rarity: "Common"},
	{ID: RelicHumanResonance, Name: "Human Resonance", Desc: "Each Human (人) Hanzi played grants +4 Mult", Cost: 5, Rarity: "Common"},
	{ID: RelicCraftAnvil, Name: "Craft Anvil", Desc: "Each Craft (工) Hanzi played grants +25 Chips & +2 Mult", Cost: 6, Rarity: "Uncommon"},
	{ID: RelicRadicalPurist, Name: "Radical Purist", Desc: "Radical Flush multiplies total Mult by x1.75", Cost: 8, Rarity: "Rare"},

	// Structure & Pivot Chain builds
	{ID: RelicPivotMaster, Name: "Pivot Master", Desc: "Pivot Chains & Grand Chains gain +60 Chips & x1.5 Mult", Cost: 8, Rarity: "Rare"},
	{ID: RelicTwinForge, Name: "Twin Anvil", Desc: "Double Forges gain +70 Chips & +6 Mult", Cost: 6, Rarity: "Uncommon"},
	{ID: RelicTriadSmith, Name: "Triad Smith", Desc: "3-char & 4-char words grant x1.6 Mult", Cost: 7, Rarity: "Rare"},
	{ID: RelicLoneWolf, Name: "Lone Wolf", Desc: "Solo 1-char plays gain +45 Chips & +4 Mult", Cost: 4, Rarity: "Common"},

	// SM-2 Memory & Economy builds
	{ID: RelicVolatileScholar, Name: "Volatile Scholar", Desc: "Each Glass (New/Leech) Hanzi adds +20 Chips & x1.25 extra", Cost: 7, Rarity: "Rare"},
	{ID: RelicGlassCannon, Name: "Glass Cannon", Desc: "Volatile Discovery compounds grant x2.0 instead of x1.5", Cost: 7, Rarity: "Rare"},
	{ID: RelicOldReliable, Name: "Old Reliable", Desc: "Each Gold (Mature) Hanzi earns an extra +$2 when proven", Cost: 5, Rarity: "Common"},
	{ID: RelicForgivingTutor, Name: "Forgiving Tutor", Desc: "Forgives your first missed Prove check each Blind", Cost: 6, Rarity: "Uncommon"},
	{ID: RelicDeepMemory, Name: "Deep Memory", Desc: "Foil (Learning) Hanzi grant +30 extra Chips", Cost: 5, Rarity: "Common"},
	{ID: RelicMerchantsInk, Name: "Merchant's Ink", Desc: "+2 Chips for every $1 you currently hold", Cost: 6, Rarity: "Uncommon"},
}

// GetRelic looks up a RelicDef by ID.
func GetRelic(id RelicID) (RelicDef, bool) {
	for _, r := range RelicCatalog {
		if r.ID == id {
			return r, true
		}
	}
	return RelicDef{}, false
}

const (
	BossNone          BossBlindID = ""
	BossToneTrial     BossBlindID = "tone_trial"
	BossPinyinVeil    BossBlindID = "pinyin_veil"
	BossComponentTrap BossBlindID = "component_trap"
	BossSilencedSuit  BossBlindID = "silenced_suit"
	BossHeavyForge    BossBlindID = "heavy_forge"
	BossBlindMaterial BossBlindID = "blind_material"
	BossTightPurse    BossBlindID = "tight_purse"
	BossHighBar       BossBlindID = "high_bar"
)

// BossBlindCatalog contains all 8 Boss Blinds.
var BossBlindCatalog = []BossBlindDef{
	{ID: BossToneTrial, Name: "The Tone Trial", Desc: "Prove checks test Tone Patterns (e.g. 4-4) instead of Meaning"},
	{ID: BossPinyinVeil, Name: "The Pinyin Veil", Desc: "Prove checks test Pinyin instead of Meaning"},
	{ID: BossComponentTrap, Name: "The Component Trap", Desc: "Distractors prioritize words sharing a Hanzi with the target"},
	{ID: BossSilencedSuit, Name: "The Silenced Suit", Desc: "One Radical Suit grants 0 Chips & no bonuses this Blind"},
	{ID: BossHeavyForge, Name: "The Heavy Forge", Desc: "Solo 1-char words score 0 — must forge 2+ char compounds"},
	{ID: BossBlindMaterial, Name: "The Fog of Memory", Desc: "Hides SM-2 Material badges on your hand cards"},
	{ID: BossTightPurse, Name: "The Tight Purse", Desc: "Only 2 Discards allowed this Blind"},
	{ID: BossHighBar, Name: "The High Bar", Desc: "1.5x Target Score, but you gain +1 Hand"},
}

// GetBossBlind looks up a BossBlindDef by ID.
func GetBossBlind(id BossBlindID) (BossBlindDef, bool) {
	for _, b := range BossBlindCatalog {
		if b.ID == id {
			return b, true
		}
	}
	return BossBlindDef{}, false
}
