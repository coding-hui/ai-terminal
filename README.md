Language : [🇺🇸 us](./README.md) | [🇨🇳 zh](./README_zh.md) | [🇯🇵 ja](./README_ja.md)

# AI-Terminal

AI-Terminal is an advanced AI-powered CLI that enhances terminal workflows through AI-driven automation and
optimization. It efficiently manages tasks such as file management, data processing, and system diagnostics.

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

Install using Homebrew:

```bash
brew install coding-hui/tap/ai-terminal
```

Or, download it:

- [Packages][releases] are available in Debian and RPM formats
- [Binaries][releases] are available for Linux, macOS, and Windows

[releases]: https://github.com/coding-hui/ai-terminal/releases

Or, build from source (requires Go 1.22+):

```sh
make build
```

Then initialize configuration:
```sh
ai configure
```

<details>
<summary>Shell Completions</summary>

All packages and archives come with pre-generated completion files for Bash,
ZSH, Fish, and PowerShell.

If you built it from source, you can generate them with:

```bash
ai completion bash -h
ai completion zsh -h
ai completion fish -h
ai completion powershell -h
```

If you use a package (like Homebrew, Debs, etc), the completions should be set
up automatically, given your shell is configured properly.

</details>

### Usage

Here are some examples of how to use AI-Terminal, grouped by functionality:

#### Chat

- **Initiate a Chat:**
  ```sh
  ai ask "What is the best way to manage Docker containers?"
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

- **Interactive Code Generation:**
  ```sh
  ai coder
  ```
  Starts interactive mode for generating code based on prompts.

- **CLI-based Code Generation:**
  ```sh
  ai ctx load /path/to/context_file
  ai coder -c session_id -p "improve comments and add unit tests"
  ```
  Load context files and specify session ID for batch processing. Supports:
  - Code improvement
  - Comment enhancement
  - Unit test generation
  - Code refactoring

- **Generate Code with Context:**
  ```sh
  ai ctx load /path/to/context_file
  ai coder "implement feature xxx"
  ```
  Load context files first to provide additional information for code generation.

#### Code Review

- **Review Code Changes:**
  ```sh
  ai review --exclude-list "*.md,*.txt"
  ```

#### Command Execution

- **Execute Shell Commands via AI:**
  ```sh
  ai exec "find all files modified in the last 7 days"
  ```
  Uses AI to interpret your instruction and execute the appropriate shell command.

- **Auto-execute Without Confirmation:**
  ```sh
  ai exec --yes "list all docker containers"
  ```
  Automatically executes the inferred command without asking for confirmation.

- **Interactive Command Mode:**
  ```sh
  ai exec --interactive
  ```
  Starts an interactive dialogue to refine and execute commands.

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

