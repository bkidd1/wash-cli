# Wash CLI

Wash is a CLI tool that helps developers explore alternative coding pathways when they're stuck. It uses AI to analyze your code and suggest different approaches you could have taken.

## Features

- Analyze individual files for optimization opportunities
- Analyze multiple files for cross-file improvements
- Explore and analyze project structure
- Continuous chat monitoring with AI analysis
- Generate summaries of coding sessions

## Installation

### Prerequisites
- Go 1.16 or higher

### From Source
```bash
git clone https://github.com/brinleekidd/wash-cli.git
cd wash-cli
go install ./cmd/wash
```

## Usage

### Analyze a Single File
```bash
wash analyze path/to/file.go
```

### Analyze Multiple Files
```bash
wash analyze-multiple file1.go file2.go
```

### Explore Project Structure
```bash
wash explore [path]
```

### Start Chat Monitoring
```bash
wash monitor
```

### Stop Chat Monitoring
```bash
wash stop-monitor
```

### Generate Summary
```bash
wash summary
```

## Project Structure

```
wash-cli/
├── cmd/
│   └── wash/           # Main CLI application
├── internal/           # Private application code
│   ├── analyzer/      # Code analysis functionality
│   ├── monitor/       # Monitoring functionality
│   ├── screenshot/    # Screenshot capture functionality
│   └── notes/         # Notes management
├── pkg/               # Public library code
│   ├── api/          # API client code
│   └── utils/        # Utility functions
├── go.mod            # Go module definition
└── go.sum            # Go dependencies checksum
```

## Development

1. Clone the repository:
```bash
git clone https://github.com/brinleekidd/wash-cli.git
cd wash-cli
```

2. Install dependencies:
```bash
go mod download
```

3. Build the project:
```bash
go build -o wash ./cmd/wash
```

4. Run in development mode:
```bash
go run ./cmd/wash
```

## Configuration

The tool stores analysis results in a `.wash-notes` folder in your home directory.

## License

MIT
