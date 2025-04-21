package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

var (
	CheckMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	XMark     = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).SetString("X")

	TextGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	TextBlue   = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	TextYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	TextDanger = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	TextPrimary   = lipgloss.NewStyle().Foreground(lipgloss.Color("-1"))
	TextSecondary = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	TextTertiary  = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

	TextLink = lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Underline(true)

	Spinner = spinner.New(
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("69"))),
		spinner.WithSpinner(spinner.Jump),
	)
)
