package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-touch/internal/types"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func getUserStats(config types.Config) (types.UserStats, error) {
	var stats types.UserStats
	_, err := os.Stat(config.Stats.FileDir)
	if err != nil {
		// returning empty if file not exists
		return stats, nil
	}
	data, err := os.ReadFile(config.Stats.FileDir)
	if err != nil {
		return stats, err
	}
	// Handle empty file
	if len(data) == 0 || len(strings.TrimSpace(string(data))) == 0 {
		stats.Sessions = make([]types.TypingSession, 0)
		return stats, nil
	}

	err = json.Unmarshal(data, &stats)
	if err != nil {
		return stats, err
	}

	// Ensure Sessions is never nil
	if stats.Sessions == nil {
		stats.Sessions = make([]types.TypingSession, 0)
	}

	return stats, err

}

type SessionResult struct {
	Error   error
	Session *types.TypingSession
	Exited  bool
}

// enum
type WelcomeAction int

const (
	StartSession WelcomeAction = iota
	Exit
)

func (w WelcomeAction) String() string {
	switch w {
	case StartSession:
		return "Lets Exercise"
	case Exit:
		return "Naah not today lets exit"
	default:
		return "Unknown Action"
	}
}

type welcomeModel struct {
	config   types.Config
	stats    types.UserStats
	cursor   int
	selected *WelcomeAction // stores the selected action
	choices  []WelcomeAction
	done     bool // indicates user made a selection
}

func newWelcomeModel(config types.Config, stats types.UserStats) welcomeModel {
	return welcomeModel{
		config: config,
		stats:  stats,
		cursor: 0,
		choices: []WelcomeAction{
			StartSession,
			Exit,
		},
	}
}

func (m welcomeModel) Init() tea.Cmd {
	return nil
}

func (m welcomeModel) View() string {
	s := "Welcome to GoTouch!\n\n"

	// Menu options
	for i, choice := range m.choices {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor points to current selection
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice.String())
	}

	s += "\nUse arrow keys to navigate, Enter to select, q to quit.\n"
	return s
}

func (m welcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1 // wrap to bottom
			}

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0 // wrap to top
			}

		case "enter", " ":
			selectedAction := m.choices[m.cursor]
			m.selected = &selectedAction
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func showWelcome(config types.Config, stats types.UserStats) (WelcomeAction, error) {
	model := newWelcomeModel(config, stats)

	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return Exit, err
	}

	welcomeModel := finalModel.(welcomeModel)
	if welcomeModel.selected != nil {
		return *welcomeModel.selected, nil
	}

	return Exit, nil
}

type sessionModel struct {
	text           string
	typedText      string
	errors         int
	cursor         int
	startTime      time.Time       // when session started
	lastKeyTime    time.Time       // last keystroke time
	keyStrokeTimes []time.Duration // time per word
	completed      bool            // session finished
	quit           bool            // user quit early
	hasStarted     bool
	width          int // terminal width
	height         int // terminal height
}

func (m sessionModel) Init() tea.Cmd {
	return nil
}

