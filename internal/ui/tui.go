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

	program := tea.NewProgram(model)
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
	has_started    bool
}

func (m sessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

	}
}

func (m sessionModel) View() string {
	if !m.hasStarted { // Add this field to your struct
		return "Press ENTER to start\n\nPress ESC or CTRL-C to exit"
	}

	var result strings.Builder

	// Display each character with appropriate color
	for i, char := range m.text {
		charStr := string(char)

		if i < len(m.typedText) {
			// Character has been typed
			if m.typedText[i] == m.text[i] {
				result.WriteString(correctStyle.Render(charStr)) // Green
			} else {
				result.WriteString(incorrectStyle.Render(charStr)) // Red
			}
		} else if i == len(m.typedText) {
			// Current cursor position
			result.WriteString(currentStyle.Render(charStr)) // Yellow/highlighted
		} else {
			// Not yet typed
			result.WriteString(normalStyle.Render(charStr)) // Default
		}
	}

	result.WriteString("\n\nPress ESC or CTRL-C to exit")

	return result.String()
}

// Placeholder for the actual typing session - you'll need to implement this
func startSession(config types.Config, text string) (types.TypingSession, error) {
	// TODO: Implement the actual typing session logic
	// For now, return a dummy session
	return types.TypingSession{}, errors.New("startSession not implemented yet")
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
