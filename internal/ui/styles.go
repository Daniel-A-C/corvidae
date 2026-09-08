package ui

import "github.com/charmbracelet/lipgloss"

var (
	CharStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7CCB"))
	PinyinStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#569CD6"))
	MeaningStyle     = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#ebebeb"))
	ExplanationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")).Italic(true).Width(60).Align(lipgloss.Center)
	HintStyle        = lipgloss.NewStyle().Faint(false)
	KeyStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4EC9B0"))
	CursorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7CCB")).Bold(true)
	ErrorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)
	CorrectStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)
	LogoStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9")).Width(60).Align(lipgloss.Left)
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
