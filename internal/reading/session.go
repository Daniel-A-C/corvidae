package reading

// Session manages the interactive state of reading a text.
type Session struct {
	Text             *Text
	SentenceIdx      int
	Progress         *ReadingProgress
	Aggressiveness   int
	CurrentWeave     WeavedSentence
	SelectedWord     *SubstitutedWord
	ShowDiscussion   bool
	AutoShowDiscuss  bool // whether discussion appears automatically between sentences
	InspectedKeys    map[string]bool
	SessionCompleted bool
}

// NewSession initializes a reading session for a text with loaded user progress.
func NewSession(text *Text, progress *ReadingProgress) *Session {
	if progress == nil {
		progress = DefaultProgress()
	}

	startIdx := progress.GetBookmark(text.ID)
	if startIdx < 0 || startIdx >= len(text.Sentences) {
		startIdx = 0
	}

	s := &Session{
		Text:            text,
		SentenceIdx:     startIdx,
		Progress:        progress,
		Aggressiveness:  progress.Aggressiveness,
		AutoShowDiscuss: true,
		InspectedKeys:   make(map[string]bool),
	}

	s.updateCurrentWeave()
	return s
}

func (s *Session) updateCurrentWeave() {
	if s.Text == nil || len(s.Text.Sentences) == 0 || s.SentenceIdx >= len(s.Text.Sentences) {
		s.CurrentWeave = WeavedSentence{}
		return
	}
	sent := s.Text.Sentences[s.SentenceIdx]
	s.CurrentWeave = WeaveSentence(sent, s.Progress, s.Aggressiveness)
	s.SelectedWord = nil
	s.InspectedKeys = make(map[string]bool)
}

// CurrentSentence returns the currently active Sentence.
func (s *Session) CurrentSentence() (Sentence, bool) {
	if s.Text == nil || s.SentenceIdx >= len(s.Text.Sentences) {
		return Sentence{}, false
	}
	return s.Text.Sentences[s.SentenceIdx], true
}

// SelectWordByKey looks up a substituted word by its shortcut key and records inspection.
func (s *Session) SelectWordByKey(key string) (*SubstitutedWord, bool) {
	word, found := s.CurrentWeave.KeyToWord[key]
	if !found {
		return nil, false
	}

	s.SelectedWord = &word
	s.InspectedKeys[key] = true

	// Record that user requested help on this word
	if s.Progress != nil {
		s.Progress.RecordWordHelpRequested(word.Translation)
		_ = s.Progress.Save()
	}

	return s.SelectedWord, true
}

// CloseWordDetail closes the active word detail popup.
func (s *Session) CloseWordDetail() {
	s.SelectedWord = nil
}

// MarkSelectedWordKnown marks the currently selected word as known.
func (s *Session) MarkSelectedWordKnown() {
	if s.SelectedWord != nil && s.Progress != nil {
		s.Progress.RecordWordMastered(s.SelectedWord.Translation)
		_ = s.Progress.Save()
	}
}

// FinishSentence updates mastery for uninspected substituted words and prepares discussion or advances.
func (s *Session) FinishSentence() {
	if s.Progress != nil && len(s.CurrentWeave.SubstitutedWords) > 0 {
		var unassisted []WordTranslation
		for _, sw := range s.CurrentWeave.SubstitutedWords {
			if !s.InspectedKeys[sw.Key] {
				unassisted = append(unassisted, sw.Translation)
			}
		}
		if len(unassisted) > 0 {
			s.Progress.RecordSentenceSuccess(unassisted)
		}
		s.Progress.SetBookmark(s.Text.ID, s.SentenceIdx)
		_ = s.Progress.Save()
	}
}

// Next advances to the next sentence, bookmarking progress.
func (s *Session) Next() bool {
	s.FinishSentence()
	s.ShowDiscussion = false
	s.SelectedWord = nil

	if s.SentenceIdx+1 < len(s.Text.Sentences) {
		s.SentenceIdx++
		s.updateCurrentWeave()
		if s.Progress != nil {
			s.Progress.SetBookmark(s.Text.ID, s.SentenceIdx)
			_ = s.Progress.Save()
		}
		return true
	}

	s.SessionCompleted = true
	return false
}

// Prev steps back to the previous sentence.
func (s *Session) Prev() bool {
	s.ShowDiscussion = false
	s.SelectedWord = nil

	if s.SentenceIdx > 0 {
		s.SentenceIdx--
		s.updateCurrentWeave()
		if s.Progress != nil {
			s.Progress.SetBookmark(s.Text.ID, s.SentenceIdx)
			_ = s.Progress.Save()
		}
		return true
	}
	return false
}

// ToggleDiscussion toggles the visibility of the grammar discussion box.
func (s *Session) ToggleDiscussion() {
	s.ShowDiscussion = !s.ShowDiscussion
	if s.ShowDiscussion {
		s.SelectedWord = nil
	}
}

// CycleAggressiveness adjusts the aggressiveness level by delta (-1 or +1) and reweaves.
func (s *Session) CycleAggressiveness(delta int) {
	newLevel := s.Aggressiveness + delta
	if newLevel < 0 {
		newLevel = 0
	}
	if newLevel > 4 {
		newLevel = 4
	}
	if newLevel != s.Aggressiveness {
		s.Aggressiveness = newLevel
		if s.Progress != nil {
			s.Progress.Aggressiveness = newLevel
			_ = s.Progress.Save()
		}
		s.updateCurrentWeave()
	}
}

// SetAggressiveness directly sets the aggressiveness level (0 to 4).
func (s *Session) SetAggressiveness(level int) {
	if level < 0 {
		level = 0
	}
	if level > 4 {
		level = 4
	}
	s.Aggressiveness = level
	if s.Progress != nil {
		s.Progress.Aggressiveness = level
		_ = s.Progress.Save()
	}
	s.updateCurrentWeave()
}

// Restart restarts the current text from sentence 0.
func (s *Session) Restart() {
	s.SentenceIdx = 0
	s.SessionCompleted = false
	s.ShowDiscussion = false
	s.SelectedWord = nil
	if s.Progress != nil {
		s.Progress.SetBookmark(s.Text.ID, 0)
		_ = s.Progress.Save()
	}
	s.updateCurrentWeave()
}
