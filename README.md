# AI-Terminal

AI-Terminal is an advanced AI-powered CLI that enhances terminal workflows through AI-driven automation and
optimization. It efficiently manages tasks such as file management, data processing, and system diagnostics.


For the Chinese version of this README, please see [README_zh.md](README_zh.md).

## Key Features

- **Contextual Assistance:** Learns from user commands to provide syntax suggestions.
- **Automated Tasks:** Recognizes repetitive task patterns and creates shortcuts.
- **Intelligent Search:** Conducts searches within files, directories, and specific file types.
- **Error Correction:** Corrects incorrect commands and suggests alternatives.
- **Custom Integrations:** Supports integration with various tools and services via plugins or APIs.

## Getting Started

### Prerequisites

- Go version v1.22.0 or higher.

### Installation

To install AI-Terminal, follow these steps:

1. Ensure you have Go version v1.22.0 or higher installed.
2. Build and install the Go binary:

```sh
make build
```

This command will generate the binary file at `./bin/ai` for Unix-based systems or `./bin/ai.exe` for Windows.

3. Initialize the AI model configuration:

```sh
ai configure
```

### Usage

Here are some examples of how to use AI-Terminal, grouped by functionality:

#### Chat

- **Initiate a Chat:**
  ```sh
  ai ask "What is the best way to manage Docker containers?"
  ```

- **Interactive Dialogue:**
  ```sh
  ai ask --interactive
  ```

- **Use a Prompt File:**
  ```sh
  ai ask --file /path/to/prompt_file.txt
  ```

- **Pipe Input:**
  ```sh
  cat some_script.go | ai ask generate unit tests
  ```

#### Code Generation

- **Auto-generate Code:**
  ```sh
  ai coder
  ```
  This command starts the auto-coding feature, which allows you to interactively generate code based on prompts.

#### Code Review

- **Review Code Changes:**
  ```sh
  ai review --exclude-list "*.md,*.txt"
  ```

#### Commit Messages

- **Generate a Commit Message:**
  ```sh
  ai commit --diff-unified 3 --lang en
  ```

## Contributing

We welcome contributions! Please see our [Contribution Guidelines](CONTRIBUTING.md) for more details.

### Changelog

Check out the [CHANGELOG.md](CHANGELOG.md) for detailed updates and changes to the project.

### License

Copyright 2024 coding-hui. Licensed under the [MIT License](LICENSE).

