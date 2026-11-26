// Package stats tracks daily study activity across Corvidae modes and computes dashboard analytics and heatmaps.
package stats

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"flashcards/internal/deck"
	"gopkg.in/yaml.v3"
)

const DefaultStatsPath = "stats.yaml"

// DailyActivity captures study counts for a single calendar date (YYYY-MM-DD).
type DailyActivity struct {
	Date             string `yaml:"date"`
	Reviews          int    `yaml:"reviews,omitempty"`
	QuizAnswered     int    `yaml:"quiz_answered,omitempty"`
	QuizCorrect      int    `yaml:"quiz_correct,omitempty"`
	MatchPairs       int    `yaml:"match_pairs,omitempty"`
	MemorizeWords    int    `yaml:"memorize_words,omitempty"`
	ReadingSentences int    `yaml:"reading_sentences,omitempty"`
	ArabicDrills     int    `yaml:"arabic_drills,omitempty"`
}

// Total returns the combined count of study actions logged on this day.
func (d *DailyActivity) Total() int {
	if d == nil {
		return 0
	}
	return d.Reviews + d.QuizAnswered + d.MatchPairs + d.MemorizeWords + d.ReadingSentences + d.ArabicDrills
}

// Store persists daily study activity indexed by YYYY-MM-DD date string.
type Store struct {
	Days     map[string]*DailyActivity `yaml:"days"`
	FilePath string                    `yaml:"-"`
}

// NewEmpty returns an empty Store pointing at filePath.
func NewEmpty(filePath string) *Store {
	if filePath == "" {
		filePath = DefaultStatsPath
	}
	return &Store{
		Days:     make(map[string]*DailyActivity),
		FilePath: filePath,
	}
}

// Load reads study statistics from filePath. If missing, an empty Store is returned without error.
func Load(filePath string) (*Store, error) {
	if filePath == "" {
		filePath = DefaultStatsPath
	}
	s := NewEmpty(filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, fmt.Errorf("failed to read stats %s: %w", filePath, err)
	}

	if err := yaml.Unmarshal(data, s); err != nil {
		return s, fmt.Errorf("failed to parse stats %s: %w", filePath, err)
	}
	if s.Days == nil {
		s.Days = make(map[string]*DailyActivity)
	}
	for k, v := range s.Days {
		if v != nil && v.Date == "" {
			v.Date = k
		}
	}
	s.FilePath = filePath
	return s, nil
}

// Save writes the Store to disk.
func (s *Store) Save() error {
	if s == nil || s.FilePath == "" {
		return nil
	}
	dir := filepath.Dir(s.FilePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create stats directory: %w", err)
		}
	}
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}
	return os.WriteFile(s.FilePath, data, 0644)
}

func (s *Store) dayEntry(now time.Time) *DailyActivity {
	if s.Days == nil {
		s.Days = make(map[string]*DailyActivity)
	}
	if now.IsZero() {
		now = time.Now()
	}
	key := now.Format("2006-01-02")
	entry, ok := s.Days[key]
	if !ok || entry == nil {
		entry = &DailyActivity{Date: key}
		s.Days[key] = entry
	}
	return entry
}

// RecordReview logs one SM-2 card review and saves.
func (s *Store) RecordReview(now time.Time) {
	if s == nil {
		return
	}
	s.dayEntry(now).Reviews++
	_ = s.Save()
}

// RecordQuiz logs one quiz answer and saves.
func (s *Store) RecordQuiz(correct bool, now time.Time) {
	if s == nil {
		return
	}
	e := s.dayEntry(now)
	e.QuizAnswered++
	if correct {
		e.QuizCorrect++
	}
	_ = s.Save()
}

// RecordQuizOverride adjusts today's quiz correct count when the user presses "I knew it".
func (s *Store) RecordQuizOverride(now time.Time) {
	if s == nil {
		return
	}
	e := s.dayEntry(now)
	if e.QuizCorrect < e.QuizAnswered {
		e.QuizCorrect++
	}
	_ = s.Save()
}

// RecordMatchPair logs one matched pair and saves.
func (s *Store) RecordMatchPair(now time.Time) {
	if s == nil {
		return
	}
	s.dayEntry(now).MatchPairs++
	_ = s.Save()
}

// RecordMemorizeWord logs one correctly recalled word in Memorize mode and saves.
func (s *Store) RecordMemorizeWord(now time.Time) {
	if s == nil {
		return
	}
	s.dayEntry(now).MemorizeWords++
	_ = s.Save()
}

// RecordReadingSentence logs one completed sentence in Learn by Reading and saves.
func (s *Store) RecordReadingSentence(now time.Time) {
	if s == nil {
		return
	}
	s.dayEntry(now).ReadingSentences++
	_ = s.Save()
}

// RecordArabicDrill logs one answered Arabic drill question and saves.
func (s *Store) RecordArabicDrill(now time.Time) {
	if s == nil {
		return
	}
	s.dayEntry(now).ArabicDrills++
	_ = s.Save()
}

// TotalForDate returns the total logged actions for a YYYY-MM-DD date.
func (s *Store) TotalForDate(date string) int {
	if s == nil || s.Days == nil {
		return 0
	}
	return s.Days[date].Total()
}

