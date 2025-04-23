# Contributing to Wash CLI

Thank you for your interest in contributing to Wash CLI! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct.

## How to Contribute

1. Fork the repository
2. Create a new branch for your feature or bugfix
3. Make your changes
4. Run tests and ensure they pass
5. Submit a pull request

## Development Setup

1. Install Go (version 1.21 or later)
2. Clone the repository:
   ```bash
   git clone https://github.com/bkidd1/wash-cli.git
   cd wash-cli
   ```
3. Install dependencies:
   ```bash
   go mod tidy
   ```

## Testing

Run the test suite:
```bash
go test ./...
```

## Code Style

- Follow the Go standard formatting
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small
- Write tests for new functionality

## Pull Request Process

1. Update the CHANGELOG.md with details of changes
2. Update the README.md if necessary
3. Ensure all tests pass
4. Submit the pull request with a clear description

## Release Process

1. Create a release branch
2. Update version numbers
3. Update CHANGELOG.md
4. Create a pull request
5. After review, merge and tag the release

## Questions?

Feel free to open an issue if you have any questions about contributing! 