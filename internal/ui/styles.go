package ui

import "github.com/charmbracelet/lipgloss"

var (
	CharStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7CCB"))
	PinyinStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#569CD6"))
	PronunciationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#569CD6"))
	MeaningStyle       = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#ebebeb"))
	ExplanationStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")).Italic(true).Width(60).Align(lipgloss.Center)
	HintStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#8E95B4"))
	KeyStyle           = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4EC9B0"))
	CursorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7CCB")).Bold(true)
	ErrorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)
	CorrectStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)
	LogoStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9")).Width(60).Align(lipgloss.Left)

	// Shared navigation, dashboard & semantic accents
	SectionTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9"))
	BreadcrumbStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F8F8F2"))
	ModePillStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8BE9FD")).Background(lipgloss.Color("#2D3047")).Padding(0, 1)
	DashboardPillStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")).Background(lipgloss.Color("#2D3047")).Padding(0, 1)
	HotStreakStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C"))
	HelpBoxStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#8BE9FD")).Padding(1, 2).Align(lipgloss.Left)

	// SM-2 Grade styles
	GradeBlackoutStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555"))
	GradeWrongStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7979"))
	GradeHardStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C"))
	GradeGoodStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4EC9B0"))
	GradeEasyStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B"))
	GradePerfectStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#69FF94"))

	// Folder selection styles
	FolderSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B"))
	FolderSubsetStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C"))

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
	QuizProgressStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)
	QuizDividerStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A"))
	QuizPromptBoxStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#BD93F9")).Padding(0, 2).Align(lipgloss.Left)
	QuizExplanationStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")).Italic(true).Align(lipgloss.Left)
	QuizSourceLabelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Bold(true)
	QuizSourcePathStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Italic(true)
	QuizCompleteBoxStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#50FA7B")).Padding(1, 3).Align(lipgloss.Left)

	// Learn by Reading styles
	ReadingHeaderStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9"))
	ReadingSourceStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Italic(true)
	ReadingCategoryBadge       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C")).Background(lipgloss.Color("#44475A")).Padding(0, 1)
	ReadingWordSubstituted     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B"))
	ReadingWordPinyin          = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Italic(true)
	ReadingKeyBadge            = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7CCB"))
	ReadingAggressivenessStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C"))
	ReadingBoxStyle            = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#BD93F9")).Padding(1, 3).Align(lipgloss.Left)
	ReadingDiscussionCard      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#50FA7B")).Padding(1, 3).Align(lipgloss.Left)
	ReadingWordDetailCard      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FF7CCB")).Padding(1, 3).Align(lipgloss.Left)

	// Matching Game styles
	MatchColumnHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9"))
	MatchCardBox           = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#6272A4")).Padding(0, 2)
	MatchCardSelectedBox   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FF7CCB")).Padding(0, 2).Bold(true)
	MatchCardMatchedBox    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#50FA7B")).Foreground(lipgloss.Color("#6272A4")).Padding(0, 2)
	MatchCardMismatchBox   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FF5555")).Padding(0, 2)

	// Dashboard Heatmap & Due Badge styles
	HeatmapLevel0Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A"))
	HeatmapLevel1Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#2D6A4F"))
	HeatmapLevel2Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#40916C"))
	HeatmapLevel3Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	HeatmapLevel4Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#95FFB0")).Bold(true)
	DueBadgeStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C"))
	SyncOnPillStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B")).Background(lipgloss.Color("#2D3047")).Padding(0, 1)

	// Morpheme Forge styles
	ForgeTitleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7CCB"))
	ForgeChipsStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8BE9FD"))
	ForgeMultStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555"))
	ForgeCoinStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F1FA8C"))
	ForgeBossPillStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555")).Background(lipgloss.Color("#3D1E2E")).Padding(0, 1)
	ForgeRelicPillStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F1FA8C")).Background(lipgloss.Color("#2D3047")).Padding(0, 1)
	ForgeCardGlassBox     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FF7CCB")).Padding(0, 1)
	ForgeCardFoilBox      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#BD93F9")).Padding(0, 1)
	ForgeCardGoldBox      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#F1FA8C")).Padding(0, 1)
	ForgeCardStagedBox    = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("#50FA7B")).Padding(0, 1)
	ForgeSuitNatureStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B"))
	ForgeSuitHumanStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7CCB"))
	ForgeSuitCraftStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C"))
	ForgeSuitCoreStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8BE9FD"))
	ForgeStageBoxStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#8BE9FD")).Padding(0, 2)
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
