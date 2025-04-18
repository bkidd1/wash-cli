# Wash CLI

Wash is a powerful command-line tool designed to help developers analyze, optimize, and monitor their code. It provides a suite of tools for code analysis, file monitoring, screenshot capture, and note-taking.

## Features

- **Code Analysis**
  - Analyze individual files for optimizations and improvements
  - Analyze project structure and organization
  - Get detailed feedback on code quality, performance, and security
  - Analyze chat history for key insights and actionable suggestions
  - Analyze chat summaries for patterns and progress tracking

- **File Monitoring**
  - Monitor files and directories for changes
  - Real-time notifications of file system events
  - Support for recursive directory watching

- **Screenshot Capture**
  - Capture screenshots of specific displays
  - Automatic timestamping and organization
  - Support for multiple displays

- **Note Taking**
  - Create and manage markdown notes
  - Automatic organization with timestamps
  - List and view existing notes

## Installation

### Prerequisites

- Go 1.16 or later
- OpenAI API key (for code analysis features)

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/brinleekidd/wash-cli.git
   cd wash-cli
   ```

2. Build the project:
   ```bash
   go build -o wash cmd/wash/main.go
   ```

3. Install the binary:
   ```bash
   sudo mv wash /usr/local/bin/
   ```

### Configuration

1. Create a configuration file at `~/.wash/wash.yaml`:
   ```yaml
   openai_key: your_openai_api_key
   ```

   Alternatively, you can set the `OPENAI_API_KEY` environment variable.

## Usage

### Code Analysis

Analyze a specific file:
```bash
wash analyze-file path/to/file.go
```

Analyze project structure:
```bash
wash analyze-project path/to/project
```

Analyze chat history:
```bash
wash analyze-chat path/to/chat.txt
```

The chat analysis provides:
- Key discussion points and decisions
- Actionable suggestions for code improvements
- Communication patterns and improvements
- Progress tracking and next steps

Analyze chat summary:
```bash
wash analyze-chat-summary path/to/summary.txt
```

The chat summary analysis provides:
- Key patterns and themes in the conversation
- Communication strengths and areas for improvement
- Actionable recommendations
- Overall progress tracking

### File Monitoring

Monitor files and directories:
```bash
wash monitor path/to/directory path/to/file
```

Press Ctrl+C to stop monitoring.

### Screenshot Capture

Capture a screenshot of a specific display:
```bash
wash screenshot 0  # Capture display 0
```

### Note Taking

Create a new note:
```bash
wash note "This is a new note"
```

List all notes:
```bash
wash list-notes
```

## Directory Structure

```
wash-cli/
├── cmd/
│   └── wash/           # Main CLI application
├── internal/
│   ├── analyzer/       # Code analysis functionality
│   ├── monitor/        # File system monitoring
│   ├── notes/          # Note-taking functionality
│   └── screenshot/     # Screenshot capture
└── pkg/
    └── config/         # Configuration management
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
