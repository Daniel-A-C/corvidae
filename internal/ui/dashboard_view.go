package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewDashboard() string {
	frameWidth := responsiveFrameWidth(m.Width, 78, 44)
	now := time.Now()
	todayKey := now.Format("2006-01-02")

	var b strings.Builder
	b.WriteString(renderSplitLine(
		SectionTitleStyle.Render("★ STUDY DASHBOARD & ACTIVITY ★"),
		ModePillStyle.Render(todayKey),
		frameWidth,
	) + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n\n")

	// 1. Streak & Totals KPI Strip
	curStreak := m.Stats.CurrentStreak(now)
	bestStreak := m.Stats.LongestStreak()
	todayTotal := m.Stats.TotalForDate(todayKey)
	sum7 := m.Stats.LastNDaysSummary(7, now)
	sum30 := m.Stats.LastNDaysSummary(30, now)

	streakText := fmt.Sprintf("Streak: %d days (Best: %d)", curStreak, bestStreak)
	if curStreak >= 3 {
		streakText = HotStreakStyle.Render("★ " + streakText)
	} else {
		streakText = KeyStyle.Render(streakText)
	}
	totalsText := fmt.Sprintf("Today: %s  •  7d: %s  •  30d: %s",
		CorrectStyle.Render(fmt.Sprintf("%d", todayTotal)),
		KeyStyle.Render(fmt.Sprintf("%d", sum7.Total)),
		KeyStyle.Render(fmt.Sprintf("%d", sum30.Total)),
	)
	b.WriteString(renderSplitLine(streakText, totalsText, frameWidth) + "\n\n")

	// 2. Activity Heatmap (responsive weeks count: 16 weeks on wide terminals, 12 on standard)
	weeks := 16
	if frameWidth < 60 {
		weeks = 12
	}
	grid := m.Stats.HeatmapGrid(weeks, now)
	dayLabels := [7]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

	b.WriteString(SectionTitleStyle.Render(fmt.Sprintf("Activity Heatmap (Last %d Weeks)", weeks)) + "\n")
	for row := 0; row < 7; row++ {
		lbl := "   "
		if row == 0 || row == 2 || row == 4 || row == 6 {
			lbl = dayLabels[row]
		}
		var cells strings.Builder
		for col := 0; col < weeks; col++ {
			cell := grid[row][col]
			if cell.Future {
				cells.WriteString("   ")
				continue
			}
			cells.WriteString(renderHeatmapCell(cell.Level) + " ")
		}
		b.WriteString(fmt.Sprintf("  %s  %s\n", HintStyle.Render(lbl), cells.String()))
	}
	legend := fmt.Sprintf("  %s %s %s %s %s %s %s",
		HintStyle.Render("Less"),
		renderHeatmapCell(0),
		renderHeatmapCell(1),
		renderHeatmapCell(2),
		renderHeatmapCell(3),
		renderHeatmapCell(4),
		HintStyle.Render("More"),
	)
	b.WriteString(legend + "\n\n")

	// 3. Card Maturity Breakdown (across decks/)
	mat := m.DashboardMaturity
	b.WriteString(SectionTitleStyle.Render("Flashcard Maturity") + HintStyle.Render(fmt.Sprintf(" (%d decks, %d cards)", mat.TotalDecks, mat.TotalCards)) + "\n")
	b.WriteString(fmt.Sprintf("  %s %d   •   %s %d   •   %s %d   •   %s %d   •   %s %d\n\n",
		DueBadgeStyle.Render("Due Today:"), mat.DueToday,
		HintStyle.Render("New:"), mat.NewCards,
		PinyinStyle.Render("Learning:"), mat.LearningCards,
		CorrectStyle.Render("Mature:"), mat.MatureCards,
		ErrorStyle.Render("Leeches:"), mat.Leeches,
	))

	// 4. 30-Day Mode Breakdown
	b.WriteString(SectionTitleStyle.Render("30-Day Breakdown") + HintStyle.Render(fmt.Sprintf(" (%d active days)", sum30.ActiveDays)) + "\n")
	quizAcc := "—"
	if sum30.QuizAnswered > 0 {
		quizAcc = fmt.Sprintf("%.0f%%", float64(sum30.QuizCorrect)/float64(sum30.QuizAnswered)*100)
	}
	b.WriteString(fmt.Sprintf("  SM-2 Reviews: %s   •   Quiz: %s (%s acc)   •   Match Pairs: %s\n",
		KeyStyle.Render(fmt.Sprintf("%d", sum30.Reviews)),
		KeyStyle.Render(fmt.Sprintf("%d", sum30.QuizAnswered)),
		quizAcc,
		KeyStyle.Render(fmt.Sprintf("%d", sum30.MatchPairs)),
	))
	b.WriteString(fmt.Sprintf("  Memorize Words: %s   •   Reading Sentences: %s   •   Arabic Drills: %s\n",
		KeyStyle.Render(fmt.Sprintf("%d", sum30.MemorizeWords)),
		KeyStyle.Render(fmt.Sprintf("%d", sum30.ReadingSentences)),
		KeyStyle.Render(fmt.Sprintf("%d", sum30.ArabicDrills)),
	))

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	footerRaw := fmt.Sprintf("[%s] Refresh  •  [%s] Settings  •  [%s] Main Menu  •  [%s] Quit",
		KeyStyle.Render("r"), KeyStyle.Render("c"), KeyStyle.Render("Esc / Enter"), KeyStyle.Render("q"))
	b.WriteString(lipgloss.NewStyle().Width(frameWidth).Align(lipgloss.Left).Render(HintStyle.Render(footerRaw)))
	return renderFrame(frameWidth, b.String())
}

func renderHeatmapCell(level int) string {
	switch level {
	case 0:
		return HeatmapLevel0Style.Render("··")
	case 1:
		return HeatmapLevel1Style.Render("░░")
	case 2:
		return HeatmapLevel2Style.Render("▒▒")
	case 3:
		return HeatmapLevel3Style.Render("▓▓")
	default:
		return HeatmapLevel4Style.Render("██")
	}
}
