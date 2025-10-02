# AI-Terminal

[🇺🇸 English](./README.md) | [🇨🇳 中文](./README_zh.md) | [🇯🇵 日本語](./README_ja.md)

AI-Terminal is an AI-powered command-line interface that enhances terminal workflows through intelligent automation and optimization.

## Features

- **Smart Assistance:** Context-aware command suggestions and completions
- **Task Automation:** Automate repetitive tasks with AI-generated shortcuts
- **Intelligent Search:** Advanced file and content search capabilities
- **Error Handling:** Command correction and alternative suggestions
- **Extensible:** Plugin system for custom integrations

## Quick Start

### Prerequisites

- Go 1.22.0 or later

### Installation

**Homebrew:**
```bash
brew install coding-hui/tap/ai-terminal
```

**Download:**
- [Packages][releases] (Debian/RPM formats)
- [Binaries][releases] (Linux/macOS/Windows)

[releases]: https://github.com/coding-hui/ai-terminal/releases

**From Source:**
```bash
make build
```

**Initialize:**
```bash
ai configure
```

### Shell Completions

Completion files are included for Bash, ZSH, Fish, and PowerShell. Generate manually with:
```bash
ai completion [bash|zsh|fish|powershell] -h
```

## Usage Examples

### Chat & Assistance
```bash
ai ask "How to optimize Docker performance?"
ai ask --file prompt.txt
echo "code content" | ai ask "analyze this code"
```

### Code Generation
```bash
# Interactive mode
ai coder

# Batch processing
ai ctx load context.txt
ai coder -c session_id -p "add error handling"
```

### Code Review
```bash
ai review --exclude-list "*.md,*.txt"
```

### Command Execution
```bash
ai exec "find large files from last week"
ai exec --yes "docker ps -a"
ai exec --interactive
```

### Commit Messages
```bash
ai commit --diff-unified 3 --lang en
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**Changelog:** [CHANGELOG.md](CHANGELOG.md)  
**License:** [MIT](LICENSE) © 2024 coding-hui

