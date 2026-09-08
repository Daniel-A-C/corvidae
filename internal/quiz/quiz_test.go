package quiz

import (
	"testing"

	"flashcards/internal/deck"
)

func TestGenerateOptions(t *testing.T) {
	target := deck.Flashcard{Character: "一", Pinyin: "yī", Meaning: "One"}
	allCards := []deck.Flashcard{
		target,
		{Character: "二", Pinyin: "èr", Meaning: "Two"},
		{Character: "三", Pinyin: "sān", Meaning: "Three"},
		{Character: "四", Pinyin: "sì", Meaning: "Four"},
		{Character: "五", Pinyin: "wǔ", Meaning: "Five"},
		{Character: "六", Pinyin: "liù", Meaning: "Six"},
		{Character: "七", Pinyin: "qī", Meaning: "Seven"},
		{Character: "二_dup", Pinyin: "èr", Meaning: "Two"}, // duplicate meaning/pinyin
	}

	options, correctIdx := GenerateOptions(target, allCards, 5)

	// 1 target + 5 unique distractors = 6 total options
	if len(options) != 6 {
		t.Fatalf("expected 6 options, got %d: %v", len(options), options)
	}

	if correctIdx < 0 || correctIdx >= len(options) {
		t.Fatalf("invalid correctIdx: %d", correctIdx)
	}

	if options[correctIdx] != "yī - One" {
		t.Errorf("expected correct answer 'yī - One', got '%s'", options[correctIdx])
	}

	// Verify no duplicates in options
	seen := make(map[string]bool)
	for _, opt := range options {
		if seen[opt] {
			t.Errorf("duplicate option found: %s", opt)
		}
		seen[opt] = true
	}
}

func TestGenerateOptions_SmallPool(t *testing.T) {
	target := deck.Flashcard{Character: "A", Meaning: "Letter A"}
	allCards := []deck.Flashcard{
		target,
		{Character: "B", Meaning: "Letter B"},
	}

	options, correctIdx := GenerateOptions(target, allCards, 5)

	if len(options) != 2 {
		t.Fatalf("expected 2 options for small pool, got %d", len(options))
	}
	if options[correctIdx] != "Letter A" {
		t.Errorf("expected 'Letter A', got '%s'", options[correctIdx])
	}
}
