package deck

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreOperations(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Initially empty directory
	dirs, err := GetDeckDirectories(tempDir)
	if err != nil {
		t.Fatalf("unexpected error getting dirs: %v", err)
	}
	if len(dirs) != 0 {
		t.Fatalf("expected 0 dirs, got %d", len(dirs))
	}

	// 2. Create subdirectories (including a hidden one)
	mandarinDir := filepath.Join(tempDir, "Mandarin")
	spanishDir := filepath.Join(tempDir, "Spanish")
	hiddenDir := filepath.Join(tempDir, ".hidden")
	if err := os.Mkdir(mandarinDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(spanishDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(hiddenDir, 0755); err != nil {
		t.Fatal(err)
	}

	dirs, err = GetDeckDirectories(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 2 || dirs[0] != "Mandarin" || dirs[1] != "Spanish" {
		t.Fatalf("expected [Mandarin, Spanish], got %v", dirs)
	}

	// 3. Create deck files in Mandarin (and a non-yaml file and a hidden yaml file)
	deckPath1 := filepath.Join(mandarinDir, "deck1.yaml")
	deckPath2 := filepath.Join(mandarinDir, "deck2.yaml")
	nonYaml := filepath.Join(mandarinDir, "notes.txt")
	hiddenYaml := filepath.Join(mandarinDir, ".draft.yaml")

	cardDeck := Deck{
		Cards: []Flashcard{
			{
				Character:   "你好",
				Pinyin:      "nǐ hǎo",
				Meaning:     "Hello",
				Explanation: "Common greeting",
			},
		},
	}

	if err := SaveDeck(deckPath1, cardDeck); err != nil {
		t.Fatalf("failed to save deck: %v", err)
	}
	if err := SaveDeck(deckPath2, cardDeck); err != nil {
		t.Fatalf("failed to save deck: %v", err)
	}
	if err := os.WriteFile(nonYaml, []byte("notes"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hiddenYaml, []byte("draft"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := GetDeckFiles(tempDir, "Mandarin")
	if err != nil {
		t.Fatalf("failed to get files: %v", err)
	}
	if len(files) != 2 || files[0] != "deck1.yaml" || files[1] != "deck2.yaml" {
		t.Fatalf("expected [deck1.yaml, deck2.yaml], got %v", files)
	}

	if count := CountDecksInDir(tempDir, "Mandarin"); count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}
	if count := CountDecksInDir(tempDir, "Spanish"); count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}

	// 4. Load Deck verification
	loaded, err := LoadDeck(deckPath1)
	if err != nil {
		t.Fatalf("failed to load deck: %v", err)
	}
	if len(loaded.Cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(loaded.Cards))
	}
	if loaded.Cards[0].Character != "你好" {
		t.Fatalf("expected '你好', got %s", loaded.Cards[0].Character)
	}
	if loaded.Cards[0].FormatQuizOption() != "nǐ hǎo - Hello" {
		t.Fatalf("unexpected quiz option format: %s", loaded.Cards[0].FormatQuizOption())
	}

	// 5. Test card without pinyin
	noPinyinCard := Flashcard{
		Character: "Bonjour",
		Meaning:   "Hello",
	}
	if noPinyinCard.FormatQuizOption() != "Hello" {
		t.Fatalf("expected 'Hello', got %s", noPinyinCard.FormatQuizOption())
	}
}
