package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name       string
	Correct    lipgloss.Style
	Incorrect  lipgloss.Style
	Current    lipgloss.Style
	Normal     lipgloss.Style
	Background lipgloss.Style
}

var (
	DefaultTheme = Theme{
		Name: "Default",
		Correct: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")), // Green
		Incorrect: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Background(lipgloss.Color("#FFCCCC")), // Red with light background
		Current: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Background(lipgloss.Color("#333333")), // Yellow with dark background
		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")), // White
		Background: lipgloss.NewStyle().
			Background(lipgloss.Color("#000000")), // Black background
	}

	DarkTheme = Theme{
		Name: "Dark",
		Correct: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00AA00")), // Darker green
		Incorrect: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AA0000")).
			Background(lipgloss.Color("#330000")), // Dark red
		Current: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFF00")), // Black on yellow
		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")), // Light gray
		Background: lipgloss.NewStyle().
			Background(lipgloss.Color("#111111")), // Very dark background
	}
)
