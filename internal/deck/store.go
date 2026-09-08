package deck

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// GetDeckDirectories retrieves all non-hidden subdirectories in the specified deck base directory.
func GetDeckDirectories(baseDir string) ([]string, error) {
	if baseDir == "" {
		baseDir = "decks"
	}
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			dirs = append(dirs, entry.Name())
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

// GetDeckFiles retrieves all YAML deck files inside the given language subdirectory.
func GetDeckFiles(baseDir, dir string) ([]string, error) {
	if baseDir == "" {
		baseDir = "decks"
	}
	dirPath := filepath.Join(baseDir, dir)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasPrefix(entry.Name(), ".") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

// CountDecksInDir counts the number of YAML decks present in a language directory.
func CountDecksInDir(baseDir, dir string) int {
	files, err := GetDeckFiles(baseDir, dir)
	if err != nil {
		return 0
	}
	return len(files)
}

// LoadDeck reads and unmarshals a YAML deck file from disk.
func LoadDeck(filePath string) (Deck, error) {
	var deck Deck
	data, err := os.ReadFile(filePath)
	if err != nil {
		return deck, err
	}
	err = yaml.Unmarshal(data, &deck)
	return deck, err
}

// SaveDeck serializes and writes a Deck to disk.
func SaveDeck(filePath string, d Deck) error {
	data, err := yaml.Marshal(d)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}
