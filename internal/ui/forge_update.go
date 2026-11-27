package ui

import (
	"path/filepath"
	"time"

	"flashcards/internal/config"
	"flashcards/internal/forge"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) openForgeMenu() {
	forgeDir := m.ForgeDir
	if forgeDir == "" {
		forgeDir = config.DefaultForgeDir
	}
	if m.ForgeLexicon == nil {
		lex, err := forge.LoadLexiconFromDir(forgeDir)
		if err != nil {
			m.StatusMessage = "Could not load Forge lexicon: " + err.Error()
			return
		}
		m.ForgeLexicon = lex
	}
	if m.ForgeProgress == nil {
		progPath := ""
		if m.Config != nil && m.Config.FilePath != "" {
			progPath = filepath.Join(forgeDir, forge.DefaultProgressFilename)
		} else if forgeDir != "" {
			progPath = filepath.Join(forgeDir, forge.DefaultProgressFilename)
		}
		prog, _ := forge.LoadProgress(progPath)
		m.ForgeProgress = prog
	}
	m.ForgeMenuCursor = 0
	m.ForgeShowInfo = false
	m.ForgeShowBreakdown = false
	m.ForgeMissPause = false
	m.State = StateForgeMenu
}

func (m *Model) startForgeRun(archetypeID string) {
	if m.ForgeLexicon == nil {
		m.openForgeMenu()
		if m.ForgeLexicon == nil {
			return
		}
	}
	m.ForgeRun = forge.NewRun(m.ForgeLexicon, m.ForgeProgress, archetypeID, 0)
	_ = m.ForgeProgress.Save()
	m.ForgeShowBreakdown = false
	m.ForgeMissPause = false
	m.ForgeShowInfo = false
	m.ForgeLastReward = 0
	m.State = StateForgePlay
}

func (m *Model) isForgeSM2SyncEnabled() bool {
	return m.Config != nil && m.Config.ForgeSM2Sync
}

func (m Model) updateForgeMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "esc", "b":
		m.State = StateModeSelect
		return m, nil
	case "S":
		if m.Config != nil {
			m.Config.ForgeSM2Sync = !m.Config.ForgeSM2Sync
			m.saveConfig()
		}
		return m, nil
	case "up", "k":
		if m.ForgeMenuCursor > 0 {
			m.ForgeMenuCursor--
		}
		return m, nil
	case "down", "j":
		if m.ForgeMenuCursor < len(forge.StarterArchetypes)-1 {
			m.ForgeMenuCursor++
		}
		return m, nil
	case "enter", " ":
		if m.ForgeMenuCursor >= 0 && m.ForgeMenuCursor < len(forge.StarterArchetypes) {
			m.startForgeRun(forge.StarterArchetypes[m.ForgeMenuCursor].ID)
		}
		return m, nil
	}

	idx := KeyToIndex(key)
	if idx >= 0 && idx < len(forge.StarterArchetypes) {
		m.ForgeMenuCursor = idx
		m.startForgeRun(forge.StarterArchetypes[idx].ID)
		return m, nil
	}
	return m, nil
}

