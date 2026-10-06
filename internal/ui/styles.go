package ui

import "github.com/charmbracelet/lipgloss"

var (
	CharStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7CCB"))
	PinyinStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#569CD6"))
	PronunciationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#569CD6"))
	MeaningStyle       = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#ebebeb"))
	ExplanationStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")).Italic(true).Width(60).Align(lipgloss.Center)
	HintStyle          = lipgloss.NewStyle().Faint(false)
	KeyStyle           = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4EC9B0"))
	CursorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7CCB")).Bold(true)
	ErrorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)
	CorrectStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)
	LogoStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9")).Width(60).Align(lipgloss.Left)

	// Arabic Academy styles
	ArabicHeaderStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C"))
	ArabicSubHeaderStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Italic(true)
	ArabicGlyphHuge         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7CCB"))
	ArabicCardBox           = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#BD93F9")).Padding(1, 2)
	ArabicFormBox           = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#6272A4")).Padding(0, 1)
	ArabicFormLabel         = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Faint(false)
	ArabicFormGlyph         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B"))
	BadgeSunStyle           = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C"))
	BadgeMoonStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8BE9FD"))
	BadgeConnectorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	BadgeNonConnectorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)

	// Marked Cards styles
	MarkedBadgeStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C")).Background(lipgloss.Color("#44475A")).Padding(0, 1)
	MarkedStarStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F1FA8C"))

	// Memorize by Options styles
	MemorizeTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9"))
	MemorizeHeaderStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C"))
	MemorizeSubStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Italic(true)
	MemorizeVerseNum      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8BE9FD"))
	MemorizeCompleted     = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))
	MemorizePreviousText  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
	MemorizeTargetBlank   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7CCB")).Background(lipgloss.Color("#44475A")).Padding(0, 1)
	MemorizeOptionBox     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#6272A4")).Padding(0, 1)
	MemorizeOptionActive  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FF7CCB")).Padding(0, 1)
	MemorizeSuccessStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B"))
	MemorizeWrongStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555"))
	MemorizeProgressStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)

	// Learn by Reading styles
	ReadingHeaderStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9"))
	ReadingSourceStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Italic(true)
	ReadingCategoryBadge       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C")).Background(lipgloss.Color("#44475A")).Padding(0, 1)
	ReadingWordSubstituted     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B"))
	ReadingWordPinyin          = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Italic(true)
	ReadingKeyBadge            = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7CCB"))
	ReadingAggressivenessStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C"))
	ReadingBoxStyle            = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#BD93F9")).Padding(1, 3).Width(78).Align(lipgloss.Center)
	ReadingDiscussionCard      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#50FA7B")).Padding(1, 3).Width(78).Align(lipgloss.Left)
	ReadingWordDetailCard      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FF7CCB")).Padding(1, 3).Width(68).Align(lipgloss.Center)
)

const AsciiLogo = `                           .-.
                          ( o>
                          / ) \
                         ='---'=
                           m m
    ____                          _       _                
   / ___|   ___    _ __  __   __ (_)   __| |   __ _    ___ 
  | |      / _ \  | '__| \ \ / / | |  / _' |  / _' |  / _ \
  | |___  | (_) | | |     \ V /  | | | (_| | | (_| | |  __/
   \____|  \___/  |_|      \_/   |_|  \__,_|  \__,_|  \___|`
