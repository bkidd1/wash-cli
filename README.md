# Wash CLI

Wash CLI is an AI-powered development assistant that helps you:
- Monitor your development workflow
- Track project progress
- Remember important details
- Analyze code and suggest improvements
- Generate summaries and documentation
- Debug and fix issues

## Installation

### Using Go Install (Recommended)

```bash
go install github.com/bkidd1/wash-cli/cmd/wash@latest
```

This will install the `wash` binary to your `$GOPATH/bin` directory. Make sure `$GOPATH/bin` is in your PATH.

### Using Homebrew (macOS/Linux)

```bash
brew install bkidd1/tap/wash-cli
```

### Manual Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/bkidd1/wash-cli.git
   cd wash-cli
   ```

2. Build the project:
   ```bash
   ./build.sh
   ```

3. Move the binary to a directory in your PATH:
   ```bash
   mv wash /usr/local/bin/  # or ~/.local/bin/ for user-local installation
   ```

## Configuration

Before using Wash CLI, you'll need to set up your OpenAI API key:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

You can add this to your shell's configuration file (e.g., `~/.bashrc`, `~/.zshrc`) to make it permanent.

## Usage

Basic commands:
```bash
wash monitor start    # Start monitoring your development workflow
wash summary         # Get a summary of today's progress
wash remember       # Save important information
wash bug            # Report and track bugs
```

For more information about a specific command, use:
```bash
wash [command] --help
```

### Code Analysis

Analyze a specific file:
```bash
wash analyze-file path/to/file.go
```