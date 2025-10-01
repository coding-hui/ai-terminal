# Change Log


<a name="v0.13.6"></a>
## [v0.13.6] - 2025-10-02
### Code Refactoring
- refactor system configuration detection


<a name="v0.13.5"></a>
## [v0.13.5] - 2025-10-02
### Code Refactoring
- remove deprecated prompt configuration


<a name="latest"></a>
## [latest] - 2025-10-02
### Bug Fixes
- fix raw and quiet mode rendering

### Code Refactoring
- refactor Ctrl+C handler into bye function
- run Bubble Tea programs without renderer

### Features
- add prompt mode cycling key binding
- refactor prompt configuration for mode-specific prefixes
- integrate AutoCoder for AI-powered command execution
- add AI-powered shell command execution


<a name="v0.13.4"></a>
## [v0.13.4] - 2025-09-27
### Bug Fixes
- exit with code 1 on config command error
- handle config command errors more gracefully


<a name="v0.13.3"></a>
## [v0.13.3] - 2025-09-13
### Bug Fixes
- add error handling and make git hook executable

### Code Refactoring
- fix Cobra command output stream setter
- refactor Options struct into separate file
- simplify SimpleChatHistoryStore, clean up test code and delete redundant files

### Features
- Support recursive directory processing in /add

### Pull Requests
- Merge pull request [#18](https://github.com/coding-hui/ai-terminal/issues/18) from clh021/feat/add-folder
- Merge pull request [#17](https://github.com/coding-hui/ai-terminal/issues/17) from clh021/main


<a name="v0.13.2"></a>
## [v0.13.2] - 2025-02-16
### Bug Fixes
- sanitize input strings using HTML unescaping
- integrate SILICONCLOUD_API_KEY and streamline clipboard logic
- refactor fence detection logic for newline and header checks
- fix autocoder failure when adding multiple files

### Code Refactoring
- enhance `commands.go` comments and use `html` package in completion functions
- refactor terminal renderer to use environment-based color detection
- refactor fence selection logic for improved accuracy
- refactor fence detection logic for improved accuracy
- consolidate user input for command execution
- use precomputed `AverageTokensPerSecond` field for token calculations
- optimize token rendering by eliminating redundant `fmt.Sprintf` calls
- refactor logging tests and add AI debug data tracking
- remove redundant `fmt.Sprintf` in user confirmation prompt
- refactor engine option naming for consistency
- refactor AI engine initialization to use `New` function

### Features
- prepare v0.13.2 release: add features & fixes, refactor code, support new models and commands
- upgrade Go dependencies, refactor command handling, replace library, and modify key bindings
- enhance thread-safety of SimpleChatHistoryStore and update test cases for concurrency
- enhance code fence handling and add related methods and documentation
- upgrade dependencies, refactor struct fields, add HTML handling, and new Markdown module
- streamline chat message rendering and output formatting
- enable clipboard support for chat output copying
- enhance design and chat model commands with new methods and templates
- enhance command functionalities for AI-assisted coding
- refactor command handlers to use single string input
- refactor welcome message and enhance code fence handling in `AutoCoder`
- integrate token usage tracking and reporting in the application
- add `/apply` command and support for direct code input in `GetEdits`
- support Volcengine Ark models and document v0.13.2 changes in changelog
- add token usage tracking and configuration options
- refactor AI engine to support new models and APIs
- auto add modify file to load context


<a name="v0.13.1"></a>
## [v0.13.1] - 2025-01-31
### Code Refactoring
- refactor `Render` function to use formatted string for `content`
- refactor coder block fence rendering to use `RenderStep`
- refactor command handler method names for consistency
- simplify `coding` function by removing redundant context checks

### Features
- implement command-line flag handling and completion support
- add `deepseek-reasoner` model with alias `ds-r1`


<a name="v0.13.0"></a>
## [v0.13.0] - 2025-01-18
### Bug Fixes
- ensure `CacheWriteToID` is checked before appending assistant messages

### Code Refactoring
- refactor `llm` package to `ai` and simplify engine configuration
- reorder message assignment after invalidation check in `SetMessages`
- remove unused code and imports
- refactor chat commands to simplify logic and improve error handling
- use format string for `confirmTitle` in `WaitForUserConfirm`
- refactor message loading logic and tests

### Features
- refactor `AutoCoder` to support context-based code generation
- refactor conversation and context management logic


<a name="v0.12.1"></a>
## [v0.12.1] - 2025-01-18
### Bug Fixes
- enhance error handling in `ListContextsByteConvoID` with descriptive wrapping
- improve HTTP error message with status code details
- improve progress bar edge case handling and validation
- fix file creation confirmation logic

### Code Refactoring
- remove unused code and simplify function parameters
- remove unused `cfg` parameter from `postRunHook`
- replace manual HTML decoding with `html.UnescapeString`
- refactor completion methods to use context and consistent naming
- refactor stream completion return type and error handling
- remove lowercase conversion of `commitPrefix`
- refactor load package into context command structure
- refactor SQL error handling and auto-increment logic
- refactor conversation store handling in Engine struct
- refactor `ChatID` to `ConversationID` for consistency
- remove `ChatStream` function from LLM engine
- refactor completion methods to use `llms.PromptValue`
- refactor imports and update cache path configuration

### Features
- replace history command with convo command and add subcommands
- refactor chat history storage and enhance conversation handling
- add context support to `Persistent` and `Invalidate` methods
- add context management commands and improve cleanup functionality
- add context support and debug utilities for conversation handling
- integrate conversation store and streamline file handling
- refactor conversation package and add SQLite context storage
- Support sqlite to save session records
- improve URL handling and content fetching logic


<a name="v0.12.0"></a>
## [v0.12.0] - 2025-01-12
### Bug Fixes
- fix HTML entity handling and improve env prefix consistency
- improve error handling and code cleanup in command execution

### Code Refactoring
- refactor confirmation logic and remove unused dependencies
- remove `needConfirm` parameter from `ApplyEdits`
- improve diff output formatting and rendering
- simplify edit confirmation logic in diff block editor
- refactor chat message handling and remove unused code
- remove unused UI history and spinner components
- refactor chat system to use `llms.ChatMessage` for message handling
- refactor console rendering for consistency and simplicity
- refactor commit message rendering and styling
- refactor commit message rendering with styled UI components
- refactor options initialization using functional pattern
- Refactor auto.coder to support command completion
- refactor error handling and improve config error messaging

### Features
- refactor and enhance chat system with new features and bug fixes
- add language support to commit command
- add support for commit message language configuration
- add support for conventional commit prefixes
- add shorthand flag for version command output


<a name="v0.1.11"></a>
## [v0.1.11] - 2025-01-05
### Code Refactoring
- improve error handling and streaming logic in LLM engine


<a name="v0.1.10"></a>
## [v0.1.10] - 2025-01-05

<a name="0.1.10"></a>
## [0.1.10] - 2025-01-05
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
- add shell completions and manpage generation support
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


[Unreleased]: https://github.com/coding-hui/ai-terminal/compare/v0.13.6...HEAD
[v0.13.6]: https://github.com/coding-hui/ai-terminal/compare/v0.13.5...v0.13.6
[v0.13.5]: https://github.com/coding-hui/ai-terminal/compare/latest...v0.13.5
[latest]: https://github.com/coding-hui/ai-terminal/compare/v0.13.4...latest
[v0.13.4]: https://github.com/coding-hui/ai-terminal/compare/v0.13.3...v0.13.4
[v0.13.3]: https://github.com/coding-hui/ai-terminal/compare/v0.13.2...v0.13.3
[v0.13.2]: https://github.com/coding-hui/ai-terminal/compare/v0.13.1...v0.13.2
[v0.13.1]: https://github.com/coding-hui/ai-terminal/compare/v0.13.0...v0.13.1
[v0.13.0]: https://github.com/coding-hui/ai-terminal/compare/v0.12.1...v0.13.0
[v0.12.1]: https://github.com/coding-hui/ai-terminal/compare/v0.12.0...v0.12.1
[v0.12.0]: https://github.com/coding-hui/ai-terminal/compare/v0.1.11...v0.12.0
[v0.1.11]: https://github.com/coding-hui/ai-terminal/compare/v0.1.10...v0.1.11
[v0.1.10]: https://github.com/coding-hui/ai-terminal/compare/0.1.10...v0.1.10
[0.1.10]: https://github.com/coding-hui/ai-terminal/compare/v0.1.9-1...0.1.10
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
