package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-touch/internal/sources"
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

func saveUserStats(config types.Config, stats types.UserStats) error {
	// Ensure the directory exists
	// Marshal to pretty JSON
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	err = os.WriteFile(config.Stats.FileDir, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// analyzeErrors compares typed text with target text and returns error patterns
func analyzeErrors(typed, target string) (errorChars []rune, problemWords []string) {
	errorCharMap := make(map[rune]int)
	problemWordMap := make(map[string]bool)

	// Split into words
	typedWords := strings.Fields(typed)
	targetWords := strings.Fields(target)

	// Compare character by character
	minLen := len(typed)
	if len(target) < minLen {
		minLen = len(target)
	}

	for i := 0; i < minLen; i++ {
		if typed[i] != target[i] {
			errorCharMap[rune(target[i])]++
		}
	}

	// Compare words to find problematic ones
	for i := 0; i < len(typedWords) && i < len(targetWords); i++ {
		if typedWords[i] != targetWords[i] {
			problemWordMap[targetWords[i]] = true
		}
	}

	// Convert maps to slices
	for char := range errorCharMap {
		errorChars = append(errorChars, char)
	}

	for word := range problemWordMap {
		problemWords = append(problemWords, word)
	}

	return errorChars, problemWords
}

type SessionResult struct {
	Error   error
	Session *types.TypingSession
	Exited  bool
}

func (s SessionResult) String() string {
	if s.Exited {
		return "Session exited without starting."
	}

	if s.Error != nil {
		return fmt.Sprintf("Session ended with error: %v", s.Error)
	}

	if s.Session == nil {
		return "No session data available."
	}

	// Format the session stats nicely
	var result strings.Builder
	result.WriteString("\n╔════════════════════════════════════════╗\n")
	result.WriteString("║          SESSION SUMMARY               ║\n")
	result.WriteString("╚════════════════════════════════════════╝\n")
	result.WriteString(fmt.Sprintf("  WPM:       %.0f\n", s.Session.WPM))
	result.WriteString(fmt.Sprintf("  Accuracy:  %.1f%%\n", s.Session.Accuracy))
	result.WriteString(fmt.Sprintf("  Errors:    %d\n", s.Session.Errors))
	result.WriteString(fmt.Sprintf("  Duration:  %s\n", formatDuration(s.Session.Duration)))
	result.WriteString("\nSession saved successfully!\n")

	return result.String()
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

// dashboardModel shows post-session statistics
type dashboardModel struct {
	config         types.Config
	currentSession types.TypingSession
	allStats       types.UserStats
	width          int
	height         int
}

func newDashboardModel(config types.Config, session types.TypingSession, stats types.UserStats) dashboardModel {
	return dashboardModel{
		config:         config,
		currentSession: session,
		allStats:       stats,
	}
}

func (m dashboardModel) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Only Enter key exits the dashboard
		if msg.String() == "enter" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m dashboardModel) View() string {
	var s strings.Builder

	s.WriteString("╔════════════════════════════════════════╗\n")
	s.WriteString("║        SESSION COMPLETE!               ║\n")
	s.WriteString("╚════════════════════════════════════════╝\n\n")

	// Current session stats
	s.WriteString("📊 Your Performance:\n")
	s.WriteString(fmt.Sprintf("   WPM:      %.0f\n", m.currentSession.WPM))
	s.WriteString(fmt.Sprintf("   Accuracy: %.1f%%\n", m.currentSession.Accuracy))
	s.WriteString(fmt.Sprintf("   Errors:   %d\n", m.currentSession.Errors))
	s.WriteString(fmt.Sprintf("   Duration: %s\n\n", formatDuration(m.currentSession.Duration)))

	// Historical stats
	if len(m.allStats.Sessions) > 0 {
		avgWPM, bestWPM, avgAccuracy := calculateHistoricalStats(m.allStats)

		s.WriteString("📈 Historical Stats:\n")
		s.WriteString(fmt.Sprintf("   Average WPM:    %.0f\n", avgWPM))
		s.WriteString(fmt.Sprintf("   Best WPM:       %.0f\n", bestWPM))
		s.WriteString(fmt.Sprintf("   Avg Accuracy:   %.1f%%\n", avgAccuracy))
		s.WriteString(fmt.Sprintf("   Total Sessions: %d\n\n", len(m.allStats.Sessions)))
	}

	s.WriteString("Press Enter to exit...")

	return s.String()
}

func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func calculateHistoricalStats(stats types.UserStats) (avgWPM, bestWPM, avgAccuracy float32) {
	if len(stats.Sessions) == 0 {
		return 0, 0, 0
	}

	var totalWPM, totalAccuracy float32
	bestWPM = 0

	for _, session := range stats.Sessions {
		totalWPM += session.WPM
		totalAccuracy += session.Accuracy
		if session.WPM > bestWPM {
			bestWPM = session.WPM
		}
	}

	avgWPM = totalWPM / float32(len(stats.Sessions))
	avgAccuracy = totalAccuracy / float32(len(stats.Sessions))

	return avgWPM, bestWPM, avgAccuracy
}

func showDashboard(config types.Config, session types.TypingSession, stats types.UserStats) error {
	model := newDashboardModel(config, session, stats)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}

type sessionModel struct {
	text             string
	typedText        string
	errors           int
	cursor           int
	startTime        time.Time       // when session started
	lastKeyTime      time.Time       // last keystroke time
	keyStrokeTimes   []time.Duration // time per word
	completed        bool            // session finished
	quit             bool            // user quit early
	hasStarted       bool
	width            int           // terminal width
	height           int           // terminal height
	viewportOffset   int           // horizontal scroll position for centered cursor
	sessionDuration  time.Duration // total session duration
	selectedDuration int           // selected duration in minutes (for setup UI)

	// LLM pregeneration fields
	isLLMSource          bool               // Flag for LLM mode
	llmSource            *sources.LLMSource // LLM source for generating text
	lastSentence         string             // Previous sentence for context
	errorPatterns        map[rune]int       // Track char error frequency
	problemWords         []string           // Words with mistakes
	generationPending    bool               // Is LLM call in progress?
	nextSentenceReady    bool               // Next sentence generated?
	nextSentenceBuffer   string             // Buffered next sentence
	generationChan       chan string        // Channel for async generation
	generationErrChan    chan error         // Channel for generation errors
	pregenerateThreshold int                // Chars before end to trigger
	currentSentenceEndPos int               // Position where current sentence ends
	config               types.Config       // Config for LLM settings
}

func (m sessionModel) Init() tea.Cmd {
	return tea.Batch(
		tea.WindowSize(),
		tickCmd(),
	)
}

// tickCmd returns a command that ticks every 100ms to update the timer
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time

type generationCompleteMsg struct {
	sentence string
}

type generationErrorMsg struct {
	err error
}

func (m sessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		// Check if session time has expired
		if m.hasStarted && !m.completed && !m.quit {
			elapsed := time.Since(m.startTime)
			if elapsed >= m.sessionDuration {
				m.completed = true
				return m, tea.Quit
			}
		}

		// Check for pregeneration trigger
		if m.isLLMSource && m.hasStarted && !m.generationPending && !m.nextSentenceReady {
			charsRemaining := m.currentSentenceEndPos - len(m.typedText)
			if charsRemaining <= m.pregenerateThreshold && charsRemaining > 0 {
				// Start async generation
				m.generationPending = true
				return m, tea.Batch(tickCmd(), m.generateNextSentenceCmd())
			}
		}

		return m, tickCmd() // Continue ticking

	case generationCompleteMsg:
		// LLM generation completed successfully - show it immediately
		m.nextSentenceBuffer = msg.sentence
		m.nextSentenceReady = true
		m.generationPending = false

		// Append the new sentence to display text immediately
		m.text += " " + msg.sentence

		return m, nil

	case generationErrorMsg:
		// LLM generation failed - fallback to dummy or just continue
		m.generationPending = false
		m.nextSentenceReady = false
		// Could log error or show message to user
		return m, nil

	case tea.KeyMsg:
		// Exit keys
		switch msg.String() {
		case "esc", "ctrl+c":
			m.quit = true
			return m, tea.Quit

		case "up":
			// Increase duration before session starts
			if !m.hasStarted {
				m.selectedDuration++
				if m.selectedDuration > 60 { // Max 60 minutes
					m.selectedDuration = 60
				}
				return m, nil
			}

		case "down":
			// Decrease duration before session starts
			if !m.hasStarted {
				m.selectedDuration--
				if m.selectedDuration < 1 { // Min 1 minute
					m.selectedDuration = 1
				}
				return m, nil
			}

		case "enter":
			// Start the session on first enter press
			if !m.hasStarted {
				m.hasStarted = true
				m.sessionDuration = time.Duration(m.selectedDuration) * time.Minute
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

			// Check if completed current sentence
			if len(m.typedText) >= m.currentSentenceEndPos {
				if m.isLLMSource {
					// LLM mode: transition to next sentence if ready
					if m.nextSentenceReady {
						// Analyze errors from the sentence just completed
						errorChars, problemWords := analyzeErrors(m.typedText[:m.currentSentenceEndPos], m.text[:m.currentSentenceEndPos])

						// Update for next sentence
						m.lastSentence = m.nextSentenceBuffer
						m.currentSentenceEndPos = len(m.text) // Move to end of newly appended text
						m.nextSentenceReady = false

						// Store error patterns for next generation
						for _, char := range errorChars {
							m.errorPatterns[char]++
						}
						m.problemWords = append(m.problemWords, problemWords...)

						return m, nil
					} else if m.generationPending {
						// Wait for generation to complete
						// Show loading indicator in View()
						return m, nil
					} else {
						// No more text available - check if we've typed everything
						if len(m.typedText) >= len(m.text) {
							m.completed = true
							return m, tea.Quit
						}
					}
				} else {
					// Normal mode: session complete
					m.completed = true
					return m, tea.Quit
				}
			}
		}
	}
	return m, nil
}

// getCurrentWPM calculates WPM based on current typing progress
func (m sessionModel) getCurrentWPM() float32 {
	if !m.hasStarted {
		return 0
	}

	elapsed := time.Since(m.startTime)
	if elapsed.Seconds() < 1 {
		return 0 // Avoid division by very small numbers
	}

	// Calculate based on characters typed so far
	words := float32(len(m.typedText)) / 5.0
	minutes := float32(elapsed.Minutes())
	return words / minutes
}

// getCurrentAccuracy calculates accuracy percentage based on current progress
func (m sessionModel) getCurrentAccuracy() float32 {
	totalChars := len(m.typedText)
	if totalChars == 0 {
		return 0
	}

	correctChars := totalChars - m.errors
	return (float32(correctChars) / float32(totalChars)) * 100
}

// generateNextSentenceCmd creates a command that generates the next sentence asynchronously
func (m sessionModel) generateNextSentenceCmd() tea.Cmd {
	return func() tea.Msg {
		// Convert error pattern map to slice
		errorChars := make([]rune, 0, len(m.errorPatterns))
		for char := range m.errorPatterns {
			errorChars = append(errorChars, char)
		}

		// Call LLM to generate next sentence
		nextSentence, err := m.llmSource.GetNextSentence(m.lastSentence, errorChars, m.problemWords)
		if err != nil {
			return generationErrorMsg{err: err}
		}

		return generationCompleteMsg{sentence: nextSentence}
	}
}

func (m sessionModel) View() string {
	if !m.hasStarted {
		return fmt.Sprintf("Session Duration: %d minutes\n\nUse UP/DOWN arrows to adjust\nPress ENTER to start\nPress ESC or CTRL-C to exit", m.selectedDuration)
	}

	// Wait for terminal dimensions before rendering
	if m.width == 0 {
		return "Loading..."
	}

	var result strings.Builder

	terminalWidth := m.width

	// Calculate and display stats header
	elapsed := time.Since(m.startTime)
	remaining := m.sessionDuration - elapsed
	currentWPM := m.getCurrentWPM()
	currentAccuracy := m.getCurrentAccuracy()

	// Display remaining time in minutes:seconds format
	remainingMinutes := int(remaining.Minutes())
	remainingSeconds := int(remaining.Seconds()) % 60

	statsHeader := fmt.Sprintf("Time Remaining: %d:%02d | WPM: %.0f | Accuracy: %.1f%% | Errors: %d",
		remainingMinutes,
		remainingSeconds,
		currentWPM,
		currentAccuracy,
		m.errors,
	)

	// Add loading indicator if generation is pending
	if m.generationPending {
		statsHeader += " | ⏳ Generating..."
	}

	statsHeader += "\n\n"
	result.WriteString(statsHeader)

	// Reserve space for margins
	displayWidth := terminalWidth - 4

	// Current cursor position (next character to type)
	cursorPos := len(m.typedText)

	// Calculate center position
	centerPos := displayWidth / 2

	// Calculate viewport window
	var viewportStart, viewportEnd int

	textLen := len(m.text)

	// If text is shorter than display width, no scrolling needed
	if textLen <= displayWidth {
		viewportStart = 0
		viewportEnd = textLen
	} else {
		// Calculate viewport to keep cursor centered
		viewportStart = cursorPos - centerPos
		viewportEnd = viewportStart + displayWidth

		// Adjust if we're at the beginning
		if viewportStart < 0 {
			viewportStart = 0
			viewportEnd = displayWidth
		}

		// Adjust if we're near the end
		if viewportEnd > textLen {
			viewportEnd = textLen
			viewportStart = textLen - displayWidth
			if viewportStart < 0 {
				viewportStart = 0
			}
		}
	}

	// Render visible characters
	// Note: cursor position on screen = (cursorPos - viewportStart)
	// The viewport calculation already handles centering
	for i := viewportStart; i < viewportEnd; i++ {
		char := m.text[i]
		charStr := string(char)

		// Skip newlines and tabs (as per user's notes)
		if char == '\n' || char == '\t' {
			continue
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
	}

	result.WriteString("\n\nPress ESC or CTRL-C to exit")

	return result.String()
}

func startSession(config types.Config, text string, textSource sources.TextSource) (types.TypingSession, error) {
	// Check if we're using LLM source
	llmSource, isLLM := textSource.(*sources.LLMSource)

	// Create initial model
	model := sessionModel{
		text:             text,
		typedText:        "",
		errors:           0,
		cursor:           0,
		keyStrokeTimes:   make([]time.Duration, 0),
		completed:        false,
		quit:             false,
		hasStarted:       false,
		width:            0, // Will be set by WindowSizeMsg
		height:           0, // Will be set by WindowSizeMsg
		viewportOffset:   0, // Start at beginning
		selectedDuration: 1, // Default to 1 minute
		sessionDuration:  0, // Will be set when session starts

		// LLM fields
		isLLMSource:           isLLM,
		llmSource:             llmSource,
		lastSentence:          text, // First sentence becomes context
		errorPatterns:         make(map[rune]int),
		problemWords:          make([]string, 0),
		generationPending:     false,
		nextSentenceReady:     false,
		nextSentenceBuffer:    "",
		generationChan:        make(chan string, 1),
		generationErrChan:     make(chan error, 1),
		pregenerateThreshold:  config.Text.LLM.PregenerateThreshold,
		currentSentenceEndPos: len(text), // Initialize to initial text length
		config:                config,
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

func Run(config types.Config, text string, textSource sources.TextSource) SessionResult {
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
		session, err := startSession(config, text, textSource)
		if err != nil {
			return SessionResult{Error: err, Session: nil, Exited: false}
		}

		// Add session to stats
		stats.Sessions = append(stats.Sessions, session)

		// Save updated stats
		err = saveUserStats(config, stats)
		if err != nil {
			// Log error but don't fail - we still show the results
			fmt.Fprintf(os.Stderr, "Warning: Failed to save stats: %v\n", err)
		}

		// Show dashboard with results
		err = showDashboard(config, session, stats)
		if err != nil {
			// Dashboard error shouldn't fail the whole thing
			fmt.Fprintf(os.Stderr, "Warning: Failed to show dashboard: %v\n", err)
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
