# GoTouch

A fast, terminal-based touch typing trainer built in Go with AI-powered adaptive learning.

## Features

- **AI-Powered Adaptive Learning**: Uses Claude AI to generate contextual typing exercises that adapt to your mistakes
- **Real-time Statistics**: Track your WPM, accuracy, and errors as you type
- **Session History**: Automatic saving of typing sessions with historical statistics
- **Terminal Theme Support**: Respects your terminal's color scheme for a seamless experience
- **Smooth UX**: Centered cursor view with smooth scrolling and polished Bubbletea UI
- **Flexible Text Sources**: Choose between LLM-generated content or dummy text
- **Configurable Sessions**: Adjust session duration with up/down arrows before starting

## Installation

### Pre-built Binaries (Recommended)

Download the latest pre-built binary for your platform from [GitHub Releases](https://github.com/YOUR_USERNAME/GoTouch/releases):

**Linux (x86_64):**
```bash
# Download the latest release
curl -LO https://github.com/YOUR_USERNAME/GoTouch/releases/latest/download/gotouch-linux-amd64

# Make it executable
chmod +x gotouch-linux-amd64

# Move to your PATH (optional)
sudo mv gotouch-linux-amd64 /usr/local/bin/gotouch
```

**macOS (Intel):**
```bash
curl -LO https://github.com/YOUR_USERNAME/GoTouch/releases/latest/download/gotouch-darwin-amd64
chmod +x gotouch-darwin-amd64
sudo mv gotouch-darwin-amd64 /usr/local/bin/gotouch
```

**macOS (Apple Silicon):**
```bash
curl -LO https://github.com/YOUR_USERNAME/GoTouch/releases/latest/download/gotouch-darwin-arm64
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
git clone https://github.com/YOUR_USERNAME/GoTouch.git
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

GoTouch uses a `config.yaml` file for configuration. On first run, it will use default settings.

### Basic Configuration

Create a `config.yaml` file in the same directory as the binary:

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
  file_dir: "user_stats.json"
```

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

   **Method 2: API Key File**
   ```bash
   # Create an api-key file in the same directory as the binary
   echo "your-api-key-here" > api-key
   ```

3. **Update config.yaml:**
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
├── main.go                 # Entry point
├── config.yaml            # Configuration file
├── internal/
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
├── user_stats.json       # Auto-generated session history
└── README.md
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

Ensure the directory where you're running GoTouch has write permissions for creating `user_stats.json`.

---

**Happy Typing!** 🚀
