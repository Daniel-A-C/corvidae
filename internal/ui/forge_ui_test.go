package ui

import (
	"path/filepath"
	"strings"
	"testing"

	"flashcards/internal/config"
	"flashcards/internal/forge"

	tea "github.com/charmbracelet/bubbletea"
)

func TestForgeEndToEndUIFlow(t *testing.T) {
	cfg := config.Default()
	cfg.FilePath = ""
	cfg.StatsFile = ""
	cfg.DecksDir = filepath.Join("..", "..", "decks")
	cfg.ForgeDir = filepath.Join("..", "..", "forge")

	m := NewWithConfig(cfg)
	// Avoid mutating repo's real forge/progress.yaml during test
	m.ForgeProgress = forge.NewEmptyProgress("")
	m.Width = 100
	m.Height = 36

	// 1. Mode Select -> press 'j' to open Morpheme Forge
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = next.(Model)
	if m.State != StateForgeMenu {
		t.Fatalf("expected StateForgeMenu after pressing 'j', got %d (status: %s)", m.State, m.StatusMessage)
	}
	menuView := m.View()
	if !strings.Contains(menuView, "MORPHEME FORGE") {
		t.Errorf("expected Forge Menu title in view")
	}

	// 2. Press 'a' to start Hogwarts Core starter deck
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = next.(Model)
	if m.State != StateForgePlay || m.ForgeRun == nil {
		t.Fatalf("expected StateForgePlay with active run, got state=%d", m.State)
	}
	if len(m.ForgeRun.Hand) != forge.HandSize {
		t.Fatalf("expected %d cards in hand, got %d", forge.HandSize, len(m.ForgeRun.Hand))
	}

	// 3. Inject a known valid 2-char compound at hand[0], hand[1] and stage it via 'a', 's', Enter
	comp := m.ForgeLexicon.Compounds[0]
	runes := []rune(comp.Word)
	for i, r := range runes {
		m.ForgeRun.Hand[i] = forge.DeckCard{ID: 900 + i, Glyph: string(r)}
	}
	m.ForgeRun.ClearStage()
	keys := []rune{'a', 's', 'd', 'f'}
	for i := range runes {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{keys[i]}})
		m = next.(Model)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.State != StateForgeProve {
		t.Fatalf("expected StateForgeProve after committing valid compound %s, got %d", comp.Word, m.State)
	}

	// 4. Answer the Prove challenge correctly
	correctKey := IndexToKey(m.ForgeRun.Challenges[0].CorrectIdx)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(correctKey)})
	m = next.(Model)
	if m.State != StateForgePlay || !m.ForgeShowBreakdown {
		t.Fatalf("expected StateForgePlay with breakdown after proving, got state=%d breakdown=%v", m.State, m.ForgeShowBreakdown)
	}
	if m.ForgeRun.BlindScore <= 0 {
		t.Errorf("expected positive BlindScore, got %d", m.ForgeRun.BlindScore)
	}

	// 5. Force Blind clear and press Enter to enter Shop
	m.ForgeRun.BlindScore = m.ForgeRun.TargetScore
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.State != StateForgeShop {
		t.Fatalf("expected StateForgeShop after clearing blind, got %d", m.State)
	}
	shopView := m.View()
	if !strings.Contains(shopView, "FORGE SHOP") {
		t.Errorf("expected FORGE SHOP in shop view")
	}

	// 6. Press Enter to advance to Big Blind
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.State != StateForgePlay || m.ForgeRun.BlindKind != forge.BlindBig {
		t.Fatalf("expected StateForgePlay at BlindBig, got state=%d blind=%v", m.State, m.ForgeRun.BlindKind)
	}
}