func (m Model) updateForgePlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	run := m.ForgeRun
	if run == nil {
		m.State = StateForgeMenu
		return m, nil
	}

	key := msg.String()

	// If showing the breakdown of a completed hand (or fizzle), Space/Enter advances
	if m.ForgeShowBreakdown {
		switch key {
		case "enter", " ":
			m.ForgeShowBreakdown = false
			if run.BlindCleared() {
				m.ForgeLastReward = run.ClaimBlindReward()
				_ = m.ForgeProgress.Save()
				m.State = StateForgeShop
			} else if run.GameOver {
				_ = m.ForgeProgress.Save()
				m.State = StateForgeRunOver
			}
			return m, nil
		case "esc":
			m.ForgeShowBreakdown = false
			if run.GameOver {
				m.State = StateForgeRunOver
			}
			return m, nil
		}
		return m, nil
	}

	switch key {
	case "v":
		m.ForgeShowInfo = !m.ForgeShowInfo
		return m, nil
	case "esc":
		if m.ForgeShowInfo {
			m.ForgeShowInfo = false
			return m, nil
		}
		if run.HasStagedCards() {
			run.ClearStage()
			return m, nil
		}
		_ = m.ForgeProgress.Save()
		m.State = StateForgeMenu
		return m, nil
	case "x":
		run.ClearStage()
		return m, nil
	case "backspace":
		run.UndoStage()
		return m, nil
	case " ":
		run.StartNewWordSlot()
		return m, nil
	case "tab":
		run.StartPivotSlot()
		return m, nil
	case ";":
		run.DiscardStaged()
		return m, nil
	case "enter":
		if !run.HasStagedCards() {
			return m, nil
		}
		hasValid := run.CommitPlay()
		if hasValid {
			m.ForgeMissPause = false
			m.State = StateForgeProve
		} else {
			// All staged words fizzled immediately
			m.ForgeShowBreakdown = true
		}
		return m, nil
	}

	// Home-row keys a-l (indices 0..8) select cards from the 9-card hand
	idx := KeyToIndex(key)
	if idx >= 0 && idx < forge.HandSize && idx < len(run.Hand) {
		run.StageCard(idx)
		return m, nil
	}
	return m, nil
}

func (m Model) updateForgeProve(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	run := m.ForgeRun
	if run == nil {
		m.State = StateForgeMenu
		return m, nil
	}

	key := msg.String()

	// Brief pause after a missed word so the player can read the correction
	if m.ForgeMissPause {
		switch key {
		case "enter", " ", "esc":
			m.ForgeMissPause = false
			if run.ProveIdx >= len(run.Challenges) {
				m.ForgeShowBreakdown = true
				m.State = StateForgePlay
			}
		}
		return m, nil
	}

	if run.ProveIdx >= len(run.Challenges) {
		m.ForgeShowBreakdown = true
		m.State = StateForgePlay
		return m, nil
	}

	ch := run.Challenges[run.ProveIdx]
	choiceIdx := -1
	switch key {
	case "1":
		choiceIdx = 0
	case "2":
		choiceIdx = 1
	case "3":
		choiceIdx = 2
	case "4":
		choiceIdx = 3
	case "5":
		choiceIdx = 4
	default:
		kIdx := KeyToIndex(key)
		if kIdx >= 0 && kIdx < len(ch.Options) {
			choiceIdx = kIdx
		}
	}

	if choiceIdx >= 0 && choiceIdx < len(ch.Options) {
		now := time.Now()
		passed, more := run.AnswerProve(choiceIdx, now, m.isForgeSM2SyncEnabled(), m.BaseDir)
		_ = m.ForgeProgress.Save()
		if passed && m.Stats != nil {
			m.Stats.RecordForgeWord(now)
		}
		if !passed {
			// Pause briefly on miss so the player sees the true Pinyin & Meaning
			m.ForgeMissPause = true
			return m, nil
		}
		if !more {
			m.ForgeShowBreakdown = true
			m.State = StateForgePlay
		}
		return m, nil
	}

	return m, nil
}

func (m Model) updateForgeShop(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	run := m.ForgeRun
	if run == nil {
		m.State = StateForgeMenu
		return m, nil
	}

	switch msg.String() {
	case "a":
		run.BuyRelic(0)
	case "s":
		run.BuyRelic(1)
	case "d":
		run.BuyDraft(0)
	case "f":
		run.BuyDraft(1)
	case "g":
		run.BuyDraft(2)
	case "h":
		run.BuyScroll()
	case "j":
		run.PruneLowestConnectivityCard()
	case "k":
		run.DuplicateBestHubCard()
	case "r":
		run.RerollShop()
	case "enter", " ":
		run.AdvanceToNextBlind()
		m.ForgeShowBreakdown = false
		m.State = StateForgePlay
	case "esc":
		_ = m.ForgeProgress.Save()
		m.State = StateForgeMenu
	}
	return m, nil
}

func (m Model) updateForgeRunOver(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		if m.ForgeRun != nil {
			m.startForgeRun(m.ForgeRun.ArchetypeID)
		} else {
			m.State = StateForgeMenu
		}
	case "enter", " ", "esc", "b":
		m.State = StateForgeMenu
	}
	return m, nil
}
