package reading

import (
	"path/filepath"
	"testing"
)

func sampleSentence() Sentence {
	return Sentence{
		ID:            "test-1",
		English:       "In the beginning was the Word, and the Word was with God, and the Word was God.",
		NaturalTarget: "太初有道，道与神同在，道就是神。",
		TargetPinyin:  "Tàichū yǒu Dào, Dào yǔ Shén tóngzài, Dào jiùshì Shén.",
		GrammarNote:   "In Chinese biblical translation, Logos is translated as 道 (Dào). Prepositional phrase 与神同在 (with God together-exist) precedes the predicate.",
		Words: []WordTranslation{
			{
				English:     "beginning",
				Target:      "太初",
				Pinyin:      "tàichū",
				Meaning:     "the primordial beginning",
				Explanation: "太 (greatest/supreme) + 初 (beginning).",
				Language:    "Mandarin",
				Difficulty:  2,
			},
			{
				English:     "Word",
				Target:      "道",
				Pinyin:      "dào",
				Meaning:     "the Word / Logos / the Way",
				Explanation: "The Word/Logos.",
				Language:    "Mandarin",
				Difficulty:  1,
			},
			{
				English:     "God",
				Target:      "神",
				Pinyin:      "shén",
				Meaning:     "God",
				Explanation: "Deity/God.",
				Language:    "Mandarin",
				Difficulty:  1,
			},
		},
	}
}

func sampleText() *Text {
	return &Text{
		ID:          "john-test",
		Title:       "Gospel of John Prologue",
		Language:    "Mandarin",
		Description: "Prologue test",
		Category:    "Scripture",
		Sentences: []Sentence{
			sampleSentence(),
			{
				ID:            "test-2",
				English:       "He was in the beginning with God.",
				NaturalTarget: "这道太初与神同在。",
				TargetPinyin:  "Zhè Dào tàichū yǔ Shén tóngzài.",
				GrammarNote:   "Notice the subject-time-predicate order in Chinese.",
				Words: []WordTranslation{
					{
						English:  "beginning",
						Target:   "太初",
						Pinyin:   "tàichū",
						Meaning:  "beginning",
						Language: "Mandarin",
					},
					{
						English:  "God",
						Target:   "神",
						Pinyin:   "shén",
						Meaning:  "God",
						Language: "Mandarin",
					},
				},
			},
		},
	}
}

func TestWeaveSentenceAggressiveness(t *testing.T) {
	sent := sampleSentence()
	progress := DefaultProgress()

	// Level 0: Pure English
	w0 := WeaveSentence(sent, progress, 0)
	if len(w0.SubstitutedWords) != 0 {
		t.Errorf("Expected 0 substitutions at aggressiveness 0, got %d", len(w0.SubstitutedWords))
	}
	if len(w0.Segments) != 1 || w0.Segments[0].Text != sent.English {
		t.Errorf("Expected single original text segment at level 0")
	}

	// Level 1: Gentle (around 1 word)
	w1 := WeaveSentence(sent, progress, 1)
	if len(w1.SubstitutedWords) == 0 {
		t.Errorf("Expected at least 1 substitution at level 1")
	}

	// Level 4: Full Immersion (all words)
	w4 := WeaveSentence(sent, progress, 4)
	if len(w4.SubstitutedWords) < 3 {
		t.Errorf("Expected all 3 words substituted at level 4, got %d", len(w4.SubstitutedWords))
	}

	// Check that keys are mapped
	for _, sw := range w4.SubstitutedWords {
		if sw.Key == "" {
			t.Errorf("Substituted word %s missing shortcut key", sw.Translation.Target)
		}
		if _, ok := w4.KeyToWord[sw.Key]; !ok {
			t.Errorf("KeyToWord missing key %s", sw.Key)
		}
	}
}

func TestReadingProgressTracking(t *testing.T) {
	tmpDir := t.TempDir()
	progFile := filepath.Join(tmpDir, "progress.yaml")

	p, err := LoadProgress(progFile)
	if err != nil {
		t.Fatalf("Failed to init progress: %v", err)
	}

	word := WordTranslation{
		English:  "beginning",
		Target:   "太初",
		Language: "Mandarin",
		Pinyin:   "tàichū",
	}

	// Initially unknown
	if status := p.GetWordStatus("Mandarin", "太初"); status != StatusUnknown {
		t.Errorf("Expected status unknown, got %v", status)
	}

	// User requests help
	p.RecordWordHelpRequested(word)
	if status := p.GetWordStatus("Mandarin", "太初"); status != StatusLearning {
		t.Errorf("Expected status learning after help requested, got %v", status)
	}

	// User recalls successfully multiple times
	p.RecordSentenceSuccess([]WordTranslation{word})
	p.RecordSentenceSuccess([]WordTranslation{word})
	p.RecordSentenceSuccess([]WordTranslation{word})

	if status := p.GetWordStatus("Mandarin", "太初"); status != StatusKnown {
		t.Errorf("Expected status known after successful recalls, got %v", status)
	}

	// Save and reload
	p.SetBookmark("john-test", 1)
	p.Aggressiveness = 3
	if err := p.Save(); err != nil {
		t.Fatalf("Failed to save progress: %v", err)
	}

	loaded, err := LoadProgress(progFile)
	if err != nil {
		t.Fatalf("Failed to reload progress: %v", err)
	}
	if loaded.GetBookmark("john-test") != 1 {
		t.Errorf("Bookmark not preserved: got %d", loaded.GetBookmark("john-test"))
	}
	if loaded.Aggressiveness != 3 {
		t.Errorf("Aggressiveness not preserved: got %d", loaded.Aggressiveness)
	}
	if loaded.GetWordStatus("Mandarin", "太初") != StatusKnown {
		t.Errorf("Word status not preserved after reload")
	}
}

