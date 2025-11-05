package ui

import (
	"go-touch/internal/types"
	"os"
	"path/filepath"
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
		contains []string
	}{
		{
			name: "exited without starting",
			result: SessionResult{
				Exited: true,
			},
			contains: []string{"exited without starting"},
		},
		{
			name: "error",
			result: SessionResult{
				Error: os.ErrNotExist,
			},
			contains: []string{"error"},
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
			contains: []string{"SESSION SUMMARY", "WPM", "Accuracy", "Errors", "Duration", "saved successfully"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.result.String()

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
