package arabic

// PositionalForms contains the 4 cursive shapes for an Arabic letter.
type PositionalForms struct {
	Isolated string
	Initial  string
	Medial   string
	Final    string
}

// FormAt returns the string form for the given position name:
// "isolated", "initial", "medial", "final".
func (p PositionalForms) FormAt(position string) string {
	switch position {
	case "isolated":
		return p.Isolated
	case "initial":
		return p.Initial
	case "medial":
		return p.Medial
	case "final":
		return p.Final
	default:
		return p.Isolated
	}
}

// ExampleWord illustrates a letter's usage in a real Arabic word.
type ExampleWord struct {
	Arabic      string
	Translit    string
	Meaning     string
	Position    string // "initial", "medial", "final", or "isolated"
	Highlight   string // the letter in context
}

// Letter represents an Arabic alphabet character with linguistic,
// phonetic, and pedagogical properties.
type Letter struct {
	ID           int
	ArabicName   string // e.g. "بَاء"
	EnglishName  string // e.g. "Baa"
	Translit     string // e.g. "b"
	IPA          string // e.g. "/b/"
	SoundDesc    string // Pronunciation description & articulation point
	Forms        PositionalForms
	IsConnector  bool   // true if connects both sides; false if non-connecting (right-joining only)
	IsSunLetter  bool   // true for Sun (Shamsi), false for Moon (Qamari)
	Emphatic     bool   // true for velarized/pharyngealized consonants (ص, ض, ط, ظ)
	Confusable   []int  // IDs of visually or phonetically confusable letters
	PedGroup     int    // Pedagogical stage (1-8)
	GroupName    string // Name of pedagogical group
	Tips         string // Memory mnemonic or articulation tip
	Examples     []ExampleWord
}

// AuxiliaryItem represents a special mark, ligature, or short vowel diacritic.
type AuxiliaryItem struct {
	Arabic      string
	Name        string
	Category    string // "Special Character" or "Harakat (Diacritic)"
	Description string
	Example     string
	Translit    string
}

// Stage represents a pedagogical group of letters for progressive learning.
type Stage struct {
	Number      int
	Name        string
	Description string
	LetterIDs   []int
}