func TestSessionFlow(t *testing.T) {
	text := sampleText()
	progress := DefaultProgress()

	s := NewSession(text, progress)
	if s.SentenceIdx != 0 {
		t.Errorf("Expected start at sentence 0, got %d", s.SentenceIdx)
	}
	if len(s.CurrentWeave.Segments) == 0 {
		t.Errorf("Expected non-empty weave segments")
	}

	// Inspect a word
	if len(s.CurrentWeave.SubstitutedWords) > 0 {
		firstKey := s.CurrentWeave.SubstitutedWords[0].Key
		sw, found := s.SelectWordByKey(firstKey)
		if !found || sw == nil {
			t.Fatalf("Failed to select word by key %s", firstKey)
		}
		if s.SelectedWord == nil {
			t.Fatalf("SelectedWord is nil")
		}
		s.CloseWordDetail()
		if s.SelectedWord != nil {
			t.Fatalf("SelectedWord should be nil after closing")
		}
	}

	// Cycle aggressiveness
	initialAgg := s.Aggressiveness
	s.CycleAggressiveness(1)
	if s.Aggressiveness != initialAgg+1 && initialAgg < 4 {
		t.Errorf("Aggressiveness cycle up failed")
	}
	s.CycleAggressiveness(-1)
	if s.Aggressiveness != initialAgg {
		t.Errorf("Aggressiveness cycle down failed")
	}

	// Advance to sentence 2
	advanced := s.Next()
	if !advanced || s.SentenceIdx != 1 {
		t.Errorf("Expected advance to sentence 1")
	}

	// Advance past end
	advanced = s.Next()
	if advanced || !s.SessionCompleted {
		t.Errorf("Expected session completion at end")
	}

	// Step back
	stepped := s.Prev()
	if !stepped || s.SentenceIdx != 0 {
		t.Errorf("Expected stepped back to 0")
	}
}

func TestSaveAndLoadText(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "sample.yaml")

	original := sampleText()
	if err := SaveText(filePath, original); err != nil {
		t.Fatalf("SaveText failed: %v", err)
	}

	loaded, err := LoadText(filePath)
	if err != nil {
		t.Fatalf("LoadText failed: %v", err)
	}

	if loaded.Title != original.Title {
		t.Errorf("Expected title %s, got %s", original.Title, loaded.Title)
	}
	if len(loaded.Sentences) != len(original.Sentences) {
		t.Errorf("Expected %d sentences, got %d", len(original.Sentences), len(loaded.Sentences))
	}
}

func TestSplitIntoSentences(t *testing.T) {
	text := "That which was from the beginning, which we have heard. This is the message we have heard from him! And we proclaim to you."
	sentences := SplitIntoSentences(text)
	if len(sentences) != 3 {
		t.Errorf("Expected 3 sentences, got %d: %v", len(sentences), sentences)
	}
}

func TestListRealTexts(t *testing.T) {
	headers, err := ListTexts("../../readingTranslationTexts", nil)
	if err != nil {
		t.Fatalf("ListTexts failed: %v", err)
	}
	if len(headers) < 2 {
		t.Fatalf("Expected at least 2 texts (1John and HarryPotterCh1), got %d", len(headers))
	}
	found1John := false
	foundHP := false
	for _, h := range headers {
		if h.ID == "1John" {
			found1John = true
			if h.SentenceCount == 0 {
				t.Errorf("1John has 0 sentences")
			}
		}
		if h.ID == "HarryPotterCh1" {
			foundHP = true
			if h.SentenceCount == 0 {
				t.Errorf("HarryPotterCh1 has 0 sentences")
			}
		}
	}
	if !found1John {
		t.Errorf("1John not found in headers")
	}
	if !foundHP {
		t.Errorf("HarryPotterCh1 not found in headers")
	}
}
