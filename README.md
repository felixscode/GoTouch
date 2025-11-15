
<div align="center">

# GOTOUCH

[![Build and Release](https://github.com/felixscode/GoTouch/actions/workflows/build-and-release.yml/badge.svg)](https://github.com/felixscode/GoTouch/actions/workflows/build-and-release.yml)
[![codecov](https://codecov.io/gh/felixscode/GoTouch/branch/main/graph/badge.svg)](https://codecov.io/gh/felixscode/GoTouch)
[![Go Version](https://img.shields.io/badge/go-1.24.4-blue.svg)](https://go.dev/dl/)
[![Release](https://img.shields.io/github/v/release/felixscode/GoTouch)](https://github.com/felixscode/GoTouch/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/felixscode/GoTouch)](https://goreportcard.com/report/github.com/felixscode/GoTouch)

**A fast, terminal-based touch typing trainer built in Go with AI-powered adaptive learning.**

![gif](./gotouch.gif)

 [Features](#features) | [Installation](#installation) | [Quick Start](#quick-start) | [LLM Setup](#llm-setup-optional-but-recommended) | [Configuration](#configuration) | [Keyboard Controls](#keyboard-controls) | [Development](#development) | [Troubleshooting](#troubleshooting)
</div>



## Features

- **AI-Powered Adaptive Learning**: Uses LLMs to generate typing exercises that adapt to your mistakes
- **Multi-Provider Support**: Anthropic Claude, OpenAI GPT, or local Ollama models
- **Real-time Statistics**: Track WPM, accuracy, and errors as you type
- **Session History**: Automatic saving with historical statistics

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

```bash
# Run (works with dummy text out of the box)
gotouch
```

Use arrow keys to set duration, Enter to start, type the text. Stats appear at the end.

## LLM Setup (Optional but Recommended)

### Anthropic Claude (Recommended)

1. Get API key from [console.anthropic.com](https://console.anthropic.com)
2. Set environment variable:
   ```bash
   export GOTOUCH_LLM_API_KEY="your-key"
   ```
3. Update `~/.config/gotouch/config.yaml`:
   ```yaml
   text:
     source: llm
     llm:
       provider: "anthropic"
       model: "claude-3-5-haiku-latest"
   ```

### OpenAI GPT

```bash
export GOTOUCH_LLM_API_KEY="your-openai-key"
```

Config:
```yaml
text:
  source: llm
  llm:
    provider: "openai"
    model: "gpt-4"  # or "gpt-3.5-turbo"
```

### Ollama (Local)

```bash
ollama pull llama2
```

Config:
```yaml
text:
  source: llm
  llm:
    provider: "ollama"
    model: "llama2"
    api_base: "http://localhost:11434"
```

## Configuration

**Default config location:**
- Linux/macOS: `~/.config/gotouch/config.yaml`
- Windows: `%APPDATA%\gotouch\config.yaml`

**Example configs available:**
- `config.anthropic.yaml`
- `config.openai.yaml`
- `config.ollama.yaml`

See [config.example.yaml](config.example.yaml) for all options.

## Keyboard Controls

**Before Session:** ↑/↓ adjust duration, Enter to start
**During Session:** Type naturally, Backspace to correct, Esc to quit
**After Session:** Enter to exit

## Development

```bash
# Run tests
go test ./...

# With API key for integration tests
export GOTOUCH_LLM_API_KEY="your-key"
go test -v ./internal/sources

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o gotouch-linux-amd64
GOOS=darwin GOARCH=arm64 go build -o gotouch-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o gotouch-windows-amd64.exe
```

## Troubleshooting

**"GOTOUCH_LLM_API_KEY not set"**: Set the environment variable or add key to `~/.config/gotouch/api-key`

**LLM falls back to dummy**: Check `fallback_to_dummy: true` in config. Set to `false` for detailed errors.

**Ollama issues**: Ensure Ollama is running (`ollama serve`). Pull model first: `ollama pull llama2`

**Switch providers**: Copy example configs or edit `config.yaml` to change `provider` field

## License

MIT License - see LICENSE file for details

**Happy Typing!**
