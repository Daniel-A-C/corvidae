package ui

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"flashcards/internal/deck"
)

// SameCard checks whether two flashcards represent the same vocabulary card based on character and meaning.
func SameCard(a, b deck.Flashcard) bool {
	if a.Character != b.Character || a.Meaning != b.Meaning {
		return false
	}
	if a.Pinyin != "" && b.Pinyin != "" && a.Pinyin != b.Pinyin {
		return false
	}
	if a.Pronunciation != "" && b.Pronunciation != "" && a.Pronunciation != b.Pronunciation {
		return false
	}
	return true
}

func (m Model) getMarkedDeckPath() string {
	return filepath.Join(m.BaseDir, "marked.yaml")
}

// DetectCardLanguage detects or infers the language of a flashcard.
func (m Model) DetectCardLanguage(card deck.Flashcard) string {
	if card.Language != "" {
		return card.Language
	}
	if card.Deck != "" {
		parts := strings.Split(filepath.ToSlash(card.Deck), "/")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}
	if card.Pinyin != "" {
		return "Mandarin"
	}
	// Check for Han characters
	for _, r := range card.Character {
		if unicode.Is(unicode.Han, r) {
			return "Mandarin"
		}
	}
	// Check for Arabic script characters
	for _, r := range card.Character {
		if unicode.Is(unicode.Arabic, r) {
			return "Arabic"
		}
	}
	// If still not identified, search baseDir to match card against existing decks
	if m.BaseDir != "" {
		dirs, err := deck.GetDeckDirectories(m.BaseDir)
		if err == nil {
			for _, langDir := range dirs {
				files, err := deck.GetAllDeckFiles(m.BaseDir, langDir)
				if err != nil {
					continue
				}
				for _, f := range files {
					fullPath := filepath.Join(m.BaseDir, langDir, f)
					d, err := deck.LoadDeck(fullPath)
					if err != nil {
						continue
					}
					for _, c := range d.Cards {
						if SameCard(c, card) {
							return langDir
						}
					}
				}
			}
		}
	}
	return ""
}

// LoadMarkedDeck loads marked cards from marked.yaml if it exists.
func (m *Model) LoadMarkedDeck() {
	markedPath := m.getMarkedDeckPath()
	d, err := deck.LoadDeck(markedPath)
	if err == nil && len(d.Cards) > 0 {
		for i := range d.Cards {
			if d.Cards[i].Language == "" {
				d.Cards[i].Language = m.DetectCardLanguage(d.Cards[i])
			}
		}
		m.MarkedCards = d.Cards
	} else {
		m.MarkedCards = nil
	}
}

// saveMarkedDeck saves the current MarkedCards slice to marked.yaml, or deletes the file if empty.
func (m *Model) saveMarkedDeck() error {
	markedPath := m.getMarkedDeckPath()
	if len(m.MarkedCards) == 0 {
		_ = os.Remove(markedPath)
		return nil
	}
	return deck.SaveDeck(markedPath, deck.Deck{Cards: m.MarkedCards})
}

// isCardMarked reports whether the given card is in MarkedCards.
func (m Model) isCardMarked(card deck.Flashcard) bool {
	for _, c := range m.MarkedCards {
		if SameCard(c, card) {
			return true
		}
	}
	return false
}

// markCard adds a card to MarkedCards and saves to disk.
func (m *Model) markCard(card deck.Flashcard) {
	m.markCardWithSource(card, "")
}

// markCardWithSource adds a card to MarkedCards preserving source file metadata and saves to disk.
func (m *Model) markCardWithSource(card deck.Flashcard, sourceFilename string) {
	if m.isCardMarked(card) {
		return
	}
	markedCard := deck.Flashcard{
		Character:     card.Character,
		Pinyin:        card.Pinyin,
		Pronunciation: card.Pronunciation,
		Meaning:       card.Meaning,
		Explanation:   card.Explanation,
		Language:      card.Language,
		Deck:          card.Deck,
	}

	if sourceFilename != "" && sourceFilename != m.getMarkedDeckPath() {
		relPath, err := filepath.Rel(m.BaseDir, sourceFilename)
		if err == nil {
			parts := strings.Split(filepath.ToSlash(relPath), "/")
			if len(parts) > 0 && markedCard.Language == "" {
				markedCard.Language = parts[0]
			}
			if markedCard.Deck == "" {
				markedCard.Deck = filepath.ToSlash(relPath)
			}
		}
	}
	if markedCard.Language == "" {
		markedCard.Language = m.DetectCardLanguage(card)
	}

	m.MarkedCards = append(m.MarkedCards, markedCard)
	_ = m.saveMarkedDeck()
}

// unmarkCard removes a card from MarkedCards and saves to disk.
func (m *Model) unmarkCard(card deck.Flashcard) bool {
	found := false
	var updated []deck.Flashcard
	for _, c := range m.MarkedCards {
		if SameCard(c, card) {
			found = true
			continue
		}
		updated = append(updated, c)
	}
	if found {
		m.MarkedCards = updated
		_ = m.saveMarkedDeck()
	}
	return found
}

// toggleCardMarked toggles whether a card is marked.
func (m *Model) toggleCardMarked(card deck.Flashcard) bool {
	return m.toggleCardMarkedWithSource(card, "")
}

// toggleCardMarkedWithSource toggles whether a card is marked, tracking source deck file.
func (m *Model) toggleCardMarkedWithSource(card deck.Flashcard, sourceFilename string) bool {
	if m.isCardMarked(card) {
		m.unmarkCard(card)
		return false
	}
	m.markCardWithSource(card, sourceFilename)
	return true
}

