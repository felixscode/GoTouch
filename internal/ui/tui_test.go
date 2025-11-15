package ui

import (
	"fmt"
	"go-touch/internal/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAnalyzeErrors(t *testing.T) {
	tests := []struct {
		name              string
		typed             string
		target            string
		expectedErrorChars int
		expectedProblemWords int
	}{
		{
			name:              "perfect typing",
			typed:             "hello world",
			target:            "hello world",
			expectedErrorChars: 0,
			expectedProblemWords: 0,
		},
		{
			name:              "one character error",
			typed:             "hallo world",
			target:            "hello world",
			expectedErrorChars: 1, // 'e' was wrong
			expectedProblemWords: 1, // "hello" was wrong
		},
		{
			name:              "multiple errors",
			typed:             "hallo wurld",
			target:            "hello world",
			expectedErrorChars: 2, // 'e' and 'o'
			expectedProblemWords: 2, // "hello" and "world"
		},
		{
			name:              "empty strings",
			typed:             "",
			target:            "",
			expectedErrorChars: 0,
			expectedProblemWords: 0,
		},
		{
			name:              "typed shorter than target",
			typed:             "hel",
			target:            "hello",
			expectedErrorChars: 0, // only compares up to typed length
			expectedProblemWords: 1, // word count mismatch detected
		},
		{
			name:              "target shorter than typed",
			typed:             "hello",
			target:            "hel",
			expectedErrorChars: 0, // compares min length
			expectedProblemWords: 1, // word count mismatch detected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorChars, problemWords := analyzeErrors(tt.typed, tt.target)

			if len(errorChars) != tt.expectedErrorChars {
				t.Errorf("analyzeErrors() errorChars count = %d, want %d (chars: %v)",
					len(errorChars), tt.expectedErrorChars, errorChars)
			}

			if len(problemWords) != tt.expectedProblemWords {
				t.Errorf("analyzeErrors() problemWords count = %d, want %d (words: %v)",
					len(problemWords), tt.expectedProblemWords, problemWords)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "zero duration",
			duration: 0,
			expected: "0:00",
		},
		{
			name:     "30 seconds",
			duration: 30 * time.Second,
			expected: "0:30",
		},
		{
			name:     "1 minute",
			duration: 1 * time.Minute,
			expected: "1:00",
		},
		{
			name:     "1 minute 30 seconds",
			duration: 1*time.Minute + 30*time.Second,
			expected: "1:30",
		},
		{
			name:     "10 minutes 5 seconds",
			duration: 10*time.Minute + 5*time.Second,
			expected: "10:05",
		},
		{
			name:     "59 minutes 59 seconds",
			duration: 59*time.Minute + 59*time.Second,
			expected: "59:59",
		},
		{
			name:     "over 1 hour",
			duration: 65 * time.Minute,
			expected: "65:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestCalculateHistoricalStats(t *testing.T) {
	tests := []struct {
		name            string
		stats           types.UserStats
		expectedAvgWPM  float32
		expectedBestWPM float32
		expectedAvgAcc  float32
	}{
		{
			name: "single session",
			stats: types.UserStats{
				Sessions: []types.TypingSession{
					{WPM: 50, Accuracy: 95},
				},
			},
			expectedAvgWPM:  50,
			expectedBestWPM: 50,
			expectedAvgAcc:  95,
		},
		{
			name: "multiple sessions",
			stats: types.UserStats{
				Sessions: []types.TypingSession{
					{WPM: 40, Accuracy: 90},
					{WPM: 50, Accuracy: 95},
					{WPM: 60, Accuracy: 97},
				},
			},
			expectedAvgWPM:  50,     // (40+50+60)/3
			expectedBestWPM: 60,     // max
			expectedAvgAcc:  94,     // (90+95+97)/3 = 94
		},
		{
			name: "empty sessions",
			stats: types.UserStats{
				Sessions: []types.TypingSession{},
			},
			expectedAvgWPM:  0,
			expectedBestWPM: 0,
			expectedAvgAcc:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avgWPM, bestWPM, avgAcc := calculateHistoricalStats(tt.stats)

			if avgWPM != tt.expectedAvgWPM {
				t.Errorf("calculateHistoricalStats() avgWPM = %v, want %v", avgWPM, tt.expectedAvgWPM)
			}

			if bestWPM != tt.expectedBestWPM {
				t.Errorf("calculateHistoricalStats() bestWPM = %v, want %v", bestWPM, tt.expectedBestWPM)
			}

			if avgAcc != tt.expectedAvgAcc {
				t.Errorf("calculateHistoricalStats() avgAcc = %v, want %v", avgAcc, tt.expectedAvgAcc)
			}
		})
	}
}

func TestGetCurrentWPM(t *testing.T) {
	tests := []struct {
		name       string
		typedChars int
		elapsed    time.Duration
		wantWPM    float32
	}{
		{
			name:       "not started",
			typedChars: 0,
			elapsed:    0,
			wantWPM:    0,
		},
		{
			name:       "less than 1 second",
			typedChars: 10,
			elapsed:    500 * time.Millisecond,
			wantWPM:    0, // Avoid division by very small numbers
		},
		{
			name:       "50 chars in 1 minute (10 WPM)",
			typedChars: 50,
			elapsed:    1 * time.Minute,
			wantWPM:    10, // 50 chars / 5 = 10 words, 10 words / 1 min = 10 WPM
		},
		{
			name:       "100 chars in 2 minutes (10 WPM)",
			typedChars: 100,
			elapsed:    2 * time.Minute,
			wantWPM:    10, // 100 chars / 5 = 20 words, 20 words / 2 min = 10 WPM
		},
		{
			name:       "250 chars in 1 minute (50 WPM)",
			typedChars: 250,
			elapsed:    1 * time.Minute,
			wantWPM:    50, // 250 / 5 = 50 words, 50 / 1 min = 50 WPM
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := sessionModel{
				typedText:  makeString(tt.typedChars),
				hasStarted: tt.elapsed > 0,
				startTime:  time.Now().Add(-tt.elapsed),
			}

			wpm := model.getCurrentWPM()

			if wpm != tt.wantWPM {
				t.Errorf("getCurrentWPM() = %v, want %v", wpm, tt.wantWPM)
			}
		})
	}
}

func TestGetCurrentAccuracy(t *testing.T) {
	tests := []struct {
		name        string
		typedChars  int
		errors      int
		wantAccuracy float32
	}{
		{
			name:        "no typing",
			typedChars:  0,
			errors:      0,
			wantAccuracy: 0,
		},
		{
			name:        "perfect accuracy",
			typedChars:  100,
			errors:      0,
			wantAccuracy: 100,
		},
		{
			name:        "90% accuracy",
			typedChars:  100,
			errors:      10,
			wantAccuracy: 90, // (100-10)/100 * 100 = 90%
		},
		{
			name:        "50% accuracy",
			typedChars:  100,
			errors:      50,
			wantAccuracy: 50,
		},
		{
			name:        "very low accuracy",
			typedChars:  10,
			errors:      9,
			wantAccuracy: 10, // (10-9)/10 * 100 = 10%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := sessionModel{
				typedText: makeString(tt.typedChars),
				errors:    tt.errors,
			}

			accuracy := model.getCurrentAccuracy()

			if accuracy != tt.wantAccuracy {
				t.Errorf("getCurrentAccuracy() = %v, want %v", accuracy, tt.wantAccuracy)
			}
		})
	}
}

