package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name         string
	Correct      lipgloss.Style
	Incorrect    lipgloss.Style
	Current      lipgloss.Style
	Normal       lipgloss.Style
	Background   lipgloss.Style
	Title        lipgloss.Style
	Subtitle     lipgloss.Style
	Border       lipgloss.Style
	Highlight    lipgloss.Style
	Muted        lipgloss.Style
	Success      lipgloss.Style
	Warning      lipgloss.Style
	Info         lipgloss.Style
	ProgressBar  lipgloss.Style
	ProgressFill lipgloss.Style
}

var (
	DefaultTheme = Theme{
		Name: "Default",
		Correct: lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")), // Terminal green - keep for typed text
		Incorrect: lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")), // Terminal red - keep for typed text
		Current: lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("6")), // Black text on cyan highlight
		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")), // Terminal white/default
		Background: lipgloss.NewStyle().
			Background(lipgloss.Color("0")), // Terminal black
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")). // Light grey
			Bold(true).
			MarginBottom(1),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Muted grey
			Italic(true),
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")). // Grey border
			Padding(1, 2),
		Highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")). // Cyan highlight
			Bold(true),
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")), // Dark grey
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")). // Light grey instead of green
			Bold(true),
		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Grey instead of yellow
			Bold(true),
		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")), // Light grey instead of blue
		ProgressBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")), // Dark grey
		ProgressFill: lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")), // Cyan fill
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
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")).
			Bold(true).
			MarginBottom(1),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("13")).
			Italic(true),
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("14")).
			Padding(1, 2),
		Highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true),
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true),
		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true),
		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")),
		ProgressBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		ProgressFill: lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")),
	}
)
