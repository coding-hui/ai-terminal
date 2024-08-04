# Change Log


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


[Unreleased]: https://github.com/coding-hui/ai-terminal/compare/v0.1.3...HEAD
[v0.1.3]: https://github.com/coding-hui/ai-terminal/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/coding-hui/ai-terminal/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/coding-hui/ai-terminal/compare/v0.1...v0.1.1
