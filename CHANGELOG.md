# Change Log


<a name="v0.1.10"></a>
## [v0.1.10] - 2025-01-05
### Bug Fixes
- correct function name typo in auto coder

### Code Refactoring
- refactor cache configuration and environment variable handling
- refactor error handling and remove unused code
- refactor LLM engine to use consolidated call options
- refactor LLM call options into centralized function
- refactor configuration management to remove viper dependency
- refactor options package and update Go version
- refactor error handling to use `display.Fatal`

### Features
- add default chat ID handling for LLM engine
- add warning message display functionality in UI

### Pull Requests
- Merge pull request [#4](https://github.com/coding-hui/ai-terminal/issues/4) from eltociear/add-japanese-readme


<a name="v0.1.9-1"></a>
## [v0.1.9-1] - 2024-09-15
### Bug Fixes
- ask cmd issue

### Code Refactoring
- refactor language handling and template specifications
- refactor ui display func
- rename ui package
- improve Git Repository Handling and Error Messages

### Features
- expand supported languages for commit message summarization
- implement interactive AI configuration command


<a name="v0.1.9"></a>
## [v0.1.9] - 2024-09-01
### Bug Fixes
- update git reset behavior during rollback

### Code Refactoring
- refactor commit command and improve messaging
- refactor dependencies and logging for consistency

### Features
- add Undo Command to Help Message
- enhance undo functionality with context and error handling
- add `/undo` command for commit rollback
- add rollback functionality for recent commits
- enhance command execution feedback
- enhance code management with new commit command
- enhance commit command with file addition support
- enhance logging and status rendering capabilities
- enhance file handling with glob support


<a name="v0.1.8"></a>
## [v0.1.8] - 2024-09-01
### Bug Fixes
- MongoDB test container setup issue

### Code Refactoring
- refactor Docker targets and improve message formatting

### Features
- enhance version command functionality


<a name="v0.1.7"></a>
## [v0.1.7] - 2024-09-01
### Code Refactoring
- refactor success messages in coding modules
- refactor confirmation handling in coders
- refactor chat stream and edit application logic
- refactor code block editing and validation
- refactor code block editing and error handling
- refactor diff block editor and enhance test coverage
- refactor suggestion handling in AutoCoder
- refactor input handling and state management
- refactor suggestion handling with Suggester interface
- refactor test suite for clarity and efficiency
- refactor and enhance application functionality

### Features
- enhance AutoCoder suggestion handling
- enhance suggestion system for AutoCoder
- enhance command suggestions in prompt


<a name="v0.1.6"></a>
## [v0.1.6] - 2024-08-29
### Bug Fixes
- fix panic issue in waitForUserConfirm method

### Code Refactoring
- refactor code quality and remove unused methods
- simplify edit application process
- refactor input handling with `textinput` package
- refactor UI dimensions handling
- enhance terminal interaction and input handling
- simplify viewport height calculation
- refactor codebase for context handling and logging
- enhance error handling and file content formatting

### Features
- enhance Git repository file listing functionality


<a name="v0.1.5"></a>
## [v0.1.5] - 2024-08-24
### Code Refactoring
- refactor test suite for clarity and maintainability
- refactor and simplify code across multiple files
- enhance user interaction handling and code clarity
- enhance UI rendering and input handling
- remove deprecated CodeEditor module
- update Assistant Name to ai in Prompts

### Features
- implement user confirmation workflow
- improve User Onboarding Experience
- improve coder UI and feedback messages
- improve Checkpoint Display and Diff Handling
- improve Edit Application and Error Handling
- implement Code Editor with Search/Replace Blocks
- add `/drop` command to clear tracked files
- optimize asauto-coderk CMD Prompts


<a name="v0.1.4"></a>
## [v0.1.4] - 2024-08-18
### Bug Fixes
- improve UI responsiveness and efficiency
- simplify UI Initialization Condition

### Code Refactoring
- improve command handling in AutoCoder
- improve Coder UI and Chat Integration
- refactor coder and LLM modules
- optimize Prompt Templates and Actions Configuration
- remove empty diff checks in commands
- integrate chat history summary logic
- remove redundant color formatting in prompts
- remove PROJECT configuration file
- remove InferenceService and Related Components
- simplify edit settings command execution
- simplify hook command handling

### Features
- release v0.1.4 with fixes and features
- improve file listing output formatting
- introduce AutoCoder for file-based chat assistance
- add User Custom Prompts to Commit Cmd
- add colored console feedback for no sessions
- add empty diff checks in CLI commands
- release v0.1.3-1 with enhancements and fixes


<a name="v0.1.3-1"></a>
## [v0.1.3-1] - 2024-08-04
### Bug Fixes
- Fix sessions loading issue

### Code Refactoring
- improve code quality and reliability
- convert summary prefix to lowercase
- Refactor CLI command for improved structure

### Features
- add Git hook commands and fix typo
- add multi-language support for commit summaries
- add auto code review command
- enhance commit message generation flexibility
- enhance UI and automate commit messages
- Implement AI-assisted commit workflow and UI
- New command line completion


<a name="v0.1.3"></a>
## [v0.1.3] - 2024-08-04
### Bug Fixes
- By default, the local mongodb server is used
- Fix chat session deletion issue
- Use vim as the default editor
- sql close

### Code Refactoring
- Simplify the ls history code
- Optimize configuration loading and binding logic
- rename flag utils
- Refactoring the Chat History Store
- rename run pkg name

### Features
- Historical sessions can be persisted to files
- Added command to delete history session
- Upgrade go version to 1.22.5
- Add history ls cmd, refactor and update existing code for CLI and datastore components


<a name="v0.1.2"></a>
## [v0.1.2] - 2024-07-14
### Bug Fixes
- ExecCompletion output json serialization issue

### Code Refactoring
- Add the output parameter, which can be markdown or raw
- update Dockerfile for CLI and infer-controller


<a name="v0.1.1"></a>
## [v0.1.1] - 2024-07-07

<a name="v0.1"></a>
## v0.1 - 2024-07-07
### Features
- add ask cmd


[Unreleased]: https://github.com/coding-hui/ai-terminal/compare/v0.1.10...HEAD
[v0.1.10]: https://github.com/coding-hui/ai-terminal/compare/v0.1.9-1...v0.1.10
[v0.1.9-1]: https://github.com/coding-hui/ai-terminal/compare/v0.1.9...v0.1.9-1
[v0.1.9]: https://github.com/coding-hui/ai-terminal/compare/v0.1.8...v0.1.9
[v0.1.8]: https://github.com/coding-hui/ai-terminal/compare/v0.1.7...v0.1.8
[v0.1.7]: https://github.com/coding-hui/ai-terminal/compare/v0.1.6...v0.1.7
[v0.1.6]: https://github.com/coding-hui/ai-terminal/compare/v0.1.5...v0.1.6
[v0.1.5]: https://github.com/coding-hui/ai-terminal/compare/v0.1.4...v0.1.5
[v0.1.4]: https://github.com/coding-hui/ai-terminal/compare/v0.1.3-1...v0.1.4
[v0.1.3-1]: https://github.com/coding-hui/ai-terminal/compare/v0.1.3...v0.1.3-1
[v0.1.3]: https://github.com/coding-hui/ai-terminal/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/coding-hui/ai-terminal/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/coding-hui/ai-terminal/compare/v0.1...v0.1.1