// CurrentStreak returns the number of consecutive days with activity ending today (or yesterday if none yet today).
func (s *Store) CurrentStreak(now time.Time) int {
	if s == nil || len(s.Days) == 0 {
		return 0
	}
	if now.IsZero() {
		now = time.Now()
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	cursor := today
	if s.TotalForDate(cursor.Format("2006-01-02")) == 0 {
		cursor = cursor.AddDate(0, 0, -1)
		if s.TotalForDate(cursor.Format("2006-01-02")) == 0 {
			return 0
		}
	}

	streak := 0
	for s.TotalForDate(cursor.Format("2006-01-02")) > 0 {
		streak++
		cursor = cursor.AddDate(0, 0, -1)
	}
	return streak
}

// LongestStreak returns the longest consecutive-day activity streak recorded.
func (s *Store) LongestStreak() int {
	if s == nil || len(s.Days) == 0 {
		return 0
	}
	var dates []string
	for k, v := range s.Days {
		if v.Total() > 0 {
			dates = append(dates, k)
		}
	}
	if len(dates) == 0 {
		return 0
	}
	sort.Strings(dates)

	best := 1
	cur := 1
	prev, err := time.Parse("2006-01-02", dates[0])
	if err != nil {
		return 1
	}
	for i := 1; i < len(dates); i++ {
		d, err := time.Parse("2006-01-02", dates[i])
		if err != nil {
			continue
		}
		diff := int(d.Sub(prev).Hours() / 24)
		if diff == 1 {
			cur++
			if cur > best {
				best = cur
			}
		} else if diff > 1 {
			cur = 1
		}
		prev = d
	}
	return best
}

// PeriodSummary aggregates activity across the last N days (inclusive of today).
type PeriodSummary struct {
	Days             int
	ActiveDays       int
	Total            int
	Reviews          int
	QuizAnswered     int
	QuizCorrect      int
	MatchPairs       int
	MemorizeWords    int
	ReadingSentences int
	ArabicDrills     int
}

// LastNDaysSummary computes totals across the last n days ending at now.
func (s *Store) LastNDaysSummary(n int, now time.Time) PeriodSummary {
	sum := PeriodSummary{Days: n}
	if s == nil || n <= 0 {
		return sum
	}
	if now.IsZero() {
		now = time.Now()
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	for i := 0; i < n; i++ {
		key := today.AddDate(0, 0, -i).Format("2006-01-02")
		e, ok := s.Days[key]
		if !ok || e == nil || e.Total() == 0 {
			continue
		}
		sum.ActiveDays++
		sum.Total += e.Total()
		sum.Reviews += e.Reviews
		sum.QuizAnswered += e.QuizAnswered
		sum.QuizCorrect += e.QuizCorrect
		sum.MatchPairs += e.MatchPairs
		sum.MemorizeWords += e.MemorizeWords
		sum.ReadingSentences += e.ReadingSentences
		sum.ArabicDrills += e.ArabicDrills
	}
	return sum
}

// HeatmapCell represents a single day cell in the terminal activity heatmap.
type HeatmapCell struct {
	Date   string
	Count  int
	Level  int  // 0 (none) to 4 (highest)
	Future bool // true for days after today in the final week column
}

// HeatmapGrid builds a 7-row (Mon=0 .. Sun=6) by `weeks`-column grid ending in the week of `now`.
func (s *Store) HeatmapGrid(weeks int, now time.Time) [][]HeatmapCell {
	if weeks < 1 {
		weeks = 12
	}
	if now.IsZero() {
		now = time.Now()
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Monday = 0, ..., Sunday = 6
	weekdayOffset := (int(today.Weekday()) + 6) % 7
	currentWeekMonday := today.AddDate(0, 0, -weekdayOffset)
	startMonday := currentWeekMonday.AddDate(0, 0, -(weeks-1)*7)

	grid := make([][]HeatmapCell, 7)
	for r := 0; r < 7; r++ {
		grid[r] = make([]HeatmapCell, weeks)
	}

	for col := 0; col < weeks; col++ {
		for row := 0; row < 7; row++ {
			day := startMonday.AddDate(0, 0, col*7+row)
			future := day.After(today)
			key := day.Format("2006-01-02")
			count := 0
			if !future {
				count = s.TotalForDate(key)
			}
			grid[row][col] = HeatmapCell{
				Date:   key,
				Count:  count,
				Level:  intensityLevel(count),
				Future: future,
			}
		}
	}
	return grid
}

func intensityLevel(count int) int {
	switch {
	case count <= 0:
		return 0
	case count <= 5:
		return 1
	case count <= 15:
		return 2
	case count <= 35:
		return 3
	default:
		return 4
	}
}

// DeckMaturity summarizes card scheduling states across all decks in baseDir.
type DeckMaturity struct {
	TotalDecks    int
	TotalCards    int
	NewCards      int
	LearningCards int
	MatureCards   int
	DueToday      int
	Leeches       int
}

// ScanDeckMaturity walks baseDir (excluding marked.yaml) and aggregates SM-2 maturity counts.
func ScanDeckMaturity(baseDir string, now time.Time) DeckMaturity {
	var m DeckMaturity
	if baseDir == "" {
		return m
	}
	if now.IsZero() {
		now = time.Now()
	}
	today := now.Format("2006-01-02")

	files, err := deck.GetAllDeckFiles(baseDir, "")
	if err != nil {
		return m
	}
	for _, rel := range files {
		if filepath.Base(rel) == "marked.yaml" {
			continue
		}
		d, err := deck.LoadDeck(filepath.Join(baseDir, rel))
		if err != nil {
			continue
		}
		m.TotalDecks++
		for _, c := range d.Cards {
			m.TotalCards++
			if c.NextReview == "" || c.NextReview <= today {
				m.DueToday++
			}
			if c.Reps == 0 && c.NextReview == "" {
				m.NewCards++
			} else if c.Interval >= 21 {
				m.MatureCards++
			} else {
				m.LearningCards++
			}
			if c.Ease > 0 && c.Ease <= 1.5 {
				m.Leeches++
			}
		}
	}
	return m
}