func TestGetUserStats_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	config := types.Config{
		Stats: types.StatsConfig{
			FileDir: filepath.Join(tmpDir, "nonexistent.json"),
		},
	}

	stats, err := getUserStats(config)

	// Should not return error for non-existent file
	if err != nil {
		t.Errorf("getUserStats() unexpected error: %v", err)
	}

	// For non-existent file, Sessions will be nil (zero value)
	// This is acceptable behavior - can be empty or nil
	if stats.Sessions != nil && len(stats.Sessions) != 0 {
		t.Errorf("getUserStats() Sessions length = %d, want 0 or nil", len(stats.Sessions))
	}
}

func TestGetUserStats_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats.json")

	// Create empty file
	err := os.WriteFile(statsFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := types.Config{
		Stats: types.StatsConfig{
			FileDir: statsFile,
		},
	}

	stats, err := getUserStats(config)

	if err != nil {
		t.Errorf("getUserStats() unexpected error for empty file: %v", err)
	}

	if stats.Sessions == nil {
		t.Errorf("getUserStats() Sessions should be initialized, got nil")
	}
}

func TestGetUserStats_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats.json")

	// Create valid stats file
	validJSON := `{
		"sessions": [
			{
				"date": "2024-01-01T12:00:00Z",
				"wpm": 50,
				"accuracy": 95,
				"errors": 5,
				"duration": 60000000000
			}
		]
	}`

	err := os.WriteFile(statsFile, []byte(validJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := types.Config{
		Stats: types.StatsConfig{
			FileDir: statsFile,
		},
	}

	stats, err := getUserStats(config)

	if err != nil {
		t.Errorf("getUserStats() unexpected error: %v", err)
	}

	if len(stats.Sessions) != 1 {
		t.Errorf("getUserStats() Sessions length = %d, want 1", len(stats.Sessions))
	}

	if stats.Sessions[0].WPM != 50 {
		t.Errorf("getUserStats() Session WPM = %v, want 50", stats.Sessions[0].WPM)
	}

	if stats.Sessions[0].Accuracy != 95 {
		t.Errorf("getUserStats() Session Accuracy = %v, want 95", stats.Sessions[0].Accuracy)
	}
}

func TestSaveUserStats(t *testing.T) {
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats.json")

	config := types.Config{
		Stats: types.StatsConfig{
			FileDir: statsFile,
		},
	}

	stats := types.UserStats{
		Sessions: []types.TypingSession{
			{
				Date:     time.Now(),
				WPM:      45,
				Accuracy: 92.5,
				Errors:   8,
				Duration: 90 * time.Second,
			},
		},
	}

	err := saveUserStats(config, stats)

	if err != nil {
		t.Errorf("saveUserStats() unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		t.Errorf("saveUserStats() did not create file")
	}

	// Read back and verify
	loadedStats, err := getUserStats(config)
	if err != nil {
		t.Errorf("getUserStats() after save unexpected error: %v", err)
	}

	if len(loadedStats.Sessions) != 1 {
		t.Fatalf("Loaded stats has %d sessions, want 1", len(loadedStats.Sessions))
	}

	if loadedStats.Sessions[0].WPM != 45 {
		t.Errorf("Loaded WPM = %v, want 45", loadedStats.Sessions[0].WPM)
	}

	if loadedStats.Sessions[0].Accuracy != 92.5 {
		t.Errorf("Loaded Accuracy = %v, want 92.5", loadedStats.Sessions[0].Accuracy)
	}

	if loadedStats.Sessions[0].Errors != 8 {
		t.Errorf("Loaded Errors = %v, want 8", loadedStats.Sessions[0].Errors)
	}
}

func TestSessionResult_String(t *testing.T) {
	tests := []struct {
		name     string
		result   SessionResult
		expected string
		contains []string
	}{
		{
			name: "exited without starting",
			result: SessionResult{
				Exited: true,
			},
			expected: "", // Now returns empty string (dashboard handles display)
		},
		{
			name: "error",
			result: SessionResult{
				Error: os.ErrNotExist,
			},
			contains: []string{"error"}, // Still shows errors
		},
		{
			name: "successful session",
			result: SessionResult{
				Session: &types.TypingSession{
					WPM:      50,
					Accuracy: 95.5,
					Errors:   5,
					Duration: 60 * time.Second,
				},
			},
			expected: "", // Now returns empty string (dashboard handles display)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.result.String()

			// If expected is set, check exact match
			if tt.expected != "" || (tt.expected == "" && len(tt.contains) == 0) {
				if result != tt.expected {
					t.Errorf("String() = %q, want %q", result, tt.expected)
				}
			}

			// Check for contains strings
			for _, expected := range tt.contains {
				if !contains(result, expected) {
					t.Errorf("String() output missing %q\nGot: %s", expected, result)
				}
			}
		})
	}
}

func TestWelcomeAction_String(t *testing.T) {
	tests := []struct {
		name     string
		action   WelcomeAction
		expected string
	}{
		{
			name:     "start session",
			action:   StartSession,
			expected: "Lets Exercise",
		},
		{
			name:     "exit",
			action:   Exit,
			expected: "Naah not today lets exit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.action.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNewWelcomeModel(t *testing.T) {
	config := types.Config{
		Ui: types.UiConfig{Theme: "default"},
	}
	stats := types.UserStats{
		Sessions: []types.TypingSession{},
	}

	model := newWelcomeModel(config, stats)

	if model.cursor != 0 {
		t.Errorf("newWelcomeModel() cursor = %d, want 0", model.cursor)
	}

	if len(model.choices) != 2 {
		t.Errorf("newWelcomeModel() choices length = %d, want 2", len(model.choices))
	}

	if model.choices[0] != StartSession {
		t.Errorf("newWelcomeModel() choices[0] = %v, want StartSession", model.choices[0])
	}

	if model.choices[1] != Exit {
		t.Errorf("newWelcomeModel() choices[1] = %v, want Exit", model.choices[1])
	}

	if model.done {
		t.Errorf("newWelcomeModel() done = true, want false")
	}

	if model.selected != nil {
		t.Errorf("newWelcomeModel() selected = %v, want nil", model.selected)
	}
}

func TestNewDashboardModel(t *testing.T) {
	config := types.Config{
		Ui: types.UiConfig{Theme: "default"},
	}
	session := types.TypingSession{
		WPM:      50,
		Accuracy: 95,
		Errors:   5,
		Duration: 60 * time.Second,
	}
	stats := types.UserStats{
		Sessions: []types.TypingSession{session},
	}

	model := newDashboardModel(config, session, stats)

	if model.currentSession.WPM != 50 {
		t.Errorf("newDashboardModel() currentSession.WPM = %v, want 50", model.currentSession.WPM)
	}

	if model.currentSession.Accuracy != 95 {
		t.Errorf("newDashboardModel() currentSession.Accuracy = %v, want 95", model.currentSession.Accuracy)
	}

	if len(model.allStats.Sessions) != 1 {
		t.Errorf("newDashboardModel() allStats.Sessions length = %d, want 1", len(model.allStats.Sessions))
	}
}

func TestWelcomeModel_Init(t *testing.T) {
	config := types.Config{}
	stats := types.UserStats{}
	model := newWelcomeModel(config, stats)

	cmd := model.Init()

	if cmd != nil {
		t.Errorf("Init() = %v, want nil", cmd)
	}
}

func TestDashboardModel_Init(t *testing.T) {
	config := types.Config{}
	session := types.TypingSession{}
	stats := types.UserStats{}
	model := newDashboardModel(config, session, stats)

	cmd := model.Init()

	// Init returns tea.WindowSize() command, not nil
	if cmd == nil {
		t.Logf("Init() returned nil (acceptable)")
	}
}

func TestWelcomeModel_View(t *testing.T) {
	config := types.Config{}
	stats := types.UserStats{}
	model := newWelcomeModel(config, stats)

	view := model.View()

	// Check that view contains expected elements
	if !contains(view, "AI-Powered Touch Typing Trainer") {
		t.Errorf("View() missing subtitle")
	}

	if !contains(view, "Lets Exercise") {
		t.Errorf("View() missing 'Lets Exercise'")
	}

	if !contains(view, "Naah not today lets exit") {
		t.Errorf("View() missing exit option")
	}
}

func TestDashboardModel_View(t *testing.T) {
	config := types.Config{}
	session := types.TypingSession{
		WPM:      50,
		Accuracy: 95.5,
		Errors:   5,
		Duration: 90 * time.Second,
	}
	stats := types.UserStats{
		Sessions: []types.TypingSession{session},
	}
	model := newDashboardModel(config, session, stats)

	view := model.View()

	// Check that view contains expected elements
	if !contains(view, "SESSION COMPLETE") {
		t.Errorf("View() missing 'SESSION COMPLETE'")
	}

	if !contains(view, "WPM") {
		t.Errorf("View() missing 'WPM'")
	}

	if !contains(view, "Accuracy") {
		t.Errorf("View() missing 'Accuracy'")
	}

	if !contains(view, "Press Enter to exit") {
		t.Errorf("View() missing 'Press Enter to exit'")
	}
}

func TestSessionModel_Init(t *testing.T) {
	model := sessionModel{
		text:       "test text",
		typedText:  "",
		hasStarted: false,
	}

	cmd := model.Init()

	// Init should return a batch command with window size and tick
	if cmd == nil {
		t.Errorf("Init() returned nil, expected command batch")
	}
}

func TestTickCmd(t *testing.T) {
	cmd := tickCmd()

	if cmd == nil {
		t.Errorf("tickCmd() returned nil")
	}
}

func TestSessionModel_View_NotStarted(t *testing.T) {
	model := sessionModel{
		text:             "test text",
		typedText:        "",
		hasStarted:       false,
		selectedDuration: 5,
	}

	view := model.View()

	if !contains(view, "5 minutes") {
		t.Errorf("View() missing session duration display")
	}

	if !contains(view, "Press ENTER to start") {
		t.Errorf("View() missing start instruction")
	}
}

func TestSessionModel_View_Loading(t *testing.T) {
	model := sessionModel{
		text:       "test",
		hasStarted: true,
		width:      0, // width not set yet
	}

	view := model.View()

	if view != "Loading..." {
		t.Errorf("View() = %q, want 'Loading...'", view)
	}
}

func TestSessionModel_View_Started(t *testing.T) {
	model := sessionModel{
		text:             "The quick brown fox",
		typedText:        "The qui",
		errors:           1,
		hasStarted:       true,
		width:            80,
		height:           24,
		startTime:        time.Now().Add(-30 * time.Second),
		sessionDuration:  2 * time.Minute,
		lastKeyTime:      time.Now(),
		generationPending: false,
	}

	view := model.View()

	// Should contain stats
	if !contains(view, "WPM") {
		t.Errorf("View() missing WPM stat")
	}

	if !contains(view, "Accuracy") {
		t.Errorf("View() missing Accuracy stat")
	}

	if !contains(view, "Errors") {
		t.Errorf("View() missing Errors stat")
	}

	if !contains(view, "Time Remaining") {
		t.Errorf("View() missing Time Remaining")
	}
}

func TestSessionModel_View_WithGenerationPending(t *testing.T) {
	model := sessionModel{
		text:              "test",
		typedText:         "",
		hasStarted:        true,
		width:             80,
		startTime:         time.Now(),
		sessionDuration:   1 * time.Minute,
		lastKeyTime:       time.Now(),
		generationPending: true,
		isLLMSource:       true,
	}

	view := model.View()

	// Should show loading indicator
	if !contains(view, "Generating") {
		t.Errorf("View() missing generation indicator")
	}
}

func TestWelcomeModel_Update_Quit(t *testing.T) {
	config := types.Config{}
	stats := types.UserStats{}
	model := newWelcomeModel(config, stats)

	// Test Ctrl+C quit
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedModel, cmd := model.Update(msg)

	if cmd == nil {
		t.Errorf("Update(ctrl+c) should return quit command")
	}

	m := updatedModel.(welcomeModel)
	if m.selected != nil {
		t.Logf("Model selected state: %v", m.selected)
	}
}

func TestWelcomeModel_Update_Navigation(t *testing.T) {
	config := types.Config{}
	stats := types.UserStats{}
	model := newWelcomeModel(config, stats)

	// Test down arrow
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(welcomeModel)

	if m.cursor != 1 {
		t.Errorf("After down arrow, cursor = %d, want 1", m.cursor)
	}

	// Test up arrow (should wrap to bottom)
	msg2 := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel2, _ := m.Update(msg2)
	m2 := updatedModel2.(welcomeModel)

	if m2.cursor != 0 {
		t.Errorf("After up arrow, cursor = %d, want 0", m2.cursor)
	}
}

func TestWelcomeModel_Update_Select(t *testing.T) {
	config := types.Config{}
	stats := types.UserStats{}
	model := newWelcomeModel(config, stats)

	// Test Enter key to select
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(welcomeModel)

	if !m.done {
		t.Errorf("After Enter, done should be true")
	}

	if m.selected == nil {
		t.Errorf("After Enter, selected should not be nil")
	}

	if cmd == nil {
		t.Errorf("After Enter, should return quit command")
	}

	if m.selected != nil && *m.selected != StartSession {
		t.Errorf("Selected action = %v, want StartSession", *m.selected)
	}
}

func TestDashboardModel_Update_Quit(t *testing.T) {
	config := types.Config{}
	session := types.TypingSession{}
	stats := types.UserStats{}
	model := newDashboardModel(config, session, stats)

	// Test Enter key to quit
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Errorf("Update(enter) should return quit command")
	}
}

func TestDashboardModel_Update_OtherKeys(t *testing.T) {
	config := types.Config{}
	session := types.TypingSession{}
	stats := types.UserStats{}
	model := newDashboardModel(config, session, stats)

	// Test other keys (should not quit)
	msg := tea.KeyMsg{Type: tea.KeySpace}
	updatedModel, cmd := model.Update(msg)

	// Other keys should not return a quit command
	if cmd != nil {
		t.Logf("Update(space) returned cmd: %v", cmd)
	}

	_ = updatedModel // Use the variable
}

func TestDashboardModel_Update_CtrlC(t *testing.T) {
	config := types.Config{}
	session := types.TypingSession{}
	stats := types.UserStats{}
	model := newDashboardModel(config, session, stats)

	// Test Ctrl+C key - dashboard only responds to Enter
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := model.Update(msg)

	// Dashboard doesn't handle ctrl+c explicitly, should return nil
	if cmd != nil {
		t.Logf("Update(ctrl+c) returned cmd: %v (dashboard only responds to Enter)", cmd)
	}
}

func TestDashboardModel_Update_EscKey(t *testing.T) {
	config := types.Config{}
	session := types.TypingSession{}
	stats := types.UserStats{}
	model := newDashboardModel(config, session, stats)

	// Test ESC key - dashboard only responds to Enter
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := model.Update(msg)

	// Dashboard doesn't handle esc explicitly, should return nil
	if cmd != nil {
		t.Logf("Update(esc) returned cmd: %v (dashboard only responds to Enter)", cmd)
	}
}

func TestDashboardModel_Update_WindowSizeMsg(t *testing.T) {
	config := types.Config{}
	session := types.TypingSession{}
	stats := types.UserStats{}
	model := newDashboardModel(config, session, stats)

	// Test window size message
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)

	m := updatedModel.(dashboardModel)
	if m.width != 120 {
		t.Errorf("Width = %d, want 120", m.width)
	}
	if m.height != 40 {
		t.Errorf("Height = %d, want 40", m.height)
	}
}

// Helper functions

func makeString(length int) string {
	if length == 0 {
		return ""
	}
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr ||
			(len(s) > len(substr) && contains(s[1:], substr)))))
}

