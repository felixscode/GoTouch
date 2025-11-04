# GoTouch

[![Build and Release](https://github.com/felixscode/GoTouch/actions/workflows/build-and-release.yml/badge.svg)](https://github.com/felixscode/GoTouch/actions/workflows/build-and-release.yml)

A fast, terminal-based touch typing trainer built in Go with AI-powered adaptive learning.

## Features

- **AI-Powered Adaptive Learning**: Uses Claude AI to generate contextual typing exercises that adapt to your mistakes
- **Real-time Statistics**: Track your WPM, accuracy, and errors as you type
- **Session History**: Automatic saving of typing sessions with historical statistics
- **Terminal Theme Support**: Respects your terminal's color scheme for a seamless experience
- **Minimal Overhead:** purely written on go for a fast and minimal experiance

## Installation

### Pre-built Binaries (Recommended)

Download the latest pre-built binary for your platform from [GitHub Releases](https://github.com/felixscode/GoTouch/releases):

**Linux (x86_64):**
```bash
# Download the latest release
curl -LO https://github.com/felixscode/GoTouch/releases/latest/download/gotouch-linux-amd64

# Make it executable
chmod +x gotouch-linux-amd64

# Move to your PATH (optional)
sudo mv gotouch-linux-amd64 /usr/local/bin/gotouch
```

**Linux (ARM64):**
```bash
curl -LO https://github.com/felixscode/GoTouch/releases/latest/download/gotouch-linux-arm64
chmod +x gotouch-linux-arm64
sudo mv gotouch-linux-arm64 /usr/local/bin/gotouch
```

**macOS (Intel):**
```bash
curl -LO https://github.com/felixscode/GoTouch/releases/latest/download/gotouch-darwin-amd64
chmod +x gotouch-darwin-amd64
sudo mv gotouch-darwin-amd64 /usr/local/bin/gotouch
```

**macOS (Apple Silicon):**
```bash
curl -LO https://github.com/felixscode/GoTouch/releases/latest/download/gotouch-darwin-arm64
chmod +x gotouch-darwin-arm64
sudo mv gotouch-darwin-arm64 /usr/local/bin/gotouch
```

**Windows:**
```powershell
# Download gotouch-windows-amd64.exe from releases and add to your PATH
```

### Building from Source

If you prefer to build from source:

```bash
# Clone the repository
git clone https://github.com/felixscode/GoTouch.git
cd GoTouch

# Build
go build -o gotouch

# Run
./gotouch
```

**Requirements:**
- Go 1.21 or later

## Quick Start

1. **Run GoTouch:**
   ```bash
   gotouch
   ```

2. **Configure session duration** using up/down arrows

3. **Press Enter** to start typing

4. **Type the displayed text** - correct characters show in green, errors in red

5. **View your stats** at the end of the session

## Configuration

GoTouch automatically manages configuration files in platform-specific directories:

**Linux/macOS:**
- Config: `~/.config/gotouch/config.yaml`
- Stats: `~/.local/share/gotouch/user_stats.json`
- API Key (optional): `~/.config/gotouch/api-key`

**Windows:**
- Config: `%APPDATA%\gotouch\config.yaml`
- Stats: `%LOCALAPPDATA%\gotouch\user_stats.json`
- API Key (optional): `%APPDATA%\gotouch\api-key`

On first run, GoTouch will automatically create a default configuration file if none exists.

### Basic Configuration

The default `config.yaml` looks like this:

```yaml
text:
  source: dummy  # Options: "dummy" or "llm"
  llm:
    model: "haiku"  # Options: "sonnet" or "haiku"
    pregenerate_threshold: 20
    fallback_to_dummy: true
    timeout_seconds: 5
    max_retries: 1
ui:
  theme: "default"  # Options: "default" or "dark"
stats:
  file_dir: "~/.local/share/gotouch/user_stats.json"  # Auto-configured
```

You can edit this file directly with your preferred text editor, or use a custom config file location with the `--config` flag.

### LLM Mode Setup (Optional)

For AI-powered adaptive typing practice:

1. **Get an Anthropic API Key:**
   - Sign up at [https://console.anthropic.com](https://console.anthropic.com)
   - Generate an API key

2. **Configure the API key** (choose one method):

   **Method 1: Environment Variable**
   ```bash
   export ANTHROPIC_API_KEY="your-api-key-here"
   ```

   **Method 2: API Key File (Recommended)**
   ```bash
   # Linux/macOS
   echo "your-api-key-here" > ~/.config/gotouch/api-key

   # Windows (PowerShell)
   echo "your-api-key-here" > $env:APPDATA\gotouch\api-key
   ```

3. **Update config.yaml:**
   ```bash
   # Linux/macOS
   nano ~/.config/gotouch/config.yaml

   # Windows
   notepad %APPDATA%\gotouch\config.yaml
   ```

   Change `source: dummy` to `source: llm`:
   ```yaml
   text:
     source: llm
     llm:
       model: "haiku"  # Fast and cost-effective
       # model: "sonnet"  # More creative, higher quality
   ```

4. **Run GoTouch** - it will now generate adaptive content based on your typing patterns!

### How LLM Mode Works

- Generates contextual, interesting sentences for typing practice
- Analyzes your typing errors in real-time
- Pre-generates next sentences that focus on characters and words you struggle with
- Maintains topic continuity for a natural reading/typing experience
- Shows generated text immediately as it becomes available

## Usage

```bash
# Start with default configuration
gotouch

# Use specific config file
gotouch --config /path/to/config.yaml
```

### Keyboard Controls

**Before Session:**
- `↑/↓` - Adjust session duration (1-60 minutes)
- `Enter` - Start the typing session
- `Esc/Ctrl+C` - Exit

**During Session:**
- Type naturally - the cursor moves as you type
- `Backspace` - Correct mistakes
- `Esc/Ctrl+C` - Quit session early

**After Session:**
- `Enter` - Exit and save stats

## Project Structure

```
GoTouch/
├── main.go                 # Entry point with CLI flag parsing
├── internal/
│   ├── config/           # Configuration management
│   │   ├── paths.go      # Platform-specific path resolution
│   │   └── default.go    # Default config generation
│   ├── sources/          # Text source implementations
│   │   ├── source.go     # Text source interface
│   │   ├── llm.go        # Claude AI integration
│   │   ├── llm_test.go   # LLM tests
│   │   └── dummy.go      # Dummy text source
│   ├── types/            # Type definitions
│   │   ├── config.go     # Configuration types
│   │   └── stats.go      # Statistics types
│   └── ui/               # Terminal UI
│       ├── tui.go        # Bubbletea application
│       └── styles.go     # Color themes
└── README.md

User Data (auto-created):
  ~/.config/gotouch/config.yaml      # Configuration (Linux/macOS)
  ~/.local/share/gotouch/user_stats.json  # Session history
  ~/.config/gotouch/api-key          # Optional API key file
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run LLM integration tests (requires API key)
export ANTHROPIC_API_KEY="your-key"
go test -v ./internal/sources

# Run specific test
go test -v ./internal/sources -run TestGetText_Integration
```

### Building

```bash
# Build for current platform
go build -o gotouch

# Build for all platforms (like GitHub releases does)
GOOS=linux GOARCH=amd64 go build -o gotouch-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o gotouch-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o gotouch-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o gotouch-windows-amd64.exe
```

### Dependencies

- [github.com/anthropics/anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go) - Claude AI SDK
- [github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [github.com/charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML configuration

## Statistics

All typing sessions are automatically saved to `user_stats.json`. The dashboard shows:

- **Current Session**: WPM, Accuracy, Errors, Duration
- **Historical Stats**: Average WPM, Best WPM, Average Accuracy, Total Sessions

## Themes

GoTouch supports terminal themes that respect your terminal's color scheme:

- **Default Theme**: Standard terminal colors
- **Dark Theme**: Bright variants for better visibility on dark backgrounds

Configure in `config.yaml`:
```yaml
ui:
  theme: "default"  # or "dark"
```

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Troubleshooting

### "ANTHROPIC_API_KEY environment variable not set"

Make sure you've set up your API key using one of the methods in [LLM Mode Setup](#llm-mode-setup-optional).

### LLM mode falls back to dummy text

Check your `config.yaml` - if `fallback_to_dummy: true`, the app will use dummy text when LLM fails. Set to `false` to see detailed error messages.

### Colors don't match my terminal theme

Make sure you're using one of the built-in themes. The app automatically uses your terminal's color palette.

### Session stats not saving

The app automatically creates the data directory (`~/.local/share/gotouch` on Linux/macOS). If you're seeing permission errors, check that your user has write access to this directory.


**Happy Typing!** 🚀
