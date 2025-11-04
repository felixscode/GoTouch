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
			Foreground(lipgloss.Color("2")), // Terminal green
		Incorrect: lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")), // Terminal red
		Current: lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("3")), // Black text on terminal yellow
		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")), // Terminal white/default
		Background: lipgloss.NewStyle().
			Background(lipgloss.Color("0")), // Terminal black
	}

	DarkTheme = Theme{
		Name: "Dark",
		Correct: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")), // Bright terminal green
		Incorrect: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")), // Bright terminal red
		Current: lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("11")), // Black on bright terminal yellow
		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")), // Bright terminal white
		Background: lipgloss.NewStyle().
			Background(lipgloss.Color("0")), // Terminal black
	}
)