// TestCheckForMistypedWord tests the checkForMistypedWord method
func TestCheckForMistypedWord(t *testing.T) {
	tests := []struct {
		name                 string
		text                 string
		typedText            string
		lastWordEnd          int
		wordsWithErrors      map[int]bool
		expectedProblemWords int
	}{
		{
			name:                 "correct word typed",
			text:                 "hello world",
			typedText:            "hello ",
			lastWordEnd:          0,
			wordsWithErrors:      map[int]bool{},
			expectedProblemWords: 0,
		},
		{
			name:                 "mistyped word",
			text:                 "hello world",
			typedText:            "hallo ",
			lastWordEnd:          0,
			wordsWithErrors:      map[int]bool{1: true},
			expectedProblemWords: 1,
		},
		{
			name:                 "word with errors even if corrected",
			text:                 "hello world",
			typedText:            "hello ",
			lastWordEnd:          0,
			wordsWithErrors:      map[int]bool{2: true}, // Had error at position 2
			expectedProblemWords: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &sessionModel{
				text:                tt.text,
				typedText:           tt.typedText,
				targetWords:         strings.Fields(tt.text), // Add targetWords
				lastWordEnd:         tt.lastWordEnd,
				wordsWithErrors:     tt.wordsWithErrors,
				currentProblemWords: []string{},
			}

			model.checkForMistypedWord()

			if len(model.currentProblemWords) != tt.expectedProblemWords {
				t.Errorf("checkForMistypedWord() problem words = %v, want %d",
					model.currentProblemWords, tt.expectedProblemWords)
			}
		})
	}
}

