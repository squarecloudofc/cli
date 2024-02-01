package ui

import "github.com/charmbracelet/lipgloss"

var (
	CheckMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	XMark     = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).SetString("X")

	GreenText  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	BlueText   = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	YellowText = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	RedText    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)
