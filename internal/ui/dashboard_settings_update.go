package ui

import (
	"time"

	"flashcards/internal/config"
	"flashcards/internal/stats"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) openDashboard() {
	m.DashboardMaturity = stats.ScanDeckMaturity(m.BaseDir, time.Now())
	m.State = StateDashboard
}

func (m Model) updateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		m.DashboardMaturity = stats.ScanDeckMaturity(m.BaseDir, time.Now())
	case "c":
		m.SettingsCursor = 0
		m.StatusMessage = ""
		m.State = StateSettings
	case "esc", "b", "enter", " ":
		m.State = StateModeSelect
	}
	return m, nil
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	const numSettings = 5

	idx := KeyToIndex(msg.String())
	if idx >= 0 && idx < numSettings {
		m.SettingsCursor = idx
		m.toggleSetting(idx)
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.SettingsCursor > 0 {
			m.SettingsCursor--
		}
	case "down", "j":
		if m.SettingsCursor < numSettings-1 {
			m.SettingsCursor++
		}
	case "enter", " ", "left", "right", "h", "l":
		m.toggleSetting(m.SettingsCursor)
	case "esc", "b":
		m.StatusMessage = ""
		m.State = StateModeSelect
	}
	return m, nil
}

func (m *Model) toggleSetting(idx int) {
	if m.Config == nil {
		m.Config = config.Default()
		m.Config.FilePath = ""
	}
	switch idx {
	case 0:
		m.Config.QuizSM2Sync = !m.Config.QuizSM2Sync
		m.saveConfig()
	case 1:
		m.Config.MatchSM2Sync = !m.Config.MatchSM2Sync
		m.saveConfig()
	case 2:
		m.Config.MatchHidePinyin = !m.Config.MatchHidePinyin
		m.MatchHidePinyin = m.Config.MatchHidePinyin
		m.saveConfig()
	case 3:
		m.ToggleMemorizeTypingMode()
	case 4:
		m.CycleMemorizeOptionCount()
	}
	if m.Config.FilePath != "" {
		m.StatusMessage = "✓ Saved to " + m.Config.FilePath
	} else {
		m.StatusMessage = "✓ Updated setting"
	}
}
