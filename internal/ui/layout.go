package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const (
	defaultFrameWidth = 76
	minFrameWidth     = 40
)

// responsiveFrameWidth computes a responsive frame width bounded by [minWidth, defaultMax]
// (or smaller if the terminal itself is narrower than minWidth).
func responsiveFrameWidth(termWidth, defaultMax, minWidth int) int {
	if termWidth <= 0 {
		return defaultMax
	}
	w := termWidth - 4
	if w > defaultMax {
		return defaultMax
	}
	if w < minWidth {
		if termWidth < minWidth {
			if termWidth < 20 {
				return 20
			}
			return termWidth
		}
		return minWidth
	}
	return w
}

// frameWidth returns the standard Corvidae frame width for the current terminal size.
func (m Model) frameWidth() int {
	return responsiveFrameWidth(m.Width, defaultFrameWidth, minFrameWidth)
}

// renderFrame wraps content in a fixed-width, left-justified container ready for lipgloss.Place.
func renderFrame(frameWidth int, content string) string {
	return lipgloss.NewStyle().
		Width(frameWidth).
		Align(lipgloss.Left).
		Render(content)
}

// renderSplitLine places left on the left and right pinned to the right edge of width.
func renderSplitLine(left, right string, width int) string {
	if right == "" {
		return left
	}
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	gap := width - leftW - rightW
	if gap >= 1 {
		return left + strings.Repeat(" ", gap) + right
	}
	return left + " " + right
}

// renderDivider renders a subtle horizontal rule spanning width columns.
func renderDivider(width int) string {
	if width <= 0 {
		return ""
	}
	return QuizDividerStyle.Render(strings.Repeat("─", width))
}

// renderProgressBar renders a styled [████░░░░] progress bar for current/total items.
func renderProgressBar(current, total, barLen int) string {
	if barLen <= 0 {
		return ""
	}
	filled := 0
	if total > 0 {
		filled = int(float64(barLen) * float64(current) / float64(total))
	}
	if filled < 0 {
		filled = 0
	}
	if filled > barLen {
		filled = barLen
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barLen-filled)
	return QuizProgressStyle.Render("[" + bar + "]")
}

// formatStreakBadge formats a streak counter, highlighting streaks of 3+ in warm gold.
func formatStreakBadge(streak int) string {
	if streak >= 3 {
		return HotStreakStyle.Render(fmt.Sprintf("★ Streak: %d", streak))
	}
	return HintStyle.Render(fmt.Sprintf("Streak: %d", streak))
}

// renderCompleteCard renders a unified bordered completion card across practice modes.
func renderCompleteCard(frameWidth int, title, bar string, stats []string, controls string) string {
	var lines []string
	lines = append(lines, CorrectStyle.Render(title))
	lines = append(lines, "")
	if bar != "" {
		lines = append(lines, bar)
	}
	for _, s := range stats {
		lines = append(lines, s)
	}
	lines = append(lines, "")
	lines = append(lines, controls)

	boxInnerWidth := frameWidth - 2
	if boxInnerWidth < 20 {
		boxInnerWidth = 20
	}
	return QuizCompleteBoxStyle.Width(boxInnerWidth).Render(strings.Join(lines, "\n"))
}

