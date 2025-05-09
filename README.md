# Wash CLI

Wash CLI is an AI-powered development assistant that helps you:
- Remember important details
- Analyze code and suggest improvements
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
wash remember       # Save important information
wash bug            # Report and track bugs
wash file          # Analyze code files
wash project       # Analyze project structure
```

For more information about a specific command, use:
```bash
wash [command] --help
```

### Code Analysis

Analyze a specific file:
```bash
wash file path/to/file.go
```

## Troubleshooting

### Common Issues

1. **Command not found**
   - Ensure the `wash` binary is in your PATH
   - Try running `which wash` to verify installation
   - For Homebrew users: `brew link wash-cli`

2. **API Key Issues**
   - Verify your OpenAI API key is correctly set
   - Check if the key has sufficient quota
   - Ensure the key has the necessary permissions

3. **Permission Errors**
   - Check file permissions in your project directory
   - Ensure you have write access to the configuration directory
   - Try running with elevated permissions if necessary

### Getting Help

If you encounter issues not covered here:
1. Check the [GitHub Issues](https://github.com/bkidd1/wash-cli/issues)
2. Create a new issue with:
   - Your operating system and version
   - Wash CLI version (`wash version`)
   - Steps to reproduce the issue
   - Relevant error messages

## Configuration Options

Wash CLI can be configured through environment variables or a configuration file.

### Environment Variables

- `OPENAI_API_KEY`: Your OpenAI API key (required)
- `WASH_LOG_LEVEL`: Set logging level (debug, info, warn, error)
- `WASH_CONFIG_DIR`: Custom configuration directory
- `WASH_CACHE_DIR`: Custom cache directory

### Configuration File

Create a `config.yaml` in your configuration directory:

```yaml
openai:
  api_key: "your-api-key"
  model: "gpt-4"  # or "gpt-3.5-turbo"
  temperature: 0.7

logging:
  level: "info"
  format: "text"  # or "json"

cache:
  enabled: true
  ttl: "24h"
```

## Contributing

We welcome contributions! Here's how to get started:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

1. Install Go 1.21 or later
2. Clone the repository
3. Run `go mod tidy` to install dependencies
4. Build with `./build.sh`
5. Run tests with `go test ./...`

### Code Style

- Follow Go standard formatting (`go fmt`)
- Write tests for new features
- Update documentation as needed
- Keep commits focused and atomic

## Performance Considerations

### Memory Usage

Wash CLI is designed to be memory-efficient:
- Uses streaming for large outputs
- Implements caching for frequent operations
- Cleans up temporary files automatically

### Response Time

Typical response times:
- Simple commands: < 1 second
- Code analysis: 2-5 seconds
- Complex queries: 5-10 seconds

To improve performance:
- Use appropriate cache settings
- Keep your OpenAI API key quota available
- Monitor system resources during heavy usage

## Security

### Data Handling

- API keys are stored securely using system keyring
- Temporary files are encrypted when containing sensitive data
- Cache is cleared automatically based on TTL settings

### Best Practices

1. **API Keys**
   - Never commit API keys to version control
   - Use environment variables for CI/CD
   - Rotate keys periodically

2. **File Permissions**
   - Configuration files are stored with restricted permissions
   - Temporary files are created with secure permissions
   - Cache files are isolated per user
3. **Network Security**
   - All API calls use HTTPS
   - Certificate validation is enforced
   - Timeout settings prevent hanging connections

For security concerns, please email brinlee0kidd@gmail.com or create a security advisory on GitHub.
