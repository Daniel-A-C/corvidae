package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type settingRow struct {
	label string
	desc  string
	value string
}

func (m Model) settingsRows() []settingRow {
	quizSync := "OFF"
	matchSync := "OFF"
	hidePinyin := "OFF"
	if m.Config != nil {
		if m.Config.QuizSM2Sync {
			quizSync = "ON"
		}
		if m.Config.MatchSM2Sync {
			matchSync = "ON"
		}
		if m.Config.MatchHidePinyin {
			hidePinyin = "ON"
		}
	}
	memMode := "Options"
	if m.MemorizeTypingMode {
		memMode = "First-Letter Typing"
	}

	return []settingRow{
		{
			label: "Quiz → SM-2 Sync",
			desc:  "Update card SM-2 intervals after Quiz answers (Good/Wrong)",
			value: quizSync,
		},
		{
			label: "Match → SM-2 Sync",
			desc:  "Update card SM-2 intervals when matching pairs (Good/Hard)",
			value: matchSync,
		},
		{
			label: "Match: Hide Pinyin",
			desc:  "Default Matching Game to hide Pinyin / pronunciation",
			value: hidePinyin,
		},
		{
			label: "Memorize Mode",
			desc:  "Default recitation style: Options or First-Letter Typing",
			value: memMode,
		},
		{
			label: "Memorize Option Count",
			desc:  "Number of choices shown in Memorize by Options (2–5)",
			value: fmt.Sprintf("%d options", m.MemorizeOptionCount),
		},
	}
}

func (m Model) viewSettings() string {
	frameWidth := m.frameWidth()
	cfgPath := "(in-memory)"
	if m.Config != nil && m.Config.FilePath != "" {
		cfgPath = m.Config.FilePath
	}

	var b strings.Builder
	b.WriteString(renderSplitLine(
		SectionTitleStyle.Render("★ SETTINGS & PERSISTENT CONFIG ★"),
		ModePillStyle.Render(cfgPath),
		frameWidth,
	) + "\n")
	b.WriteString(HintStyle.Render("Changes are saved automatically to config.yaml") + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n\n")

	rows := m.settingsRows()
	for i, row := range rows {
		key := IndexToKey(i)
		keyBadge := KeyStyle.Render(fmt.Sprintf("[%s]", key))
		cursor := "  "
		label := row.label
		if m.SettingsCursor == i {
			cursor = "> "
			label = CursorStyle.Render(label)
		}

		valStyled := ModePillStyle.Render(row.value)
		if row.value == "ON" || row.value == "First-Letter Typing" {
			valStyled = SyncOnPillStyle.Render(row.value)
		}

		leftPart := fmt.Sprintf("%s%s  %s", CursorStyle.Render(cursor), keyBadge, label)
		b.WriteString(renderSplitLine(leftPart, valStyled, frameWidth) + "\n")
		b.WriteString(fmt.Sprintf("        %s\n\n", HintStyle.Render(truncateToWidth(row.desc, frameWidth-8))))
	}

	if m.StatusMessage != "" {
		b.WriteString(CorrectStyle.Render(m.StatusMessage) + "\n")
	}

	b.WriteString(renderDivider(frameWidth) + "\n")
	footerRaw := fmt.Sprintf("[%s] Toggle / Cycle  •  [%s] Navigate  •  [%s] Main Menu  •  [%s] Quit",
		KeyStyle.Render("a-g / Space / Enter"), KeyStyle.Render("↑↓ / jk"), KeyStyle.Render("Esc"), KeyStyle.Render("q"))
	b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))
	return renderFrame(frameWidth, b.String())
}