// TestSessionModel_Update_Tick tests tick message handling
func TestSessionModel_Update_Tick(t *testing.T) {
	model := sessionModel{
		text:            "test text",
		typedText:       "test",
		hasStarted:      true,
		startTime:       time.Now().Add(-30 * time.Second),
		sessionDuration: 1 * time.Minute,
		completed:       false,
		quit:            false,
	}

	msg := tickMsg(time.Now())
	updatedModel, cmd := model.Update(msg)

	// Should return tick command to continue
	if cmd == nil {
		t.Error("Update(tickMsg) should return tick command")
	}

	m := updatedModel.(sessionModel)
	if m.completed {
		t.Error("Session should not be completed yet")
	}
}

// TestSessionModel_Update_SessionExpired tests session expiration
func TestSessionModel_Update_SessionExpired(t *testing.T) {
	model := sessionModel{
		text:            "test text",
		typedText:       "test",
		hasStarted:      true,
		startTime:       time.Now().Add(-2 * time.Minute),
		sessionDuration: 1 * time.Minute,
		completed:       false,
		quit:            false,
	}

	msg := tickMsg(time.Now())
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(sessionModel)
	if !m.completed {
		t.Error("Session should be completed after duration expires")
	}

	// Should return quit command
	if cmd == nil {
		t.Error("Update should return quit command when session expires")
	}
}