func (m sessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Exit keys
		switch msg.String() {
		case "esc", "ctrl+c":
			m.quit = true
			return m, tea.Quit

		case "enter":
			// Start the session on first enter press
			if !m.hasStarted {
				m.hasStarted = true
				m.startTime = time.Now()
				m.lastKeyTime = m.startTime
				return m, nil
			}

		case "backspace":
			if len(m.typedText) > 0 {
				m.typedText = m.typedText[:len(m.typedText)-1]
			}
			return m, nil
		}

		// Only process character input if session has started
		if !m.hasStarted {
			return m, nil
		}

		// Handle regular character input
		key := msg.String()

		// Only process single character keys or space
		if len(key) == 1 || key == "space" {
			if key == "space" {
				key = " "
			}

			// Add the character to typed text
			m.typedText += key

			// Track timing
			currentTime := time.Now()
			m.keyStrokeTimes = append(m.keyStrokeTimes, currentTime.Sub(m.lastKeyTime))
			m.lastKeyTime = currentTime

			// Check if character is incorrect
			if len(m.typedText) <= len(m.text) {
				if m.typedText[len(m.typedText)-1] != m.text[len(m.typedText)-1] {
					m.errors++
				}
			}

			// Check if completed
			if m.typedText == m.text {
				m.completed = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m sessionModel) View() string {
	if !m.hasStarted {
		return "Press ENTER to start\n\nPress ESC or CTRL-C to exit"
	}

	var result strings.Builder

	// Set default width if not set yet
	maxWidth := m.width
	if maxWidth == 0 {
		maxWidth = 80 // Default fallback
	}

	// Reserve space for margins and the footer
	maxWidth -= 4 // Some padding

	currentLineWidth := 0

	// Display each character with appropriate color, wrapping at terminal width
	for i, char := range m.text {
		charStr := string(char)

		// Handle newlines in the source text
		if char == '\n' {
			result.WriteString("\n")
			currentLineWidth = 0
			continue
		}

		// Wrap to next line if we exceed terminal width
		if currentLineWidth >= maxWidth && char != ' ' {
			result.WriteString("\n")
			currentLineWidth = 0
		}

		// Apply styling based on typing status
		var styledChar string
		if i < len(m.typedText) {
			// Character has been typed
			if m.typedText[i] == m.text[i] {
				styledChar = DefaultTheme.Correct.Render(charStr) // Green
			} else {
				styledChar = DefaultTheme.Incorrect.Render(charStr) // Red
			}
		} else if i == len(m.typedText) {
			// Current cursor position
			styledChar = DefaultTheme.Current.Render(charStr) // Yellow/highlighted
		} else {
			// Not yet typed
			styledChar = DefaultTheme.Normal.Render(charStr) // Default
		}

		result.WriteString(styledChar)

		// Update line width (account for visual width)
		if char == ' ' {
			currentLineWidth++
			// Break line after space if needed
			if currentLineWidth >= maxWidth {
				result.WriteString("\n")
				currentLineWidth = 0
			}
		} else {
			currentLineWidth++
		}
	}

	result.WriteString("\n\nPress ESC or CTRL-C to exit")

	return result.String()
}

func startSession(config types.Config, text string) (types.TypingSession, error) {
	// Create initial model
	model := sessionModel{
		text:           text,
		typedText:      "",
		errors:         0,
		cursor:         0,
		keyStrokeTimes: make([]time.Duration, 0),
		completed:      false,
		quit:           false,
		hasStarted:     false,
		width:          0, // Will be set by WindowSizeMsg
		height:         0, // Will be set by WindowSizeMsg
	}

	// Run the Bubbletea program with alternate screen
	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return types.TypingSession{}, err
	}

	// Cast back to sessionModel
	session := finalModel.(sessionModel)

	// If user quit early, return error
	if session.quit {
		return types.TypingSession{}, errors.New("session cancelled by user")
	}

	// Calculate duration
	duration := session.lastKeyTime.Sub(session.startTime)

	// Calculate WPM (Words Per Minute)
	// Standard: 1 word = 5 characters
	words := float32(len(session.text)) / 5.0
	minutes := duration.Minutes()
	wpm := words / float32(minutes)

	// Calculate accuracy
	totalChars := len(session.typedText)
	correctChars := totalChars - session.errors
	accuracy := float32(0)
	if totalChars > 0 {
		accuracy = (float32(correctChars) / float32(totalChars)) * 100
	}

	// Create the typing session result
	result := types.TypingSession{
		Date:     session.startTime,
		WPM:      wpm,
		Accuracy: accuracy,
		Errors:   session.errors,
		Duration: duration,
	}

	return result, nil
}

func Run(config types.Config, text string) SessionResult {
	stats, err := getUserStats(config)
	if err != nil {
		return SessionResult{
			Error:   err,
			Session: nil,
			Exited:  false,
		}
	}

	action, err := showWelcome(config, stats)
	if err != nil {
		return SessionResult{
			Error:   err,
			Session: nil,
			Exited:  false,
		}
	}

	switch action {
	case StartSession:
		session, err := startSession(config, text)
		if err != nil {
			return SessionResult{Error: err, Session: nil, Exited: false}
		}
		return SessionResult{Error: nil, Session: &session, Exited: false}

	case Exit:
		return SessionResult{Error: nil, Session: nil, Exited: true}

	default:
		return SessionResult{
			Error:   errors.New("unknown action for welcome screen (this is a bug)"),
			Session: nil,
			Exited:  false,
		}
	}
}
