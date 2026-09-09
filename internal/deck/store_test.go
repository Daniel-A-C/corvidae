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

func TestBalatroDecksConsolidated(t *testing.T) {
	decksPath := filepath.Join("..", "..", "decks", "Mandarin")
	if _, err := os.Stat(decksPath); os.IsNotExist(err) {
		t.Skip("decks directory not present at expected relative path")
	}

	balatroFiles := []string{
		"balatro1.yaml",
		"balatro2.yaml",
		"balatro3.yaml",
		"balatro4.yaml",
		"balatro5.yaml",
		"balatro6.yaml",
	}

	seen := make(map[string]string)
	totalCards := 0

	for _, file := range balatroFiles {
		fullPath := filepath.Join(decksPath, file)
		d, err := LoadDeck(fullPath)
		if err != nil {
			t.Fatalf("failed to load %s: %v", file, err)
		}
		if len(d.Cards) == 0 {
			t.Errorf("expected %s to have cards, but it was empty", file)
		}
		totalCards += len(d.Cards)

		for _, card := range d.Cards {
			if card.Character == "" {
				t.Errorf("%s: card missing character", file)
			}
			if card.Meaning == "" {
				t.Errorf("%s: card %s missing meaning", file, card.Character)
			}
			if origFile, exists := seen[card.Character]; exists {
				t.Errorf("duplicate card %s found in %s and %s", card.Character, origFile, file)
			}
			seen[card.Character] = file
		}
	}

	if totalCards != 74 {
		t.Fatalf("expected 74 total cards across balatro decks, got %d", totalCards)
	}

	// Verify old balatro files (7 and 8) are removed
	for i := 7; i <= 8; i++ {
		oldPath := filepath.Join(decksPath, "balatro"+string(rune('0'+i))+".yaml")
		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			t.Errorf("expected old deck %s to be deleted", oldPath)
		}
	}
}

func TestMiscDecks(t *testing.T) {
	decksPath := filepath.Join("..", "..", "decks", "Mandarin")
	if _, err := os.Stat(decksPath); os.IsNotExist(err) {
		t.Skip("decks directory not present at expected relative path")
	}

	miscFiles := []string{
		"misc.yaml",
		"misc2.yaml",
		"misc3.yaml",
		"misc4.yaml",
	}

	seen := make(map[string]string)
	totalCards := 0

	for _, file := range miscFiles {
		fullPath := filepath.Join(decksPath, file)
		d, err := LoadDeck(fullPath)
		if err != nil {
			t.Fatalf("failed to load %s: %v", file, err)
		}
		if len(d.Cards) == 0 {
			t.Errorf("expected %s to have cards, but it was empty", file)
		}
		totalCards += len(d.Cards)

		for _, card := range d.Cards {
			if card.Character == "" {
				t.Errorf("%s: card missing character", file)
			}
			if card.Meaning == "" {
				t.Errorf("%s: card %s missing meaning", file, card.Character)
			}
			if origFile, exists := seen[card.Character]; exists {
				t.Errorf("duplicate card %s found in %s and %s", card.Character, origFile, file)
			}
			seen[card.Character] = file
		}
	}

	if totalCards != 40 {
		t.Fatalf("expected 40 total cards across 4 misc decks, got %d", totalCards)
	}
}