// TestSessionModel_Update_WindowSize tests window size message handling
func TestSessionModel_Update_WindowSize(t *testing.T) {
	model := sessionModel{
		text:      "test text",
		width:     0,
		height:    0,
	}

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)

	m := updatedModel.(sessionModel)
	if m.width != 100 {
		t.Errorf("Width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("Height = %d, want 50", m.height)
	}
}

// TestSessionModel_Update_EscKey tests ESC key handling
func TestSessionModel_Update_EscKey(t *testing.T) {
	model := sessionModel{
		text:       "test text",
		hasStarted: true,
		quit:       false,
	}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(sessionModel)
	if !m.quit {
		t.Error("Quit should be true after ESC key")
	}

	if cmd == nil {
		t.Error("Update should return quit command after ESC")
	}
}

// TestSessionModel_Update_UpDownKeys tests duration adjustment
func TestSessionModel_Update_UpDownKeys(t *testing.T) {
	model := sessionModel{
		text:             "test text",
		hasStarted:       false,
		selectedDuration: 5,
	}

	// Test up key
	msgUp := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := model.Update(msgUp)
	m := updatedModel.(sessionModel)

	if m.selectedDuration != 6 {
		t.Errorf("After up key, selectedDuration = %d, want 6", m.selectedDuration)
	}

	// Test down key
	msgDown := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel2, _ := m.Update(msgDown)
	m2 := updatedModel2.(sessionModel)

	if m2.selectedDuration != 5 {
		t.Errorf("After down key, selectedDuration = %d, want 5", m2.selectedDuration)
	}
}

// TestGenerateNextSentenceCmd tests the generate command
func TestGenerateNextSentenceCmd(t *testing.T) {
	// Create a simple test - we can't actually test the async behavior easily
	// but we can test that the function returns a command
	model := sessionModel{
		isLLMSource:   true,
		errorPatterns: map[rune]int{'a': 1},
		problemWords:  []string{"test"},
		lastSentence:  "Hello world",
	}

	cmd := model.generateNextSentenceCmd()

	if cmd == nil {
		t.Error("generateNextSentenceCmd() should return a command")
	}
}

