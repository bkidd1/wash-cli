# Wash - AI-Powered Code Pathway Analysis

Wash is a VSCode extension that helps developers explore alternative coding pathways when they're stuck. It uses GPT-4 Vision to analyze your code and suggest different approaches you could have taken.

## Project Structure

```
wash/
├── src/                    # VSCode extension source code
│   ├── extension.ts        # Main extension code
│   └── test/              # Extension tests
├── backend/               # Backend server code
│   ├── server.ts          # Local development server
│   ├── lambda.ts          # AWS Lambda handler
│   ├── package.json       # Backend dependencies
│   └── tsconfig.json      # Backend TypeScript config
├── package.json           # Extension dependencies
└── tsconfig.json          # Extension TypeScript config
```

## Features

- 🔍 Analyze current code implementation
- 💡 Suggest alternative coding pathways
- 🔒 Secure API key storage
- 📊 Clean, readable analysis display

## Development Setup

### Prerequisites

- Node.js 18+
- VSCode
- AWS CLI (for deployment)

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/wash.git
   cd wash
   ```

2. Install dependencies:
   ```bash
   # Install extension dependencies
   npm install
   
   # Install backend dependencies
   cd backend
   npm install
   ```

3. Set up environment variables:
   ```bash
   # Create .env.local in backend directory
   cd backend
   cp .env .env.local
   # Edit .env.local with your OpenAI API key
   ```

4. Start the backend server:
   ```bash
   cd backend
   npm start
   ```

5. Open the extension in a new VSCode window:
   ```bash
   /Applications/Visual\ Studio\ Code.app/Contents/Resources/app/bin/code --new-window --extensionDevelopmentPath="$(pwd)"
   ```

6. Run the extension:
   - Press F5 in VSCode to launch the extension in a new window
   - Use the "Wash: Analyze Coding Pathways" command to test

### Deployment

1. Set up AWS credentials:
   ```bash
   aws configure
   ```

2. Deploy the backend:
   ```bash
   cd backend
   ./deploy.sh
   ```

3. Update the extension's API URL in `src/extension.ts`

4. Package the extension:
   ```bash
   npm run package
   ```

## Security

- API keys are stored securely in AWS Secrets Manager
- All API calls are authenticated
- Rate limiting is implemented on the backend

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT

## Configuration

1. Open the Command Palette (Cmd/Ctrl + Shift + P)
2. Type "Wash: Configure OpenAI API Key"
3. Enter your OpenAI API key when prompted

## Usage

1. Open the file you want to analyze
2. Open the Command Palette (Cmd/Ctrl + Shift + P)
3. Type "Wash: Analyze Coding Pathways"
4. Wait for the analysis to complete
5. Review the suggested alternative approaches in the new panel

## Security

Your OpenAI API key is stored securely in VSCode's configuration system and is never transmitted anywhere except to OpenAI's API.

## Requirements

- VSCode 1.99.0 or higher
- OpenAI API key with access to GPT-4 Vision

## Extension Settings

Include if your extension adds any VS Code settings through the `contributes.configuration` extension point.

For example:

This extension contributes the following settings:

* `myExtension.enable`: Enable/disable this extension.
* `myExtension.thing`: Set to `blah` to do something.

## Known Issues

Calling out known issues can help limit users opening duplicate issues against your extension.

## Release Notes

Users appreciate release notes as you update your extension.

### 1.0.0

Initial release of ...

### 1.0.1

Fixed issue #.

### 1.1.0

Added features X, Y, and Z.

---

## Following extension guidelines

Ensure that you've read through the extensions guidelines and follow the best practices for creating your extension.

* [Extension Guidelines](https://code.visualstudio.com/api/references/extension-guidelines)

## Working with Markdown

You can author your README using Visual Studio Code. Here are some useful editor keyboard shortcuts:

* Split the editor (`Cmd+\` on macOS or `Ctrl+\` on Windows and Linux).
* Toggle preview (`Shift+Cmd+V` on macOS or `Shift+Ctrl+V` on Windows and Linux).
* Press `Ctrl+Space` (Windows, Linux, macOS) to see a list of Markdown snippets.

## For more information

* [Visual Studio Code's Markdown Support](http://code.visualstudio.com/docs/languages/markdown)
* [Markdown Syntax Reference](https://help.github.com/articles/markdown-basics/)

**Enjoy!**
