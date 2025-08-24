# MN Language Extension for VS Code

This extension provides syntax highlighting support for the MN programming language in Visual Studio Code.

## Features

- Syntax highlighting for MN language files (`.mn` extension)
- Support for:
  - Keywords (функц, зарла, буц, хэрэв, etc.)
  - Functions (санамсаргүйТоо, таахТоглоом, etc.)
  - Types (тоо64, тоо, хоосон)
  - Strings
  - Numbers
  - Comments
  - Operators

## Installation

### For Users
1. Install from VS Code Marketplace (coming soon)
2. Or install from VSIX:
   - Download the latest `.vsix` file from releases
   - In VS Code, go to Extensions (Ctrl+Shift+X)
   - Click the "..." menu and select "Install from VSIX"
   - Select the downloaded `.vsix` file

### For Developers
1. Clone this repository
2. Install dependencies:
   ```bash
   npm install
   ```
3. Install VS Code Extension Manager:
   ```bash
   npm install -g @vscode/vsce
   ```
4. Build the extension:
   ```bash
   vsce package
   ```
5. Install the extension:
   - In VS Code, go to Extensions (Ctrl+Shift+X)
   - Click the "..." menu and select "Install from VSIX"
   - Select the generated `.vsix` file

## Development

### Project Structure
- `syntaxes/mn.tmLanguage.json` - Syntax highlighting rules
- `language-configuration.json` - Language configuration
- `package.json` - Extension manifest

### Building and Testing
1. Press `F5` in VS Code to start debugging
2. A new VS Code window will open with the extension loaded
3. Open any `.mn` file to test syntax highlighting

### Publishing
1. Update version in `package.json`
2. Build the extension:
   ```bash
   vsce package
   ```
3. Publish to VS Code Marketplace:
   ```bash
   vsce publish
   ```

## Requirements

- VS Code version 1.85.0 or higher
- Node.js and npm for development

## Known Issues

None at the moment.

## Release Notes

### 0.0.1

Initial release with basic syntax highlighting support.
