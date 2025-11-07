package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (cleanup func())
		wantErr     bool
		wantContain string
	}{
		{
			name: "linux with XDG_CONFIG_HOME",
			setup: func() func() {
				if runtime.GOOS != "linux" {
					t.Skip("Skipping linux-specific test")
				}
				oldXDG := os.Getenv("XDG_CONFIG_HOME")
				os.Setenv("XDG_CONFIG_HOME", "/tmp/test-config")
				return func() {
					if oldXDG != "" {
						os.Setenv("XDG_CONFIG_HOME", oldXDG)
					} else {
						os.Unsetenv("XDG_CONFIG_HOME")
					}
				}
			},
			wantErr:     false,
			wantContain: "gotouch",
		},
		{
			name: "linux without XDG_CONFIG_HOME",
			setup: func() func() {
				if runtime.GOOS != "linux" {
					t.Skip("Skipping linux-specific test")
				}
				oldXDG := os.Getenv("XDG_CONFIG_HOME")
				os.Unsetenv("XDG_CONFIG_HOME")
				return func() {
					if oldXDG != "" {
						os.Setenv("XDG_CONFIG_HOME", oldXDG)
					}
				}
			},
			wantErr:     false,
			wantContain: ".config/gotouch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			got, err := GetConfigDir()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Errorf("GetConfigDir() returned empty string")
			}
			if tt.wantContain != "" && got != "" {
				if !contains(got, tt.wantContain) {
					t.Errorf("GetConfigDir() = %v, want to contain %v", got, tt.wantContain)
				}
			}
		})
	}
}

func TestGetDataDir(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (cleanup func())
		wantErr     bool
		wantContain string
	}{
		{
			name: "linux with XDG_DATA_HOME",
			setup: func() func() {
				if runtime.GOOS != "linux" {
					t.Skip("Skipping linux-specific test")
				}
				oldXDG := os.Getenv("XDG_DATA_HOME")
				os.Setenv("XDG_DATA_HOME", "/tmp/test-data")
				return func() {
					if oldXDG != "" {
						os.Setenv("XDG_DATA_HOME", oldXDG)
					} else {
						os.Unsetenv("XDG_DATA_HOME")
					}
				}
			},
			wantErr:     false,
			wantContain: "gotouch",
		},
		{
			name: "linux without XDG_DATA_HOME",
			setup: func() func() {
				if runtime.GOOS != "linux" {
					t.Skip("Skipping linux-specific test")
				}
				oldXDG := os.Getenv("XDG_DATA_HOME")
				os.Unsetenv("XDG_DATA_HOME")
				return func() {
					if oldXDG != "" {
						os.Setenv("XDG_DATA_HOME", oldXDG)
					}
				}
			},
			wantErr:     false,
			wantContain: ".local/share/gotouch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			got, err := GetDataDir()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDataDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Errorf("GetDataDir() returned empty string")
			}
			if tt.wantContain != "" && got != "" {
				if !contains(got, tt.wantContain) {
					t.Errorf("GetDataDir() = %v, want to contain %v", got, tt.wantContain)
				}
			}
		})
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		dir     string
		wantErr bool
	}{
		{
			name:    "create new directory",
			dir:     filepath.Join(tmpDir, "test-dir"),
			wantErr: false,
		},
		{
			name:    "create nested directory",
			dir:     filepath.Join(tmpDir, "nested", "test-dir"),
			wantErr: false,
		},
		{
			name:    "existing directory",
			dir:     tmpDir,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnsureDir(tt.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify directory was created
				if _, err := os.Stat(tt.dir); os.IsNotExist(err) {
					t.Errorf("EnsureDir() did not create directory %s", tt.dir)
				}
			}
		})
	}
}

func TestFindConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test config file
	testConfigPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(testConfigPath, []byte("test: true"), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	tests := []struct {
		name         string
		explicitPath string
		wantErr      bool
		wantEmpty    bool
	}{
		{
			name:         "explicit path exists",
			explicitPath: testConfigPath,
			wantErr:      false,
			wantEmpty:    false,
		},
		{
			name:         "explicit path not found",
			explicitPath: filepath.Join(tmpDir, "nonexistent.yaml"),
			wantErr:      true,
			wantEmpty:    false,
		},
		{
			name:         "no explicit path",
			explicitPath: "",
			wantErr:      false,
			wantEmpty:    true, // May not find a config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindConfigFile(tt.explicitPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindConfigFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.wantEmpty && got == "" {
				t.Errorf("FindConfigFile() returned empty string when file should exist")
			}
		})
	}
}

func TestGetDefaultConfigPath(t *testing.T) {
	got, err := GetDefaultConfigPath()
	if err != nil {
		t.Errorf("GetDefaultConfigPath() error = %v", err)
		return
	}
	if got == "" {
		t.Errorf("GetDefaultConfigPath() returned empty string")
	}
	if !contains(got, "config.yaml") {
		t.Errorf("GetDefaultConfigPath() = %v, want to contain 'config.yaml'", got)
	}
}

func TestGetDefaultStatsPath(t *testing.T) {
	got, err := GetDefaultStatsPath()
	if err != nil {
		t.Errorf("GetDefaultStatsPath() error = %v", err)
		return
	}
	if got == "" {
		t.Errorf("GetDefaultStatsPath() returned empty string")
	}
	if !contains(got, "user_stats.json") {
		t.Errorf("GetDefaultStatsPath() = %v, want to contain 'user_stats.json'", got)
	}
}

func TestFindAPIKeyFile(t *testing.T) {
	// Test with api-key file in current directory
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	os.Chdir(tmpDir)

	// Create api-key file in current directory
	apiKeyPath := "api-key"
	if err := os.WriteFile(apiKeyPath, []byte("test-key"), 0644); err != nil {
		t.Fatalf("Failed to create api-key file: %v", err)
	}
	defer os.Remove(apiKeyPath)

	got := FindAPIKeyFile()
	if got == "" {
		t.Error("FindAPIKeyFile() should find api-key file in current directory")
	}
}

func TestFindAPIKeyFile_NotFound(t *testing.T) {
	// Test when no api-key file exists
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	os.Chdir(tmpDir)

	got := FindAPIKeyFile()
	if got != "" {
		t.Errorf("FindAPIKeyFile() should return empty when file not found, got %q", got)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		len(s) > len(substr)+1 && containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
