package ui

import (
	"fmt"
	"strings"
)

type helpBinding struct {
	key  string
	desc string
}

func (m Model) canToggleHelpWithQuestionMark() bool {
	switch m.State {
	case StateDirSelect, StateReading:
		return false
	default:
		return true
	}
}

func (m Model) viewHelpOverlay() string {
	frameWidth := responsiveFrameWidth(m.Width, 72, 40)
	boxInnerWidth := frameWidth - 2
	if boxInnerWidth < 24 {
		boxInnerWidth = 24
	}
	contentWidth := boxInnerWidth - 4 // Padding(1, 2)
	if contentWidth < 20 {
		contentWidth = 20
	}

	title, bindings := m.contextHelpBindings()

	var b strings.Builder
	b.WriteString(renderSplitLine(SectionTitleStyle.Render("? Keyboard Shortcuts"), ModePillStyle.Render(title), contentWidth) + "\n")
	b.WriteString(renderDivider(contentWidth) + "\n\n")

	for _, hb := range bindings {
		keyPart := KeyStyle.Render(fmt.Sprintf("%-16s", hb.key))
		b.WriteString(fmt.Sprintf("  %s %s\n", keyPart, hb.desc))
	}

	b.WriteString("\n" + renderDivider(contentWidth) + "\n")
	b.WriteString(SectionTitleStyle.Render("Hummingbird 30-Key Layout") + HintStyle.Render(" (groups of 5)") + "\n")
	b.WriteString("  " + HintStyle.Render("Home:   ") + KeyStyle.Render("a s d f g") + HintStyle.Render("  │  ") + KeyStyle.Render("h j k l ;") + "\n")
	b.WriteString("  " + HintStyle.Render("Bottom: ") + KeyStyle.Render("z x c v b") + HintStyle.Render("  │  ") + KeyStyle.Render("n m , . /") + "\n")
	b.WriteString("  " + HintStyle.Render("Top:    ") + KeyStyle.Render("q w e r t") + HintStyle.Render("  │  ") + KeyStyle.Render("y u i o p") + "\n\n")

	dismiss := fmt.Sprintf("[%s] Close Help", KeyStyle.Render("Esc / ? / F1 / Space"))
	b.WriteString(HintStyle.Render(dismiss))

	return HelpBoxStyle.Width(boxInnerWidth).Render(b.String())
}

func (m Model) contextHelpBindings() (string, []helpBinding) {
	switch m.State {
	case StateModeSelect:
		return "Main Menu", []helpBinding{
			{"a - h", "Jump directly to a study mode"},
			{"↑ / ↓ / j / k", "Move mode selection cursor"},
			{"Enter / Space", "Confirm selected mode"},
			{"m", "Practice Marked (starred) cards"},
			{"i", "Open Study Dashboard & Heatmap"},
			{"c", "Open Settings (config.yaml)"},
			{"q / Ctrl+C", "Quit Corvidae"},
		}
	case StateDirSelect:
		return "Folders", []helpBinding{
			{"Hummingbird key", "Open language or topic folder"},
			{"Shift + key", "Toggle selection of entire folder"},
			{"Tab", "Start practice with selected decks"},
			{"u / Ctrl+D", "Deselect all decks across folders"},
			{"m", "Practice Marked Cards"},
			{"Esc / b", "Go up a level / Main Menu"},
		}
	case StateDeckSelect:
		return "Deck Select", []helpBinding{
			{"Hummingbird / Space", "Toggle deck selection"},
			{"Shift + Enter", "Toggle all decks in this folder"},
			{"Enter / Tab", "Start practice with selected decks"},
			{"u / Ctrl+D", "Deselect all decks"},
			{"Esc / b", "Back to folders"},
		}
	case StateReview:
		return "SM-2 Review", []helpBinding{
			{"Spacebar", "Reveal answer"},
			{"d / f / g", "Grade: Blackout / Wrong / Hard"},
			{"h / j / k", "Grade: Good / Easy / Perfect"},
			{"e", "Toggle card explanation"},
			{"m / u", "Mark or unmark card"},
			{"r", "Cram / review all cards (when done)"},
			{"Esc / q", "Main Menu / Quit"},
		}
	case StateQuiz:
		return "Quiz Mode", []helpBinding{
			{"a-g or 1-5", "Select multiple-choice answer"},
			{"Space / Enter", "Advance after feedback"},
			{"k / w", "I knew it! (override misclick)"},
			{"S (Shift+S)", "Toggle Quiz → SM-2 Sync"},
			{"m / u", "Mark or unmark card"},
			{"r", "Retry quiz (on completion)"},
			{"Esc / q", "Return to Main Menu"},
		}
	case StateMatch:
		return "Matching Game", []helpBinding{
			{"a-g or 1-5", "Select left column (Character)"},
			{"h-; or 6-0", "Select right column (Meaning)"},
			{"p", "Toggle Pinyin visibility"},
			{"S (Shift+S)", "Toggle Match → SM-2 Sync"},
			{"m", "Mark selected or last matched card"},
			{"Esc / q", "Return to Main Menu"},
		}
	case StateMemorizeSelectText, StateMemorizePortionSelect, StateMemorize, StateMemorizeComplete:
		return "Memorize", []helpBinding{
			{"a-g or 1-5", "Choose next word (Options mode)"},
			{"a-z", "Type first letter (Typing mode)"},
			{"t / Tab", "Toggle Options ↔ First-Letter Typing"},
			{"o / 2-5", "Change number of options (2–5)"},
			{"d", "Select next due sections (Portion Select)"},
			{"k / w", "I knew it! (override mistake)"},
			{"Tab / ← → / [ ]", "Adjust portion Start / End bounds"},
			{"r / Esc", "Restart portion / Go back"},
		}
	case StateDashboard:
		return "Study Dashboard", []helpBinding{
			{"r", "Refresh card maturity & stats"},
			{"c", "Open Settings"},
			{"Esc / Enter", "Return to Main Menu"},
			{"q", "Quit Corvidae"},
		}
	case StateSettings:
		return "Settings", []helpBinding{
			{"a - g / Space", "Toggle or cycle selected setting"},
			{"↑ / ↓ / j / k", "Move settings cursor"},
			{"Esc / b", "Return to Main Menu"},
			{"q", "Quit Corvidae"},
		}
	case StateReadingSelectText, StateReading, StateReadingWordDetail, StateReadingDiscussion, StateReadingComplete:
		return "Learn by Reading", []helpBinding{
			{"Space / Enter / n", "Next sentence / Discussion"},
			{"p / ←", "Previous sentence"},
			{"Hummingbird key", "Inspect substituted word"},
			{"h / d", "Toggle Grammar & Discussion card"},
			{"- / + / 0", "Tune substitution aggressiveness"},
			{"m / k (in word)", "Mark as flashcard / Mark known"},
		}
	default:
		return "Arabic Academy", []helpBinding{
			{"Hummingbird key", "Select activity, stage, or answer"},
			{"← → / h l", "Browse letters or auxiliary marks"},
			{"Tab", "Switch Letters ↔ Auxiliary tabs"},
			{"Space / Enter", "Advance after drill feedback"},
			{"Esc / b", "Back to Academy menu"},
		}
	}
}