// fitBlockHeight ensures s has exactly height lines by padding with empty lines or clamping excess lines.
func fitBlockHeight(s string, height int) string {
	if height <= 0 {
		return ""
	}
	var lines []string
	if s != "" {
		lines = strings.Split(s, "\n")
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

// wrapPathWithIndent wraps a file path into at most maxLines lines within width,
// preferring to break after '/' separators and applying a hanging indent under prefix.
func wrapPathWithIndent(prefix, path string, width, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	prefixW := runewidth.StringWidth(prefix)
	avail := width - prefixW
	if avail < 10 {
		avail = 10
	}

	pathLines := splitPathIntoLines(path, avail, maxLines)
	indent := strings.Repeat(" ", prefixW)

	out := make([]string, 0, len(pathLines))
	for i, line := range pathLines {
		if i == 0 {
			out = append(out, QuizSourceLabelStyle.Render(prefix)+QuizSourcePathStyle.Render(line))
		} else {
			out = append(out, indent+QuizSourcePathStyle.Render(line))
		}
	}
	return strings.Join(out, "\n")
}

// splitPathIntoLines breaks path into up to maxLines chunks of at most lineWidth visual columns,
// preferring to break after '/' characters.
func splitPathIntoLines(path string, lineWidth, maxLines int) []string {
	if lineWidth <= 0 || maxLines <= 0 {
		return nil
	}
	parts := strings.Split(path, "/")
	tokens := make([]string, 0, len(parts))
	for i, p := range parts {
		if i < len(parts)-1 {
			tokens = append(tokens, p+"/")
		} else if p != "" {
			tokens = append(tokens, p)
		}
	}

	var lines []string
	current := ""
	currentW := 0

	flush := func() {
		if current != "" {
			lines = append(lines, current)
			current = ""
			currentW = 0
		}
	}

	for _, tok := range tokens {
		tokW := runewidth.StringWidth(tok)
		if currentW > 0 && currentW+tokW > lineWidth {
			flush()
		}
		if tokW <= lineWidth {
			current += tok
			currentW += tokW
			continue
		}
		for _, r := range tok {
			rw := runewidth.RuneWidth(r)
			if currentW > 0 && currentW+rw > lineWidth {
				flush()
			}
			current += string(r)
			currentW += rw
		}
	}
	flush()

	if len(lines) == 0 {
		return []string{""}
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		last := lines[maxLines-1]
		lines[maxLines-1] = truncateToWidth(last, lineWidth)
	}
	return lines
}

// wrapLabeledText wraps a labeled feedback line (e.g., "Meaning: ...") with hanging indent under the label.
func wrapLabeledText(leadIndent, label, value string, valueStyle lipgloss.Style, frameWidth int) []string {
	prefix := leadIndent + label
	prefixW := runewidth.StringWidth(prefix)
	avail := frameWidth - prefixW
	if avail < 12 {
		avail = 12
	}

	wrapped := wrapPlainText(value, avail)
	hangIndent := strings.Repeat(" ", prefixW)
	out := make([]string, len(wrapped))
	for i, line := range wrapped {
		if i == 0 {
			out[i] = prefix + valueStyle.Render(line)
		} else {
			out[i] = hangIndent + valueStyle.Render(line)
		}
	}
	return out
}

// wrapPlainText wraps plain text to maxWidth visual columns on word boundaries,
// falling back to rune-level splits for words longer than maxWidth.
func wrapPlainText(text string, maxWidth int) []string {
	if maxWidth <= 0 || text == "" {
		return []string{text}
	}
	if runewidth.StringWidth(text) <= maxWidth {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	current := ""
	currentW := 0

	for _, w := range words {
		ww := runewidth.StringWidth(w)
		if currentW == 0 {
			if ww <= maxWidth {
				current = w
				currentW = ww
			} else {
				for _, r := range w {
					rw := runewidth.RuneWidth(r)
					if currentW > 0 && currentW+rw > maxWidth {
						lines = append(lines, current)
						current = ""
						currentW = 0
					}
					current += string(r)
					currentW += rw
				}
			}
			continue
		}
		if currentW+1+ww <= maxWidth {
			current += " " + w
			currentW += 1 + ww
		} else {
			lines = append(lines, current)
			current = ""
			currentW = 0
			if ww <= maxWidth {
				current = w
				currentW = ww
			} else {
				for _, r := range w {
					rw := runewidth.RuneWidth(r)
					if currentW > 0 && currentW+rw > maxWidth {
						lines = append(lines, current)
						current = ""
						currentW = 0
					}
					current += string(r)
					currentW += rw
				}
			}
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// truncateToWidth truncates s to at most maxWidth visual columns, appending "…" if truncated.
func truncateToWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}
	if maxWidth == 1 {
		return "…"
	}
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > maxWidth-1 {
			break
		}
		b.WriteRune(r)
		w += rw
	}
	b.WriteRune('…')
	return b.String()
}
