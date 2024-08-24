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

Build and install the Go binary:

```sh
make build
```

### Usage

Usage Examples:

1. **Initiate a Chat:**
   ```sh
   ai ask "What is the best way to manage Docker containers?"
   ```

2. **Generate a Commit Message:**
   ```sh
   ai commit --diff-unified 3 --lang en
   ```

3. **Auto-generate Code:**
   ```sh
   ai coder
   ```

4. **Review Code Changes:**
   ```sh
   ai review --exclude-list "*.md,*.txt"
   ```

## Contributing

Contribution guidelines will be provided soon. For more information, run `make help`.

## License

Copyright 2024 coding-hui. Licensed under the MIT License.