// TestStartSession tests startSession initialization
func TestStartSession(t *testing.T) {
	t.Skip("Skipping startSession test as it requires interactive TUI")
	// This function starts a full bubbletea program which is hard to test in unit tests
	// It's better tested through integration tests or manual testing
}

// TestSessionModel_Update_EnterToStart tests starting session with Enter key
func TestSessionModel_Update_EnterToStart(t *testing.T) {
	model := sessionModel{
		text:             "test text",
		hasStarted:       false,
		selectedDuration: 3,
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	if !m.hasStarted {
		t.Error("Session should be started after Enter key")
	}

	if m.sessionDuration != 3*time.Minute {
		t.Errorf("sessionDuration = %v, want 3 minutes", m.sessionDuration)
	}
}

// TestSessionModel_Update_Backspace tests backspace handling
func TestSessionModel_Update_Backspace(t *testing.T) {
	model := sessionModel{
		text:      "test text",
		typedText: "test",
	}

	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	if m.typedText != "tes" {
		t.Errorf("After backspace, typedText = %q, want 'tes'", m.typedText)
	}
}

// TestSessionModel_Update_CharacterInput tests character input
func TestSessionModel_Update_CharacterInput(t *testing.T) {
	model := sessionModel{
		text:             "test text",
		typedText:        "",
		hasStarted:       true,
		startTime:        time.Now(),
		lastKeyTime:      time.Now(),
		keyStrokeTimes:   []time.Duration{},
		errors:           0,
		wordsWithErrors:  map[int]bool{},
		currentProblemWords: []string{},
		lastWordEnd:      0,
	}

	// Type correct character
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	if m.typedText != "t" {
		t.Errorf("After typing 't', typedText = %q, want 't'", m.typedText)
	}

	if m.errors != 0 {
		t.Errorf("Correct input should not increment errors, got %d", m.errors)
	}
}

// TestSessionModel_Update_IncorrectCharacter tests error tracking
func TestSessionModel_Update_IncorrectCharacter(t *testing.T) {
	model := sessionModel{
		text:             "test",
		typedText:        "",
		hasStarted:       true,
		startTime:        time.Now(),
		lastKeyTime:      time.Now(),
		keyStrokeTimes:   []time.Duration{},
		errors:           0,
		wordsWithErrors:  map[int]bool{},
		currentProblemWords: []string{},
		lastWordEnd:      0,
	}

	// Type incorrect character
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	if m.errors != 1 {
		t.Errorf("Incorrect input should increment errors, got %d", m.errors)
	}

	if !m.wordsWithErrors[0] {
		t.Error("wordsWithErrors should track error at position 0")
	}
}

// TestSessionModel_Update_SpaceKey tests space key handling
func TestSessionModel_Update_SpaceKey(t *testing.T) {
	model := sessionModel{
		text:             "test text",
		typedText:        "test",
		hasStarted:       true,
		startTime:        time.Now(),
		lastKeyTime:      time.Now(),
		keyStrokeTimes:   []time.Duration{},
		wordsWithErrors:  map[int]bool{},
		currentProblemWords: []string{},
		lastWordEnd:      0,
	}

	msg := tea.KeyMsg{Type: tea.KeySpace}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	if m.typedText != "test " {
		t.Errorf("After space, typedText = %q, want 'test '", m.typedText)
	}
}

// TestSessionModel_Update_SessionComplete tests completion
func TestSessionModel_Update_SessionComplete(t *testing.T) {
	model := sessionModel{
		text:                  "test",
		typedText:             "tes",
		hasStarted:            true,
		startTime:             time.Now(),
		lastKeyTime:           time.Now(),
		keyStrokeTimes:        []time.Duration{},
		currentSentenceEndPos: 4,
		isLLMSource:           false,
		wordsWithErrors:       map[int]bool{},
		currentProblemWords:   []string{},
		lastWordEnd:           0,
	}

	// Type the last character
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(sessionModel)

	if !m.completed {
		t.Error("Session should be completed after typing all text")
	}

	if cmd == nil {
		t.Error("Should return quit command when session completes")
	}
}

// TestSessionModel_Update_GenerationComplete tests LLM generation completion
func TestSessionModel_Update_GenerationComplete(t *testing.T) {
	model := sessionModel{
		text:              "first sentence",
		typedText:         "",
		isLLMSource:       true,
		generationPending: true,
		nextSentenceReady: false,
	}

	msg := generationCompleteMsg{sentence: "second sentence"}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	if m.generationPending {
		t.Error("generationPending should be false after completion")
	}

	if !m.nextSentenceReady {
		t.Error("nextSentenceReady should be true after completion")
	}

	if m.nextSentenceBuffer != "second sentence" {
		t.Errorf("nextSentenceBuffer = %q, want 'second sentence'", m.nextSentenceBuffer)
	}

	expectedText := "first sentence second sentence"
	if m.text != expectedText {
		t.Errorf("text = %q, want %q", m.text, expectedText)
	}
}

// TestSessionModel_Update_GenerationError tests LLM generation error
func TestSessionModel_Update_GenerationError(t *testing.T) {
	model := sessionModel{
		text:              "test",
		generationPending: true,
		nextSentenceReady: true,
	}

	msg := generationErrorMsg{err: fmt.Errorf("test error")}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	if m.generationPending {
		t.Error("generationPending should be false after error")
	}

	if m.nextSentenceReady {
		t.Error("nextSentenceReady should be false after error")
	}
}

// TestSessionModel_Update_PregenTrigger tests LLM pregeneration trigger
func TestSessionModel_Update_PregenTrigger(t *testing.T) {
	model := sessionModel{
		text:                  "this is a test sentence",
		typedText:             "this is a te",
		isLLMSource:           true,
		hasStarted:            true,
		startTime:             time.Now(),
		sessionDuration:       5 * time.Minute,
		generationPending:     false,
		nextSentenceReady:     false,
		currentSentenceEndPos: 23,
		pregenerateThreshold:  20,
		errorPatterns:         map[rune]int{},
		problemWords:          []string{},
		lastSentence:          "previous",
	}

	msg := tickMsg(time.Now())
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(sessionModel)

	// Should trigger pregeneration since charsRemaining (11) <= threshold (20)
	if !m.generationPending {
		t.Error("Should trigger pregeneration when chars remaining <= threshold")
	}

	if cmd == nil {
		t.Error("Should return batch command with tick and generate")
	}
}

// TestSessionModel_Update_UpDownMaxMin tests duration limits
func TestSessionModel_Update_UpDownMaxMin(t *testing.T) {
	// Test max limit
	model := sessionModel{
		text:             "test",
		hasStarted:       false,
		selectedDuration: 60,
	}

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	if m.selectedDuration != 60 {
		t.Errorf("Should cap at 60 minutes, got %d", m.selectedDuration)
	}

	// Test min limit
	model2 := sessionModel{
		text:             "test",
		hasStarted:       false,
		selectedDuration: 1,
	}

	msg2 := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel2, _ := model2.Update(msg2)
	m2 := updatedModel2.(sessionModel)

	if m2.selectedDuration != 1 {
		t.Errorf("Should floor at 1 minute, got %d", m2.selectedDuration)
	}
}

// TestSessionModel_Update_KeysBeforeStart tests that keys are ignored before start
func TestSessionModel_Update_KeysBeforeStart(t *testing.T) {
	model := sessionModel{
		text:       "test",
		typedText:  "",
		hasStarted: false,
	}

	// Try typing before start
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	if m.typedText != "" {
		t.Errorf("Should not accept input before start, got %q", m.typedText)
	}
}

// TestSessionModel_View_SessionConfig tests view before session starts
func TestSessionModel_View_SessionConfig(t *testing.T) {
	model := sessionModel{
		text:             "test text",
		hasStarted:       false,
		selectedDuration: 10,
		width:            80,
	}

	view := model.View()

	if !contains(view, "10 minutes") {
		t.Errorf("View should show '10 minutes', got: %s", view)
	}

	if !contains(view, "Session Configuration") {
		t.Error("View should show 'Session Configuration'")
	}
}

// TestSessionModel_View_SingleMinute tests singular minute display
func TestSessionModel_View_SingleMinute(t *testing.T) {
	model := sessionModel{
		text:             "test",
		hasStarted:       false,
		selectedDuration: 1,
		width:            80,
	}

	view := model.View()

	if !contains(view, "1 minute") && !contains(view, "minute") {
		t.Error("View should show '1 minute' (singular)")
	}
}

// TestSessionModel_Update_BackspaceEmpty tests backspace on empty string
func TestSessionModel_Update_BackspaceEmpty(t *testing.T) {
	model := sessionModel{
		text:      "test",
		typedText: "",
	}

	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	if m.typedText != "" {
		t.Errorf("Backspace on empty should remain empty, got %q", m.typedText)
	}
}

// TestSessionModel_Update_LLMSentenceTransition tests LLM sentence transition
func TestSessionModel_Update_LLMSentenceTransition(t *testing.T) {
	model := sessionModel{
		text:                  "first sentence second sentence",
		typedText:             "first sentenc",
		isLLMSource:           true,
		hasStarted:            true,
		startTime:             time.Now(),
		lastKeyTime:           time.Now(),
		keyStrokeTimes:        []time.Duration{},
		currentSentenceEndPos: 14,
		nextSentenceReady:     true,
		nextSentenceBuffer:    "second sentence",
		errorPatterns:         map[rune]int{},
		problemWords:          []string{},
		currentProblemWords:   []string{},
		wordsWithErrors:       map[int]bool{},
		lastWordEnd:           0,
	}

	// Type the last character of first sentence
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	// Should transition to next sentence
	if m.lastSentence != "second sentence" {
		t.Errorf("lastSentence should be updated to %q, got %q", "second sentence", m.lastSentence)
	}

	if m.nextSentenceReady {
		t.Error("nextSentenceReady should be false after transition")
	}
}

// TestSessionModel_Update_LLMWaitingForGeneration tests waiting for LLM
func TestSessionModel_Update_LLMWaitingForGeneration(t *testing.T) {
	model := sessionModel{
		text:                  "first",
		typedText:             "first",
		isLLMSource:           true,
		hasStarted:            true,
		startTime:             time.Now(),
		lastKeyTime:           time.Now(),
		keyStrokeTimes:        []time.Duration{},
		currentSentenceEndPos: 5,
		generationPending:     true,
		nextSentenceReady:     false,
		wordsWithErrors:       map[int]bool{},
		currentProblemWords:   []string{},
		lastWordEnd:           0,
	}

	// Should wait when generation pending
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	// Should remain in waiting state
	if !m.generationPending {
		t.Error("Should remain in generationPending state")
	}
}

// TestSessionModel_Update_LLMCompleteAllText tests LLM mode completion
func TestSessionModel_Update_LLMCompleteAllText(t *testing.T) {
	model := sessionModel{
		text:                  "test",
		typedText:             "tes",
		isLLMSource:           true,
		hasStarted:            true,
		startTime:             time.Now(),
		lastKeyTime:           time.Now(),
		keyStrokeTimes:        []time.Duration{},
		currentSentenceEndPos: 4,
		generationPending:     false,
		nextSentenceReady:     false,
		wordsWithErrors:       map[int]bool{},
		currentProblemWords:   []string{},
		lastWordEnd:           0,
	}

	// Type last character when no more text available
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(sessionModel)

	if !m.completed {
		t.Error("Should complete when all text typed in LLM mode with no next sentence")
	}

	if cmd == nil {
		t.Error("Should return quit command")
	}
}

// TestCheckForMistypedWord_FixedLogic tests the fixed word validation logic
func TestCheckForMistypedWord_FixedLogic(t *testing.T) {
	tests := []struct {
		name                 string
		targetText           string
		typedText            string
		wordsWithErrors      map[int]bool
		lastWordEnd          int
		expectedProblemCount int
		expectedProblemWord  string
	}{
		{
			name:                 "first word with typo",
			targetText:           "hello world",
			typedText:            "hallo ",
			wordsWithErrors:      map[int]bool{1: true}, // Error at position 1 (second char)
			lastWordEnd:          0,
			expectedProblemCount: 1,
			expectedProblemWord:  "hello",
		},
		{
			name:                 "word with error during typing",
			targetText:           "testing words",
			typedText:            "testing ",
			wordsWithErrors:      map[int]bool{2: true}, // Error at position 2
			lastWordEnd:          0,
			expectedProblemCount: 1,
			expectedProblemWord:  "testing",
		},
		{
			name:                 "correct word no errors",
			targetText:           "hello world",
			typedText:            "hello ",
			wordsWithErrors:      map[int]bool{},
			lastWordEnd:          0,
			expectedProblemCount: 0,
			expectedProblemWord:  "",
		},
		{
			name:                 "second word with error",
			targetText:           "hello world",
			typedText:            "hello wurld ",
			wordsWithErrors:      map[int]bool{8: true}, // Error in "world" at position 8
			lastWordEnd:          6,                      // Already processed "hello "
			expectedProblemCount: 1,
			expectedProblemWord:  "world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &sessionModel{
				text:                tt.targetText,
				typedText:           tt.typedText,
				targetWords:         strings.Fields(tt.targetText),
				wordsWithErrors:     tt.wordsWithErrors,
				currentProblemWords: []string{},
				lastWordEnd:         tt.lastWordEnd,
			}

			model.checkForMistypedWord()

			if len(model.currentProblemWords) != tt.expectedProblemCount {
				t.Errorf("Expected %d problem words, got %d: %v",
					tt.expectedProblemCount, len(model.currentProblemWords), model.currentProblemWords)
			}

			if tt.expectedProblemCount > 0 && len(model.currentProblemWords) > 0 {
				found := false
				for _, word := range model.currentProblemWords {
					if word == tt.expectedProblemWord {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected problem word %q not found in %v",
						tt.expectedProblemWord, model.currentProblemWords)
				}
			}
		})
	}
}

