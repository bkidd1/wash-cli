# Wash CLI

Wash is a powerful command-line tool designed to help developers analyze, optimize, and monitor their code. It provides a suite of tools for code analysis, file monitoring, screenshot capture, and note-taking.

## Installation

### Using Homebrew (macOS)

```bash
brew tap brinleekidd/tap
brew install wash
```

### Using Binary Releases

1. Download the appropriate binary for your platform from the [Releases](https://github.com/brinleekidd/wash-cli/releases) page.

2. Make the binary executable:
   ```bash
   chmod +x wash
   ```

3. Move the binary to a directory in your PATH:
   ```bash
   sudo mv wash /usr/local/bin/
   ```

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

## Configuration

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

### Chat Monitoring

Start chat monitoring:
```bash
wash chat start
```

Stop chat monitoring:
```bash
wash chat stop
```

### Screenshot Capture

Capture a screenshot:
```bash
wash screenshot 0  # Capture display 0
```

### Note Taking

Create a new note:
```bash
wash note "This is a new note"
```

### Version Information

Check the version:
```bash
wash version
```

## Features

- **Code Analysis**
  - Analyze individual files for optimizations and improvements
  - Analyze project structure and organization
  - Get detailed feedback on code quality, performance, and security

- **Chat Monitoring**
  - Monitor and analyze chat interactions
  - Generate insights and suggestions
  - Track decisions and changes

- **Screenshot Capture**
  - Capture screenshots of specific displays
  - Automatic timestamping and organization
  - Support for multiple displays

- **Note Taking**
  - Create and manage markdown notes
  - Automatic organization with timestamps
  - List and view existing notes

## Platform Support

Wash supports the following platforms:
- macOS (Intel and Apple Silicon)
- Linux (amd64)
- Windows (amd64)

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
