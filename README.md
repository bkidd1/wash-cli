# Wash - AI-Powered Code Pathway Analysis

Wash is a VSCode extension that helps developers explore alternative coding pathways when they're stuck. It uses GPT-4 Vision to analyze your code and suggest different approaches you could have taken.

## Features

- 🔍 Analyze current code implementation
- 💡 Suggest alternative coding pathways
- 🔒 Secure API key storage
- 📊 Clean, readable analysis display

## Installation

1. Install the extension from the VSCode marketplace
2. Configure your OpenAI API key (see below)
3. Start analyzing your code!

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

## License

MIT

## Contributing

Feel free to open issues or submit pull requests!

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