// getCurrentQuizCardRef returns the active card reference.
func (m Model) getCurrentQuizCardRef() (deck.CardRef, bool) {
	if m.CurrentIndex >= len(m.ActiveCards) {
		return deck.CardRef{}, false
	}
	return m.ActiveCards[m.CurrentIndex], true
}

// getCurrentQuizCard returns the flashcard currently displayed in quiz mode.
func (m Model) getCurrentQuizCard() (deck.Flashcard, bool) {
	ref, ok := m.getCurrentQuizCardRef()
	if !ok {
		return deck.Flashcard{}, false
	}
	d, ok := m.Decks[ref.Filename]
	if !ok || ref.OrigIdx >= len(d.Cards) {
		return deck.Flashcard{}, false
	}
	return d.Cards[ref.OrigIdx], true
}

// toggleCurrentCardMarked toggles the marked status of the active quiz card.
func (m *Model) toggleCurrentCardMarked() bool {
	ref, ok := m.getCurrentQuizCardRef()
	if !ok {
		return false
	}
	d, ok := m.Decks[ref.Filename]
	if !ok || ref.OrigIdx >= len(d.Cards) {
		return false
	}
	card := d.Cards[ref.OrigIdx]
	return m.toggleCardMarkedWithSource(card, ref.Filename)
}

// unmarkCurrentCard unmarks the active quiz card.
func (m *Model) unmarkCurrentCard() bool {
	card, ok := m.getCurrentQuizCard()
	if !ok {
		return false
	}
	return m.unmarkCard(card)
}

// isPlayingMarkedDeck reports whether the currently running session is for the marked deck.
func (m Model) isPlayingMarkedDeck() bool {
	if m.PlayingMarkedDeck {
		return true
	}
	markedPath := m.getMarkedDeckPath()
	return m.SelectedFiles[markedPath]
}

// startMarkedPractice starts a practice session (quiz or review) using the marked cards deck.
func (m *Model) startMarkedPractice() (tea.Model, tea.Cmd) {
	m.LoadMarkedDeck()
	if len(m.MarkedCards) == 0 {
		m.StatusMessage = "No marked cards yet! During quiz mode, press [m] to mark cards."
		return *m, nil
	}
	m.StatusMessage = ""
	markedPath := m.getMarkedDeckPath()
	m.Decks = map[string]deck.Deck{
		markedPath: {Cards: m.MarkedCards},
	}
	m.SelectedFiles = map[string]bool{
		markedPath: true,
	}
	m.PlayingMarkedDeck = true

	if m.State == StateDirSelect && m.Mode == ModeReview {
		m.SetupReview()
		m.State = StateReview
	} else {
		m.Mode = ModeQuiz
		m.SetupQuiz()
		m.State = StateQuiz
	}
	return *m, nil
}

// getLanguageDistractorCards gathers additional flashcards matching the given language,
// prioritizing the preferred deck/folder if available.
func (m Model) getLanguageDistractorCards(lang, preferredDeck string, needed int) []deck.Flashcard {
	var extra []deck.Flashcard
	markedName := filepath.Base(m.getMarkedDeckPath())
	loadedSet := make(map[string]bool)

	// 1. If preferredDeck is provided and exists, load cards from it first
	if preferredDeck != "" {
		prefPath := filepath.Join(m.BaseDir, preferredDeck)
		if d, err := deck.LoadDeck(prefPath); err == nil {
			loadedSet[prefPath] = true
			for _, c := range d.Cards {
				extra = append(extra, c)
				if len(extra) >= needed+15 {
					return extra
				}
			}
		}
	}

	// 2. Find matching language directory
	langDir := lang
	dirs, err := deck.GetDeckDirectories(m.BaseDir)
	if err == nil {
		for _, d := range dirs {
			if strings.EqualFold(d, lang) {
				langDir = d
				break
			}
		}
	}

	deckFiles, err := deck.GetAllDeckFiles(m.BaseDir, langDir)
	if err != nil || len(deckFiles) == 0 {
		return extra
	}

	// 3. Load cards from the language directory decks
	for _, file := range deckFiles {
		fullPath := filepath.Join(m.BaseDir, langDir, file)
		if loadedSet[fullPath] || file == markedName || filepath.Base(file) == markedName {
			continue
		}
		d, err := deck.LoadDeck(fullPath)
		if err != nil {
			continue
		}
		for _, c := range d.Cards {
			extra = append(extra, c)
			if len(extra) >= needed+15 {
				return extra
			}
		}
	}

	return extra
}

// getExtraDistractorCards gathers additional flashcards from any available deck as a final fallback.
func (m Model) getExtraDistractorCards(needed int) []deck.Flashcard {
	var extra []deck.Flashcard
	deckFiles, err := deck.GetAllDeckFiles(m.BaseDir, "")
	if err != nil {
		return nil
	}
	markedName := filepath.Base(m.getMarkedDeckPath())
	for _, file := range deckFiles {
		if file == markedName || filepath.Base(file) == markedName {
			continue
		}
		fullPath := filepath.Join(m.BaseDir, file)
		d, err := deck.LoadDeck(fullPath)
		if err != nil {
			continue
		}
		for _, c := range d.Cards {
			extra = append(extra, c)
			if len(extra) >= needed+10 {
				return extra
			}
		}
	}
	return extra
}
