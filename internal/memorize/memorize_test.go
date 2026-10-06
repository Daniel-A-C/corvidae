package memorize

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFilemonYaml(t *testing.T) {
	yamlPath := filepath.Join("..", "..", "memorizationTexts", "filemon.yaml")
	loaded, err := LoadText(yamlPath)
	if err != nil {
		t.Fatalf("expected to load filemon.yaml, got error: %v", err)
	}
	if loaded.Title != "Epístola a Filemón" {
		t.Errorf("expected title 'Epístola a Filemón', got '%s'", loaded.Title)
	}
	if len(loaded.Words) != 423 {
		t.Errorf("expected 423 words, got %d", len(loaded.Words))
	}

	// Verify each word has 5 options and includes the target word
	for i, item := range loaded.Words {
		if len(item.Options) != 5 {
			t.Errorf("item %d (%s) has %d options, expected 5", i, item.Word, len(item.Options))
		}
		found := false
		for _, opt := range item.Options {
			if opt == item.Word {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("item %d (%s) target word not found in options %v", i, item.Word, item.Options)
		}
	}
}

func TestShuffleOptions(t *testing.T) {
	item := WordItem{
		Word:    "Pablo,",
		Options: []string{"Pablo,", "Pedro,", "Juan,", "Lucas,", "Santiago,"},
	}
	shuffled, correctIdx := ShuffleOptions(item)
	if len(shuffled) != 5 {
		t.Fatalf("expected 5 shuffled options, got %d", len(shuffled))
	}
	if correctIdx < 0 || correctIdx >= 5 {
		t.Fatalf("invalid correctIdx: %d", correctIdx)
	}
	if shuffled[correctIdx] != "Pablo," {
		t.Errorf("expected shuffled[%d] to be 'Pablo,', got '%s'", correctIdx, shuffled[correctIdx])
	}
}

func TestListTextsAndAutoGenerate(t *testing.T) {
	tempDir := t.TempDir()

	// Place a raw text file in tempDir
	txtContent := "1 El principio de la sabiduría es el temor del Señor.\n2 Los necios desprecian la corrección."
	err := os.WriteFile(filepath.Join(tempDir, "proverbios.txt"), []byte(txtContent), 0644)
	if err != nil {
		t.Fatalf("failed to write proverbios.txt: %v", err)
	}

	headers, err := ListTexts(tempDir)
	if err != nil {
		t.Fatalf("ListTexts failed: %v", err)
	}
	if len(headers) != 1 {
		t.Fatalf("expected 1 text, got %d", len(headers))
	}
	if headers[0].Title != "proverbios" {
		t.Errorf("expected title 'proverbios', got '%s'", headers[0].Title)
	}
	if headers[0].WordCount == 0 {
		t.Errorf("expected WordCount > 0, got %d", headers[0].WordCount)
	}

	// Check that proverbios.yaml was auto-generated
	if _, err := os.Stat(filepath.Join(tempDir, "proverbios.yaml")); err != nil {
		t.Errorf("expected proverbios.yaml to be auto-generated, got err: %v", err)
	}
}

func TestSaveAndLoadProgress(t *testing.T) {
	tempDir := t.TempDir()
	savePath := filepath.Join(tempDir, "test.yaml")

	text := &Text{
		Title:    "Test Text",
		Progress: 15,
		Words: []WordItem{
			{Word: "Hello", Options: []string{"Hello", "World", "Hi", "Greetings", "Hey"}},
		},
	}

	if err := SaveText(savePath, text); err != nil {
		t.Fatalf("SaveText failed: %v", err)
	}

	loaded, err := LoadText(savePath)
	if err != nil {
		t.Fatalf("LoadText failed: %v", err)
	}
	if loaded.Progress != 15 {
		t.Errorf("expected Progress 15, got %d", loaded.Progress)
	}
}

func TestGetSections(t *testing.T) {
	yamlPath := filepath.Join("..", "..", "memorizationTexts", "filemon.yaml")
	loaded, err := LoadText(yamlPath)
	if err != nil {
		t.Fatalf("expected to load filemon.yaml: %v", err)
	}

	sections := loaded.GetSections()
	if len(sections) != 25 {
		t.Fatalf("expected 25 verse sections in Filemón, got %d", len(sections))
	}

	// Verify sections are contiguous and non-empty
	expectedStart := 0
	for i, sec := range sections {
		if sec.Index != i {
			t.Errorf("section %d has Index %d", i, sec.Index)
		}
		if sec.StartIdx != expectedStart {
			t.Errorf("section %d expected StartIdx %d, got %d", i, expectedStart, sec.StartIdx)
		}
		if sec.EndIdx <= sec.StartIdx {
			t.Errorf("section %d has non-positive length (Start=%d, End=%d)", i, sec.StartIdx, sec.EndIdx)
		}
		if sec.WordCount != (sec.EndIdx - sec.StartIdx) {
			t.Errorf("section %d WordCount mismatch", i)
		}
		if sec.Preview == "" {
			t.Errorf("section %d preview is empty", i)
		}
		expectedStart = sec.EndIdx
	}

	if expectedStart != len(loaded.Words) {
		t.Errorf("expected final EndIdx to be %d, got %d", len(loaded.Words), expectedStart)
	}

	// Check specific verses: Verse 15 and Verse 20
	sec15 := sections[14]
	if sec15.Number != "15" || sec15.Label != "Verse 15" {
		t.Errorf("expected Verse 15 at index 14, got %s", sec15.Label)
	}
	sec20 := sections[19]
	if sec20.Number != "20" || sec20.Label != "Verse 20" {
		t.Errorf("expected Verse 20 at index 19, got %s", sec20.Label)
	}
}