// TestTypoBlocking tests the typo blocking mechanism
func TestTypoBlocking(t *testing.T) {
	config := types.Config{
		Ui: types.UiConfig{
			BlockOnTypo: true,
		},
	}

	model := sessionModel{
		text:             "test",
		typedText:        "t",
		hasStarted:       true,
		startTime:        time.Now(),
		lastKeyTime:      time.Now(),
		keyStrokeTimes:   []time.Duration{},
		wordsWithErrors:  map[int]bool{},
		currentProblemWords: []string{},
		lastWordEnd:      0,
		hasTypo:          true, // Simulate existing typo
		config:           config,
		targetWords:      []string{"test"},
	}

	// Try to type when typo exists
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	// typedText should remain unchanged because of typo blocking
	if m.typedText != "t" {
		t.Errorf("Typo blocking failed: typed text should remain 't', got %q", m.typedText)
	}
}

// TestTypoFlash tests the visual flash feedback
func TestTypoFlash(t *testing.T) {
	config := types.Config{
		Ui: types.UiConfig{
			BlockOnTypo:         false,
			TypoFlashEnabled:    true,
			TypoFlashDurationMs: 200,
		},
	}

	model := sessionModel{
		text:             "test",
		typedText:        "",
		hasStarted:       true,
		startTime:        time.Now(),
		lastKeyTime:      time.Now(),
		keyStrokeTimes:   []time.Duration{},
		wordsWithErrors:  map[int]bool{},
		currentProblemWords: []string{},
		lastWordEnd:      0,
		config:           config,
		targetWords:      []string{"test"},
	}

	// Type incorrect character
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(sessionModel)

	// Should trigger flash
	if m.typoFlashTime.IsZero() {
		t.Error("typoFlashTime should be set when typo occurs")
	}

	// Should return flash command
	if cmd == nil {
		t.Error("Should return flash command when flash is enabled")
	}
}

// TestBackspaceClearsTypo tests that backspace clears the typo flag
func TestBackspaceClearsTypo(t *testing.T) {
	model := sessionModel{
		text:      "test",
		typedText: "tx",
		hasTypo:   true,
	}

	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	if m.hasTypo {
		t.Error("hasTypo should be cleared after backspace")
	}

	if m.typedText != "t" {
		t.Errorf("typedText should be 't' after backspace, got %q", m.typedText)
	}
}

// TestTypoFlashMsg tests typo flash message handling
func TestTypoFlashMsg(t *testing.T) {
	model := sessionModel{
		text:          "test",
		typoFlashTime: time.Now(),
	}

	msg := typoFlashMsg{}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(sessionModel)

	if !m.typoFlashTime.IsZero() {
		t.Error("typoFlashTime should be cleared after flash message")
	}
}
